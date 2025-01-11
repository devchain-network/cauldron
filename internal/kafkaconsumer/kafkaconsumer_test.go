package kafkaconsumer_test

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/IBM/sarama"
	"github.com/IBM/sarama/mocks"
	"github.com/devchain-network/cauldron/internal/cerrors"
	"github.com/devchain-network/cauldron/internal/kafkaconsumer"
	"github.com/devchain-network/cauldron/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockJSONLogHandler struct{}

func (h *MockJSONLogHandler) Enabled(_ context.Context, _ slog.Level) bool {
	return true
}

func (h *MockJSONLogHandler) Handle(_ context.Context, record slog.Record) error {
	return nil
}

func (h *MockJSONLogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h
}

func (h *MockJSONLogHandler) WithGroup(name string) slog.Handler {
	return h
}

type MockStorer struct {
	mock.Mock
}

func (m *MockStorer) GitHubStore(payload *storage.GitHubWebhookData) error {
	args := m.Called(payload)
	return args.Error(0)
}

func (m *MockStorer) Ping() error {
	return nil
}

type MockPartitionConsumer struct {
	mock.Mock
}

func (m *MockPartitionConsumer) Messages() <-chan *sarama.ConsumerMessage {
	args := m.Called()
	return args.Get(0).(<-chan *sarama.ConsumerMessage)
}

func (m *MockPartitionConsumer) Errors() <-chan *sarama.ConsumerError {
	args := m.Called()
	return args.Get(0).(<-chan *sarama.ConsumerError)
}

func (m *MockPartitionConsumer) Close() error {
	return m.Called().Error(0)
}

func getLogger() *slog.Logger {
	return slog.New(new(MockJSONLogHandler))
}

func TestNew_ErrorNilLogger(t *testing.T) {
	consumer, err := kafkaconsumer.New()
	assert.Error(t, err)
	assert.ErrorIs(t, err, cerrors.ErrValueRequired)
	assert.Nil(t, consumer)
}

func TestNew_ErrorNilStorage(t *testing.T) {
	consumer, err := kafkaconsumer.New(
		kafkaconsumer.WithLogger(getLogger()),
	)
	assert.Error(t, err)
	assert.ErrorIs(t, err, cerrors.ErrValueRequired)
	assert.Nil(t, consumer)
}

func TestNew_ErrorEmptyTopic(t *testing.T) {
	db := new(MockStorer)
	db.On("GitHubStore", mock.Anything).Return(nil)

	consumer, err := kafkaconsumer.New(
		kafkaconsumer.WithLogger(getLogger()),
		kafkaconsumer.WithStorage(db),
	)
	assert.Error(t, err)
	assert.ErrorIs(t, err, cerrors.ErrValueRequired)
	assert.Nil(t, consumer)
}

func TestNew_ErrorInvalidBrokers(t *testing.T) {
	db := new(MockStorer)
	db.On("GitHubStore", mock.Anything).Return(nil)

	consumer, err := kafkaconsumer.New(
		kafkaconsumer.WithLogger(getLogger()),
		kafkaconsumer.WithStorage(db),
		kafkaconsumer.WithTopic("test"),
		kafkaconsumer.WithBrokers([]string{"foo"}),
	)
	assert.Error(t, err)
	assert.Nil(t, consumer)
}

func TestNew_ErrorInvalidPartitionNegative(t *testing.T) {
	db := new(MockStorer)
	db.On("GitHubStore", mock.Anything).Return(nil)

	consumer, err := kafkaconsumer.New(
		kafkaconsumer.WithLogger(getLogger()),
		kafkaconsumer.WithStorage(db),
		kafkaconsumer.WithTopic("test"),
		kafkaconsumer.WithPartition(-1),
	)
	assert.Error(t, err)
	assert.ErrorIs(t, err, cerrors.ErrInvalid)
	assert.Nil(t, consumer)
}

func TestNew_ErrorInvalidPartition(t *testing.T) {
	db := new(MockStorer)
	db.On("GitHubStore", mock.Anything).Return(nil)

	consumer, err := kafkaconsumer.New(
		kafkaconsumer.WithLogger(getLogger()),
		kafkaconsumer.WithStorage(db),
		kafkaconsumer.WithTopic("test"),
		kafkaconsumer.WithPartition(2147483648),
	)
	assert.Error(t, err)
	assert.ErrorIs(t, err, cerrors.ErrInvalid)
	assert.Nil(t, consumer)
}

func TestNew_ErrorInvalidDialTimeout(t *testing.T) {
	db := new(MockStorer)
	db.On("GitHubStore", mock.Anything).Return(nil)

	consumer, err := kafkaconsumer.New(
		kafkaconsumer.WithLogger(getLogger()),
		kafkaconsumer.WithStorage(db),
		kafkaconsumer.WithTopic("test"),
		kafkaconsumer.WithDialTimeout(-1*time.Second),
	)
	assert.Error(t, err)
	assert.ErrorIs(t, err, cerrors.ErrInvalid)
	assert.Nil(t, consumer)
}

func TestNew_ErrorInvalidReadTimeout(t *testing.T) {
	db := new(MockStorer)
	db.On("GitHubStore", mock.Anything).Return(nil)

	consumer, err := kafkaconsumer.New(
		kafkaconsumer.WithLogger(getLogger()),
		kafkaconsumer.WithStorage(db),
		kafkaconsumer.WithTopic("test"),
		kafkaconsumer.WithReadTimeout(-1*time.Second),
	)
	assert.Error(t, err)
	assert.ErrorIs(t, err, cerrors.ErrInvalid)
	assert.Nil(t, consumer)
}

func TestNew_ErrorInvalidWriteTimeout(t *testing.T) {
	db := new(MockStorer)
	db.On("GitHubStore", mock.Anything).Return(nil)

	consumer, err := kafkaconsumer.New(
		kafkaconsumer.WithLogger(getLogger()),
		kafkaconsumer.WithStorage(db),
		kafkaconsumer.WithTopic("test"),
		kafkaconsumer.WithWriteTimeout(-1*time.Second),
	)
	assert.Error(t, err)
	assert.ErrorIs(t, err, cerrors.ErrInvalid)
	assert.Nil(t, consumer)
}

func TestNew_ErrorZeroBackoff(t *testing.T) {
	db := new(MockStorer)
	db.On("GitHubStore", mock.Anything).Return(nil)

	consumer, err := kafkaconsumer.New(
		kafkaconsumer.WithLogger(getLogger()),
		kafkaconsumer.WithStorage(db),
		kafkaconsumer.WithTopic("test"),
		kafkaconsumer.WithBackoff(0),
	)
	assert.Error(t, err)
	assert.ErrorIs(t, err, cerrors.ErrValueRequired)
	assert.Nil(t, consumer)
}

func TestNew_ErrorInvalidBackoff(t *testing.T) {
	db := new(MockStorer)
	db.On("GitHubStore", mock.Anything).Return(nil)

	consumer, err := kafkaconsumer.New(
		kafkaconsumer.WithLogger(getLogger()),
		kafkaconsumer.WithStorage(db),
		kafkaconsumer.WithTopic("test"),
		kafkaconsumer.WithBackoff(-1),
	)
	assert.Error(t, err)
	assert.ErrorIs(t, err, cerrors.ErrInvalid)
	assert.Nil(t, consumer)
}

func TestNew_ErrorInvalidBackoffLong(t *testing.T) {
	db := new(MockStorer)
	db.On("GitHubStore", mock.Anything).Return(nil)

	consumer, err := kafkaconsumer.New(
		kafkaconsumer.WithLogger(getLogger()),
		kafkaconsumer.WithStorage(db),
		kafkaconsumer.WithTopic("test"),
		kafkaconsumer.WithBackoff(2*time.Minute),
	)
	assert.Error(t, err)
	assert.ErrorIs(t, err, cerrors.ErrInvalid)
	assert.Nil(t, consumer)
}

func TestNew_ErrorInvalidMaxRetriesNegative(t *testing.T) {
	db := new(MockStorer)
	db.On("GitHubStore", mock.Anything).Return(nil)

	consumer, err := kafkaconsumer.New(
		kafkaconsumer.WithLogger(getLogger()),
		kafkaconsumer.WithStorage(db),
		kafkaconsumer.WithTopic("test"),
		kafkaconsumer.WithMaxRetries(-1),
	)
	assert.Error(t, err)
	assert.ErrorIs(t, err, cerrors.ErrInvalid)
	assert.Nil(t, consumer)
}

func TestNew_ErrorInvalidMaxRetries(t *testing.T) {
	db := new(MockStorer)
	db.On("GitHubStore", mock.Anything).Return(nil)

	consumer, err := kafkaconsumer.New(
		kafkaconsumer.WithLogger(getLogger()),
		kafkaconsumer.WithStorage(db),
		kafkaconsumer.WithTopic("test"),
		kafkaconsumer.WithMaxRetries(256),
	)
	assert.Error(t, err)
	assert.ErrorIs(t, err, cerrors.ErrInvalid)
	assert.Nil(t, consumer)
}

func TestNew_Defaults(t *testing.T) {
	db := new(MockStorer)
	db.On("GitHubStore", mock.Anything).Return(nil)

	consumer, err := kafkaconsumer.New(
		kafkaconsumer.WithLogger(getLogger()),
		kafkaconsumer.WithStorage(db),
		kafkaconsumer.WithBrokers([]string{"127.0.0.1:9094"}),
		kafkaconsumer.WithTopic("test"),
	)
	assert.NoError(t, err)
	assert.NotNil(t, consumer)
	assert.Equal(t, consumer.DialTimeout, 30*time.Second)
	assert.Equal(t, consumer.ReadTimeout, 30*time.Second)
	assert.Equal(t, consumer.WriteTimeout, 30*time.Second)
	assert.Equal(t, consumer.Backoff, 2*time.Second)
	assert.Equal(t, consumer.MaxRetries, uint8(10))
	assert.Equal(t, consumer.Brokers, []string{"127.0.0.1:9094"})
}

func TestPing_Success(t *testing.T) {
	db := new(MockStorer)
	db.On("GitHubStore", mock.Anything).Return(nil)

	consumer, err := kafkaconsumer.New(
		kafkaconsumer.WithLogger(getLogger()),
		kafkaconsumer.WithStorage(db),
		kafkaconsumer.WithTopic("test"),
	)

	assert.NoError(t, err)
	assert.NotNil(t, consumer)

	mockConsumer := mocks.NewConsumer(t, nil)

	consumer.ConsumerConfig = func() *sarama.Config {
		return mocks.NewTestConfig()
	}

	consumer.ConsumerFactory = func(brokers []string, config *sarama.Config) (sarama.Consumer, error) {
		return mockConsumer, nil
	}

	err = consumer.Ping()
	assert.NoError(t, err)
}
