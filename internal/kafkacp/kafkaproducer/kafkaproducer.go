package kafkaproducer

import (
	"fmt"
	"log/slog"
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
	Logger       *slog.Logger
	KafkaBrokers kafkacp.KafkaBrokers
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	MaxRetries   uint8
	Backoff      time.Duration
}

func (p Producer) checkRequired() error {
	if p.Logger == nil {
		return fmt.Errorf("kafkaproducer.New Logger error: [%w]", cerrors.ErrValueRequired)
	}
	if !p.KafkaBrokers.Valid() {
		return fmt.Errorf("kafkaproducer.New KafkaBrokers error: [%w]", cerrors.ErrInvalid)
	}

	return nil
}

// Option represents option function type.
type Option func(*Producer) error

// WithLogger sets logger.
func WithLogger(l *slog.Logger) Option {
	return func(p *Producer) error {
		if l == nil {
			return fmt.Errorf("kafkaproducer.WithLogger error: [%w]", cerrors.ErrValueRequired)
		}
		p.Logger = l

		return nil
	}
}

// WithKafkaBrokers sets kafka brokers list.
func WithKafkaBrokers(brokers kafkacp.KafkaBrokers) Option {
	return func(p *Producer) error {
		if !brokers.Valid() {
			return fmt.Errorf("kafkaproducer.WithKafkaBrokers error: [%w]", cerrors.ErrInvalid)
		}

		p.KafkaBrokers = brokers

		return nil
	}
}

// WithMaxRetries sets max retries value.
func WithMaxRetries(i int) Option {
	return func(p *Producer) error {
		if i > 255 || i < 0 {
			return fmt.Errorf("kafkaproducer.WithMaxRetries error: [%w]", cerrors.ErrInvalid)
		}
		p.MaxRetries = uint8(i)

		return nil
	}
}

// WithBackoff sets backoff duration.
func WithBackoff(d time.Duration) Option {
	return func(p *Producer) error {
		if d == 0 {
			return fmt.Errorf("kafkaproducer.WithBackoff error: [%w]", cerrors.ErrValueRequired)
		}
		p.Backoff = d

		return nil
	}
}

// WithDialTimeout sets dial timeout.
func WithDialTimeout(d time.Duration) Option {
	return func(p *Producer) error {
		if d < 0 {
			return fmt.Errorf("kafkaproducer.WithDialTimeout error: [%w]", cerrors.ErrInvalid)
		}
		p.DialTimeout = d

		return nil
	}
}

// WithReadTimeout sets read timeout.
func WithReadTimeout(d time.Duration) Option {
	return func(p *Producer) error {
		if d < 0 {
			return fmt.Errorf("kafkaproducer.WithReadTimeout error: [%w]", cerrors.ErrInvalid)
		}
		p.ReadTimeout = d

		return nil
	}
}

// WithWriteTimeout sets write timeout.
func WithWriteTimeout(d time.Duration) Option {
	return func(p *Producer) error {
		if d < 0 {
			return fmt.Errorf("kafkaproducer.WithWriteTimeout error: [%w]", cerrors.ErrInvalid)
		}
		p.WriteTimeout = d

		return nil
	}
}

// New instantiates new kafka producer.
func New(options ...Option) (sarama.AsyncProducer, error) {
	producer := new(Producer)
	producer.DialTimeout = DefaultDialTimeout
	producer.ReadTimeout = DefaultReadTimeout
	producer.WriteTimeout = DefaultWriteTimeout
	producer.MaxRetries = DefaultMaxRetries
	producer.Backoff = DefaultBackoff

	for _, option := range options {
		if err := option(producer); err != nil {
			return nil, fmt.Errorf("kafkaproducer.New option error: [%w]", err)
		}
	}

	if err := producer.checkRequired(); err != nil {
		return nil, err
	}

	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true
	config.Net.DialTimeout = producer.DialTimeout
	config.Net.ReadTimeout = producer.ReadTimeout
	config.Net.WriteTimeout = producer.WriteTimeout

	var kafkaProducer sarama.AsyncProducer
	var kafkaProducerErr error
	backoff := producer.Backoff

	for i := range producer.MaxRetries {
		kafkaProducer, kafkaProducerErr = sarama.NewAsyncProducer(producer.KafkaBrokers.ToStringSlice(), config)
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
		return nil, fmt.Errorf("kafkaproducer.New sarama.NewAsyncProducer error: [%w]", kafkaProducerErr)
	}

	return kafkaProducer, nil
}
