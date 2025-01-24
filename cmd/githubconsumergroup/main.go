package main

import (
	"context"
	"fmt"
	"log"

	"github.com/devchain-network/cauldron/internal/kafkacp/kafkaconsumergroup"
	"github.com/devchain-network/cauldron/internal/slogger"
	"github.com/devchain-network/cauldron/internal/storage"
	"github.com/devchain-network/cauldron/internal/storage/githubstorage"
	"github.com/vigo/getenv"
)

// const (
// 	defaultKafkaConsumerTopic = "github"
// )

// Run ...
func Run() error {
	logLevel := getenv.String("LOG_LEVEL", slogger.DefaultLogLevel)
	// kafkaTopic := getenv.String("KC_TOPIC_GITHUB", defaultKafkaConsumerTopic)
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

// import (

// 	kafkaConsumer, err := kafkaconsumergroup.New(
// 		kafkaconsumergroup.WithLogger(logger),
// 		kafkaconsumergroup.WithStorage(db),
// 	)
// 	if err != nil {
// 		return fmt.Errorf("kafka consumer instantiate error: [%w]", err)
// 	}
//
// 	for {
// 		if err = kafkaConsumer.SaramaConsumerGroup.Consume(
// 			ctx, []string{*kafkaTopic}, kafkaConsumer.ConsumerGroupHandler,
// 		); err != nil {
// 			logger.Error("error", "error", err)
//
// 			continue
// 		}
// 	}
//
// 	// return nil
// }
//
