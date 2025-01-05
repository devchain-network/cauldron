package kafkaconsumer

import (
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"time"

	"github.com/IBM/sarama"
	"github.com/devchain-network/cauldron/internal/cerrors"
	"github.com/devchain-network/cauldron/internal/slogger"
	"github.com/vigo/getenv"
)

// constants.
const (
	loggerDefaultLevel = "INFO"

	kkDefaultTopic        = "deneme"
	kkDefaultBroker1      = "127.0.0.1:9094"
	kkDefaultDialTimeout  = 30 * time.Second
	kkDefaultReadTimeout  = 30 * time.Second
	kkDefaultWriteTimeout = 30 * time.Second

	kkDefaultBackoff = 2 * time.Second
	kkMaxRetries     = 10
)

// KafkaConsumer defines kafka consumer behaviours.
type KafkaConsumer interface {
	Start() error
}

var _ KafkaConsumer = (*Consumer)(nil) // compile time proof

// Consumer represents kafa consumer setup.
type Consumer struct {
	Logger       *slog.Logger
	Topic        string
	Brokers      []string
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	Backoff      time.Duration
	MaxRetries   uint8
	Partition    int32
}

// Start starts consumer.
func (c Consumer) Start() error {
	config := c.getConfig()

	var consumer sarama.Consumer
	var consumerErr error
	backoff := kkDefaultBackoff

	for i := range kkMaxRetries {
		consumer, consumerErr = sarama.NewConsumer(c.Brokers, config)
		if consumerErr == nil {
			break
		}

		c.Logger.Error(
			"can not connect broker",
			"error",
			consumerErr,
			"retry",
			fmt.Sprintf("%d/%d", i, c.MaxRetries),
			"backoff",
			backoff.String(),
		)
		time.Sleep(backoff)
		backoff *= 2
	}

	if consumerErr != nil {
		return fmt.Errorf("new consumer error: [%w]", consumerErr)
	}
	defer func() { _ = consumer.Close() }()

	partitionConsumer, errpc := consumer.ConsumePartition(c.Topic, c.Partition, sarama.OffsetNewest)
	if errpc != nil {
		return fmt.Errorf("partition consumer error: [%w]", errpc)
	}
	defer func() { _ = partitionConsumer.Close() }()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	c.Logger.Info("consuming messages from", "topic", c.Topic)

	messageChan := make(chan *sarama.ConsumerMessage, 10)
	defer close(messageChan)

	numWorkers := runtime.NumCPU()
	c.Logger.Info("starting workers", "count", numWorkers)

	var wg sync.WaitGroup
	for i := range numWorkers {
		wg.Add(1)
		go c.worker(i, messageChan, &wg)
	}

	go func() {
		for {
			select {
			case msg := <-partitionConsumer.Messages():
				if msg != nil {
					messageChan <- msg
				}
			case err := <-partitionConsumer.Errors():
				c.Logger.Error("partition consumer error", "err", err)
			case <-signals:
				c.Logger.Info("shutting down message producer")

				return
			}
		}
	}()

	wg.Wait()
	c.Logger.Info("all workers stopped")

	return nil
}

func (c Consumer) getConfig() *sarama.Config {
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true
	config.Net.DialTimeout = c.DialTimeout
	config.Net.ReadTimeout = c.ReadTimeout
	config.Net.WriteTimeout = c.WriteTimeout

	return config
}

func (c Consumer) worker(id int, messages <-chan *sarama.ConsumerMessage, wg *sync.WaitGroup) {
	defer wg.Done()

	for msg := range messages {
		for _, header := range msg.Headers {
			c.Logger.Info("header", "key", string(header.Key), "value", string(header.Value))
		}

		c.Logger.Info(
			"received",
			"worker id", id,
			"key", string(msg.Key),
			"value", string(msg.Value),
			"offset", msg.Offset,
			"partition", msg.Partition,
		)

		// process message here...
	}
}

// Option represents option function type.
type Option func(*Consumer) error

// WithLogger sets logger.
func WithLogger(l *slog.Logger) Option {
	return func(consumer *Consumer) error {
		if l == nil {
			return fmt.Errorf("consumer logger error: [%w]", cerrors.ErrValueRequired)
		}
		consumer.Logger = l

		return nil
	}
}

// WithTopic sets topic.
func WithTopic(s string) Option {
	return func(consumer *Consumer) error {
		if s == "" {
			return fmt.Errorf("consumer topic error: [%w]", cerrors.ErrValueRequired)
		}
		consumer.Topic = s

		return nil
	}
}

// WithBrokers sets brokers list.
func WithBrokers(brokers []string) Option {
	return func(consumer *Consumer) error {
		if brokers == nil {
			return fmt.Errorf("consumer brokers error: [%w]", cerrors.ErrValueRequired)
		}

		consumer.Brokers = make([]string, len(brokers))
		copy(consumer.Brokers, brokers)

		return nil
	}
}

// WithPartition sets partition.
func WithPartition(i int) Option {
	return func(consumer *Consumer) error {
		if i > 2147483647 || i < -2147483648 {
			return fmt.Errorf("consumer partition error: [%w]", cerrors.ErrInvalid)
		}
		consumer.Partition = int32(i)

		return nil
	}
}

// WithDialTimeout sets dial timeout.
func WithDialTimeout(d time.Duration) Option {
	return func(consumer *Consumer) error {
		consumer.DialTimeout = d

		return nil
	}
}

// WithReadTimeout sets read timeout.
func WithReadTimeout(d time.Duration) Option {
	return func(consumer *Consumer) error {
		consumer.ReadTimeout = d

		return nil
	}
}

// WithWriteTimeout sets write timeout.
func WithWriteTimeout(d time.Duration) Option {
	return func(consumer *Consumer) error {
		consumer.WriteTimeout = d

		return nil
	}
}

// WithBackoff sets backoff duration.
func WithBackoff(d time.Duration) Option {
	return func(consumer *Consumer) error {
		if d == 0 {
			return fmt.Errorf("consumer backoff error: [%w]", cerrors.ErrValueRequired)
		}
		consumer.Backoff = d

		return nil
	}
}

// WithMaxRetries sets max retries value.
func WithMaxRetries(i int) Option {
	return func(consumer *Consumer) error {
		if i > 255 || i < 0 {
			return fmt.Errorf("consumer max retries error: [%w]", cerrors.ErrInvalid)
		}
		consumer.MaxRetries = uint8(i)

		return nil
	}
}

// New instantiates new kafka consumer instance.
func New(options ...Option) (*Consumer, error) {
	consumer := new(Consumer)

	for _, option := range options {
		if err := option(consumer); err != nil {
			return nil, fmt.Errorf("consumer option error: [%w]", err)
		}
	}

	if consumer.Logger == nil {
		return nil, fmt.Errorf("consumer logger error: [%w]", cerrors.ErrValueRequired)
	}

	if consumer.Topic == "" {
		return nil, fmt.Errorf("consumer topic error: [%w]", cerrors.ErrValueRequired)
	}

	return consumer, nil
}

// Run runs kafa consumer.
func Run() error {
	logLevel := getenv.String("LOG_LEVEL", loggerDefaultLevel)
	partition := getenv.Int("KK_PARTITION", 0)
	topic := getenv.String("KK_TOPIC", kkDefaultTopic)
	broker1 := getenv.TCPAddr("KK_BROKER_1", kkDefaultBroker1)

	dialTimeout := getenv.Duration("KK_DIAL_TIMEOUT", kkDefaultDialTimeout)
	readTimeout := getenv.Duration("KK_READ_TIMEOUT", kkDefaultReadTimeout)
	writeTimeout := getenv.Duration("KK_WRITE_TIMEOUT", kkDefaultWriteTimeout)
	backoff := getenv.Duration("KK_BACKOFF", kkDefaultBackoff)
	maxRetries := getenv.Int("KK_MAX_RETRIES", kkMaxRetries)
	if err := getenv.Parse(); err != nil {
		return fmt.Errorf("kafka consumer run error, getenv: [%w]", err)
	}

	logger, errlg := slogger.New(
		slogger.WithLogLevelName(*logLevel),
	)
	if errlg != nil {
		return fmt.Errorf("run error, logger: [%w]", errlg)
	}

	brokers := []string{*broker1}

	kafkaCons, errkc := New(
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
	if errkc != nil {
		return fmt.Errorf("kafka consumer run error, new kafka consumer: [%w]", errkc)
	}

	if err := kafkaCons.Start(); err != nil {
		return fmt.Errorf("run error, consumer.Run: [%w]", err)
	}

	return nil
}
