package kafkagithubconsumer

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/IBM/sarama"
	"github.com/devchain-network/cauldron/internal/cerrors"
	"github.com/devchain-network/cauldron/internal/kafkacp"
	"github.com/devchain-network/cauldron/internal/kafkacp/kafkaconsumer"
	"github.com/devchain-network/cauldron/internal/storage/githubstorage"
)

var _ kafkaconsumer.KafkaConsumeStorer = (*Consumer)(nil) // compile time proof

// Consumer ...
type Consumer struct {
	Logger     *slog.Logger
	Config     *sarama.Config
	Storage    githubstorage.GitHubPingStorer
	Backoff    time.Duration
	MaxRetries uint8
}

// Consume ...
func (c Consumer) Consume(topic kafkacp.KafkaTopicIdentifier, partition int32) error {
	c.Logger.Info("info", "topic", topic, "partition", partition)

	return nil
}

// Store ...
func (c Consumer) Store(message *sarama.ConsumerMessage) error {
	c.Logger.Info("message", "message", message)

	return nil
}

func (c Consumer) checkRequired() error {
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
func WithStorage(st githubstorage.GitHubPingStorer) Option {
	return func(c *Consumer) error {
		if st == nil {
			return fmt.Errorf("kafkagithubconsumer.WithStorage error: [%w]", cerrors.ErrValueRequired)
		}
		c.Storage = st

		return nil
	}
}

// New instantiates new kafka github consumer instance.
func New(options ...Option) (*Consumer, error) {
	consumer := new(Consumer)
	consumer.Backoff = kafkacp.DefaultKafkaConsumerBackoff
	consumer.MaxRetries = kafkacp.DefaultKafkaConsumerMaxRetries

	for _, option := range options {
		if err := option(consumer); err != nil {
			return nil, fmt.Errorf("kafkagithubconsumer.New option error: [%w]", err)
		}
	}

	if err := consumer.checkRequired(); err != nil {
		return nil, err
	}

	return consumer, nil
}
