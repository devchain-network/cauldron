package kafkaproducer_test

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/IBM/sarama"
	"github.com/IBM/sarama/mocks"
	"github.com/devchain-network/cauldron/internal/cerrors"
	"github.com/devchain-network/cauldron/internal/kafkacp"
	"github.com/devchain-network/cauldron/internal/kafkacp/kafkaproducer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

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

type mockProducerFactory struct {
	mock.Mock
}

func (m *mockProducerFactory) NewAsyncProducer(brokers []string, config *sarama.Config) (sarama.AsyncProducer, error) {
	args := m.Called(brokers, config)
	return args.Get(0).(sarama.AsyncProducer), args.Error(1)
}

func TestNew_MissingRequiredFields(t *testing.T) {
	producer, err := kafkaproducer.New()
	assert.ErrorIs(t, err, cerrors.ErrValueRequired)
	assert.Nil(t, producer)
}

func TestNew_InvalidKafkaBrokers(t *testing.T) {
	var kafkaBrokers kafkacp.KafkaBrokers
	kafkaBrokers.AddFromString("invalid")

	logger := slog.New(new(mockLogger))
	producer, err := kafkaproducer.New(
		kafkaproducer.WithLogger(logger),
		kafkaproducer.WithKafkaBrokers(kafkaBrokers),
	)
	assert.ErrorIs(t, err, cerrors.ErrInvalid)
	assert.Nil(t, producer)
}

func TestNew_InvalidMaxRetries(t *testing.T) {
	logger := slog.New(new(mockLogger))
	producer, err := kafkaproducer.New(
		kafkaproducer.WithLogger(logger),
		kafkaproducer.WithMaxRetries(300),
	)
	assert.ErrorIs(t, err, cerrors.ErrInvalid)
	assert.Nil(t, producer)
}

func TestNew_InvalidBackoff(t *testing.T) {
	logger := slog.New(new(mockLogger))
	producer, err := kafkaproducer.New(
		kafkaproducer.WithLogger(logger),
		kafkaproducer.WithBackoff(0),
	)
	assert.ErrorIs(t, err, cerrors.ErrValueRequired)
	assert.Nil(t, producer)
}

func TestNew_InvalidDialTimeout(t *testing.T) {
	logger := slog.New(new(mockLogger))
	producer, err := kafkaproducer.New(
		kafkaproducer.WithLogger(logger),
		kafkaproducer.WithDialTimeout(-1*time.Second),
	)
	assert.ErrorIs(t, err, cerrors.ErrInvalid)
	assert.Nil(t, producer)
}

func TestNew_InvalidReadTimeout(t *testing.T) {
	logger := slog.New(new(mockLogger))
	producer, err := kafkaproducer.New(
		kafkaproducer.WithLogger(logger),
		kafkaproducer.WithReadTimeout(-1*time.Second),
	)
	assert.ErrorIs(t, err, cerrors.ErrInvalid)
	assert.Nil(t, producer)
}

func TestNew_InvalidWriteTimeout(t *testing.T) {
	logger := slog.New(new(mockLogger))
	producer, err := kafkaproducer.New(
		kafkaproducer.WithLogger(logger),
		kafkaproducer.WithWriteTimeout(-1*time.Second),
	)
	assert.ErrorIs(t, err, cerrors.ErrInvalid)
	assert.Nil(t, producer)
}

func TestNew_WithNilProducerFactoryFunc(t *testing.T) {
	logger := slog.New(new(mockLogger))
	producer, err := kafkaproducer.New(
		kafkaproducer.WithLogger(logger),
		kafkaproducer.WithSaramaProducerFactoryFunc(nil),
	)
	assert.ErrorIs(t, err, cerrors.ErrValueRequired)
	assert.Nil(t, producer)
}

func TestNew_WithSaramaProducerFactoryFunc_Error(t *testing.T) {
	logger := slog.New(new(mockLogger))

	mockConfig := mocks.NewTestConfig()
	mockProducer := mocks.NewAsyncProducer(t, mockConfig)

	mockFactory := &mockProducerFactory{}
	mockFactory.On("NewAsyncProducer", mock.Anything, mock.Anything).
		Return(mockProducer, sarama.ErrOutOfBrokers)

	producer, err := kafkaproducer.New(
		kafkaproducer.WithLogger(logger),
		kafkaproducer.WithSaramaProducerFactoryFunc(mockFactory.NewAsyncProducer),
		kafkaproducer.WithMaxRetries(1),
		kafkaproducer.WithBackoff(100*time.Millisecond),
	)

	assert.Error(t, err)
	assert.Nil(t, producer)

	mockFactory.AssertNumberOfCalls(t, "NewAsyncProducer", 1)
	mockFactory.AssertExpectations(t)
}

func TestNew_NoLogger(t *testing.T) {
	var kafkaBrokers kafkacp.KafkaBrokers
	kafkaBrokers.AddFromString("127.0.0.1:9094")

	producer, err := kafkaproducer.New(
		kafkaproducer.WithKafkaBrokers(kafkaBrokers),
		kafkaproducer.WithMaxRetries(2),
		kafkaproducer.WithBackoff(time.Second),
		kafkaproducer.WithDialTimeout(5*time.Second),
		kafkaproducer.WithReadTimeout(5*time.Second),
		kafkaproducer.WithWriteTimeout(5*time.Second),
	)
	assert.ErrorIs(t, err, cerrors.ErrValueRequired)
	assert.Nil(t, producer)
}

func TestNew_NilLogger(t *testing.T) {
	var kafkaBrokers kafkacp.KafkaBrokers
	kafkaBrokers.AddFromString("127.0.0.1:9094")

	producer, err := kafkaproducer.New(
		kafkaproducer.WithLogger(nil),
		kafkaproducer.WithKafkaBrokers(kafkaBrokers),
		kafkaproducer.WithMaxRetries(2),
		kafkaproducer.WithBackoff(time.Second),
		kafkaproducer.WithDialTimeout(5*time.Second),
		kafkaproducer.WithReadTimeout(5*time.Second),
		kafkaproducer.WithWriteTimeout(5*time.Second),
	)
	assert.ErrorIs(t, err, cerrors.ErrValueRequired)
	assert.Nil(t, producer)
}

func TestNew_Success(t *testing.T) {
	logger := slog.New(new(mockLogger))

	mockConfig := mocks.NewTestConfig()
	mockProducer := mocks.NewAsyncProducer(t, mockConfig)

	mockFactory := &mockProducerFactory{}
	mockFactory.On("NewAsyncProducer", mock.Anything, mock.Anything).
		Return(mockProducer, sarama.ErrOutOfBrokers).
		Twice()
	mockFactory.On("NewAsyncProducer", mock.Anything, mock.Anything).Return(mockProducer, nil).Once()

	producer, err := kafkaproducer.New(
		kafkaproducer.WithLogger(logger),
		kafkaproducer.WithSaramaProducerFactoryFunc(mockFactory.NewAsyncProducer),
		kafkaproducer.WithMaxRetries(3),
		kafkaproducer.WithBackoff(100*time.Millisecond),
	)

	assert.NoError(t, err)
	assert.NotNil(t, producer)

	mockFactory.AssertNumberOfCalls(t, "NewAsyncProducer", 3)
	mockFactory.AssertExpectations(t)
}
