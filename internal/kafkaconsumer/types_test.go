package kafkaconsumer_test

import (
	"testing"

	"github.com/devchain-network/cauldron/internal/cerrors"
	"github.com/devchain-network/cauldron/internal/kafkaconsumer"
	"github.com/stretchr/testify/assert"
)

func TestTCPAddrs_List(t *testing.T) {
	tests := []struct {
		name   string
		input  kafkaconsumer.TCPAddrs
		output []string
	}{
		{
			name:   "Valid addresses",
			input:  kafkaconsumer.TCPAddrs("127.0.0.1:9092,192.168.1.1:9093"),
			output: []string{"127.0.0.1:9092", "192.168.1.1:9093"},
		},
		{
			name:   "Empty string",
			input:  kafkaconsumer.TCPAddrs(""),
			output: []string{""},
		},
		{
			name:   "Invalid address",
			input:  kafkaconsumer.TCPAddrs("127.0.0.1:9092,invalid-address"),
			output: []string{"127.0.0.1:9092"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.input.List()
			assert.Equal(t, tt.output, result)
		})
	}
}

func TestIsKafkaTopicValid_Success(t *testing.T) {
	tests := []struct {
		name   string
		topic  kafkaconsumer.KafkaTopicIdentifier
		hasErr bool
	}{
		{name: "Valid GitHub Topic", topic: kafkaconsumer.KafkaTopicIdentifierGitHub, hasErr: false},
		{name: "Valid GitLab Topic", topic: kafkaconsumer.KafkaTopicIdentifierGitLab, hasErr: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := kafkaconsumer.IsKafkaTopicValid(tt.topic)
			assert.NoError(t, err)
		})
	}
}

func TestIsKafkaTopicValid_Error(t *testing.T) {
	tests := []struct {
		name   string
		topic  kafkaconsumer.KafkaTopicIdentifier
		expErr error
	}{
		{name: "Empty Topic", topic: "", expErr: cerrors.ErrValueRequired},
		{name: "Invalid Topic", topic: "unknown", expErr: cerrors.ErrInvalid},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := kafkaconsumer.IsKafkaTopicValid(tt.topic)
			assert.Error(t, err)
			assert.ErrorIs(t, err, tt.expErr)
		})
	}
}

func TestIsBrokersAreValid_Success(t *testing.T) {
	tests := []struct {
		name    string
		brokers []string
	}{
		{name: "Single Valid Broker", brokers: []string{"127.0.0.1:9092"}},
		{name: "Multiple Valid Brokers", brokers: []string{"127.0.0.1:9092", "localhost:9093"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := kafkaconsumer.IsBrokersAreValid(tt.brokers)
			assert.NoError(t, err)
		})
	}
}

func TestIsBrokersAreValid_Error(t *testing.T) {
	tests := []struct {
		name    string
		brokers []string
		expErr  error
	}{
		{name: "Nil Brokers", brokers: nil, expErr: cerrors.ErrValueRequired},
		{name: "Empty Brokers", brokers: []string{}, expErr: cerrors.ErrValueRequired},
		{name: "Empty Brokers - empty string", brokers: []string{""}, expErr: cerrors.ErrValueRequired},
		{name: "Invalid Broker Address", brokers: []string{"invalid-addr:9092"}, expErr: cerrors.ErrInvalid},
		{
			name:    "Mixed Valid and Invalid Brokers",
			brokers: []string{"127.0.0.1:9092", "invalid-addr"},
			expErr:  cerrors.ErrInvalid,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := kafkaconsumer.IsBrokersAreValid(tt.brokers)
			assert.Error(t, err)
			assert.ErrorIs(t, err, tt.expErr)
		})
	}
}
