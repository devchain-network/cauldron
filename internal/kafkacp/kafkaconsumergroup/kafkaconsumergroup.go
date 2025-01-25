package kafkaconsumergroup

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
	DefaultDialTimeout  = 30 * time.Second
	DefaultReadTimeout  = 30 * time.Second
	DefaultWriteTimeout = 30 * time.Second
	DefaultBackoff      = 2 * time.Second
	DefaultMaxRetries   = 10
)

var _ sarama.ConsumerGroupHandler = (*Consumer)(nil) // compile time proof

// Consumer represents kafa group consumer setup.
type Consumer struct {
	KafkaGroupName string
	Logger         *slog.Logger
	Storage        storage.PingStorer
	// SaramaConfig                   *sarama.Config
	SaramaConsumerGroupFactoryFunc SaramaConsumerGroupFactoryFunc
	SaramaConsumerGroup            sarama.ConsumerGroup
	Topic                          kafkacp.KafkaTopicIdentifier
	KafkaBrokers                   kafkacp.KafkaBrokers
	KafkaVersion                   sarama.KafkaVersion
	DialTimeout                    time.Duration
	ReadTimeout                    time.Duration
	WriteTimeout                   time.Duration
	Backoff                        time.Duration
	MaxRetries                     uint8
	NumberOfWorkers                int
}

// SaramaConsumerGroupFactoryFunc is a factory function.
type SaramaConsumerGroupFactoryFunc func([]string, string, *sarama.Config) (sarama.ConsumerGroup, error)

func (c *Consumer) checkRequired() error {
	if c.Logger == nil {
		return fmt.Errorf("kafka consumer group check required, Logger error: [%w]", cerrors.ErrValueRequired)
	}

	if c.Storage == nil {
		return fmt.Errorf("kafka consumer group check required, Storage error: [%w]", cerrors.ErrValueRequired)
	}

	if c.KafkaGroupName == "" {
		return fmt.Errorf("kafka consumer group check required, KafkaGroupName error: [%w]", cerrors.ErrValueRequired)
	}

	if !c.Topic.Valid() {
		return fmt.Errorf("kafka consumer group check required, Topic error: [%w]", cerrors.ErrInvalid)
	}

	return nil
}

// Setup ...
func (Consumer) Setup(_ sarama.ConsumerGroupSession) error { return nil }

// Cleanup ...
func (Consumer) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }

// ConsumeClaim ...
func (Consumer) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		fmt.Println(msg.Topic, msg.Partition, msg.Offset, string(msg.Key), string(msg.Value))

		sess.MarkMessage(msg, "")
	}

	return nil
}

// StartConsume ...
func (c Consumer) StartConsume() error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()

		for {
			select {
			case err, ok := <-c.SaramaConsumerGroup.Errors():
				if !ok {
					c.Logger.Info("error chan closed, exiting error handler")

					return
				}
				c.Logger.Error("group errrrrrrrrr", "error", err)
			case <-ctx.Done():
				c.Logger.Info("context canceled, stopping error handler")

				return
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		topics := []string{c.Topic.String()}

		for {
			if ctx.Err() != nil {
				c.Logger.Info("context canceled, stopping consumer loop")

				return
			}

			if err := c.SaramaConsumerGroup.Consume(ctx, topics, c); err != nil {
				if ctx.Err() != nil {
					c.Logger.Info("consume stopped due to context cancellation")

					return
				}

				c.Logger.Error("erroooo", "error", err)
			}
		}
	}()

	<-ctx.Done()
	wg.Wait()

	if err := c.SaramaConsumerGroup.Close(); err != nil {
		return fmt.Errorf("failed to close consumer group, error: [%w]", err)
	}

	c.Logger.Info("all workers are stopped")

	return nil
}

// Option represents option function type.
type Option func(*Consumer) error

// WithLogger sets logger.
func WithLogger(l *slog.Logger) Option {
	return func(c *Consumer) error {
		if l == nil {
			return fmt.Errorf("kafka consumer group WithLogger error: [%w]", cerrors.ErrValueRequired)
		}
		c.Logger = l

		return nil
	}
}

// WithStorage sets storage value.
func WithStorage(st storage.PingStorer) Option {
	return func(c *Consumer) error {
		if st == nil {
			return fmt.Errorf("kafka consumer group WithStorage error: [%w]", cerrors.ErrValueRequired)
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
			return fmt.Errorf("kafka consumer group WithTopic error: [%w]", cerrors.ErrInvalid)
		}
		c.Topic = kt

		return nil
	}
}

// WithKafkaVersion sets kafka version.
func WithKafkaVersion(s string) Option {
	return func(c *Consumer) error {
		version, err := sarama.ParseKafkaVersion(s)
		if err != nil {
			return fmt.Errorf("kafka consumer group WithKafkaVersion error: [%w][%w]", err, cerrors.ErrInvalid)
		}

		c.KafkaVersion = version

		return nil
	}
}

// WithKafkaBrokers sets kafka brokers list.
func WithKafkaBrokers(brokers string) Option {
	return func(c *Consumer) error {
		var kafkaBrokers kafkacp.KafkaBrokers
		kafkaBrokers.AddFromString(brokers)
		if !kafkaBrokers.Valid() {
			return fmt.Errorf("kafka consumer group WithKafkaBrokers error: [%w]", cerrors.ErrInvalid)
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
			return fmt.Errorf("kafka consumer group WithReadTimeout error: [%w]", cerrors.ErrInvalid)
		}
		c.ReadTimeout = d

		return nil
	}
}

// WithWriteTimeout sets write timeout.
func WithWriteTimeout(d time.Duration) Option {
	return func(c *Consumer) error {
		if d < 0 {
			return fmt.Errorf("kafka consumer group WithWriteTimeout error: [%w]", cerrors.ErrInvalid)
		}
		c.WriteTimeout = d

		return nil
	}
}

// WithBackoff sets backoff duration.
func WithBackoff(d time.Duration) Option {
	return func(c *Consumer) error {
		if d == 0 {
			return fmt.Errorf("kafka consumer group WithBackoff error: [%w]", cerrors.ErrValueRequired)
		}

		if d < 0 || d > time.Minute {
			return fmt.Errorf("kafka consumer group WithBackoff error: [%w]", cerrors.ErrInvalid)
		}

		c.Backoff = d

		return nil
	}
}

// WithMaxRetries sets max retries value.
func WithMaxRetries(i int) Option {
	return func(c *Consumer) error {
		if i > math.MaxUint8 || i < 0 {
			return fmt.Errorf("kafka consumer group WithMaxRetries error: [%w]", cerrors.ErrInvalid)
		}
		c.MaxRetries = uint8(i)

		return nil
	}
}

// WithKafkaGroupName sets kafka consumer group name.
func WithKafkaGroupName(s string) Option {
	return func(c *Consumer) error {
		if s == "" {
			return fmt.Errorf("kafka consumer group WithKafkaGroupName error: [%w]", cerrors.ErrValueRequired)
		}
		c.KafkaGroupName = s

		return nil
	}
}

// WithSaramaConsumerGroupFactoryFunc sets a custom factory function for Sarama consumer group.
func WithSaramaConsumerGroupFactoryFunc(factory SaramaConsumerGroupFactoryFunc) Option {
	return func(c *Consumer) error {
		if factory == nil {
			return fmt.Errorf(
				"kafka consumer group WithSaramaConsumerGroupFactoryFunc error: [%w]",
				cerrors.ErrValueRequired,
			)
		}

		c.SaramaConsumerGroupFactoryFunc = factory

		return nil
	}
}

// New instantiates new kafka github consumer group instance.
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
	consumer.KafkaVersion = sarama.V3_9_0_0
	consumer.SaramaConsumerGroupFactoryFunc = sarama.NewConsumerGroup

	for _, option := range options {
		if err := option(consumer); err != nil {
			return nil, fmt.Errorf("kafka consumer group option error: [%w]", err)
		}
	}

	if err := consumer.checkRequired(); err != nil {
		return nil, err
	}

	config := sarama.NewConfig()
	config.Net.DialTimeout = consumer.DialTimeout
	config.Net.ReadTimeout = consumer.ReadTimeout
	config.Net.WriteTimeout = consumer.WriteTimeout
	config.Version = sarama.V3_9_0_0
	config.Consumer.Return.Errors = true

	var saramaConsumerGroup sarama.ConsumerGroup
	var saramaConsumerGroupErr error
	backoff := consumer.Backoff

	for i := range consumer.MaxRetries {
		saramaConsumerGroup, saramaConsumerGroupErr = consumer.SaramaConsumerGroupFactoryFunc(
			consumer.KafkaBrokers.ToStringSlice(),
			consumer.KafkaGroupName,
			config,
		)

		if saramaConsumerGroupErr == nil {
			break
		}

		consumer.Logger.Error(
			"can not connect to",
			"brokers", consumer.KafkaBrokers,
			"error", saramaConsumerGroupErr,
			"retry", fmt.Sprintf("%d/%d", i, consumer.MaxRetries),
			"backoff", backoff.String(),
		)
		time.Sleep(backoff)
		backoff *= 2
	}
	if saramaConsumerGroupErr != nil {
		return nil, fmt.Errorf("kafka consumer group, group error: [%w]", saramaConsumerGroupErr)
	}

	consumer.Logger.Info("successfully connected to", "broker", consumer.KafkaBrokers)

	consumer.SaramaConsumerGroup = saramaConsumerGroup

	return consumer, nil
}
