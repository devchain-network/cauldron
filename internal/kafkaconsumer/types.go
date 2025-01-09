package kafkaconsumer

import (
	"strings"

	"github.com/vigo/getenv"
)

// constants.
const (
	KafkaTopicIdentifierGitHub KafkaTopicIdentifier = "github"
	KafkaTopicIdentifierGitLab KafkaTopicIdentifier = "gitlab"
)

// TCPAddrs represents comma separated tcp addr list.
type TCPAddrs string

// List validates and return list of tcp addrs.
func (t TCPAddrs) List() []string {
	var addrs []string
	for _, addr := range strings.Split(string(t), ",") {
		if _, err := getenv.ValidateTCPNetworkAddress(addr); err == nil {
			addrs = append(addrs, addr)
		}
	}

	return addrs
}

// KafkaTopicIdentifier represents custom type.
type KafkaTopicIdentifier string

func (s KafkaTopicIdentifier) String() string {
	return string(s)
}
