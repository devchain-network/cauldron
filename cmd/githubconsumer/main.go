package main

import (
	"context"
	"fmt"
	"log"

	"github.com/devchain-network/cauldron/internal/kafkacp"
	"github.com/devchain-network/cauldron/internal/kafkacp/kafkaconsumer"
	"github.com/devchain-network/cauldron/internal/slogger"
	"github.com/devchain-network/cauldron/internal/storage"
	"github.com/devchain-network/cauldron/internal/storage/githubstorage"
	"github.com/vigo/getenv"
)

const (
	defaultKafkaConsumerTopic = "github"
)

// Run runs kafa github consumer.
func Run() error {
	logLevel := getenv.String("LOG_LEVEL", slogger.DefaultLogLevel)
	brokersList := getenv.String("KCP_BROKERS", kafkacp.DefaultKafkaBrokers)

	kafkaTopic := getenv.String("KC_TOPIC_GITHUB", defaultKafkaConsumerTopic)
	kafkaPartition := getenv.Int("KC_PARTITION", kafkaconsumer.DefaultPartition)
	kafkaDialTimeout := getenv.Duration("KC_DIAL_TIMEOUT", kafkaconsumer.DefaultDialTimeout)
	kafkaReadTimeout := getenv.Duration("KC_READ_TIMEOUT", kafkaconsumer.DefaultReadTimeout)
	kafkaWriteTimeout := getenv.Duration("KC_WRITE_TIMEOUT", kafkaconsumer.DefaultWriteTimeout)
	kafkaBackoff := getenv.Duration("KC_BACKOFF", kafkaconsumer.DefaultBackoff)
	kafkaMaxRetries := getenv.Int("KC_MAX_RETRIES", kafkaconsumer.DefaultMaxRetries)

	databaseURL := getenv.String("DATABASE_URL", "")
	if err := getenv.Parse(); err != nil {
		return fmt.Errorf("environment variable parse error: [%w]", err)
	}

	logger, err := slogger.New(
		slogger.WithLogLevelName(*logLevel),
	)
	if err != nil {
		return fmt.Errorf("logger instantiate error: [%w]", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), storage.DefaultDBPingTimeout)
	defer cancel()

	db, err := githubstorage.New(
		ctx,
		githubstorage.WithDatabaseDSN(*databaseURL),
		githubstorage.WithLogger(logger),
	)
	if err != nil {
		return fmt.Errorf("github storage instantiate error: [%w]", err)
	}

	if err = db.Ping(ctx, storage.DefaultDBPingMaxRetries, storage.DefaultDBPingBackoff); err != nil {
		return fmt.Errorf("github storage ping error: [%w]", err)
	}
	defer func() {
		logger.Info("github storage - closing pgx pool")
		db.Pool.Close()
	}()

	kafkaGitHubConsumer, err := kafkaconsumer.New(
		kafkaconsumer.WithLogger(logger),
		kafkaconsumer.WithStorage(db),
		kafkaconsumer.WithKafkaBrokers(*brokersList),
		kafkaconsumer.WithDialTimeout(*kafkaDialTimeout),
		kafkaconsumer.WithReadTimeout(*kafkaReadTimeout),
		kafkaconsumer.WithWriteTimeout(*kafkaWriteTimeout),
		kafkaconsumer.WithBackoff(*kafkaBackoff),
		kafkaconsumer.WithMaxRetries(*kafkaMaxRetries),
		kafkaconsumer.WithTopic(*kafkaTopic),
		kafkaconsumer.WithPartition(*kafkaPartition),
	)
	if err != nil {
		return fmt.Errorf("github kafka consumer instantiate error: [%w]", err)
	}

	defer func() { _ = kafkaGitHubConsumer.SaramaConsumer.Close() }()

	if err = kafkaGitHubConsumer.Consume(); err != nil {
		return fmt.Errorf("github kafka consumer consume error: [%w]", err)
	}

	return nil
}

func main() {
	if err := Run(); err != nil {
		log.Fatal(err)
	}
}
