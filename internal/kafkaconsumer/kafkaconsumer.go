package kafkaconsumer

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/IBM/sarama"
	"github.com/devchain-network/cauldron/internal/cerrors"
	"github.com/devchain-network/cauldron/internal/storage"
	"github.com/go-playground/webhooks/v6/github"
	"github.com/google/uuid"
	"github.com/vigo/getenv"
)

// TCPAddrs represents comma separated tcp addr list.
type TCPAddrs string

// List validates and return list of tcp addrs.
func (t TCPAddrs) List() []string {
	var addrs []string
	for _, addr := range strings.Split(string(t), ",") {
		if _, err := getenv.ValidateTCPNetworkAddress(addr); err == nil {
			addrs = append(addrs, addr)
		}
	}

	return addrs
}

// KafkaTopicIdentifier represents custom type.
type KafkaTopicIdentifier string

func (s KafkaTopicIdentifier) String() string {
	return string(s)
}

// constants.
const (
	DefaultKafkaBrokers = "127.0.0.1:9094"

	DefaultKafkaConsumerPartition    = 0
	DefaultKafkaConsumerDialTimeout  = 30 * time.Second
	DefaultKafkaConsumerReadTimeout  = 30 * time.Second
	DefaultKafkaConsumerWriteTimeout = 30 * time.Second

	DefaultKafkaConsumerBackoff    = 2 * time.Second
	DefaultKafkaConsumerMaxRetries = 10

	KafkaTopicIdentifierGitHub KafkaTopicIdentifier = "github"
	KafkaTopicIdentifierGitLab KafkaTopicIdentifier = "gitlab"
)

// KafkaConsumer defines kafka consumer behaviours.
type KafkaConsumer interface {
	Start() error
}

var _ KafkaConsumer = (*Consumer)(nil) // compile time proof

// Consumer represents kafa consumer setup.
type Consumer struct {
	Logger       *slog.Logger
	Storage      storage.Storer
	Topic        string
	Brokers      []string
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	Backoff      time.Duration
	MaxRetries   uint8
	Partition    int32
}

// Start starts consumer.
func (c Consumer) Start() error {
	config := c.getConfig()

	var consumer sarama.Consumer
	var consumerErr error
	backoff := c.Backoff

	for i := range c.MaxRetries {
		consumer, consumerErr = sarama.NewConsumer(c.Brokers, config)
		if consumerErr == nil {
			break
		}

		c.Logger.Error(
			"can not connect broker",
			"error",
			consumerErr,
			"retry",
			fmt.Sprintf("%d/%d", i, c.MaxRetries),
			"backoff",
			backoff.String(),
		)
		time.Sleep(backoff)
		backoff *= 2
	}

	if consumerErr != nil {
		return fmt.Errorf("kafkaconsumer.Consumer.Start error: [%w]", consumerErr)
	}
	defer func() { _ = consumer.Close() }()

	partitionConsumer, err := consumer.ConsumePartition(c.Topic, c.Partition, sarama.OffsetNewest)
	if err != nil {
		return fmt.Errorf("kafkaconsumer.Consumer consumer.ConsumePartition error: [%w]", err)
	}
	defer func() { _ = partitionConsumer.Close() }()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	c.Logger.Info("consuming messages from", "topic", c.Topic)

	messageChan := make(chan *sarama.ConsumerMessage, 10)
	defer close(messageChan)

	numWorkers := runtime.NumCPU()
	c.Logger.Info("starting workers", "count", numWorkers)

	var wg sync.WaitGroup
	for i := range numWorkers {
		wg.Add(1)
		go c.worker(i, messageChan, &wg)
	}

	go func() {
		for {
			select {
			case msg := <-partitionConsumer.Messages():
				if msg != nil {
					messageChan <- msg
				}
			case err := <-partitionConsumer.Errors():
				c.Logger.Error("partition consumer error", "error", err)
			case <-signals:
				c.Logger.Info("shutting down message producer")

				return
			}
		}
	}()

	wg.Wait()
	c.Logger.Info("all workers stopped")

	return nil
}

func (c Consumer) getConfig() *sarama.Config {
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true
	config.Net.DialTimeout = c.DialTimeout
	config.Net.ReadTimeout = c.ReadTimeout
	config.Net.WriteTimeout = c.WriteTimeout

	return config
}

func (c Consumer) storeGitHubMessage(msg *sarama.ConsumerMessage) error {
	deliveryID, err := uuid.Parse(string(msg.Key))
	if err != nil {
		return fmt.Errorf("kafkaconsumer.storeGitHubMessage deliveryID error: [%w]", err)
	}

	targetID, err := strconv.ParseUint(string(msg.Headers[2].Value), 10, 64)
	if err != nil {
		return fmt.Errorf("kafkaconsumer.storeGitHubMessage targetID error: [%w]", err)
	}

	hookID, err := strconv.ParseUint(string(msg.Headers[3].Value), 10, 64)
	if err != nil {
		return fmt.Errorf("kafkaconsumer.storeGitHubMessage hookID error: [%w]", err)
	}

	target := string(msg.Headers[1].Value)
	event := github.Event(string(msg.Headers[0].Value))
	offset := msg.Offset
	partition := msg.Partition

	var payload any

	switch event { //nolint:exhaustive
	case github.IssuesEvent:
		var pl github.IssuesPayload
		if err = json.Unmarshal(msg.Value, &pl); err != nil {
			return fmt.Errorf("kafkaconsumer.storeGitHubMessage github.IssuesPayload error: [%w]", err)
		}
		payload = pl
	case github.IssueCommentEvent:
		var pl github.IssueCommentPayload
		if err = json.Unmarshal(msg.Value, &pl); err != nil {
			return fmt.Errorf("kafkaconsumer.storeGitHubMessage github.IssueCommentPayload error: [%w]", err)
		}
		payload = pl
	case github.CreateEvent:
		var pl github.CreatePayload
		if err = json.Unmarshal(msg.Value, &pl); err != nil {
			return fmt.Errorf("kafkaconsumer.storeGitHubMessage github.CreatePayload error: [%w]", err)
		}
		payload = pl
	case github.DeleteEvent:
		var pl github.DeletePayload
		if err = json.Unmarshal(msg.Value, &pl); err != nil {
			return fmt.Errorf("kafkaconsumer.storeGitHubMessage github.DeletePayload error: [%w]", err)
		}
		payload = pl
	case github.PushEvent:
		var pl github.PushPayload
		if err = json.Unmarshal(msg.Value, &pl); err != nil {
			return fmt.Errorf("kafkaconsumer.storeGitHubMessage github.PushPayload error: [%w]", err)
		}
		payload = pl
	}

	var userID int64
	var userLogin string

	switch payload := payload.(type) {
	case github.IssuesPayload:
		userID = payload.Sender.ID
		userLogin = payload.Sender.Login
	case github.IssueCommentPayload:
		userID = payload.Sender.ID
		userLogin = payload.Sender.Login
	case github.CreatePayload:
		userID = payload.Sender.ID
		userLogin = payload.Sender.Login
	case github.DeletePayload:
		userID = payload.Sender.ID
		userLogin = payload.Sender.Login
	case github.PushPayload:
		userID = payload.Sender.ID
		userLogin = payload.Sender.Login
	}

	storagePayload := storage.GitHubWebhook{
		DeliveryID: deliveryID,
		Event:      event,
		Target:     target,
		TargetID:   targetID,
		HookID:     hookID,
		Offset:     offset,
		Partition:  partition,
		UserID:     userID,
		UserLogin:  userLogin,
		Payload:    payload,
	}

	if err = c.Storage.GitHubStore(&storagePayload); err != nil {
		return fmt.Errorf("kafkaconsumer.storeGitHubMessage Storage.GitHubStore error: [%w]", err)
	}

	return nil
}

func (c Consumer) worker(id int, messages <-chan *sarama.ConsumerMessage, wg *sync.WaitGroup) {
	defer wg.Done()

	for msg := range messages {
		switch KafkaTopicIdentifier(c.Topic) {
		case KafkaTopicIdentifierGitHub:
			if err := c.storeGitHubMessage(msg); err != nil {
				c.Logger.Error("store github message error", "error", err, "worker id", id)

				continue
			}
		case KafkaTopicIdentifierGitLab:
			fmt.Println("parse GitLab kafka message")
		default:
			fmt.Println("unknown topic identifier")
		}

		c.Logger.Info("github messages successfully stored to db", "worker id", id)
	}
}

// Option represents option function type.
type Option func(*Consumer) error

// WithLogger sets logger.
func WithLogger(l *slog.Logger) Option {
	return func(consumer *Consumer) error {
		if l == nil {
			return fmt.Errorf("kafkaconsumer.WithLogger error: [%w]", cerrors.ErrValueRequired)
		}
		consumer.Logger = l

		return nil
	}
}

// WithTopic sets topic name.
func WithTopic(s string) Option {
	return func(consumer *Consumer) error {
		if s == "" {
			return fmt.Errorf("kafkaconsumer.WithLogger error: [%w]", cerrors.ErrValueRequired)
		}
		consumer.Topic = s

		return nil
	}
}

// WithBrokers sets brokers list.
func WithBrokers(brokers []string) Option {
	return func(consumer *Consumer) error {
		if brokers == nil {
			return fmt.Errorf("kafkaconsumer.WithBrokers error: [%w]", cerrors.ErrValueRequired)
		}

		consumer.Brokers = make([]string, len(brokers))
		copy(consumer.Brokers, brokers)

		return nil
	}
}

// WithPartition sets partition.
func WithPartition(i int) Option {
	return func(consumer *Consumer) error {
		if i > 2147483647 || i < -2147483648 {
			return fmt.Errorf("kafkaconsumer.WithPartition error: [%w]", cerrors.ErrInvalid)
		}
		consumer.Partition = int32(i)

		return nil
	}
}

// WithDialTimeout sets dial timeout.
func WithDialTimeout(d time.Duration) Option {
	return func(consumer *Consumer) error {
		consumer.DialTimeout = d

		return nil
	}
}

// WithReadTimeout sets read timeout.
func WithReadTimeout(d time.Duration) Option {
	return func(consumer *Consumer) error {
		consumer.ReadTimeout = d

		return nil
	}
}

// WithWriteTimeout sets write timeout.
func WithWriteTimeout(d time.Duration) Option {
	return func(consumer *Consumer) error {
		consumer.WriteTimeout = d

		return nil
	}
}

// WithBackoff sets backoff duration.
func WithBackoff(d time.Duration) Option {
	return func(consumer *Consumer) error {
		if d == 0 {
			return fmt.Errorf("kafkaconsumer.WithBackoff error: [%w]", cerrors.ErrValueRequired)
		}
		consumer.Backoff = d

		return nil
	}
}

// WithMaxRetries sets max retries value.
func WithMaxRetries(i int) Option {
	return func(consumer *Consumer) error {
		if i > 255 || i < 0 {
			return fmt.Errorf("kafkaconsumer.WithMaxRetries error: [%w]", cerrors.ErrInvalid)
		}
		consumer.MaxRetries = uint8(i)

		return nil
	}
}

// WithStorage sets storage value.
func WithStorage(st storage.Storer) Option {
	return func(consumer *Consumer) error {
		if st == nil {
			return fmt.Errorf("kafkaconsumer.WithStorage error: [%w]", cerrors.ErrValueRequired)
		}
		consumer.Storage = st

		return nil
	}
}

// New instantiates new kafka consumer instance.
func New(options ...Option) (*Consumer, error) {
	consumer := new(Consumer)

	for _, option := range options {
		if err := option(consumer); err != nil {
			return nil, fmt.Errorf("kafkaconsumer.New option error: [%w]", err)
		}
	}

	if consumer.Logger == nil {
		return nil, fmt.Errorf("kafkaconsumer.New consumer.Logger error: [%w]", cerrors.ErrValueRequired)
	}
	if consumer.Storage == nil {
		return nil, fmt.Errorf("kafkaconsumer.New consumer.Pool error: [%w]", cerrors.ErrValueRequired)
	}

	if consumer.Topic == "" {
		return nil, fmt.Errorf("kafkaconsumer.New consumer.Topic error: [%w]", cerrors.ErrValueRequired)
	}

	return consumer, nil
}
