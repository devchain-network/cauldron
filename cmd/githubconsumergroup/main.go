package main

import (
	"context"
	"fmt"
	"log"

	"github.com/devchain-network/cauldron/internal/kafkacp"
	"github.com/devchain-network/cauldron/internal/kafkacp/kafkaconsumergroup"
	"github.com/devchain-network/cauldron/internal/slogger"
	"github.com/devchain-network/cauldron/internal/storage"
	"github.com/devchain-network/cauldron/internal/storage/githubstorage"
	"github.com/vigo/getenv"
)

// Run runs kafa github consumer group.
func Run() error {
	logLevel := getenv.String("LOG_LEVEL", slogger.DefaultLogLevel)
	brokersList := getenv.String("KCP_BROKERS", kafkacp.DefaultKafkaBrokers)
	kafkaTopic := getenv.String("KC_TOPIC_GITHUB", "")
	kafkaConsumerGroup := getenv.String("KCG_NAME", "github-group")
	kafkaDialTimeout := getenv.Duration("KC_DIAL_TIMEOUT", kafkaconsumergroup.DefaultDialTimeout)
	kafkaReadTimeout := getenv.Duration("KC_READ_TIMEOUT", kafkaconsumergroup.DefaultReadTimeout)
	kafkaWriteTimeout := getenv.Duration("KC_WRITE_TIMEOUT", kafkaconsumergroup.DefaultWriteTimeout)
	kafkaBackoff := getenv.Duration("KC_BACKOFF", kafkaconsumergroup.DefaultBackoff)
	kafkaMaxRetries := getenv.Int("KC_MAX_RETRIES", kafkaconsumergroup.DefaultMaxRetries)
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

	kafkaGitHubConsumer, err := kafkaconsumergroup.New(
		kafkaconsumergroup.WithLogger(logger),
		kafkaconsumergroup.WithStorage(db),
		kafkaconsumergroup.WithKafkaBrokers(*brokersList),
		kafkaconsumergroup.WithDialTimeout(*kafkaDialTimeout),
		kafkaconsumergroup.WithReadTimeout(*kafkaReadTimeout),
		kafkaconsumergroup.WithWriteTimeout(*kafkaWriteTimeout),
		kafkaconsumergroup.WithBackoff(*kafkaBackoff),
		kafkaconsumergroup.WithMaxRetries(*kafkaMaxRetries),
		kafkaconsumergroup.WithTopic(*kafkaTopic),
		kafkaconsumergroup.WithKafkaGroupName(*kafkaConsumerGroup),
	)
	if err != nil {
		return fmt.Errorf("github kafka group consumer instantiate error: [%w]", err)
	}

	defer func() { _ = kafkaGitHubConsumer.SaramaConsumerGroup.Close() }()

	if err = kafkaGitHubConsumer.StartConsume(); err != nil {
		return fmt.Errorf("github kafka group consumer start consume error: [%w]", err)
	}

	return nil
}

func main() {
	if err := Run(); err != nil {
		log.Fatal(err)
	}
}
