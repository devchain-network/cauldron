package apiserver

import (
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/IBM/sarama"
	"github.com/devchain-network/cauldron/internal/slogger"
	"github.com/go-playground/webhooks/v6/github"
	"github.com/valyala/fasthttp"
	"github.com/vigo/getenv"
)

// Run runs the server.
func Run() error {
	listenAddr := getenv.TCPAddr("LISTEN_ADDR", serverDefaultListenAddr)
	logLevel := getenv.String("LOG_LEVEL", loggerDefaultLevel)
	githubHMACSecret := getenv.String("GITHUB_HMAC_SECRET", "notset")
	kafkaGitHubTopic := getenv.String("KK_GITHUB_TOPIC", kkDefaultGitHubTopic)
	kafkaBroker1 := getenv.TCPAddr("KK_BROKER_1", kkDefaultBroker1)
	producerMessageQueueSize := getenv.Int("KK_PRODUCER_QUEUE_SIZE", kkDefaultQueueSize)

	if err := getenv.Parse(); err != nil {
		return fmt.Errorf("run error, getenv: [%w]", err)
	}

	githubWebhook, err := github.New(github.Options.Secret(*githubHMACSecret))
	if err != nil {
		return fmt.Errorf("run error, githubWebhook: [%w]", err)
	}

	logger, err := slogger.New(
		slogger.WithLogLevelName(*logLevel),
	)
	if err != nil {
		return fmt.Errorf("run error, logger: [%w]", err)
	}

	kafkaBrokers := []string{*kafkaBroker1}

	kafkaConfig := sarama.NewConfig()
	kafkaConfig.Producer.RequiredAcks = sarama.WaitForAll
	kafkaConfig.Producer.Return.Successes = true
	kafkaConfig.Producer.Return.Errors = true

	kafkaProducer, errkp := sarama.NewAsyncProducer(kafkaBrokers, kafkaConfig)
	if errkp != nil {
		return fmt.Errorf("run error, kafkaProducer: [%w]", errkp)
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
		topic:              *kafkaGitHubTopic,
	}

	numMessageWorkers := runtime.NumCPU()
	logger.Info(
		"number of message workers",
		"count",
		numMessageWorkers,
		"producer queue size",
		*producerMessageQueueSize,
	)

	for i := range numMessageWorkers {
		go commonHandlerOpts.messageWorker(i)
	}

	server, err := New(
		WithLogger(logger),
		WithListenAddr(*listenAddr),
		WithKafkaGitHubTopic(*kafkaGitHubTopic),
		WithKafkaBrokers(kafkaBrokers),
		WithHTTPHandler(fasthttp.MethodGet, "/healthz", healthCheckHandler),
		WithHTTPHandler(fasthttp.MethodPost, "/v1/webhook/github", githubWebhookHandler(&githubHandlerOpts)),
	)
	if err != nil {
		return fmt.Errorf("run error, server: [%w]", err)
	}

	ch := make(chan struct{})

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
		return fmt.Errorf("run error, server start: [%w]", errStop)
	}

	<-ch
	logger.Info("all clear, goodbye")

	return nil
}
