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
