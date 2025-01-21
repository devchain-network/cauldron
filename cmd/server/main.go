package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"

	"github.com/IBM/sarama"
	"github.com/devchain-network/cauldron/internal/apiserver"
	"github.com/devchain-network/cauldron/internal/kafkacp"
	"github.com/devchain-network/cauldron/internal/kafkacp/kafkaproducer"
	"github.com/devchain-network/cauldron/internal/slogger"
	"github.com/devchain-network/cauldron/internal/transport/http/githubwebhookhandler"
	"github.com/devchain-network/cauldron/internal/transport/http/healthcheckhandler"
	"github.com/valyala/fasthttp"
	"github.com/vigo/getenv"
)

// default values.
const (
	kpGitHubDefaultQueueSize = 100
)

// Run runs the server.
func Run() error {
	logLevel := getenv.String("LOG_LEVEL", slogger.DefaultLogLevel)
	listenAddr := getenv.TCPAddr("LISTEN_ADDR", apiserver.ServerDefaultListenAddr)
	githubHMACSecret := getenv.String("GITHUB_HMAC_SECRET", "")

	brokersList := getenv.String("KCP_BROKERS", kafkacp.DefaultKafkaBrokers)

	dialTimeout := getenv.Duration("KP_DIAL_TIMEOUT", kafkaproducer.DefaultDialTimeout)
	readTimeout := getenv.Duration("KP_READ_TIMEOUT", kafkaproducer.DefaultReadTimeout)
	writeTimeout := getenv.Duration("KP_WRITE_TIMEOUT", kafkaproducer.DefaultWriteTimeout)
	backoff := getenv.Duration("KP_BACKOFF", kafkaproducer.DefaultBackoff)
	maxRetries := getenv.Int("KP_MAX_RETRIES", kafkaproducer.DefaultMaxRetries)
	githubWebhookMessageQueueSize := getenv.Int("KP_GITHUB_MESSAGE_QUEUE_SIZE", kpGitHubDefaultQueueSize)

	if err := getenv.Parse(); err != nil {
		return fmt.Errorf("environment variable parse error: [%w]", err)
	}

	logger, err := slogger.New(
		slogger.WithLogLevelName(*logLevel),
	)
	if err != nil {
		return fmt.Errorf("logger instantiate error: [%w]", err)
	}

	var kafkaBrokers kafkacp.KafkaBrokers
	kafkaBrokers.AddFromString(*brokersList)

	kafkaProducer, err := kafkaproducer.New(
		kafkaproducer.WithLogger(logger),
		kafkaproducer.WithKafkaBrokers(kafkaBrokers),
		kafkaproducer.WithMaxRetries(*maxRetries),
		kafkaproducer.WithBackoff(*backoff),
		kafkaproducer.WithDialTimeout(*dialTimeout),
		kafkaproducer.WithReadTimeout(*readTimeout),
		kafkaproducer.WithWriteTimeout(*writeTimeout),
	)
	if err != nil {
		return fmt.Errorf("kafka producer instantiate error: [%w]", err)
	}

	defer kafkaProducer.AsyncClose()

	logger.Info("connected to kafka brokers", "addrs", kafkaBrokers)

	githubWebhookMessageQueue := make(chan *sarama.ProducerMessage, *githubWebhookMessageQueueSize)

	numMessageWorkers := runtime.NumCPU()
	logger.Info(
		"number of message workers",
		"count", numMessageWorkers,
		"github webhook message queue size", *githubWebhookMessageQueueSize,
	)

	healthCheckHandler, err := healthcheckhandler.New(
		healthcheckhandler.WithVersion(apiserver.ServerVersion),
	)
	if err != nil {
		return fmt.Errorf("health check http handler instantiate error: [%w]", err)
	}

	githubWebhookHandler, err := githubwebhookhandler.New(
		githubwebhookhandler.WithLogger(logger),
		githubwebhookhandler.WithTopic(kafkacp.KafkaTopicIdentifierGitHub),
		githubwebhookhandler.WithWebhookSecret(*githubHMACSecret),
		githubwebhookhandler.WithProducerGitHubMessageQueue(githubWebhookMessageQueue),
	)
	if err != nil {
		return fmt.Errorf("github webhook http handler instantiate error: [%w]", err)
	}

	server, err := apiserver.New(
		apiserver.WithLogger(logger),
		apiserver.WithListenAddr(*listenAddr),
		apiserver.WithKafkaGitHubTopic(kafkacp.KafkaTopicIdentifierGitHub),
		apiserver.WithKafkaBrokers(kafkaBrokers),
		apiserver.WithHTTPHandler(fasthttp.MethodGet, "/healthz", healthCheckHandler.Handle),
		apiserver.WithHTTPHandler(fasthttp.MethodPost, "/v1/webhook/github", githubWebhookHandler.Handle),
	)
	if err != nil {
		return fmt.Errorf("api server instantiate error: [%w]", err)
	}

	doneChannel := make(chan struct{})

	var wg sync.WaitGroup

	for i := range numMessageWorkers {
		wg.Add(1)
		go func() {
			defer func() {
				wg.Done()
				logger.Info("terminating worker", "id", i)
			}()

			func() {
				for msg := range githubWebhookMessageQueue {
					kafkaProducer.Input() <- msg

					select {
					case success := <-kafkaProducer.Successes():
						logger.Info(
							"message sent",
							"worker", i,
							"topic", success.Topic,
							"partition", success.Partition,
							"offset", success.Offset,
						)
					case err := <-kafkaProducer.Errors():
						logger.Error("message send error", "error", err)
					}
				}
			}()
		}()
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
		<-sig

		if errStop := server.Stop(); err != nil {
			logger.Error("api server stop error: [%w]", "error", errStop)
		}
		close(githubWebhookMessageQueue)
		close(doneChannel)
	}()

	if errStop := server.Start(); err != nil {
		return fmt.Errorf("api server start error: [%w]", errStop)
	}

	<-doneChannel
	wg.Wait()
	logger.Info("terminating api server, all clear")

	return nil
}

func main() {
	if err := Run(); err != nil {
		log.Fatal(err)
	}
}
