package kafkaconsumer

import (
	"github.com/IBM/sarama"
	"github.com/devchain-network/cauldron/internal/kafkacp"
)

// KafkaConsumeStorer defines kafka consumer/storer behaviours.
type KafkaConsumeStorer interface {
	Consume(topic kafkacp.KafkaTopicIdentifier, partition int32) error
	Store(message *sarama.ConsumerMessage) error
}

// GetDefaultConfig return consumer config with default values.
func GetDefaultConfig() *sarama.Config {
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true
	config.Net.DialTimeout = kafkacp.DefaultKafkaConsumerDialTimeout
	config.Net.ReadTimeout = kafkacp.DefaultKafkaConsumerReadTimeout
	config.Net.WriteTimeout = kafkacp.DefaultKafkaConsumerWriteTimeout

	return config
}

// GetDefaultConsumerFunc ..
func GetDefaultConsumerFunc( //nolint:ireturn
	brokers kafkacp.KafkaBrokers,
	config *sarama.Config,
) (sarama.Consumer, error) {
	return sarama.NewConsumer(brokers.ToStringSlice(), config) //nolint:wrapcheck
}

// ConfigFunc represents config function type.
type ConfigFunc func() *sarama.Config

// ConsumerFunc represents consumer function type.
type ConsumerFunc func(brokers kafkacp.KafkaBrokers, config *sarama.Config) (sarama.Consumer, error)
