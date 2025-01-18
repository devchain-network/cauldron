package kafkacp

import (
	"slices"
	"time"
)

// defaults.
const (
	DefaultKafkaBrokers = "127.0.0.1:9094"

	DefaultKafkaProducerBackoff    = 2 * time.Second
	DefaultKafkaProducerMaxRetries = 10

	DefaultKafkaConsumerBackoff    = 2 * time.Second
	DefaultKafkaConsumerMaxRetries = 10

	KafkaTopicIdentifierGitHub    KafkaTopicIdentifier = "github"
	KafkaTopicIdentifierGitLab    KafkaTopicIdentifier = "gitlab"
	KafkaTopicIdentifierBitBucket KafkaTopicIdentifier = "bitbucket"
)

var validKafkaTopicIdentifiers = []KafkaTopicIdentifier{
	KafkaTopicIdentifierGitHub,
	KafkaTopicIdentifierGitLab,
	KafkaTopicIdentifierBitBucket,
}

// KafkaTopicIdentifier represents custom type for kafka topic names.
type KafkaTopicIdentifier string

func (s KafkaTopicIdentifier) String() string {
	return string(s)
}

// Valid checks and validates kafka topic name.
func (s KafkaTopicIdentifier) Valid() bool {
	if s == "" {
		return false
	}

	return slices.Contains(validKafkaTopicIdentifiers, s)
}
