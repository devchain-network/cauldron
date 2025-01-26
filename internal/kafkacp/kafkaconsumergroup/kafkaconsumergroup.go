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
	KafkaGroupName                 string
	Logger                         *slog.Logger
	SaramaConsumerGroupFactoryFunc SaramaConsumerGroupFactoryFunc
	ProcessMessageFunc             ProcessMessageFunc
	MessageQueue                   chan *sarama.ConsumerMessage
	SaramaConsumerGroup            sarama.ConsumerGroup
	SaramaConsumerGroupHandler     sarama.ConsumerGroupHandler
	Topic                          kafkacp.KafkaTopicIdentifier
	KafkaBrokers                   kafkacp.KafkaBrokers
	KafkaVersion                   sarama.KafkaVersion
	DialTimeout                    time.Duration
	ReadTimeout                    time.Duration
	WriteTimeout                   time.Duration
	Backoff                        time.Duration
	MaxRetries                     uint8
	MessageBufferSize              int
	NumberOfWorkers                int
}

// SaramaConsumerGroupFactoryFunc is a factory function.
type SaramaConsumerGroupFactoryFunc func([]string, string, *sarama.Config) (sarama.ConsumerGroup, error)

// ProcessMessageFunc is a factory function for callers.
type ProcessMessageFunc func(ctx context.Context, msg *sarama.ConsumerMessage) error

func (c *Consumer) checkRequired() error {
	if c.Logger == nil {
		return fmt.Errorf(
			"[kafkaconsumergroup.checkRequired] Logger error: [%w, 'nil' received]",
			cerrors.ErrValueRequired,
		)
	}

	if c.ProcessMessageFunc == nil {
		return fmt.Errorf(
			"[kafkaconsumergroup.checkRequired] ProcessMessageFunc error: [%w, 'nil' received]",
			cerrors.ErrValueRequired,
		)
	}

	if c.KafkaGroupName == "" {
		return fmt.Errorf(
			"[kafkaconsumergroup.checkRequired] KafkaGroupName error: [%w, empty string received]",
			cerrors.ErrValueRequired,
		)
	}

	if !c.Topic.Valid() {
		return fmt.Errorf(
			"[kafkaconsumergroup.checkRequired] Topic error: [%w, false received]",
			cerrors.ErrInvalid,
		)
	}

	return nil
}

// Setup implements sarama ConsumerGroupHandler interface.
func (Consumer) Setup(_ sarama.ConsumerGroupSession) error { return nil }

// Cleanup implements sarama ConsumerGroupHandler interface.
func (Consumer) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }

// ConsumeClaim implements sarama ConsumerGroupHandler interface.
func (c Consumer) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		c.MessageQueue <- msg

		sess.MarkMessage(msg, "")
	}

	return nil
}

// StartConsume consumes message from kafka.
func (c Consumer) StartConsume() error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	topics := []string{c.Topic.String()}

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
				c.Logger.Error("kafka consumer group error", "error", err)
			case <-ctx.Done():
				c.Logger.Info("context canceled, stopping error handler")

				return
			}
		}
	}()

	c.Logger.Info("starting workers", "count", c.NumberOfWorkers)
	for i := range c.NumberOfWorkers {
		wg.Add(1)

		go func() {
			defer func() {
				wg.Done()
				c.Logger.Info("worker stopped", "id", i)
			}()

			for {
				select {
				case msg, ok := <-c.MessageQueue:
					if !ok {
						return
					}

					if err := c.ProcessMessageFunc(ctx, msg); err != nil {
						c.Logger.Error("kafka consumer group process message", "error", err, "worker", i)

						continue
					}

					c.Logger.Info(
						"message is stored to database",
						"worker", i,
						"topic", msg.Topic,
						"partition", msg.Partition,
						"offset", msg.Offset,
						"key", string(msg.Key),
					)

				case <-ctx.Done():
					return
				}
			}
		}()
	}

	wg.Add(1)
	go func() {
		defer func() {
			wg.Done()
			close(c.MessageQueue)
		}()

		for {
			if ctx.Err() != nil {
				c.Logger.Info("context canceled, stopping consumer loop")

				return
			}

			if err := c.SaramaConsumerGroup.Consume(ctx, topics, c.SaramaConsumerGroupHandler); err != nil {
				if ctx.Err() != nil {
					c.Logger.Info("consume stopped due to context cancellation")

					return
				}

				c.Logger.Error("kafka consume group consume", "error", err)
			}
		}
	}()

	<-ctx.Done()
	wg.Wait()

	if err := c.SaramaConsumerGroup.Close(); err != nil {
		return fmt.Errorf(
			"[kafkaconsumergroup.StartConsume][SaramaConsumerGroup.Close] error: [%w]",
			err,
		)
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
			return fmt.Errorf(
				"[kafkaconsumergroup.WithLogger] error: [%w, 'nil' received]",
				cerrors.ErrValueRequired,
			)
		}
		c.Logger = l

		return nil
	}
}

// WithTopic sets topic name to consume.
func WithTopic(s string) Option {
	return func(c *Consumer) error {
		kt := kafkacp.KafkaTopicIdentifier(s)
		if !kt.Valid() {
			return fmt.Errorf(
				"[kafkaconsumergroup.WithTopic] error: [%w, '%s' received]",
				cerrors.ErrInvalid, s,
			)
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
			return fmt.Errorf(
				"[kafkaconsumergroup.WithKafkaVersion] error: [(%w) %w, '%s' received]",
				err, cerrors.ErrInvalid, s,
			)
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
			return fmt.Errorf(
				"[kafkaconsumergroup.WithKafkaBrokers] error: [%w, '%s' received]",
				cerrors.ErrInvalid, brokers,
			)
		}

		c.KafkaBrokers = kafkaBrokers

		return nil
	}
}

// WithDialTimeout sets dial timeout.
func WithDialTimeout(d time.Duration) Option {
	return func(c *Consumer) error {
		if d < 0 {
			return fmt.Errorf(
				"[kafkaconsumergroup.WithDialTimeout] error: [%w, '%s' received, must > 0]",
				cerrors.ErrInvalid, d,
			)
		}
		c.DialTimeout = d

		return nil
	}
}

// WithReadTimeout sets read timeout.
func WithReadTimeout(d time.Duration) Option {
	return func(c *Consumer) error {
		if d < 0 {
			return fmt.Errorf(
				"[kafkaconsumergroup.WithReadTimeout] error: [%w, '%s' received, must > 0]",
				cerrors.ErrInvalid, d,
			)
		}
		c.ReadTimeout = d

		return nil
	}
}

// WithWriteTimeout sets write timeout.
func WithWriteTimeout(d time.Duration) Option {
	return func(c *Consumer) error {
		if d < 0 {
			return fmt.Errorf(
				"[kafkaconsumergroup.WithWriteTimeout] error: [%w, '%s' received, must > 0]",
				cerrors.ErrInvalid, d,
			)
		}
		c.WriteTimeout = d

		return nil
	}
}

// WithBackoff sets backoff duration.
func WithBackoff(d time.Duration) Option {
	return func(c *Consumer) error {
		if d == 0 {
			return fmt.Errorf(
				"[kafkaconsumergroup.WithBackoff] error: [%w, '%s' received, 0 is not allowed]",
				cerrors.ErrValueRequired, d,
			)
		}

		if d < 0 || d > time.Minute {
			return fmt.Errorf(
				"[kafkaconsumergroup.WithBackoff] error: [%w, '%s' received, must > 0 or < minute]",
				cerrors.ErrInvalid, d,
			)
		}

		c.Backoff = d

		return nil
	}
}

// WithMaxRetries sets max retries value.
func WithMaxRetries(i int) Option {
	return func(c *Consumer) error {
		if i > math.MaxUint8 || i < 0 {
			return fmt.Errorf(
				"[kafkaconsumergroup.WithMaxRetries] error: [%w, '%d' received, must < %d or > 0]",
				cerrors.ErrInvalid, i, math.MaxUint8,
			)
		}
		c.MaxRetries = uint8(i)

		return nil
	}
}

// WithKafkaGroupName sets kafka consumer group name.
func WithKafkaGroupName(s string) Option {
	return func(c *Consumer) error {
		if s == "" {
			return fmt.Errorf(
				"[kafkaconsumergroup.WithKafkaGroupName] error: [%w, empty string received]",
				cerrors.ErrValueRequired,
			)
		}
		c.KafkaGroupName = s

		return nil
	}
}

// WithSaramaConsumerGroupFactoryFunc sets a custom factory function for Sarama consumer group.
func WithSaramaConsumerGroupFactoryFunc(fn SaramaConsumerGroupFactoryFunc) Option {
	return func(c *Consumer) error {
		if fn == nil {
			return fmt.Errorf(
				"[kafkaconsumergroup.WithSaramaConsumerGroupFactoryFunc] error: [%w, 'nil' received]",
				cerrors.ErrValueRequired,
			)
		}

		c.SaramaConsumerGroupFactoryFunc = fn

		return nil
	}
}

// WithProcessMessageFunc sets the message processor.
func WithProcessMessageFunc(fn ProcessMessageFunc) Option {
	return func(c *Consumer) error {
		if fn == nil {
			return fmt.Errorf(
				"[kafkaconsumergroup.WithProcessMessageFunc] error: [%w, 'nil' received]",
				cerrors.ErrValueRequired,
			)
		}
		c.ProcessMessageFunc = fn

		return nil
	}
}

// WithSaramaConsumerGroupHandler sets sarama consumer group handler.
func WithSaramaConsumerGroupHandler(handler sarama.ConsumerGroupHandler) Option {
	return func(c *Consumer) error {
		if handler == nil {
			return fmt.Errorf(
				"[kafkaconsumergroup.WithSaramaConsumerGroupHandler] error: [%w, 'nil' received]",
				cerrors.ErrValueRequired,
			)
		}
		c.SaramaConsumerGroupHandler = handler

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
	consumer.MessageBufferSize = consumer.NumberOfWorkers * 10
	consumer.KafkaVersion = sarama.V3_9_0_0
	consumer.SaramaConsumerGroupFactoryFunc = sarama.NewConsumerGroup
	consumer.MessageQueue = make(chan *sarama.ConsumerMessage, consumer.MessageBufferSize)

	for _, option := range options {
		if err := option(consumer); err != nil {
			return nil, err
		}
	}

	if err := consumer.checkRequired(); err != nil {
		return nil, err
	}

	if consumer.SaramaConsumerGroupHandler == nil {
		consumer.SaramaConsumerGroupHandler = consumer
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
		return nil, fmt.Errorf(
			"[kafkaconsumergroup.New][SaramaConsumerGroupFactoryFunc] error: [%w]",
			saramaConsumerGroupErr,
		)
	}

	consumer.Logger.Info("successfully connected to", "broker", consumer.KafkaBrokers)

	consumer.SaramaConsumerGroup = saramaConsumerGroup

	return consumer, nil
}
