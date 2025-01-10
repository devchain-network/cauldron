package apiserver

import (
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"

	"github.com/IBM/sarama"
	"github.com/devchain-network/cauldron/internal/apiserver/githubhandleroptions"
	"github.com/devchain-network/cauldron/internal/apiserver/httphandleroptions"
	"github.com/devchain-network/cauldron/internal/kafkaconsumer"
	"github.com/devchain-network/cauldron/internal/slogger"
	"github.com/go-playground/webhooks/v6/github"
	"github.com/valyala/fasthttp"
	"github.com/vigo/getenv"
)

// Run runs the server.
func Run() error {
	listenAddr := getenv.TCPAddr("LISTEN_ADDR", serverDefaultListenAddr)
	logLevel := getenv.String("LOG_LEVEL", slogger.DefaultLogLevel)
	githubHMACSecret := getenv.String("GITHUB_HMAC_SECRET", "")
	kafkaTopicGitHub := getenv.String("KP_TOPIC_GITHUB", kpDefaultGitHubTopic)
	brokersList := getenv.String("KCP_BROKERS", kafkaconsumer.DefaultKafkaBrokers)
	producerMessageQueueSize := getenv.Int("KP_PRODUCER_QUEUE_SIZE", kpDefaultQueueSize)
	backoff := getenv.Duration("KC_BACKOFF", kafkaconsumer.DefaultKafkaConsumerBackoff)
	maxRetries := getenv.Int("KC_MAX_RETRIES", kafkaconsumer.DefaultKafkaConsumerMaxRetries)
	if err := getenv.Parse(); err != nil {
		return fmt.Errorf("apiserver.Run getenv.Parse error: [%w]", err)
	}

	githubWebhook, err := github.New(github.Options.Secret(*githubHMACSecret))
	if err != nil {
		return fmt.Errorf("apiserver.Run github.New error: [%w]", err)
	}

	logger, err := slogger.New(
		slogger.WithLogLevelName(*logLevel),
	)
	if err != nil {
		return fmt.Errorf("apiserver.Run slogger.New error: [%w]", err)
	}

	kafkaBrokers := kafkaconsumer.TCPAddrs(*brokersList).List()

	kafkaConfig := sarama.NewConfig()
	kafkaConfig.Producer.RequiredAcks = sarama.WaitForAll
	kafkaConfig.Producer.Return.Successes = true
	kafkaConfig.Producer.Return.Errors = true

	var kafkaProducer sarama.AsyncProducer
	var kafkaProducerErr error

	for i := range *maxRetries {
		kafkaProducer, kafkaProducerErr = sarama.NewAsyncProducer(kafkaBrokers, kafkaConfig)
		if kafkaProducerErr == nil {
			break
		}

		logger.Error(
			"can not connect broker",
			"error", kafkaProducerErr,
			"retry", fmt.Sprintf("%d/%d", i, *maxRetries),
			"backoff", backoff.String(),
		)
		time.Sleep(*backoff)
		*backoff *= 2
	}
	if kafkaProducerErr != nil {
		return fmt.Errorf("apiserver.Run sarama.NewAsyncProducer error: [%w]", err)
	}

	defer func() { _ = kafkaProducer.Close() }()

	logger.Info("connected to kafka brokers", "addrs", kafkaBrokers)

	producerMessageQueue := make(chan *sarama.ProducerMessage, *producerMessageQueueSize)

	handlerOptions, err := httphandleroptions.New(
		httphandleroptions.WithLogger(logger),
		httphandleroptions.WithKafkaProducer(kafkaProducer),
		httphandleroptions.WithProducerMessageQueue(producerMessageQueue),
	)
	if err != nil {
		return fmt.Errorf("apiserver.Run httphandleroptions.New error: [%w]", err)
	}

	githubHandlerOptions, err := githubhandleroptions.New(
		githubhandleroptions.WithWebhook(githubWebhook),
		githubhandleroptions.WithCommonHandler(handlerOptions),
		githubhandleroptions.WithTopic(*kafkaTopicGitHub),
	)
	if err != nil {
		return fmt.Errorf("apiserver.Run githubhandleroptions.New error: [%w]", err)
	}

	numMessageWorkers := runtime.NumCPU()
	logger.Info(
		"number of message workers",
		"count", numMessageWorkers,
		"producer queue size", *producerMessageQueueSize,
	)

	server, err := New(
		WithLogger(logger),
		WithListenAddr(*listenAddr),
		WithKafkaGitHubTopic(*kafkaTopicGitHub),
		WithKafkaBrokers(kafkaBrokers),
		WithHTTPHandler(fasthttp.MethodGet, "/healthz", HealthCheckHandler),
		WithHTTPHandler(fasthttp.MethodPost, "/v1/webhook/github", GitHubWebhookHandler(githubHandlerOptions)),
	)
	if err != nil {
		return fmt.Errorf("apiserver.Run apiserver.New error: [%w]", err)
	}

	ch := make(chan struct{})

	var wg sync.WaitGroup
	for i := range numMessageWorkers {
		wg.Add(1)
		go func() {
			defer func() {
				wg.Done()
				logger.Info("terminating worker", "worker id", i)
			}()

			handlerOptions.MessageWorker(i)
		}()
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
		<-sig

		logger.Info("shutting down the server")
		if errStop := server.Stop(); err != nil {
			logger.Error("server stop error: [%w]", "error", errStop)
		}
		handlerOptions.Shutdown()
		close(ch)
	}()

	if errStop := server.Start(); err != nil {
		return fmt.Errorf("apiserver.Run server.Start error: [%w]", errStop)
	}

	<-ch
	wg.Wait()
	logger.Info("exiting apiserver, all clear")

	return nil
}
