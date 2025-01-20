package apiserver

import (
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"

	"github.com/IBM/sarama"
	"github.com/devchain-network/cauldron/internal/kafkacp"
	"github.com/devchain-network/cauldron/internal/kafkacp/kafkaproducer"
	"github.com/devchain-network/cauldron/internal/slogger"
	"github.com/devchain-network/cauldron/internal/transport/http/githubwebhookhandler"
	"github.com/devchain-network/cauldron/internal/transport/http/healthcheckhandler"
	"github.com/valyala/fasthttp"
	"github.com/vigo/getenv"
)

// Run runs the server.
func Run() error {
	listenAddr := getenv.TCPAddr("LISTEN_ADDR", serverDefaultListenAddr)
	logLevel := getenv.String("LOG_LEVEL", slogger.DefaultLogLevel)
	brokersList := getenv.String("KCP_BROKERS", kafkacp.DefaultKafkaBrokers)
	backoff := getenv.Duration("KC_BACKOFF", kafkacp.DefaultKafkaProducerBackoff)
	maxRetries := getenv.Int("KC_MAX_RETRIES", kafkacp.DefaultKafkaProducerMaxRetries)
	githubHMACSecret := getenv.String("GITHUB_HMAC_SECRET", "")
	githubWebhookMessageQueueSize := getenv.Int("KP_GITHUB_MESSAGE_QUEUE_SIZE", kpDefaultQueueSize)
	if err := getenv.Parse(); err != nil {
		return fmt.Errorf("apiserver.Run getenv.Parse error: [%w]", err)
	}

	logger, err := slogger.New(
		slogger.WithLogLevelName(*logLevel),
	)
	if err != nil {
		return fmt.Errorf("apiserver.Run slogger.New error: [%w]", err)
	}

	var kafkaBrokers kafkacp.KafkaBrokers
	kafkaBrokers.AddFromString(*brokersList)

	kafkaProducer, err := kafkaproducer.New(
		kafkaproducer.WithLogger(logger),
		kafkaproducer.WithKafkaBrokers(kafkaBrokers),
		kafkaproducer.WithMaxRetries(*maxRetries),
		kafkaproducer.WithBackoff(*backoff),
	)
	if err != nil {
		return fmt.Errorf("apiserver.Run kafkaproducer.New error: [%w]", err)
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
		healthcheckhandler.WithVersion(serverVersion),
	)
	if err != nil {
		return fmt.Errorf("apiserver.Run healthcheckhandler.New error: [%w]", err)
	}

	githubWebhookHandler, err := githubwebhookhandler.New(
		githubwebhookhandler.WithLogger(logger),
		githubwebhookhandler.WithTopic(kafkacp.KafkaTopicIdentifierGitHub),
		githubwebhookhandler.WithWebhookSecret(*githubHMACSecret),
		githubwebhookhandler.WithProducerGitHubMessageQueue(githubWebhookMessageQueue),
	)
	if err != nil {
		return fmt.Errorf("apiserver.Run githubwebhookhandler.New error: [%w]", err)
	}

	server, err := New(
		WithLogger(logger),
		WithListenAddr(*listenAddr),
		WithKafkaGitHubTopic(kafkacp.KafkaTopicIdentifierGitHub),
		WithKafkaBrokers(kafkaBrokers),
		WithHTTPHandler(fasthttp.MethodGet, "/healthz", healthCheckHandler.Handle),
		WithHTTPHandler(fasthttp.MethodPost, "/v1/webhook/github", githubWebhookHandler.Handle),
	)
	if err != nil {
		return fmt.Errorf("apiserver.Run apiserver.New error: [%w]", err)
	}

	doneChannel := make(chan struct{})

	var wg sync.WaitGroup

	for i := range numMessageWorkers {
		wg.Add(1)
		go func() {
			defer func() {
				wg.Done()
				logger.Info("terminating worker", "worker id", i)
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
			logger.Error("server stop error: [%w]", "error", errStop)
		}
		close(githubWebhookMessageQueue)
		close(doneChannel)
	}()

	if errStop := server.Start(); err != nil {
		return fmt.Errorf("apiserver.Run server.Start error: [%w]", errStop)
	}

	<-doneChannel
	wg.Wait()
	logger.Info("exiting apiserver, all clear")

	return nil
}
