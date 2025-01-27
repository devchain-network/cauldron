package kafkaproducer

import (
	"fmt"
	"log/slog"
	"math"
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

// Producer holds required arguments.
type Producer struct {
	Logger                    *slog.Logger
	SaramaProducerFactoryFunc SaramaProducerFactoryFunc
	KafkaBrokers              kafkacp.KafkaBrokers
	DialTimeout               time.Duration
	ReadTimeout               time.Duration
	WriteTimeout              time.Duration
	MaxRetries                uint8
	Backoff                   time.Duration
}

// SaramaProducerFactoryFunc is a factory function.
type SaramaProducerFactoryFunc func([]string, *sarama.Config) (sarama.AsyncProducer, error)

func (p Producer) checkRequired() error {
	if p.Logger == nil {
		return fmt.Errorf(
			"[kafkaproducer.checkRequired] Logger error: [%w, 'nil' received]",
			cerrors.ErrValueRequired,
		)
	}

	return nil
}

// Option represents option function type.
type Option func(*Producer) error

// WithLogger sets logger.
func WithLogger(l *slog.Logger) Option {
	return func(p *Producer) error {
		if l == nil {
			return fmt.Errorf(
				"[kafkaproducer.WithLogger] error: [%w, 'nil' received]",
				cerrors.ErrValueRequired,
			)
		}
		p.Logger = l

		return nil
	}
}

// WithKafkaBrokers sets kafka brokers list.
func WithKafkaBrokers(brokers string) Option {
	return func(p *Producer) error {
		var kafkaBrokers kafkacp.KafkaBrokers
		kafkaBrokers.AddFromString(brokers)

		if !kafkaBrokers.Valid() {
			return fmt.Errorf(
				"[kafkaproducer.WithKafkaBrokers] error: [%w, '%s' received]",
				cerrors.ErrInvalid, brokers,
			)
		}

		p.KafkaBrokers = kafkaBrokers

		return nil
	}
}

// WithMaxRetries sets max retries value.
func WithMaxRetries(i int) Option {
	return func(p *Producer) error {
		if i > math.MaxUint8 || i < 0 {
			return fmt.Errorf(
				"[kafkaproducer.WithMaxRetries] error: [%w, '%d' received, must < %d or > 0]",
				cerrors.ErrInvalid, i, math.MaxUint8,
			)
		}
		p.MaxRetries = uint8(i)

		return nil
	}
}

// WithBackoff sets backoff duration.
func WithBackoff(d time.Duration) Option {
	return func(p *Producer) error {
		if d == 0 {
			return fmt.Errorf(
				"[kafkaproducer.WithBackoff] error: [%w, '%s' received, 0 is not allowed]",
				cerrors.ErrValueRequired, d,
			)
		}

		if d < 0 || d > time.Minute {
			return fmt.Errorf(
				"[kafkaproducer.WithBackoff] error: [%w, '%s' received, must > 0 or < minute]",
				cerrors.ErrInvalid, d,
			)
		}

		p.Backoff = d

		return nil
	}
}

// WithDialTimeout sets dial timeout.
func WithDialTimeout(d time.Duration) Option {
	return func(p *Producer) error {
		if d < 0 {
			return fmt.Errorf(
				"[kafkaproducer.WithDialTimeout] error: [%w, '%s' received, must > 0]",
				cerrors.ErrInvalid, d,
			)
		}
		p.DialTimeout = d

		return nil
	}
}

// WithReadTimeout sets read timeout.
func WithReadTimeout(d time.Duration) Option {
	return func(p *Producer) error {
		if d < 0 {
			return fmt.Errorf(
				"[kafkaproducer.WithReadTimeout] error: [%w, '%s' received, must > 0]",
				cerrors.ErrInvalid, d,
			)
		}
		p.ReadTimeout = d

		return nil
	}
}

// WithWriteTimeout sets write timeout.
func WithWriteTimeout(d time.Duration) Option {
	return func(p *Producer) error {
		if d < 0 {
			return fmt.Errorf(
				"[kafkaproducer.WithWriteTimeout] error: [%w, '%s' received, must > 0]",
				cerrors.ErrInvalid, d,
			)
		}
		p.WriteTimeout = d

		return nil
	}
}

// WithSaramaProducerFactoryFunc sets a custom factory function for creating Sarama producers.
func WithSaramaProducerFactoryFunc(fn SaramaProducerFactoryFunc) Option {
	return func(p *Producer) error {
		if fn == nil {
			return fmt.Errorf(
				"[kafkaproducer.WithSaramaProducerFactoryFunc] error: [%w, 'nil' received]",
				cerrors.ErrValueRequired,
			)
		}
		p.SaramaProducerFactoryFunc = fn

		return nil
	}
}

// New instantiates new kafka producer.
func New(options ...Option) (sarama.AsyncProducer, error) {
	producer := new(Producer)

	var kafkaBrokers kafkacp.KafkaBrokers
	kafkaBrokers.AddFromString(kafkacp.DefaultKafkaBrokers)

	producer.KafkaBrokers = kafkaBrokers
	producer.DialTimeout = DefaultDialTimeout
	producer.ReadTimeout = DefaultReadTimeout
	producer.WriteTimeout = DefaultWriteTimeout
	producer.MaxRetries = DefaultMaxRetries
	producer.Backoff = DefaultBackoff
	producer.SaramaProducerFactoryFunc = sarama.NewAsyncProducer

	for _, option := range options {
		if err := option(producer); err != nil {
			return nil, err
		}
	}

	if err := producer.checkRequired(); err != nil {
		return nil, err
	}

	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true
	config.Producer.Compression = sarama.CompressionSnappy
	config.Net.DialTimeout = producer.DialTimeout
	config.Net.ReadTimeout = producer.ReadTimeout
	config.Net.WriteTimeout = producer.WriteTimeout

	var kafkaProducer sarama.AsyncProducer
	var kafkaProducerErr error
	backoff := producer.Backoff

	for i := range producer.MaxRetries {
		kafkaProducer, kafkaProducerErr = producer.SaramaProducerFactoryFunc(
			producer.KafkaBrokers.ToStringSlice(),
			config,
		)
		if kafkaProducerErr == nil {
			break
		}

		producer.Logger.Error(
			"can not connect broker",
			"error", kafkaProducerErr,
			"retry", fmt.Sprintf("%d/%d", i, producer.MaxRetries),
			"backoff", backoff.String(),
		)
		time.Sleep(backoff)
		backoff *= 2
	}
	if kafkaProducerErr != nil {
		return nil, fmt.Errorf(
			"[kafkaproducer.New][SaramaProducerFactoryFunc] error: [%w]",
			kafkaProducerErr,
		)
	}

	return kafkaProducer, nil
}
