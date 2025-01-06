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
	kafkaTopicGitHub := getenv.String("KP_TOPIC_GITHUB", kkDefaultGitHubTopic)
	brokersList := getenv.String("KCP_BROKERS", kafkaconsumer.DefaultKafkaBrokers)
	producerMessageQueueSize := getenv.Int("KP_PRODUCER_QUEUE_SIZE", kkDefaultQueueSize)
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
			"error",
			kafkaProducerErr,
			"retry",
			fmt.Sprintf("%d/%d", i, *maxRetries),
			"backoff",
			backoff.String(),
		)
		time.Sleep(*backoff)
		*backoff *= 2
	}
	if kafkaProducerErr != nil {
		return fmt.Errorf("apiserver.Run sarama.NewAsyncProducer error: [%w]", err)
	}

	defer func() { _ = kafkaProducer.Close() }()

	logger.Info("connected to kafka brokers", "addrs", kafkaBrokers)

	commonHandlerOpts := httpHandlerOptions{
		logger:               logger,
		kafkaProducer:        kafkaProducer,
		producerMessageQueue: make(chan *sarama.ProducerMessage, *producerMessageQueueSize),
	}
	githubHandlerOpts := githubHandlerOptions{
		httpHandlerOptions: commonHandlerOpts,
		webhook:            githubWebhook,
		topic:              *kafkaTopicGitHub,
	}

	numMessageWorkers := runtime.NumCPU()
	logger.Info(
		"number of message workers",
		"count",
		numMessageWorkers,
		"producer queue size",
		*producerMessageQueueSize,
	)

	server, err := New(
		WithLogger(logger),
		WithListenAddr(*listenAddr),
		WithKafkaGitHubTopic(*kafkaTopicGitHub),
		WithKafkaBrokers(kafkaBrokers),
		WithHTTPHandler(fasthttp.MethodGet, "/healthz", healthCheckHandler),
		WithHTTPHandler(fasthttp.MethodPost, "/v1/webhook/github", githubWebhookHandler(&githubHandlerOpts)),
	)
	if err != nil {
		return fmt.Errorf("apiserver.Run apiserver.New error: [%w]", err)
	}

	ch := make(chan struct{})

	var wg sync.WaitGroup
	for i := range numMessageWorkers {
		wg.Add(1)
		go commonHandlerOpts.messageWorker(i, &wg)
	}

	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
		<-sig

		logger.Info("shutting down the server")
		if errStop := server.Stop(); err != nil {
			logger.Error("server stop error: [%w]", "error", errStop)
		}
		commonHandlerOpts.shutdown()
		close(ch)
	}()

	if errStop := server.Start(); err != nil {
		return fmt.Errorf("apiserver.Run server.Start error: [%w]", errStop)
	}

	<-ch
	wg.Wait()
	logger.Info("all clear, goodbye")

	return nil
}
