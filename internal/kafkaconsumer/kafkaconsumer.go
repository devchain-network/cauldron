package kafkaconsumer

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"time"

	"github.com/IBM/sarama"
	"github.com/devchain-network/cauldron/internal/cerrors"
	"github.com/devchain-network/cauldron/internal/storage"
)

// defaults values.
const (
	DefaultKafkaBrokers = "127.0.0.1:9094"

	DefaultKafkaConsumerPartition    = 0
	DefaultKafkaConsumerDialTimeout  = 30 * time.Second
	DefaultKafkaConsumerReadTimeout  = 30 * time.Second
	DefaultKafkaConsumerWriteTimeout = 30 * time.Second

	DefaultKafkaConsumerBackoff    = 2 * time.Second
	DefaultKafkaConsumerMaxRetries = 10
)

// KafkaConsumer defines kafka consumer behaviours.
type KafkaConsumer interface {
	Start() error
	Ping() error
	Worker(workerID int, messages <-chan *sarama.ConsumerMessage)
}

// GitHubKafkaConsumer defines Kafka GitHub message consumer.
type GitHubKafkaConsumer interface {
	KafkaConsumer
	StoreGitHubMessage(msg *sarama.ConsumerMessage) error
}

var (
	_ KafkaConsumer       = (*Consumer)(nil) // compile time proof
	_ GitHubKafkaConsumer = (*Consumer)(nil) // compile time proof
)

// Consumer represents kafa consumer setup.
type Consumer struct {
	Logger       *slog.Logger
	Storage      storage.Storer
	Consumer     sarama.Consumer
	Topic        string
	Brokers      []string
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	Backoff      time.Duration
	MaxRetries   uint8
	Partition    int32
}

// Option represents option function type.
type Option func(*Consumer) error

// Ping checks kafka consumer availability and sets consumer instance.
func (c *Consumer) Ping() error {
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
			"error", consumerErr,
			"retry", fmt.Sprintf("%d/%d", i, c.MaxRetries),
			"backoff", backoff.String(),
		)
		time.Sleep(backoff)
		backoff *= 2
	}

	if consumerErr != nil {
		return fmt.Errorf("kafkaconsumer.Consumer.Start error: [%w]", consumerErr)
	}

	c.Consumer = consumer

	return nil
}

// Start starts consumer.
func (c Consumer) Start() error {
	partitionConsumer, err := c.Consumer.ConsumePartition(c.Topic, c.Partition, sarama.OffsetNewest)
	if err != nil {
		return fmt.Errorf("kafkaconsumer.Consumer consumer.ConsumePartition error: [%w]", err)
	}
	defer func() { _ = partitionConsumer.Close() }()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c.Logger.Info("consuming messages from", "topic", c.Topic)

	messageBufferSize := runtime.NumCPU() * 10
	messageChan := make(chan *sarama.ConsumerMessage, messageBufferSize)

	numWorkers := runtime.NumCPU()
	c.Logger.Info("starting workers", "count", numWorkers)

	ch := make(chan struct{})
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()

		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt)
		<-sig

		c.Logger.Info("exiting message signal listener")
		cancel()
		close(ch)
	}()

	for i := range numWorkers {
		wg.Add(1)
		go func() {
			defer func() {
				wg.Done()
				c.Logger.Info("terminating worker", "worker id", i)
			}()
			c.Worker(i, messageChan)
		}()
	}

	wg.Add(1)
	go func() {
		defer func() {
			close(messageChan)
			wg.Done()
			c.Logger.Info("exiting message consumer")
		}()

		for {
			select {
			case msg := <-partitionConsumer.Messages():
				if msg != nil {
					messageChan <- msg
				}
			case err := <-partitionConsumer.Errors():
				c.Logger.Error("partition consumer error", "error", err)
			case <-ctx.Done():
				c.Logger.Info("shutting down message consumer")

				return
			}
		}
	}()

	<-ch
	wg.Wait()
	c.Logger.Info("all workers stopped")

	return nil
}

// Worker drains message queue.
func (c Consumer) Worker(workerID int, messages <-chan *sarama.ConsumerMessage) {
	for msg := range messages {
		switch KafkaTopicIdentifier(c.Topic) {
		case KafkaTopicIdentifierGitHub:
			if err := c.StoreGitHubMessage(msg); err != nil {
				c.Logger.Error("store github message error", "error", err, "worker id", workerID)

				continue
			}
		case KafkaTopicIdentifierGitLab:
			fmt.Println("parse GitLab kafka message, not implemented yet")
		default:
			fmt.Println("unknown topic identifier")
		}

		c.Logger.Info("github messages successfully stored to db", "worker id", workerID)
	}
}

func (c Consumer) getConfig() *sarama.Config {
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true
	config.Net.DialTimeout = c.DialTimeout
	config.Net.ReadTimeout = c.ReadTimeout
	config.Net.WriteTimeout = c.WriteTimeout

	return config
}

// WithLogger sets logger.
func WithLogger(l *slog.Logger) Option {
	return func(consumer *Consumer) error {
		if l == nil {
			return fmt.Errorf("kafkaconsumer.WithLogger consumer.Logger error: [%w]", cerrors.ErrValueRequired)
		}
		consumer.Logger = l

		return nil
	}
}

// WithTopic sets topic name.
func WithTopic(s string) Option {
	return func(consumer *Consumer) error {
		if s == "" {
			return fmt.Errorf("kafkaconsumer.WithLogger consumer.Topic error: [%w]", cerrors.ErrValueRequired)
		}
		consumer.Topic = s

		return nil
	}
}

// WithBrokers sets brokers list.
func WithBrokers(brokers []string) Option {
	return func(consumer *Consumer) error {
		if brokers == nil {
			return fmt.Errorf("kafkaconsumer.WithBrokers consumer.Brokers error: [%w]", cerrors.ErrValueRequired)
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
			return fmt.Errorf("kafkaconsumer.WithPartition consumer.Partition error: [%w]", cerrors.ErrInvalid)
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
			return fmt.Errorf("kafkaconsumer.WithBackoff consumer.Backoff error: [%w]", cerrors.ErrValueRequired)
		}
		consumer.Backoff = d

		return nil
	}
}

// WithMaxRetries sets max retries value.
func WithMaxRetries(i int) Option {
	return func(consumer *Consumer) error {
		if i > 255 || i < 0 {
			return fmt.Errorf("kafkaconsumer.WithMaxRetries consumer.MaxRetries error: [%w]", cerrors.ErrInvalid)
		}
		consumer.MaxRetries = uint8(i)

		return nil
	}
}

// WithStorage sets storage value.
func WithStorage(st storage.Storer) Option {
	return func(consumer *Consumer) error {
		if st == nil {
			return fmt.Errorf("kafkaconsumer.WithStorage consumer.Storage error: [%w]", cerrors.ErrValueRequired)
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
		return nil, fmt.Errorf("kafkaconsumer.New consumer.Storage error: [%w]", cerrors.ErrValueRequired)
	}

	if consumer.Topic == "" {
		return nil, fmt.Errorf("kafkaconsumer.New consumer.Topic error: [%w]", cerrors.ErrValueRequired)
	}

	return consumer, nil
}
