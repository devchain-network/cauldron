package kafkacp

import (
	"slices"
	"strings"
	"time"

	"github.com/vigo/getenv"
)

// defaults.
const (
	DefaultKafkaBrokers = "127.0.0.1:9094"

	DefaultKafkaProducerDialTimeout  = 30 * time.Second
	DefaultKafkaProducerReadTimeout  = 30 * time.Second
	DefaultKafkaProducerWriteTimeout = 30 * time.Second
	DefaultKafkaProducerBackoff      = 2 * time.Second
	DefaultKafkaProducerMaxRetries   = 10

	DefaultKafkaConsumerPartition    = 0
	DefaultKafkaConsumerDialTimeout  = 30 * time.Second
	DefaultKafkaConsumerReadTimeout  = 30 * time.Second
	DefaultKafkaConsumerWriteTimeout = 30 * time.Second
	DefaultKafkaConsumerBackoff      = 2 * time.Second
	DefaultKafkaConsumerMaxRetries   = 10

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

// TCPAddr represents tcp address as string.
type TCPAddr string

func (s TCPAddr) String() string {
	return string(s)
}

// Valid checks and validates kafka topic name.
func (s TCPAddr) Valid() bool {
	if s == "" {
		return false
	}

	if _, err := getenv.ValidateTCPNetworkAddress(s.String()); err != nil {
		return false
	}

	return true
}

// KafkaBrokers represents list of kafka broker tcp addresses.
type KafkaBrokers []TCPAddr

func (k KafkaBrokers) String() string {
	ss := make([]string, len(k))
	for i, broker := range k {
		ss[i] = broker.String()
	}

	return strings.Join(ss, ",")
}

// Valid checks if all TCPAddr elements in KafkaBrokers are valid.
func (k KafkaBrokers) Valid() bool {
	if k == nil {
		return false
	}

	for _, broker := range k {
		if !broker.Valid() {
			return false
		}
	}

	return true
}

// ToStringSlice converts KafkaBrokers to a slice of strings.
func (k KafkaBrokers) ToStringSlice() []string {
	result := make([]string, len(k))
	for i, broker := range k {
		result[i] = broker.String()
	}

	return result
}

// AddFromString populates KafkaBrokers from comma-separated string.
func (k *KafkaBrokers) AddFromString(brokers string) {
	brokerSlice := strings.Split(brokers, ",")
	for _, addr := range brokerSlice {
		brokerAddr := TCPAddr(addr)
		if brokerAddr.Valid() {
			*k = append(*k, brokerAddr)
		}
	}
}
