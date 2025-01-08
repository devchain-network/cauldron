package kafkaconsumer

import (
	"fmt"

	"github.com/devchain-network/cauldron/internal/slogger"
	"github.com/vigo/getenv"
)

// Run runs kafa consumer.
func Run() error {
	logLevel := getenv.String("LOG_LEVEL", slogger.DefaultLogLevel)

	partition := getenv.Int("KC_PARTITION", DefaultKafkaConsumerPartition)
	topic := getenv.String("KC_TOPIC", "")

	brokersList := getenv.String("KCP_BROKERS", DefaultKafkaBrokers)

	dialTimeout := getenv.Duration("KC_DIAL_TIMEOUT", DefaultKafkaConsumerDialTimeout)
	readTimeout := getenv.Duration("KC_READ_TIMEOUT", DefaultKafkaConsumerReadTimeout)
	writeTimeout := getenv.Duration("KC_WRITE_TIMEOUT", DefaultKafkaConsumerWriteTimeout)
	backoff := getenv.Duration("KC_BACKOFF", DefaultKafkaConsumerBackoff)
	maxRetries := getenv.Int("KC_MAX_RETRIES", DefaultKafkaConsumerMaxRetries)
	databaseURL := getenv.String("DATABASE_URL", "")

	if err := getenv.Parse(); err != nil {
		return fmt.Errorf("kafkaconsumer.Run getenv.Parse error: [%w]", err)
	}

	logger, err := slogger.New(
		slogger.WithLogLevelName(*logLevel),
	)
	if err != nil {
		return fmt.Errorf("kafkaconsumer.Run slogger.New error: [%w]", err)
	}

	// ctx := context.Background()
	// pgPool, err := db.New(ctx, *databaseURL)
	// if err != nil {
	// 	return fmt.Errorf("apiserver.Run db.New error: [%w]", err)
	// }
	// defer pgPool.Close()
	fmt.Println("*databaseURL", *databaseURL)

	brokers := TCPAddrs(*brokersList).List()

	kafkaConsumer, err := New(
		WithLogger(logger),
		WithTopic(*topic),
		WithPartition(*partition),
		WithBrokers(brokers),
		WithDialTimeout(*dialTimeout),
		WithReadTimeout(*readTimeout),
		WithWriteTimeout(*writeTimeout),
		WithBackoff(*backoff),
		WithMaxRetries(*maxRetries),
	)
	if err != nil {
		return fmt.Errorf("kafkaconsumer.Run kafkaconsumer.New error: [%w]", err)
	}

	if err = kafkaConsumer.Start(); err != nil {
		return fmt.Errorf("kafkaconsumer.Run kafkaconsumer.Start error: [%w]", err)
	}

	return nil
}
