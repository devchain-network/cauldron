package kafkaconsumer_test

import (
	"context"
	"log/slog"
	"os"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/IBM/sarama"
	"github.com/IBM/sarama/mocks"
	"github.com/devchain-network/cauldron/internal/cerrors"
	"github.com/devchain-network/cauldron/internal/kafkacp"
	"github.com/devchain-network/cauldron/internal/kafkacp/kafkaconsumer"
	"github.com/devchain-network/cauldron/internal/slogger/mockslogger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockConsumerFactory struct {
	mock.Mock
}

func (m *mockConsumerFactory) NewConsumer(brokers []string, config *sarama.Config) (sarama.Consumer, error) {
	args := m.Called(brokers, config)
	return args.Get(0).(sarama.Consumer), args.Error(1)
}

var (
	mockLog            = slog.New(new(mockslogger.MockLogger))
	mockProcessMessage = func(ctx context.Context, msg *sarama.ConsumerMessage) error {
		return nil
	}
)

func TestNew_MissingRequiredFields(t *testing.T) {
	consumer, err := kafkaconsumer.New()

	assert.ErrorIs(t, err, cerrors.ErrValueRequired)
	assert.Nil(t, consumer)
}

func TestNew_NilLogger(t *testing.T) {
	consumer, err := kafkaconsumer.New(
		kafkaconsumer.WithLogger(nil),
	)

	assert.ErrorIs(t, err, cerrors.ErrValueRequired)
	assert.Nil(t, consumer)
}

func TestNew_NoProcessMessageFunc(t *testing.T) {
	logger := mockLog

	consumer, err := kafkaconsumer.New(
		kafkaconsumer.WithLogger(logger),
	)

	assert.ErrorIs(t, err, cerrors.ErrValueRequired)
	assert.Nil(t, consumer)
}

func TestNew_NilProcessMessageFunc(t *testing.T) {
	logger := mockLog

	consumer, err := kafkaconsumer.New(
		kafkaconsumer.WithLogger(logger),
		kafkaconsumer.WithProcessMessageFunc(nil),
	)

	assert.ErrorIs(t, err, cerrors.ErrValueRequired)
	assert.Nil(t, consumer)
}

func TestNew_EmptyTopic(t *testing.T) {
	logger := mockLog

	consumer, err := kafkaconsumer.New(
		kafkaconsumer.WithLogger(logger),
		kafkaconsumer.WithProcessMessageFunc(mockProcessMessage),
	)

	assert.ErrorIs(t, err, cerrors.ErrInvalid)
	assert.Nil(t, consumer)
}

func TestNew_InvalidTopic(t *testing.T) {
	logger := mockLog

	consumer, err := kafkaconsumer.New(
		kafkaconsumer.WithLogger(logger),
		kafkaconsumer.WithProcessMessageFunc(mockProcessMessage),
		kafkaconsumer.WithTopic("invalid"),
	)

	assert.ErrorIs(t, err, cerrors.ErrInvalid)
	assert.Nil(t, consumer)
}

func TestNew_InvalidPartition(t *testing.T) {
	logger := mockLog

	consumer, err := kafkaconsumer.New(
		kafkaconsumer.WithLogger(logger),
		kafkaconsumer.WithProcessMessageFunc(mockProcessMessage),
		kafkaconsumer.WithTopic(kafkacp.KafkaTopicIdentifierGitHub.String()),
		kafkaconsumer.WithPartition(2147483648),
	)

	assert.ErrorIs(t, err, cerrors.ErrInvalid)
	assert.Nil(t, consumer)
}

func TestNew_InvalidBrokers(t *testing.T) {
	logger := mockLog

	consumer, err := kafkaconsumer.New(
		kafkaconsumer.WithLogger(logger),
		kafkaconsumer.WithProcessMessageFunc(mockProcessMessage),
		kafkaconsumer.WithTopic(kafkacp.KafkaTopicIdentifierGitHub.String()),
		kafkaconsumer.WithKafkaBrokers("invalid"),
	)

	assert.ErrorIs(t, err, cerrors.ErrInvalid)
	assert.Nil(t, consumer)
}

func TestNew_InvalidDialTimeout(t *testing.T) {
	logger := mockLog

	consumer, err := kafkaconsumer.New(
		kafkaconsumer.WithLogger(logger),
		kafkaconsumer.WithProcessMessageFunc(mockProcessMessage),
		kafkaconsumer.WithTopic(kafkacp.KafkaTopicIdentifierGitHub.String()),
		kafkaconsumer.WithKafkaBrokers("127.0.0.1:9094"),
		kafkaconsumer.WithDialTimeout(-1*time.Second),
	)

	assert.ErrorIs(t, err, cerrors.ErrInvalid)
	assert.Nil(t, consumer)
}

func TestNew_InvalidReadTimeout(t *testing.T) {
	logger := mockLog

	consumer, err := kafkaconsumer.New(
		kafkaconsumer.WithLogger(logger),
		kafkaconsumer.WithProcessMessageFunc(mockProcessMessage),
		kafkaconsumer.WithTopic(kafkacp.KafkaTopicIdentifierGitHub.String()),
		kafkaconsumer.WithKafkaBrokers("127.0.0.1:9094"),
		kafkaconsumer.WithReadTimeout(-1*time.Second),
	)

	assert.ErrorIs(t, err, cerrors.ErrInvalid)
	assert.Nil(t, consumer)
}

func TestNew_InvalidWriteTimeout(t *testing.T) {
	logger := mockLog

	consumer, err := kafkaconsumer.New(
		kafkaconsumer.WithLogger(logger),
		kafkaconsumer.WithProcessMessageFunc(mockProcessMessage),
		kafkaconsumer.WithTopic(kafkacp.KafkaTopicIdentifierGitHub.String()),
		kafkaconsumer.WithKafkaBrokers("127.0.0.1:9094"),
		kafkaconsumer.WithWriteTimeout(-1*time.Second),
	)

	assert.ErrorIs(t, err, cerrors.ErrInvalid)
	assert.Nil(t, consumer)
}

func TestNew_ZeroBackoff(t *testing.T) {
	logger := mockLog

	consumer, err := kafkaconsumer.New(
		kafkaconsumer.WithLogger(logger),
		kafkaconsumer.WithProcessMessageFunc(mockProcessMessage),
		kafkaconsumer.WithTopic(kafkacp.KafkaTopicIdentifierGitHub.String()),
		kafkaconsumer.WithKafkaBrokers("127.0.0.1:9094"),
		kafkaconsumer.WithBackoff(0),
	)

	assert.ErrorIs(t, err, cerrors.ErrValueRequired)
	assert.Nil(t, consumer)
}

func TestNew_InvalidBackoff(t *testing.T) {
	logger := mockLog

	consumer, err := kafkaconsumer.New(
		kafkaconsumer.WithLogger(logger),
		kafkaconsumer.WithProcessMessageFunc(mockProcessMessage),
		kafkaconsumer.WithTopic(kafkacp.KafkaTopicIdentifierGitHub.String()),
		kafkaconsumer.WithKafkaBrokers("127.0.0.1:9094"),
		kafkaconsumer.WithBackoff(2*time.Minute),
	)

	assert.ErrorIs(t, err, cerrors.ErrInvalid)
	assert.Nil(t, consumer)
}

func TestNew_InvalidMaxRetries(t *testing.T) {
	logger := mockLog

	consumer, err := kafkaconsumer.New(
		kafkaconsumer.WithLogger(logger),
		kafkaconsumer.WithProcessMessageFunc(mockProcessMessage),
		kafkaconsumer.WithTopic(kafkacp.KafkaTopicIdentifierGitHub.String()),
		kafkaconsumer.WithKafkaBrokers("127.0.0.1:9094"),
		kafkaconsumer.WithMaxRetries(256),
	)

	assert.ErrorIs(t, err, cerrors.ErrInvalid)
	assert.Nil(t, consumer)
}

func TestNew_NilSaramaConsumerFactoryFunc(t *testing.T) {
	logger := mockLog

	consumer, err := kafkaconsumer.New(
		kafkaconsumer.WithLogger(logger),
		kafkaconsumer.WithProcessMessageFunc(mockProcessMessage),
		kafkaconsumer.WithTopic(kafkacp.KafkaTopicIdentifierGitHub.String()),
		kafkaconsumer.WithKafkaBrokers("127.0.0.1:9094"),
		kafkaconsumer.WithSaramaConsumerFactoryFunc(nil),
	)

	assert.ErrorIs(t, err, cerrors.ErrValueRequired)
	assert.Nil(t, consumer)
}

func TestNew_WithSaramaConsumerFactoryFunc_Error(t *testing.T) {
	logger := mockLog

	mockConfig := mocks.NewTestConfig()
	mockSarama := mocks.NewConsumer(t, mockConfig)

	mockFactory := &mockConsumerFactory{}
	mockFactory.On("NewConsumer", mock.Anything, mock.Anything).Return(mockSarama, sarama.ErrOutOfBrokers)

	consumer, err := kafkaconsumer.New(
		kafkaconsumer.WithLogger(logger),
		kafkaconsumer.WithProcessMessageFunc(mockProcessMessage),
		kafkaconsumer.WithTopic(kafkacp.KafkaTopicIdentifierGitHub.String()),
		kafkaconsumer.WithBackoff(100*time.Millisecond),
		kafkaconsumer.WithSaramaConsumerFactoryFunc(mockFactory.NewConsumer),
		kafkaconsumer.WithMaxRetries(1),
	)

	assert.Error(t, err)
	assert.Nil(t, consumer)

	mockFactory.AssertNumberOfCalls(t, "NewConsumer", 1)
	mockFactory.AssertExpectations(t)
}

func TestNew_WithSaramaConsumerFactoryFunc_Success(t *testing.T) {
	logger := mockLog

	mockConfig := mocks.NewTestConfig()
	mockSarama := mocks.NewConsumer(t, mockConfig)

	mockFactory := &mockConsumerFactory{}
	mockFactory.On("NewConsumer", mock.Anything, mock.Anything).Return(mockSarama, sarama.ErrOutOfBrokers).Twice()
	mockFactory.On("NewConsumer", mock.Anything, mock.Anything).Return(mockSarama, nil).Once()

	consumer, err := kafkaconsumer.New(
		kafkaconsumer.WithLogger(logger),
		kafkaconsumer.WithProcessMessageFunc(mockProcessMessage),
		kafkaconsumer.WithTopic(kafkacp.KafkaTopicIdentifierGitHub.String()),
		kafkaconsumer.WithPartition(0),
		kafkaconsumer.WithDialTimeout(10*time.Second),
		kafkaconsumer.WithReadTimeout(10*time.Second),
		kafkaconsumer.WithWriteTimeout(10*time.Second),
		kafkaconsumer.WithBackoff(100*time.Millisecond),
		kafkaconsumer.WithSaramaConsumerFactoryFunc(mockFactory.NewConsumer),
		kafkaconsumer.WithMaxRetries(3),
	)

	assert.NoError(t, err)
	assert.NotNil(t, consumer)

	mockFactory.AssertNumberOfCalls(t, "NewConsumer", 3)
	mockFactory.AssertExpectations(t)
}

func TestConsumer_Consume_Success(t *testing.T) {
	logger := mockLog

	mockConfig := mocks.NewTestConfig()
	mockSarama := mocks.NewConsumer(t, mockConfig)
	mockSarama.ExpectConsumePartition(kafkacp.KafkaTopicIdentifierGitHub.String(), 0, sarama.OffsetNewest).YieldMessage(
		&sarama.ConsumerMessage{
			Value:     []byte(`{"test": "message"}`),
			Topic:     kafkacp.KafkaTopicIdentifierGitHub.String(),
			Partition: 0,
		},
	)

	mockFactory := &mockConsumerFactory{}
	mockFactory.On("NewConsumer", mock.Anything, mock.Anything).Return(mockSarama, nil).Once()

	consumer, err := kafkaconsumer.New(
		kafkaconsumer.WithLogger(logger),
		kafkaconsumer.WithProcessMessageFunc(mockProcessMessage),
		kafkaconsumer.WithTopic(kafkacp.KafkaTopicIdentifierGitHub.String()),
		kafkaconsumer.WithSaramaConsumerFactoryFunc(mockFactory.NewConsumer),
		kafkaconsumer.WithMaxRetries(1),
	)

	assert.NoError(t, err)
	assert.NotNil(t, consumer)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()

		err := consumer.Consume()
		assert.NoError(t, err)
	}()
	time.Sleep(100 * time.Millisecond)

	process, _ := os.FindProcess(syscall.Getpid())
	_ = process.Signal(os.Interrupt)

	mockFactory.AssertNumberOfCalls(t, "NewConsumer", 1)
	mockFactory.AssertExpectations(t)
	wg.Wait()
}

func TestConsumer_Consume_PartitionConsumeError(t *testing.T) {
	logger := mockLog

	mockConfig := mocks.NewTestConfig()
	mockSarama := mocks.NewConsumer(t, mockConfig)
	mockSarama.ExpectConsumePartition(kafkacp.KafkaTopicIdentifierGitHub.String(), 0, sarama.OffsetNewest).
		YieldError(sarama.ErrOutOfBrokers)

	mockFactory := &mockConsumerFactory{}
	mockFactory.On("NewConsumer", mock.Anything, mock.Anything).Return(mockSarama, nil).Once()

	consumer, err := kafkaconsumer.New(
		kafkaconsumer.WithLogger(logger),
		kafkaconsumer.WithProcessMessageFunc(mockProcessMessage),
		kafkaconsumer.WithTopic(kafkacp.KafkaTopicIdentifierGitHub.String()),
		kafkaconsumer.WithSaramaConsumerFactoryFunc(mockFactory.NewConsumer),
		kafkaconsumer.WithMaxRetries(1),
	)
	assert.NoError(t, err)
	assert.NotNil(t, consumer)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()

		err = consumer.Consume()
		assert.NoError(t, err)
	}()

	time.Sleep(100 * time.Millisecond)

	process, _ := os.FindProcess(syscall.Getpid())
	_ = process.Signal(os.Interrupt)

	mockFactory.AssertNumberOfCalls(t, "NewConsumer", 1)
	mockFactory.AssertExpectations(t)
	wg.Wait()
}
