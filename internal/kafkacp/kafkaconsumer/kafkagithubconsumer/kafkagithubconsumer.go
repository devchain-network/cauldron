package kafkagithubconsumer

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/IBM/sarama"
	"github.com/devchain-network/cauldron/internal/cerrors"
	"github.com/devchain-network/cauldron/internal/kafkacp"
	"github.com/devchain-network/cauldron/internal/kafkacp/kafkaconsumer"
	"github.com/devchain-network/cauldron/internal/storage"
	"github.com/devchain-network/cauldron/internal/storage/githubstorage"
	"github.com/google/uuid"
)

var _ kafkaconsumer.KafkaConsumer = (*Consumer)(nil) // compile time proof

// PrepareGitHubPayloadFunc represents header extract and prepagre payload function type.
type PrepareGitHubPayloadFunc func(msg *sarama.ConsumerMessage) (*githubstorage.GitHub, error)

// PrepareGitHubPayload extracts required values from kafka message header.
func PrepareGitHubPayload(msg *sarama.ConsumerMessage) (*githubstorage.GitHub, error) {
	githubStorage := new(githubstorage.GitHub)

	deliveryID, err := uuid.Parse(string(msg.Key))
	if err != nil {
		return nil, fmt.Errorf("kafkaconsumer.PrepareGitHubPayload deliveryID error: [%w]", err)
	}
	githubStorage.DeliveryID = deliveryID

	var targetID uint64
	var targetIDErr error

	var hookID uint64
	var hookIDErr error

	var userID int64
	var userIDErr error

	for _, header := range msg.Headers {
		key := string(header.Key)
		value := string(header.Value)

		switch key {
		case "event":
			githubStorage.Event = value
		case "target-type":
			githubStorage.Target = value
		case "target-id":
			targetID, targetIDErr = strconv.ParseUint(value, 10, 64)
			if targetIDErr != nil {
				return nil, fmt.Errorf("kafkaconsumer.PrepareGitHubPayload targetID error: [%w]", targetIDErr)
			}
			githubStorage.TargetID = targetID
		case "hook-id":
			hookID, hookIDErr = strconv.ParseUint(value, 10, 64)
			if hookIDErr != nil {
				return nil, fmt.Errorf("kafkaconsumer.PrepareGitHubPayload hookID error: [%w]", hookIDErr)
			}
			githubStorage.HookID = hookID
		case "sender-login":
			githubStorage.UserLogin = value
		case "sender-id":
			userID, userIDErr = strconv.ParseInt(value, 10, 64)
			if userIDErr != nil {
				return nil, fmt.Errorf("kafkaconsumer.PrepareGitHubPayload userID error: [%w]", userIDErr)
			}
			githubStorage.UserID = userID
		}
	}

	return githubStorage, nil
}

// Consumer represents kafa consumer setup.
type Consumer struct {
	ConfigFunc        kafkaconsumer.ConfigFunc
	ConsumerFunc      kafkaconsumer.ConsumerFunc
	Logger            *slog.Logger
	Storage           storage.PingStorer
	SaramaConsumer    sarama.Consumer
	KafkaBrokers      kafkacp.KafkaBrokers
	Backoff           time.Duration
	MaxRetries        uint8
	MessageBufferSize int
	NumberOfWorkers   int
}

// Consume consumes message and stores it to database.
func (c Consumer) Consume(topic kafkacp.KafkaTopicIdentifier, partition int32) error {
	partitionConsumer, err := c.SaramaConsumer.ConsumePartition(topic.String(), partition, sarama.OffsetNewest)
	if err != nil {
		return fmt.Errorf("kafkagithubconsumer.Consume c.SaramaConsumer.ConsumePartition error: [%w]", err)
	}
	defer func() { _ = partitionConsumer.Close() }()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c.Logger.Info("consuming messages from", "topic", topic, "partition", partition)

	messagesQueue := make(chan *sarama.ConsumerMessage, c.MessageBufferSize)
	c.Logger.Info("starting workers", "count", c.NumberOfWorkers)

	done := make(chan struct{})
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()

		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt)
		<-sig

		c.Logger.Info("interrupt received, exiting signal listener")
		cancel()
		close(done)
	}()

	for i := range c.NumberOfWorkers {
		wg.Add(1)
		go func() {
			defer func() {
				wg.Done()
				c.Logger.Info("terminating worker", "id", i)
			}()

			for msg := range messagesQueue {
				if err = c.Storage.Store(ctx, msg); err != nil {
					c.Logger.Error("github c.Storage.Store", "error", err, "worker", i)

					continue
				}

				c.Logger.Info("message stored to github storage", "worker", i)
			}
		}()
	}

	wg.Add(1)
	go func() {
		defer func() {
			close(messagesQueue)
			wg.Done()
			c.Logger.Info("exiting message consumer")
		}()

		for {
			select {
			case msg := <-partitionConsumer.Messages():
				if msg != nil {
					messagesQueue <- msg
				}
			case err := <-partitionConsumer.Errors():
				c.Logger.Error("partition consumer error", "error", err)
			case <-ctx.Done():
				c.Logger.Info("shutting down message consumer")

				return
			}
		}
	}()

	<-done
	wg.Wait()

	c.Logger.Info("all workers are stopped")

	return nil
}

func (c *Consumer) checkRequired() error {
	if c.Logger == nil {
		return fmt.Errorf("kafkagithubconsumer.New Logger error: [%w]", cerrors.ErrValueRequired)
	}

	if c.Storage == nil {
		return fmt.Errorf("kafkagithubconsumer.New Storage error: [%w]", cerrors.ErrValueRequired)
	}

	return nil
}

// Option represents option function type.
type Option func(*Consumer) error

// WithLogger sets logger.
func WithLogger(l *slog.Logger) Option {
	return func(c *Consumer) error {
		if l == nil {
			return fmt.Errorf("kafkagithubconsumer.WithLogger error: [%w]", cerrors.ErrValueRequired)
		}
		c.Logger = l

		return nil
	}
}

// WithStorage sets storage value.
func WithStorage(st storage.PingStorer) Option {
	return func(c *Consumer) error {
		if st == nil {
			return fmt.Errorf("kafkagithubconsumer.WithStorage error: [%w]", cerrors.ErrValueRequired)
		}
		c.Storage = st

		return nil
	}
}

// WithConfigFunc sets config function.
func WithConfigFunc(f kafkaconsumer.ConfigFunc) Option {
	return func(c *Consumer) error {
		if f != nil {
			c.ConfigFunc = f
		}

		return nil
	}
}

// WithConsumerFunc sets consumer function.
func WithConsumerFunc(f kafkaconsumer.ConsumerFunc) Option {
	return func(c *Consumer) error {
		if f != nil {
			c.ConsumerFunc = f
		}

		return nil
	}
}

// New instantiates new kafka github consumer instance.
func New(options ...Option) (*Consumer, error) {
	consumer := new(Consumer)
	consumer.Backoff = kafkacp.DefaultKafkaConsumerBackoff
	consumer.MaxRetries = kafkacp.DefaultKafkaConsumerMaxRetries
	consumer.ConfigFunc = kafkaconsumer.GetDefaultConfig
	consumer.ConsumerFunc = kafkaconsumer.GetDefaultConsumerFunc
	consumer.NumberOfWorkers = runtime.NumCPU()
	consumer.MessageBufferSize = consumer.NumberOfWorkers * 10

	for _, option := range options {
		if err := option(consumer); err != nil {
			return nil, fmt.Errorf("kafkagithubconsumer.New option error: [%w]", err)
		}
	}

	if err := consumer.checkRequired(); err != nil {
		return nil, err
	}

	var sconsumer sarama.Consumer
	var sconsumerErr error
	backoff := consumer.Backoff

	for i := range consumer.MaxRetries {
		sconsumer, sconsumerErr = consumer.ConsumerFunc(consumer.KafkaBrokers, consumer.ConfigFunc())
		if sconsumerErr == nil {
			break
		}

		consumer.Logger.Error(
			"can not connect to",
			"brokers", consumer.KafkaBrokers,
			"error", sconsumerErr,
			"retry", fmt.Sprintf("%d/%d", i, consumer.MaxRetries),
			"backoff", backoff.String(),
		)
		time.Sleep(backoff)
		backoff *= 2
	}

	if sconsumerErr != nil {
		return nil, fmt.Errorf("kafkagithubconsumer.New error: [%w]", sconsumerErr)
	}

	consumer.SaramaConsumer = sconsumer

	return consumer, nil
}
