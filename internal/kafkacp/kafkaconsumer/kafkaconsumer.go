package kafkaconsumer

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"time"

	"github.com/IBM/sarama"
	"github.com/devchain-network/cauldron/internal/cerrors"
	"github.com/devchain-network/cauldron/internal/kafkacp"
	"github.com/devchain-network/cauldron/internal/storage"
)

// defaults.
const (
	DefaultPartition    = 0
	DefaultDialTimeout  = 30 * time.Second
	DefaultReadTimeout  = 30 * time.Second
	DefaultWriteTimeout = 30 * time.Second
	DefaultBackoff      = 2 * time.Second
	DefaultMaxRetries   = 10
)

var _ KafkaConsumer = (*Consumer)(nil) // compile time proof

// KafkaConsumer defines kafka consumer behaviours.
type KafkaConsumer interface {
	Consume() error
}

// SaramaConsumerFactoryFunc is a factory function.
type SaramaConsumerFactoryFunc func([]string, *sarama.Config) (sarama.Consumer, error)

// Consumer represents kafa consumer setup.
type Consumer struct {
	Topic                     kafkacp.KafkaTopicIdentifier
	Logger                    *slog.Logger
	Storage                   storage.PingStorer
	SaramaConsumer            sarama.Consumer
	SaramaConsumerFactoryFunc SaramaConsumerFactoryFunc
	KafkaBrokers              kafkacp.KafkaBrokers
	DialTimeout               time.Duration
	ReadTimeout               time.Duration
	WriteTimeout              time.Duration
	Backoff                   time.Duration
	MaxRetries                uint8
	Partition                 int32
	MessageBufferSize         int
	NumberOfWorkers           int
}

func (c *Consumer) checkRequired() error {
	if c.Logger == nil {
		return fmt.Errorf("kafka consumer check required, Logger error: [%w]", cerrors.ErrValueRequired)
	}

	if c.Storage == nil {
		return fmt.Errorf("kafka consumer check required, Storage error: [%w]", cerrors.ErrValueRequired)
	}

	if !c.Topic.Valid() {
		return fmt.Errorf("kafka consumer check required, Topic error: [%w]", cerrors.ErrInvalid)
	}

	return nil
}

// Consume consumes message and stores it to database.
func (c Consumer) Consume() error {
	partitionConsumer, err := c.SaramaConsumer.ConsumePartition(c.Topic.String(), c.Partition, sarama.OffsetNewest)
	if err != nil {
		return fmt.Errorf("kafka consumer partition consumer instantiation error: [%w]", err)
	}
	defer func() { _ = partitionConsumer.Close() }()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c.Logger.Info("consuming messages from", "topic", c.Topic, "partition", c.Partition)

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
				if err = c.Storage.MessageStore(ctx, msg); err != nil {
					c.Logger.Error("kafka consumer message store", "error", err, "worker", i)

					continue
				}

				c.Logger.Info("message is stored to database", "worker", i)
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

// Option represents option function type.
type Option func(*Consumer) error

// WithLogger sets logger.
func WithLogger(l *slog.Logger) Option {
	return func(c *Consumer) error {
		if l == nil {
			return fmt.Errorf("kafka consumer WithLogger error: [%w]", cerrors.ErrValueRequired)
		}
		c.Logger = l

		return nil
	}
}

// WithStorage sets storage value.
func WithStorage(st storage.PingStorer) Option {
	return func(c *Consumer) error {
		if st == nil {
			return fmt.Errorf("kafka consumer WithStorage error: [%w]", cerrors.ErrValueRequired)
		}
		c.Storage = st

		return nil
	}
}

// WithTopic sets topic name to consume.
func WithTopic(s string) Option {
	return func(c *Consumer) error {
		kt := kafkacp.KafkaTopicIdentifier(s)
		if !kt.Valid() {
			return fmt.Errorf("kafka consumer WithTopic error: [%w]", cerrors.ErrInvalid)
		}
		c.Topic = kt

		return nil
	}
}

// WithPartition sets partition.
func WithPartition(i int) Option {
	return func(c *Consumer) error {
		if i < 0 || i > math.MaxInt32 {
			return fmt.Errorf("kafka consumer WithPartition error: [%w]", cerrors.ErrInvalid)
		}
		c.Partition = int32(i)

		return nil
	}
}

// WithKafkaBrokers sets kafka brokers list.
func WithKafkaBrokers(brokers string) Option {
	return func(c *Consumer) error {
		var kafkaBrokers kafkacp.KafkaBrokers
		kafkaBrokers.AddFromString(brokers)
		if !kafkaBrokers.Valid() {
			return fmt.Errorf("kafka consumer WithKafkaBrokers error: [%w]", cerrors.ErrInvalid)
		}

		c.KafkaBrokers = kafkaBrokers

		return nil
	}
}

// WithDialTimeout sets dial timeout.
func WithDialTimeout(d time.Duration) Option {
	return func(c *Consumer) error {
		if d < 0 {
			return fmt.Errorf("kafka consumer WithDialTimeout error: [%w]", cerrors.ErrInvalid)
		}
		c.DialTimeout = d

		return nil
	}
}

// WithReadTimeout sets read timeout.
func WithReadTimeout(d time.Duration) Option {
	return func(c *Consumer) error {
		if d < 0 {
			return fmt.Errorf("kafka consumer WithReadTimeout error: [%w]", cerrors.ErrInvalid)
		}
		c.ReadTimeout = d

		return nil
	}
}

// WithWriteTimeout sets write timeout.
func WithWriteTimeout(d time.Duration) Option {
	return func(c *Consumer) error {
		if d < 0 {
			return fmt.Errorf("kafka consumer WithWriteTimeout error: [%w]", cerrors.ErrInvalid)
		}
		c.WriteTimeout = d

		return nil
	}
}

// WithBackoff sets backoff duration.
func WithBackoff(d time.Duration) Option {
	return func(c *Consumer) error {
		if d == 0 {
			return fmt.Errorf("kafka consumer WithBackoff error: [%w]", cerrors.ErrValueRequired)
		}

		if d < 0 || d > time.Minute {
			return fmt.Errorf("kafka consumer WithBackoff error: [%w]", cerrors.ErrInvalid)
		}

		c.Backoff = d

		return nil
	}
}

// WithMaxRetries sets max retries value.
func WithMaxRetries(i int) Option {
	return func(c *Consumer) error {
		if i > math.MaxUint8 || i < 0 {
			return fmt.Errorf("kafka consumer WithMaxRetries error: [%w]", cerrors.ErrInvalid)
		}
		c.MaxRetries = uint8(i)

		return nil
	}
}

// WithSaramaConsumerFactoryFunc sets a custom factory function for creating Sarama consumers.
func WithSaramaConsumerFactoryFunc(factory SaramaConsumerFactoryFunc) Option {
	return func(c *Consumer) error {
		if factory == nil {
			return fmt.Errorf("kafka consumer WithSaramaConsumerFactoryFunc error: [%w]", cerrors.ErrValueRequired)
		}
		c.SaramaConsumerFactoryFunc = factory

		return nil
	}
}

// New instantiates new kafka github consumer instance.
func New(options ...Option) (*Consumer, error) {
	consumer := new(Consumer)

	var kafkaBrokers kafkacp.KafkaBrokers
	kafkaBrokers.AddFromString(kafkacp.DefaultKafkaBrokers)

	consumer.KafkaBrokers = kafkaBrokers
	consumer.DialTimeout = DefaultDialTimeout
	consumer.ReadTimeout = DefaultReadTimeout
	consumer.WriteTimeout = DefaultWriteTimeout
	consumer.Backoff = DefaultBackoff
	consumer.MaxRetries = DefaultMaxRetries
	consumer.NumberOfWorkers = runtime.NumCPU()
	consumer.MessageBufferSize = consumer.NumberOfWorkers * 10
	consumer.SaramaConsumerFactoryFunc = sarama.NewConsumer

	for _, option := range options {
		if err := option(consumer); err != nil {
			return nil, fmt.Errorf("kafka consumer option error: [%w]", err)
		}
	}

	if err := consumer.checkRequired(); err != nil {
		return nil, err
	}

	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true
	config.Net.DialTimeout = consumer.DialTimeout
	config.Net.ReadTimeout = consumer.ReadTimeout
	config.Net.WriteTimeout = consumer.WriteTimeout

	var sconsumer sarama.Consumer
	var sconsumerErr error
	backoff := consumer.Backoff

	for i := range consumer.MaxRetries {
		sconsumer, sconsumerErr = consumer.SaramaConsumerFactoryFunc(consumer.KafkaBrokers.ToStringSlice(), config)
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
		return nil, fmt.Errorf("kafka consumer NewConsumer error: [%w]", sconsumerErr)
	}

	consumer.Logger.Info("successfully connected to", "broker", consumer.KafkaBrokers)

	consumer.SaramaConsumer = sconsumer

	return consumer, nil
}
