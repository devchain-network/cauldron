package kafkacp_test

import (
	"testing"

	"github.com/devchain-network/cauldron/internal/kafkacp"
	"github.com/stretchr/testify/assert"
)

func TestKafkaTopicIdentifier_String(t *testing.T) {
	tests := []struct {
		name string
		id   kafkacp.KafkaTopicIdentifier
		want string
	}{
		{
			name: "GitHub topic",
			id:   kafkacp.KafkaTopicIdentifier("github"),
			want: "github",
		},
		{
			name: "Empty topic",
			id:   kafkacp.KafkaTopicIdentifier(""),
			want: "",
		},
		{
			name: "Custom topic",
			id:   kafkacp.KafkaTopicIdentifier("custom"),
			want: "custom",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.id.String()
			assert.Equal(t, tt.want, got, "KafkaTopicIdentifier.String() should return the correct string")
		})
	}
}

func TestKafkaTopicIdentifier_Valid(t *testing.T) {
	tests := []struct {
		name string
		id   kafkacp.KafkaTopicIdentifier
		want bool
	}{
		{
			name: "Valid GitHub topic",
			id:   kafkacp.KafkaTopicIdentifier("github"),
			want: true,
		},
		{
			name: "Valid GitLab topic",
			id:   kafkacp.KafkaTopicIdentifier("gitlab"),
			want: true,
		},
		{
			name: "Valid BitBucket topic",
			id:   kafkacp.KafkaTopicIdentifier("bitbucket"),
			want: true,
		},
		{
			name: "Invalid empty topic",
			id:   kafkacp.KafkaTopicIdentifier(""),
			want: false,
		},
		{
			name: "Invalid custom topic",
			id:   kafkacp.KafkaTopicIdentifier("custom"),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.id.Valid()
			assert.Equal(t, tt.want, got, "KafkaTopicIdentifier.Valid() validation failed")
		})
	}
}

func TestTCPAddr_String(t *testing.T) {
	tests := []struct {
		name string
		addr kafkacp.TCPAddr
		want string
	}{
		{
			name: "valid address",
			addr: kafkacp.TCPAddr("127.0.0.1:9092"),
			want: "127.0.0.1:9092",
		},
		{
			name: "empty address",
			addr: kafkacp.TCPAddr(""),
			want: "",
		},
		{
			name: "custom address",
			addr: kafkacp.TCPAddr("192.168.1.1:8080"),
			want: "192.168.1.1:8080",
		},
		{
			name: "any string",
			addr: kafkacp.TCPAddr("foo"),
			want: "foo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.addr.String()
			assert.Equal(t, tt.want, got, "TCPAddr.String() should return the correct string")
		})
	}
}

func TestTCPAddr_Valid(t *testing.T) {
	tests := []struct {
		name string
		addr kafkacp.TCPAddr
		want bool
	}{
		{
			name: "Valid address",
			addr: kafkacp.TCPAddr("127.0.0.1:9092"),
			want: true,
		},
		{
			name: "Invalid address (bad format)",
			addr: kafkacp.TCPAddr("127.0.0.1"),
			want: false,
		},
		{
			name: "Invalid empty address",
			addr: kafkacp.TCPAddr(""),
			want: false,
		},
		{
			name: "Invalid address",
			addr: kafkacp.TCPAddr("invalid"),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.addr.Valid()
			assert.Equal(t, tt.want, got, "TCPAddr.Valid() validation failed")
		})
	}
}

func TestKafkaBrokers_String(t *testing.T) {
	tests := []struct {
		name    string
		brokers kafkacp.KafkaBrokers
		want    string
	}{
		{
			name: "Single broker",
			brokers: kafkacp.KafkaBrokers{
				kafkacp.TCPAddr("127.0.0.1:9094"),
			},
			want: "127.0.0.1:9094",
		},
		{
			name: "Multiple brokers",
			brokers: kafkacp.KafkaBrokers{
				kafkacp.TCPAddr("127.0.0.1:9094"),
				kafkacp.TCPAddr("192.168.1.1:9092"),
			},
			want: "127.0.0.1:9094,192.168.1.1:9092",
		},
		{
			name:    "Empty brokers list",
			brokers: kafkacp.KafkaBrokers{},
			want:    "",
		},
		{
			name: "Includes invalid address (representation only)",
			brokers: kafkacp.KafkaBrokers{
				kafkacp.TCPAddr("127.0.0.1:9094"),
				kafkacp.TCPAddr("invalid-address"),
			},
			want: "127.0.0.1:9094,invalid-address",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.brokers.String()
			assert.Equal(t, tt.want, got, "KafkaBrokers.String() should return the correct concatenated string")
		})
	}
}

func TestKafkaBrokers_Valid(t *testing.T) {
	tests := []struct {
		name    string
		brokers kafkacp.KafkaBrokers
		want    bool
	}{
		{
			name: "All valid addresses",
			brokers: kafkacp.KafkaBrokers{
				kafkacp.TCPAddr("127.0.0.1:9094"),
				kafkacp.TCPAddr("192.168.1.1:9092"),
			},
			want: true,
		},
		{
			name: "Some invalid addresses",
			brokers: kafkacp.KafkaBrokers{
				kafkacp.TCPAddr("127.0.0.1:9094"),
				kafkacp.TCPAddr("invalid-address"),
			},
			want: false,
		},
		{
			name: "All invalid addresses",
			brokers: kafkacp.KafkaBrokers{
				kafkacp.TCPAddr("invalid-address"),
				kafkacp.TCPAddr(""),
			},
			want: false,
		},
		{
			name:    "Empty list",
			brokers: kafkacp.KafkaBrokers{},
			want:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.brokers.Valid()
			assert.Equal(t, tt.want, got, "KafkaBrokers.Valid() validation failed")
		})
	}
}

func TestKafkaBrokers_ToStringSlice(t *testing.T) {
	tests := []struct {
		name    string
		brokers kafkacp.KafkaBrokers
		want    []string
	}{
		{
			name: "Single broker",
			brokers: kafkacp.KafkaBrokers{
				kafkacp.TCPAddr("127.0.0.1:9094"),
			},
			want: []string{"127.0.0.1:9094"},
		},
		{
			name: "Multiple brokers",
			brokers: kafkacp.KafkaBrokers{
				kafkacp.TCPAddr("127.0.0.1:9094"),
				kafkacp.TCPAddr("192.168.1.1:9092"),
			},
			want: []string{"127.0.0.1:9094", "192.168.1.1:9092"},
		},
		{
			name:    "Empty brokers list",
			brokers: kafkacp.KafkaBrokers{},
			want:    []string{},
		},
		{
			name: "Includes invalid broker address (representation only)",
			brokers: kafkacp.KafkaBrokers{
				kafkacp.TCPAddr("127.0.0.1:9094"),
				kafkacp.TCPAddr("invalid-address"),
			},
			want: []string{"127.0.0.1:9094", "invalid-address"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.brokers.ToStringSlice()
			assert.Equal(t, tt.want, got, "KafkaBrokers.ToStringSlice() should return the correct slice of strings")
		})
	}
}

func TestKafkaBrokers_AddFromString(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		want      kafkacp.KafkaBrokers
		wantValid bool
	}{
		{
			name:      "single valid broker",
			input:     "127.0.0.1:9094",
			want:      kafkacp.KafkaBrokers{kafkacp.TCPAddr("127.0.0.1:9094")},
			wantValid: true,
		},
		{
			name:  "multiple valid brokers",
			input: "127.0.0.1:9094,192.168.1.1:9092",
			want: kafkacp.KafkaBrokers{
				kafkacp.TCPAddr("127.0.0.1:9094"),
				kafkacp.TCPAddr("192.168.1.1:9092"),
			},
			wantValid: true,
		},
		{
			name:      "empty input string",
			input:     "",
			want:      kafkacp.KafkaBrokers(nil),
			wantValid: false,
		},
		{
			name:      "includes invalid broker",
			input:     "127.0.0.1:9094,invalid-address",
			want:      kafkacp.KafkaBrokers{kafkacp.TCPAddr("127.0.0.1:9094")},
			wantValid: true,
		},
		{
			name:      "all invalid brokers",
			input:     "invalid-address1,invalid-address2",
			want:      kafkacp.KafkaBrokers(nil),
			wantValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var brokers kafkacp.KafkaBrokers
			brokers.AddFromString(tt.input)

			assert.Equal(t, tt.want, brokers, "KafkaBrokers.AddFromString() should populate correctly")

			if tt.wantValid {
				assert.True(t, brokers.Valid(), "KafkaBrokers.AddFromString() should produce a valid brokers list")
			} else {
				assert.False(t, brokers.Valid(), "KafkaBrokers.AddFromString() should produce an invalid brokers list")
			}
		})
	}
}
