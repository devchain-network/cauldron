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

// ProcessMessageFunc is a factory function for callers.
type ProcessMessageFunc func(ctx context.Context, msg *sarama.ConsumerMessage) error

// Consumer represents kafa consumer setup.
type Consumer struct {
	Topic                     kafkacp.KafkaTopicIdentifier
	Logger                    *slog.Logger
	SaramaConsumer            sarama.Consumer
	SaramaConsumerFactoryFunc SaramaConsumerFactoryFunc
	ProcessMessageFunc        ProcessMessageFunc
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
		return fmt.Errorf(
			"[kafkaconsumer.checkRequired] Logger error: [%w, 'nil' received]",
			cerrors.ErrValueRequired,
		)
	}

	if c.ProcessMessageFunc == nil {
		return fmt.Errorf(
			"[kafkaconsumer.checkRequired] ProcessMessageFunc error: [%w, 'nil' received]",
			cerrors.ErrValueRequired,
		)
	}

	if !c.Topic.Valid() {
		return fmt.Errorf(
			"[kafkaconsumer.checkRequired] Topic error: [%w, false received]",
			cerrors.ErrInvalid,
		)
	}

	return nil
}

// Consume consumes kafka message with using partition consumer.
func (c Consumer) Consume() error {
	partitionConsumer, err := c.SaramaConsumer.ConsumePartition(c.Topic.String(), c.Partition, sarama.OffsetNewest)
	if err != nil {
		return fmt.Errorf("[kafkaconsumer.Consume][SaramaConsumer.ConsumePartition] error: [%w]", err)
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
				if err = c.ProcessMessageFunc(ctx, msg); err != nil {
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
			return fmt.Errorf(
				"[kafkaconsumer.WithLogger] error: [%w, 'nil' received]",
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
				"[kafkaconsumer.WithTopic] error: [%w, '%s' received]",
				cerrors.ErrInvalid, s,
			)
		}
		c.Topic = kt

		return nil
	}
}

// WithPartition sets partition.
func WithPartition(i int) Option {
	return func(c *Consumer) error {
		if i < 0 || i > math.MaxInt32 {
			return fmt.Errorf(
				"[kafkaconsumer.WithPartition] error: [%w, '%d' received, must > 0 or must < %d ]",
				cerrors.ErrInvalid, i, math.MaxInt32,
			)
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
			return fmt.Errorf(
				"[kafkaconsumer.WithKafkaBrokers] error: [%w, '%s' received]",
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
				"[kafkaconsumer.WithDialTimeout] error: [%w, '%s' received, must > 0]",
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
				"[kafkaconsumer.WithReadTimeout] error: [%w, '%s' received, must > 0]",
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
				"[kafkaconsumer.WithWriteTimeout] error: [%w, '%s' received, must > 0]",
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
				"[kafkaconsumer.WithBackoff] error: [%w, '%s' received, 0 is not allowed]",
				cerrors.ErrValueRequired, d,
			)
		}

		if d < 0 || d > time.Minute {
			return fmt.Errorf(
				"[kafkaconsumer.WithBackoff] error: [%w, '%s' received, must > 0 or < minute]",
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
				"[kafkaconsumer.WithMaxRetries] error: [%w, '%d' received, must < %d or > 0]",
				cerrors.ErrInvalid, i, math.MaxUint8,
			)
		}
		c.MaxRetries = uint8(i)

		return nil
	}
}

// WithSaramaConsumerFactoryFunc sets a custom factory function for creating Sarama consumers.
func WithSaramaConsumerFactoryFunc(fn SaramaConsumerFactoryFunc) Option {
	return func(c *Consumer) error {
		if fn == nil {
			return fmt.Errorf(
				"[kafkaconsumer.WithSaramaConsumerFactoryFunc] error: [%w, 'nil' received]",
				cerrors.ErrValueRequired,
			)
		}
		c.SaramaConsumerFactoryFunc = fn

		return nil
	}
}

// WithProcessMessageFunc sets the message processor.
func WithProcessMessageFunc(fn ProcessMessageFunc) Option {
	return func(c *Consumer) error {
		if fn == nil {
			return fmt.Errorf(
				"[kafkaconsumer.WithProcessMessageFunc] error: [%w, 'nil' received]",
				cerrors.ErrValueRequired,
			)
		}
		c.ProcessMessageFunc = fn

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
			return nil, err
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
		sconsumer, sconsumerErr = consumer.SaramaConsumerFactoryFunc(
			consumer.KafkaBrokers.ToStringSlice(),
			config,
		)
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
		return nil, fmt.Errorf(
			"[kafkaconsumer.New][SaramaConsumerFactoryFunc] error: [%w]",
			sconsumerErr,
		)
	}

	consumer.Logger.Info("successfully connected to", "broker", consumer.KafkaBrokers)

	consumer.SaramaConsumer = sconsumer

	return consumer, nil
}
