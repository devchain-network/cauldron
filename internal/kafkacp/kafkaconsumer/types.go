package kafkaconsumer

// import (
// 	"fmt"
// 	"slices"
// 	"strings"
//
// 	"github.com/IBM/sarama"
// 	"github.com/devchain-network/cauldron/internal/cerrors"
// 	"github.com/vigo/getenv"
// )
//
// // constants.
// const (
// 	KafkaTopicIdentifierGitHub KafkaTopicIdentifier = "github"
// 	KafkaTopicIdentifierGitLab KafkaTopicIdentifier = "gitlab"
// )
//
// var validKafkaTopicIdentifiers = []KafkaTopicIdentifier{
// 	KafkaTopicIdentifierGitHub,
// 	KafkaTopicIdentifierGitLab,
// }
//
// // ConsumerFactoryFunc is a type for handling consumer.
// type ConsumerFactoryFunc func(brokers []string, config *sarama.Config) (sarama.Consumer, error)
//
// // ConsumerConfigFactoryFunc is a type for handling config.
// type ConsumerConfigFactoryFunc func() *sarama.Config
//
// // TCPAddrs represents comma separated tcp addr list.
// type TCPAddrs string
//
// // List validates and return list of tcp addrs.
// func (t TCPAddrs) List() []string {
// 	var addrs []string
// 	for _, addr := range strings.Split(string(t), ",") {
// 		if _, err := getenv.ValidateTCPNetworkAddress(addr); err == nil {
// 			addrs = append(addrs, addr)
// 		}
// 	}
//
// 	return addrs
// }
//
// // KafkaTopicIdentifier represents custom type.
// type KafkaTopicIdentifier string
//
// func (s KafkaTopicIdentifier) String() string {
// 	return string(s)
// }
//
// // IsKafkaTopicValid validates if given topic is supported.
// func IsKafkaTopicValid(s KafkaTopicIdentifier) error {
// 	if s.String() == "" {
// 		return fmt.Errorf("%s error: [%w]", s, cerrors.ErrValueRequired)
// 	}
//
// 	if !slices.Contains(validKafkaTopicIdentifiers, s) {
// 		return fmt.Errorf("%s error: [%w]", s, cerrors.ErrInvalid)
// 	}
//
// 	return nil
// }
//
// // IsBrokersAreValid validates if the given brokers are pointing valid tcp addrs.
// func IsBrokersAreValid(brokers []string) error {
// 	if brokers == nil {
// 		return fmt.Errorf("brokers error: [%w]", cerrors.ErrValueRequired)
// 	}
//
// 	if len(brokers) == 0 {
// 		return fmt.Errorf("brokers error: [%w]", cerrors.ErrValueRequired)
// 	}
//
// 	for _, broker := range brokers {
// 		if broker == "" {
// 			return fmt.Errorf("%s error: [%w]", broker, cerrors.ErrValueRequired)
// 		}
// 		if _, err := getenv.ValidateTCPNetworkAddress(broker); err != nil {
// 			return fmt.Errorf("%s error: [%w][%w]", broker, err, cerrors.ErrInvalid)
// 		}
// 	}
//
// 	return nil
// }
