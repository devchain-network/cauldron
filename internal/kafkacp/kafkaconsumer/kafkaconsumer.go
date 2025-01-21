package kafkaconsumer

import (
	"github.com/IBM/sarama"
	"github.com/devchain-network/cauldron/internal/kafkacp"
)

// KafkaConsumer defines kafka consumer behaviours.
type KafkaConsumer interface {
	Consume() error
}

// GetDefaultConfig returns consumer config with default values.
func GetDefaultConfig() *sarama.Config {
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true
	config.Net.DialTimeout = kafkacp.DefaultKafkaConsumerDialTimeout
	config.Net.ReadTimeout = kafkacp.DefaultKafkaConsumerReadTimeout
	config.Net.WriteTimeout = kafkacp.DefaultKafkaConsumerWriteTimeout

	return config
}

// GetDefaultConsumerFunc is a default consumer function.
func GetDefaultConsumerFunc(brokers kafkacp.KafkaBrokers, config *sarama.Config) (sarama.Consumer, error) {
	return sarama.NewConsumer(brokers.ToStringSlice(), config) //nolint:wrapcheck
}

// ConfigFunc represents config function type.
type ConfigFunc func() *sarama.Config

// ConsumerFunc represents consumer function type.
type ConsumerFunc func(brokers kafkacp.KafkaBrokers, config *sarama.Config) (sarama.Consumer, error)
