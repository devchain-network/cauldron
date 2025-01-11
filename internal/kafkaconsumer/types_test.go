package kafkaconsumer_test

import (
	"testing"

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
