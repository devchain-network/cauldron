package kafkaconsumergroup_test

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/IBM/sarama"
	"github.com/devchain-network/cauldron/internal/cerrors"
	"github.com/devchain-network/cauldron/internal/kafkacp/kafkaconsumergroup"
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
	consumer, err := kafkaconsumergroup.New()

	assert.ErrorIs(t, err, cerrors.ErrValueRequired)
	assert.Nil(t, consumer)
}

func TestNew_NilLogger(t *testing.T) {
	consumer, err := kafkaconsumergroup.New(
		kafkaconsumergroup.WithLogger(nil),
	)

	assert.ErrorIs(t, err, cerrors.ErrValueRequired)
	assert.Nil(t, consumer)
}

func TestNew_NoStorage(t *testing.T) {
	logger := slog.New(new(mockLogger))

	consumer, err := kafkaconsumergroup.New(
		kafkaconsumergroup.WithLogger(logger),
	)

	assert.ErrorIs(t, err, cerrors.ErrValueRequired)
	assert.Nil(t, consumer)
}

func TestNew_NilStorage(t *testing.T) {
	logger := slog.New(new(mockLogger))

	consumer, err := kafkaconsumergroup.New(
		kafkaconsumergroup.WithLogger(logger),
		kafkaconsumergroup.WithStorage(nil),
	)

	assert.ErrorIs(t, err, cerrors.ErrValueRequired)
	assert.Nil(t, consumer)
}

func TestNew_NoGroupName(t *testing.T) {
	logger := slog.New(new(mockLogger))
	storage := new(mockStorage)

	consumer, err := kafkaconsumergroup.New(
		kafkaconsumergroup.WithLogger(logger),
		kafkaconsumergroup.WithStorage(storage),
	)

	assert.ErrorIs(t, err, cerrors.ErrValueRequired)
	assert.Nil(t, consumer)
}

func TestNew_EmptyGroupName(t *testing.T) {
	logger := slog.New(new(mockLogger))
	storage := new(mockStorage)

	consumer, err := kafkaconsumergroup.New(
		kafkaconsumergroup.WithLogger(logger),
		kafkaconsumergroup.WithStorage(storage),
		kafkaconsumergroup.WithKafkaGroupName(""),
	)

	assert.ErrorIs(t, err, cerrors.ErrValueRequired)
	assert.Nil(t, consumer)
}

func TestNew_EmptyTopic(t *testing.T) {
	logger := slog.New(new(mockLogger))
	storage := new(mockStorage)

	consumer, err := kafkaconsumergroup.New(
		kafkaconsumergroup.WithLogger(logger),
		kafkaconsumergroup.WithStorage(storage),
		kafkaconsumergroup.WithKafkaGroupName("github-group"),
	)

	assert.ErrorIs(t, err, cerrors.ErrInvalid)
	assert.Nil(t, consumer)
}

func TestNew_InvalidTopic(t *testing.T) {
	logger := slog.New(new(mockLogger))
	storage := new(mockStorage)

	consumer, err := kafkaconsumergroup.New(
		kafkaconsumergroup.WithLogger(logger),
		kafkaconsumergroup.WithStorage(storage),
		kafkaconsumergroup.WithKafkaGroupName("github-group"),
		kafkaconsumergroup.WithTopic("invalid"),
	)

	assert.ErrorIs(t, err, cerrors.ErrInvalid)
	assert.Nil(t, consumer)
}

func TestNew_InvalidBrokers(t *testing.T) {
	logger := slog.New(new(mockLogger))
	storage := new(mockStorage)

	consumer, err := kafkaconsumergroup.New(
		kafkaconsumergroup.WithLogger(logger),
		kafkaconsumergroup.WithStorage(storage),
		kafkaconsumergroup.WithKafkaGroupName("github-group"),
		kafkaconsumergroup.WithTopic("github"),
		kafkaconsumergroup.WithKafkaBrokers("invalid"),
	)

	assert.ErrorIs(t, err, cerrors.ErrInvalid)
	assert.Nil(t, consumer)
}

func TestNew_InvalidDialTimeout(t *testing.T) {
	logger := slog.New(new(mockLogger))
	storage := new(mockStorage)

	consumer, err := kafkaconsumergroup.New(
		kafkaconsumergroup.WithLogger(logger),
		kafkaconsumergroup.WithStorage(storage),
		kafkaconsumergroup.WithKafkaGroupName("github-group"),
		kafkaconsumergroup.WithTopic("github"),
		kafkaconsumergroup.WithKafkaBrokers("127.0.0.1:9094"),
		kafkaconsumergroup.WithDialTimeout(-1*time.Second),
	)

	assert.ErrorIs(t, err, cerrors.ErrInvalid)
	assert.Nil(t, consumer)
}

func TestNew_InvalidReadTimeout(t *testing.T) {
	logger := slog.New(new(mockLogger))
	storage := new(mockStorage)

	consumer, err := kafkaconsumergroup.New(
		kafkaconsumergroup.WithLogger(logger),
		kafkaconsumergroup.WithStorage(storage),
		kafkaconsumergroup.WithKafkaGroupName("github-group"),
		kafkaconsumergroup.WithTopic("github"),
		kafkaconsumergroup.WithKafkaBrokers("127.0.0.1:9094"),
		kafkaconsumergroup.WithReadTimeout(-1*time.Second),
	)

	assert.ErrorIs(t, err, cerrors.ErrInvalid)
	assert.Nil(t, consumer)
}

func TestNew_InvalidWriteTimeout(t *testing.T) {
	logger := slog.New(new(mockLogger))
	storage := new(mockStorage)

	consumer, err := kafkaconsumergroup.New(
		kafkaconsumergroup.WithLogger(logger),
		kafkaconsumergroup.WithStorage(storage),
		kafkaconsumergroup.WithKafkaGroupName("github-group"),
		kafkaconsumergroup.WithTopic("github"),
		kafkaconsumergroup.WithKafkaBrokers("127.0.0.1:9094"),
		kafkaconsumergroup.WithWriteTimeout(-1*time.Second),
	)

	assert.ErrorIs(t, err, cerrors.ErrInvalid)
	assert.Nil(t, consumer)
}

func TestNew_ZeroBackoff(t *testing.T) {
	logger := slog.New(new(mockLogger))
	storage := new(mockStorage)

	consumer, err := kafkaconsumergroup.New(
		kafkaconsumergroup.WithLogger(logger),
		kafkaconsumergroup.WithStorage(storage),
		kafkaconsumergroup.WithKafkaGroupName("github-group"),
		kafkaconsumergroup.WithTopic("github"),
		kafkaconsumergroup.WithKafkaBrokers("127.0.0.1:9094"),
		kafkaconsumergroup.WithBackoff(0),
	)

	assert.ErrorIs(t, err, cerrors.ErrValueRequired)
	assert.Nil(t, consumer)
}

func TestNew_InvalidBackoff(t *testing.T) {
	logger := slog.New(new(mockLogger))
	storage := new(mockStorage)

	consumer, err := kafkaconsumergroup.New(
		kafkaconsumergroup.WithLogger(logger),
		kafkaconsumergroup.WithStorage(storage),
		kafkaconsumergroup.WithKafkaGroupName("github-group"),
		kafkaconsumergroup.WithTopic("github"),
		kafkaconsumergroup.WithKafkaBrokers("127.0.0.1:9094"),
		kafkaconsumergroup.WithBackoff(2*time.Minute),
	)

	assert.ErrorIs(t, err, cerrors.ErrInvalid)
	assert.Nil(t, consumer)
}

func TestNew_InvalidMaxRetries(t *testing.T) {
	logger := slog.New(new(mockLogger))
	storage := new(mockStorage)

	consumer, err := kafkaconsumergroup.New(
		kafkaconsumergroup.WithLogger(logger),
		kafkaconsumergroup.WithStorage(storage),
		kafkaconsumergroup.WithKafkaGroupName("github-group"),
		kafkaconsumergroup.WithTopic("github"),
		kafkaconsumergroup.WithKafkaBrokers("127.0.0.1:9094"),
		kafkaconsumergroup.WithMaxRetries(256),
	)

	assert.ErrorIs(t, err, cerrors.ErrInvalid)
	assert.Nil(t, consumer)
}

func TestNew_InvalidKafkaVersion(t *testing.T) {
	logger := slog.New(new(mockLogger))
	storage := new(mockStorage)

	consumer, err := kafkaconsumergroup.New(
		kafkaconsumergroup.WithLogger(logger),
		kafkaconsumergroup.WithStorage(storage),
		kafkaconsumergroup.WithKafkaGroupName("github-group"),
		kafkaconsumergroup.WithTopic("github"),
		kafkaconsumergroup.WithKafkaBrokers("127.0.0.1:9094"),
		kafkaconsumergroup.WithKafkaVersion("1111"),
	)

	assert.ErrorIs(t, err, cerrors.ErrInvalid)
	assert.Nil(t, consumer)
}

func TestNew_NilSaramaConsumerGroupFactoryFunc(t *testing.T) {
	logger := slog.New(new(mockLogger))
	storage := new(mockStorage)

	consumer, err := kafkaconsumergroup.New(
		kafkaconsumergroup.WithLogger(logger),
		kafkaconsumergroup.WithStorage(storage),
		kafkaconsumergroup.WithKafkaGroupName("github-group"),
		kafkaconsumergroup.WithTopic("github"),
		kafkaconsumergroup.WithKafkaBrokers("127.0.0.1:9094"),
		kafkaconsumergroup.WithKafkaVersion("3.9.0"),
		kafkaconsumergroup.WithDialTimeout(5*time.Second),
		kafkaconsumergroup.WithReadTimeout(5*time.Second),
		kafkaconsumergroup.WithWriteTimeout(5*time.Second),
		kafkaconsumergroup.WithBackoff(1*time.Second),
		kafkaconsumergroup.WithMaxRetries(2),
		kafkaconsumergroup.WithSaramaConsumerGroupFactoryFunc(nil),
	)

	assert.ErrorIs(t, err, cerrors.ErrValueRequired)
	assert.Nil(t, consumer)
}

// type mockConsumerGroup struct {
// 	mock.Mock
// }
//
// func (m *mockConsumerGroup) NewConsumerGroup(
// 	brokers []string,
// 	groupName string,
// 	config *sarama.Config,
// ) (sarama.ConsumerGroup, error) {
// 	args := m.Called(brokers, groupName, config)
// 	return args.Get(0).(sarama.ConsumerGroup), args.Error(1)
// }

func TestNew_NilSaramaConsumerGroupFactoryFunc_Error(t *testing.T) {
	logger := slog.New(new(mockLogger))
	storage := new(mockStorage)

	consumer, err := kafkaconsumergroup.New(
		kafkaconsumergroup.WithLogger(logger),
		kafkaconsumergroup.WithStorage(storage),
		kafkaconsumergroup.WithKafkaGroupName("github-group"),
		kafkaconsumergroup.WithTopic("github"),
		kafkaconsumergroup.WithBackoff(100*time.Millisecond),
		kafkaconsumergroup.WithMaxRetries(1),
	)

	assert.Nil(t, consumer)
	assert.Error(t, err)
}

type mockConsumerGroup struct {
	mock.Mock
}

func (m *mockConsumerGroup) Consume(ctx context.Context, topics []string, handler sarama.ConsumerGroupHandler) error {
	args := m.Called(ctx, topics, handler)
	return args.Error(0)
}

func (m *mockConsumerGroup) Errors() <-chan error {
	args := m.Called()
	return args.Get(0).(<-chan error)
}

func (m *mockConsumerGroup) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *mockConsumerGroup) Pause(partitions map[string][]int32) {
	m.Called(partitions)
}

func (m *mockConsumerGroup) Resume(partitions map[string][]int32) {
	m.Called(partitions)
}

func (m *mockConsumerGroup) PauseAll() {
	m.Called()
}

func (m *mockConsumerGroup) ResumeAll() {
	m.Called()
}

type mockConsumerGroupFactory struct {
	mock.Mock
}

func (m *mockConsumerGroupFactory) CreateConsumerGroup(
	brokers []string,
	groupName string,
	config *sarama.Config,
) (sarama.ConsumerGroup, error) {
	args := m.Called(brokers, groupName, config)
	return args.Get(0).(sarama.ConsumerGroup), args.Error(1)
}

func TestNew_NilSaramaConsumerGroupFactoryFunc_Success(t *testing.T) {
	logger := slog.New(new(mockLogger))
	storage := new(mockStorage)

	consumerGroup := &mockConsumerGroup{}
	consumerGroupFactory := &mockConsumerGroupFactory{}
	consumerGroupFactory.On(
		"CreateConsumerGroup",
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(consumerGroup, nil).Once()

	consumer, err := kafkaconsumergroup.New(
		kafkaconsumergroup.WithLogger(logger),
		kafkaconsumergroup.WithStorage(storage),
		kafkaconsumergroup.WithKafkaGroupName("github-group"),
		kafkaconsumergroup.WithTopic("github"),
		kafkaconsumergroup.WithBackoff(100*time.Millisecond),
		kafkaconsumergroup.WithMaxRetries(1),
		kafkaconsumergroup.WithSaramaConsumerGroupFactoryFunc(consumerGroupFactory.CreateConsumerGroup),
	)

	assert.NotNil(t, consumer)
	assert.NoError(t, err)
	consumerGroupFactory.AssertNumberOfCalls(t, "CreateConsumerGroup", 1)
	consumerGroupFactory.AssertExpectations(t)
}

// func TestNew(t *testing.T) {
// mockConsumerGroup.AssertNumberOfCalls(t, "NewConsumerGroup", 1)
// mockConsumerGroup.AssertExpectations(t)

// 	// kversion, err := sarama.ParseKafkaVersion("V2_0_0_0")
// 	// fmt.Println("err", err)
// 	// fmt.Println("kversion", kversion)
// 	fmt.Println(sarama.V3_9_0_0)
// 	cg, err := kafkaconsumergroup.New()
// 	assert.Nil(t, err)
// 	assert.NotNil(t, cg)
// }

// func TestNew_Version(t *testing.T) {
// 	kversion, err := sarama.ParseKafkaVersion("3.9.0")
// 	fmt.Println("kversion", kversion)
// 	fmt.Printf("%T\n", kversion)
// 	fmt.Println("err", err)
// }
