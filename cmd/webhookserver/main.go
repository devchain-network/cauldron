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
	"github.com/devchain-network/cauldron/internal/kafkacp"
	"github.com/devchain-network/cauldron/internal/kafkacp/kafkaproducer"
	"github.com/devchain-network/cauldron/internal/slogger"
	"github.com/devchain-network/cauldron/internal/transport/http/githubwebhookhandler"
	"github.com/devchain-network/cauldron/internal/transport/http/gitlabwebhookhandler"
	"github.com/devchain-network/cauldron/internal/transport/http/healthcheckhandler"
	"github.com/devchain-network/cauldron/internal/webhookserver"
	"github.com/valyala/fasthttp"
	"github.com/vigo/getenv"
)

// default values.
const (
	kpGitHubDefaultQueueSize = 100
	kpGitLabDefaultQueueSize = 100
)

// Run runs the webhookserver.
func Run() error {
	logLevel := getenv.String("LOG_LEVEL", slogger.DefaultLogLevel)
	listenAddr := getenv.TCPAddr("LISTEN_ADDR", webhookserver.ServerDefaultListenAddr)
	serverReadTimeout := getenv.Duration("SERVER_READ_TIMEOUT", webhookserver.ServerDefaultReadTimeout)
	serverWriteTimeout := getenv.Duration("SERVER_WRITE_TIMEOUT", webhookserver.ServerDefaultWriteTimeout)
	serverIdleTimeout := getenv.Duration("SERVER_IDLE_TIMEOUT", webhookserver.ServerDefaultIdleTimeout)

	githubHMACSecret := getenv.String("GITHUB_HMAC_SECRET", "")
	gitlabHMACSecret := getenv.String("GITLAB_HMAC_SECRET", "")

	brokersList := getenv.String("KCP_BROKERS", kafkacp.DefaultKafkaBrokers)

	kafkaProducerDialTimeout := getenv.Duration("KP_DIAL_TIMEOUT", kafkaproducer.DefaultDialTimeout)
	kafkaProducerReadTimeout := getenv.Duration("KP_READ_TIMEOUT", kafkaproducer.DefaultReadTimeout)
	kafkaProducerWriteTimeout := getenv.Duration("KP_WRITE_TIMEOUT", kafkaproducer.DefaultWriteTimeout)
	kafkaProducerBackoff := getenv.Duration("KP_BACKOFF", kafkaproducer.DefaultBackoff)
	kafkaProducerMaxRetries := getenv.Int("KP_MAX_RETRIES", kafkaproducer.DefaultMaxRetries)
	kafkaProducerGithubWebhookMessageQueueSize := getenv.Int("KP_GITHUB_MESSAGE_QUEUE_SIZE", kpGitHubDefaultQueueSize)
	kafkaProducerGitlabWebhookMessageQueueSize := getenv.Int("KP_GITLAB_MESSAGE_QUEUE_SIZE", kpGitLabDefaultQueueSize)

	if err := getenv.Parse(); err != nil {
		return fmt.Errorf("environment variable parse error: [%w]", err)
	}

	logger, err := slogger.New(
		slogger.WithLogLevelName(*logLevel),
	)
	if err != nil {
		return fmt.Errorf("logger instantiate error: [%w]", err)
	}

	kafkaProducer, err := kafkaproducer.New(
		kafkaproducer.WithLogger(logger),
		kafkaproducer.WithKafkaBrokers(*brokersList),
		kafkaproducer.WithMaxRetries(*kafkaProducerMaxRetries),
		kafkaproducer.WithBackoff(*kafkaProducerBackoff),
		kafkaproducer.WithDialTimeout(*kafkaProducerDialTimeout),
		kafkaproducer.WithReadTimeout(*kafkaProducerReadTimeout),
		kafkaproducer.WithWriteTimeout(*kafkaProducerWriteTimeout),
	)
	if err != nil {
		return fmt.Errorf("kafka producer instantiate error: [%w]", err)
	}

	defer kafkaProducer.AsyncClose()

	logger.Info("connected to kafka brokers", "addrs", *brokersList)

	githubWebhookMessageQueue := make(chan *sarama.ProducerMessage, *kafkaProducerGithubWebhookMessageQueueSize)
	gitlabWebhookMessageQueue := make(chan *sarama.ProducerMessage, *kafkaProducerGitlabWebhookMessageQueueSize)

	numMessageWorkers := runtime.NumCPU()
	logger.Info(
		"number of message workers",
		"count", numMessageWorkers,
		"github webhook message queue size", *kafkaProducerGithubWebhookMessageQueueSize,
		"gitlab webhook message queue size", *kafkaProducerGitlabWebhookMessageQueueSize,
	)

	healthCheckHandler, err := healthcheckhandler.New(
		healthcheckhandler.WithVersion(webhookserver.ServerVersion),
	)
	if err != nil {
		return fmt.Errorf("health check http handler instantiate error: [%w]", err)
	}

	githubWebhookHandler, err := githubwebhookhandler.New(
		githubwebhookhandler.WithLogger(logger),
		githubwebhookhandler.WithTopic(kafkacp.KafkaTopicIdentifierGitHub.String()),
		githubwebhookhandler.WithWebhookSecret(*githubHMACSecret),
		githubwebhookhandler.WithProducerGitHubMessageQueue(githubWebhookMessageQueue),
	)
	if err != nil {
		return fmt.Errorf("github webhook http handler instantiate error: [%w]", err)
	}

	gitlabWebhookHandler, err := gitlabwebhookhandler.New(
		gitlabwebhookhandler.WithLogger(logger),
		gitlabwebhookhandler.WithTopic(kafkacp.KafkaTopicIdentifierGitLab.String()),
		gitlabwebhookhandler.WithWebhookSecret(*gitlabHMACSecret),
		gitlabwebhookhandler.WithProducerGitLabMessageQueue(gitlabWebhookMessageQueue),
	)
	if err != nil {
		return fmt.Errorf("gitlab webhook http handler instantiate error: [%w]", err)
	}

	server, err := webhookserver.New(
		webhookserver.WithLogger(logger),
		webhookserver.WithListenAddr(*listenAddr),
		webhookserver.WithReadTimeout(*serverReadTimeout),
		webhookserver.WithWriteTimeout(*serverWriteTimeout),
		webhookserver.WithIdleTimeout(*serverIdleTimeout),
		webhookserver.WithKafkaGitHubTopic(kafkacp.KafkaTopicIdentifierGitHub.String()),
		webhookserver.WithKafkaBrokers(*brokersList),
		webhookserver.WithHTTPHandler(fasthttp.MethodGet, "/healthz", healthCheckHandler.Handle),
		webhookserver.WithHTTPHandler(fasthttp.MethodPost, "/v1/webhook/github", githubWebhookHandler.Handle),
		webhookserver.WithHTTPHandler(fasthttp.MethodPost, "/v1/webhook/gitlab", gitlabWebhookHandler.Handle),
	)
	if err != nil {
		return fmt.Errorf("api server instantiate error: [%w]", err)
	}

	doneChannel := make(chan struct{})

	var wg sync.WaitGroup

	// GitHub message workers
	for i := range numMessageWorkers {
		wg.Add(1)
		go func() {
			defer func() {
				wg.Done()
				logger.Info("terminating github worker", "id", i)
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

	// GitLab message workers
	for i := range numMessageWorkers {
		wg.Add(1)
		go func() {
			defer func() {
				wg.Done()
				logger.Info("terminating gitlab worker", "id", i)
			}()

			func() {
				for msg := range gitlabWebhookMessageQueue {
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
			logger.Error("webhookserver stop error: [%w]", "error", errStop)
		}
		close(githubWebhookMessageQueue)
		close(gitlabWebhookMessageQueue)
		close(doneChannel)
	}()

	if errStop := server.Start(); err != nil {
		return fmt.Errorf("webhookserver start error: [%w]", errStop)
	}

	<-doneChannel
	wg.Wait()
	logger.Info("terminating webhookserver, all clear")

	return nil
}

func main() {
	if err := Run(); err != nil {
		log.Fatal(err)
	}
}
