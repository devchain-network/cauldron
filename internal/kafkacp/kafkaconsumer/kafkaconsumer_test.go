package kafkaconsumer_test

import (
	"context"
	"log/slog"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/IBM/sarama"
	"github.com/IBM/sarama/mocks"
	"github.com/devchain-network/cauldron/internal/cerrors"
	"github.com/devchain-network/cauldron/internal/kafkacp/kafkaconsumer"
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

type mockLogger struct{}

func (h *mockLogger) Enabled(_ context.Context, _ slog.Level) bool {
	return true
}

func (h *mockLogger) Handle(_ context.Context, record slog.Record) error {
	return nil
}

func (h *mockLogger) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h
}

func (h *mockLogger) WithGroup(name string) slog.Handler {
	return h
}

type mockStorage struct {
	mock.Mock
}

func (m *mockStorage) MessageStore(ctx context.Context, msg *sarama.ConsumerMessage) error {
	args := m.Called(ctx, msg)
	return args.Error(0)
}

func (m *mockStorage) Ping(ctx context.Context, maxRetries uint8, backoff time.Duration) error {
	args := m.Called(ctx, maxRetries, backoff)
	return args.Error(0)
}

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

func TestNew_NoStorage(t *testing.T) {
	logger := slog.New(new(mockLogger))

	consumer, err := kafkaconsumer.New(
		kafkaconsumer.WithLogger(logger),
	)

	assert.ErrorIs(t, err, cerrors.ErrValueRequired)
	assert.Nil(t, consumer)
}

func TestNew_NilStorage(t *testing.T) {
	logger := slog.New(new(mockLogger))

	consumer, err := kafkaconsumer.New(
		kafkaconsumer.WithLogger(logger),
		kafkaconsumer.WithStorage(nil),
	)

	assert.ErrorIs(t, err, cerrors.ErrValueRequired)
	assert.Nil(t, consumer)
}

func TestNew_EmptyTopic(t *testing.T) {
	logger := slog.New(new(mockLogger))
	storage := new(mockStorage)

	consumer, err := kafkaconsumer.New(
		kafkaconsumer.WithLogger(logger),
		kafkaconsumer.WithStorage(storage),
	)

	assert.ErrorIs(t, err, cerrors.ErrInvalid)
	assert.Nil(t, consumer)
}

func TestNew_InvalidTopic(t *testing.T) {
	logger := slog.New(new(mockLogger))
	storage := new(mockStorage)

	consumer, err := kafkaconsumer.New(
		kafkaconsumer.WithLogger(logger),
		kafkaconsumer.WithStorage(storage),
		kafkaconsumer.WithTopic("invalid"),
	)

	assert.ErrorIs(t, err, cerrors.ErrInvalid)
	assert.Nil(t, consumer)
}

func TestNew_InvalidPartition(t *testing.T) {
	logger := slog.New(new(mockLogger))
	storage := new(mockStorage)

	consumer, err := kafkaconsumer.New(
		kafkaconsumer.WithLogger(logger),
		kafkaconsumer.WithStorage(storage),
		kafkaconsumer.WithTopic("github"),
		kafkaconsumer.WithPartition(2147483648),
	)

	assert.ErrorIs(t, err, cerrors.ErrInvalid)
	assert.Nil(t, consumer)
}

func TestNew_InvalidBrokers(t *testing.T) {
	logger := slog.New(new(mockLogger))
	storage := new(mockStorage)

	consumer, err := kafkaconsumer.New(
		kafkaconsumer.WithLogger(logger),
		kafkaconsumer.WithStorage(storage),
		kafkaconsumer.WithTopic("github"),
		kafkaconsumer.WithKafkaBrokers("invalid"),
	)

	assert.ErrorIs(t, err, cerrors.ErrInvalid)
	assert.Nil(t, consumer)
}

func TestNew_InvalidDialTimeout(t *testing.T) {
	logger := slog.New(new(mockLogger))
	storage := new(mockStorage)

	consumer, err := kafkaconsumer.New(
		kafkaconsumer.WithLogger(logger),
		kafkaconsumer.WithStorage(storage),
		kafkaconsumer.WithTopic("github"),
		kafkaconsumer.WithKafkaBrokers("127.0.0.1:9094"),
		kafkaconsumer.WithDialTimeout(-1*time.Second),
	)

	assert.ErrorIs(t, err, cerrors.ErrInvalid)
	assert.Nil(t, consumer)
}

func TestNew_InvalidReadTimeout(t *testing.T) {
	logger := slog.New(new(mockLogger))
	storage := new(mockStorage)

	consumer, err := kafkaconsumer.New(
		kafkaconsumer.WithLogger(logger),
		kafkaconsumer.WithStorage(storage),
		kafkaconsumer.WithTopic("github"),
		kafkaconsumer.WithKafkaBrokers("127.0.0.1:9094"),
		kafkaconsumer.WithReadTimeout(-1*time.Second),
	)

	assert.ErrorIs(t, err, cerrors.ErrInvalid)
	assert.Nil(t, consumer)
}

func TestNew_InvalidWriteTimeout(t *testing.T) {
	logger := slog.New(new(mockLogger))
	storage := new(mockStorage)

	consumer, err := kafkaconsumer.New(
		kafkaconsumer.WithLogger(logger),
		kafkaconsumer.WithStorage(storage),
		kafkaconsumer.WithTopic("github"),
		kafkaconsumer.WithKafkaBrokers("127.0.0.1:9094"),
		kafkaconsumer.WithWriteTimeout(-1*time.Second),
	)

	assert.ErrorIs(t, err, cerrors.ErrInvalid)
	assert.Nil(t, consumer)
}

func TestNew_ZeroBackoff(t *testing.T) {
	logger := slog.New(new(mockLogger))
	storage := new(mockStorage)

	consumer, err := kafkaconsumer.New(
		kafkaconsumer.WithLogger(logger),
		kafkaconsumer.WithStorage(storage),
		kafkaconsumer.WithTopic("github"),
		kafkaconsumer.WithKafkaBrokers("127.0.0.1:9094"),
		kafkaconsumer.WithBackoff(0),
	)

	assert.ErrorIs(t, err, cerrors.ErrValueRequired)
	assert.Nil(t, consumer)
}

func TestNew_InvalidBackoff(t *testing.T) {
	logger := slog.New(new(mockLogger))
	storage := new(mockStorage)

	consumer, err := kafkaconsumer.New(
		kafkaconsumer.WithLogger(logger),
		kafkaconsumer.WithStorage(storage),
		kafkaconsumer.WithTopic("github"),
		kafkaconsumer.WithKafkaBrokers("127.0.0.1:9094"),
		kafkaconsumer.WithBackoff(2*time.Minute),
	)

	assert.ErrorIs(t, err, cerrors.ErrInvalid)
	assert.Nil(t, consumer)
}

func TestNew_InvalidMaxRetries(t *testing.T) {
	logger := slog.New(new(mockLogger))
	storage := new(mockStorage)

	consumer, err := kafkaconsumer.New(
		kafkaconsumer.WithLogger(logger),
		kafkaconsumer.WithStorage(storage),
		kafkaconsumer.WithTopic("github"),
		kafkaconsumer.WithKafkaBrokers("127.0.0.1:9094"),
		kafkaconsumer.WithMaxRetries(256),
	)

	assert.ErrorIs(t, err, cerrors.ErrInvalid)
	assert.Nil(t, consumer)
}

func TestNew_NilSaramaConsumerFactor(t *testing.T) {
	logger := slog.New(new(mockLogger))
	storage := new(mockStorage)

	consumer, err := kafkaconsumer.New(
		kafkaconsumer.WithLogger(logger),
		kafkaconsumer.WithStorage(storage),
		kafkaconsumer.WithTopic("github"),
		kafkaconsumer.WithKafkaBrokers("127.0.0.1:9094"),
		kafkaconsumer.WithSaramaConsumerFactoryFunc(nil),
	)

	assert.ErrorIs(t, err, cerrors.ErrValueRequired)
	assert.Nil(t, consumer)
}

func TestNew_WithSaramaConsumerFactoryFunc_Error(t *testing.T) {
	logger := slog.New(new(mockLogger))
	storage := new(mockStorage)

	mockConfig := mocks.NewTestConfig()
	mockSarama := mocks.NewConsumer(t, mockConfig)

	mockFactory := &mockConsumerFactory{}
	mockFactory.On("NewConsumer", mock.Anything, mock.Anything).Return(mockSarama, sarama.ErrOutOfBrokers)

	consumer, err := kafkaconsumer.New(
		kafkaconsumer.WithLogger(logger),
		kafkaconsumer.WithStorage(storage),
		kafkaconsumer.WithTopic("github"),
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
	logger := slog.New(new(mockLogger))
	storage := new(mockStorage)

	mockConfig := mocks.NewTestConfig()
	mockSarama := mocks.NewConsumer(t, mockConfig)

	mockFactory := &mockConsumerFactory{}
	mockFactory.On("NewConsumer", mock.Anything, mock.Anything).Return(mockSarama, sarama.ErrOutOfBrokers).Once()
	mockFactory.On("NewConsumer", mock.Anything, mock.Anything).Return(mockSarama, sarama.ErrOutOfBrokers).Once()
	mockFactory.On("NewConsumer", mock.Anything, mock.Anything).Return(mockSarama, nil).Once()

	consumer, err := kafkaconsumer.New(
		kafkaconsumer.WithLogger(logger),
		kafkaconsumer.WithStorage(storage),
		kafkaconsumer.WithTopic("github"),
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
	logger := slog.New(new(mockLogger))
	storage := new(mockStorage)
	storage.On("MessageStore", mock.Anything, mock.Anything).Return(nil)

	mockConfig := mocks.NewTestConfig()
	mockSarama := mocks.NewConsumer(t, mockConfig)
	mockSarama.ExpectConsumePartition("github", 0, sarama.OffsetNewest).YieldMessage(
		&sarama.ConsumerMessage{
			Value:     []byte(`{"test": "message"}`),
			Topic:     "github",
			Partition: 0,
		},
	)

	mockFactory := &mockConsumerFactory{}
	mockFactory.On("NewConsumer", mock.Anything, mock.Anything).Return(mockSarama, nil).Once()

	consumer, err := kafkaconsumer.New(
		kafkaconsumer.WithLogger(logger),
		kafkaconsumer.WithStorage(storage),
		kafkaconsumer.WithTopic("github"),
		kafkaconsumer.WithSaramaConsumerFactoryFunc(mockFactory.NewConsumer),
		kafkaconsumer.WithMaxRetries(1),
	)

	assert.NoError(t, err)
	assert.NotNil(t, consumer)

	go func() {
		err := consumer.Consume()
		assert.NoError(t, err)
	}()
	time.Sleep(1 * time.Second)

	mockFactory.AssertNumberOfCalls(t, "NewConsumer", 1)
	mockFactory.AssertExpectations(t)
	storage.AssertNumberOfCalls(t, "MessageStore", 1)
	storage.AssertExpectations(t)
}

func TestConsumer_Consume_PartitionConsumeError(t *testing.T) {
	logger := slog.New(new(mockLogger))
	storage := new(mockStorage)

	mockConfig := mocks.NewTestConfig()
	mockSarama := mocks.NewConsumer(t, mockConfig)
	mockSarama.ExpectConsumePartition("github", 0, sarama.OffsetNewest).YieldError(sarama.ErrOutOfBrokers)

	mockFactory := &mockConsumerFactory{}
	mockFactory.On("NewConsumer", mock.Anything, mock.Anything).Return(mockSarama, nil).Once()

	consumer, err := kafkaconsumer.New(
		kafkaconsumer.WithLogger(logger),
		kafkaconsumer.WithStorage(storage),
		kafkaconsumer.WithTopic("github"),
		kafkaconsumer.WithSaramaConsumerFactoryFunc(mockFactory.NewConsumer),
		kafkaconsumer.WithMaxRetries(1),
	)
	assert.NoError(t, err)
	assert.NotNil(t, consumer)

	go func() {
		err = consumer.Consume()
		assert.NoError(t, err)
	}()

	time.Sleep(100 * time.Millisecond)
	process, _ := os.FindProcess(syscall.Getpid())
	_ = process.Signal(os.Interrupt)

	mockFactory.AssertNumberOfCalls(t, "NewConsumer", 1)
	mockFactory.AssertExpectations(t)
}
