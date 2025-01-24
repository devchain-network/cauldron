package kafkaconsumergroup

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"sync"

	"github.com/IBM/sarama"
	"github.com/devchain-network/cauldron/internal/cerrors"
	"github.com/devchain-network/cauldron/internal/kafkacp"
	"github.com/devchain-network/cauldron/internal/storage"
)

var _ sarama.ConsumerGroupHandler = (*Consumer)(nil) // compile time proof

// Consumer ...
type Consumer struct {
	Logger              *slog.Logger
	Storage             storage.PingStorer
	SaramaConsumerGroup sarama.ConsumerGroup
	Topic               kafkacp.KafkaTopicIdentifier
	KafkaBrokers        kafkacp.KafkaBrokers
	NumberOfWorkers     int
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

// New ...
func New(options ...Option) (*Consumer, error) {
	consumer := new(Consumer)

	var kafkaBrokers kafkacp.KafkaBrokers
	kafkaBrokers.AddFromString(kafkacp.DefaultKafkaBrokers)

	consumer.KafkaBrokers = kafkaBrokers
	consumer.Topic = kafkacp.KafkaTopicIdentifier("github")
	consumer.NumberOfWorkers = 5

	for _, option := range options {
		if err := option(consumer); err != nil {
			return nil, fmt.Errorf("kafka consumer group option error: [%w]", err)
		}
	}

	config := sarama.NewConfig()
	config.Version = sarama.V3_9_0_0
	config.Consumer.Return.Errors = true

	group, err := sarama.NewConsumerGroup(consumer.KafkaBrokers.ToStringSlice(), "my-group", config)
	if err != nil {
		return nil, fmt.Errorf("kafka group consumer group error: [%w]", err)
	}

	consumer.SaramaConsumerGroup = group

	return consumer, nil
}
