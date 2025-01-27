package kafkaconsumergroup_test

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/IBM/sarama"
	"github.com/devchain-network/cauldron/internal/cerrors"
	"github.com/devchain-network/cauldron/internal/kafkacp"
	"github.com/devchain-network/cauldron/internal/kafkacp/kafkaconsumergroup"
	"github.com/devchain-network/cauldron/internal/slogger/mockslogger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var mockProcessMessage = func(ctx context.Context, msg *sarama.ConsumerMessage) error {
	return nil
}
var mockLog = slog.New(new(mockslogger.MockLogger))

// mockConsumerGroup ----------------------------------------------------------
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

// mockConsumerGroupFactory ---------------------------------------------------
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

func TestNew_NoProcessMessageFunc(t *testing.T) {
	logger := mockLog

	consumer, err := kafkaconsumergroup.New(
		kafkaconsumergroup.WithLogger(logger),
	)

	assert.ErrorIs(t, err, cerrors.ErrValueRequired)
	assert.Nil(t, consumer)
}

func TestNew_NilProcessMessageFunc(t *testing.T) {
	logger := mockLog

	consumer, err := kafkaconsumergroup.New(
		kafkaconsumergroup.WithLogger(logger),
		kafkaconsumergroup.WithProcessMessageFunc(nil),
	)

	assert.ErrorIs(t, err, cerrors.ErrValueRequired)
	assert.Nil(t, consumer)
}

func TestNew_NoGroupName(t *testing.T) {
	logger := mockLog

	consumer, err := kafkaconsumergroup.New(
		kafkaconsumergroup.WithLogger(logger),
		kafkaconsumergroup.WithProcessMessageFunc(mockProcessMessage),
	)

	assert.ErrorIs(t, err, cerrors.ErrValueRequired)
	assert.Nil(t, consumer)
}

func TestNew_EmptyGroupName(t *testing.T) {
	logger := mockLog

	consumer, err := kafkaconsumergroup.New(
		kafkaconsumergroup.WithLogger(logger),
		kafkaconsumergroup.WithProcessMessageFunc(mockProcessMessage),
		kafkaconsumergroup.WithKafkaGroupName(""),
	)

	assert.ErrorIs(t, err, cerrors.ErrValueRequired)
	assert.Nil(t, consumer)
}

func TestNew_NoTopic(t *testing.T) {
	logger := mockLog

	consumer, err := kafkaconsumergroup.New(
		kafkaconsumergroup.WithLogger(logger),
		kafkaconsumergroup.WithProcessMessageFunc(mockProcessMessage),
		kafkaconsumergroup.WithKafkaGroupName("github-group"),
	)

	assert.ErrorIs(t, err, cerrors.ErrInvalid)
	assert.Nil(t, consumer)
}

func TestNew_InvalidTopic(t *testing.T) {
	logger := mockLog

	consumer, err := kafkaconsumergroup.New(
		kafkaconsumergroup.WithLogger(logger),
		kafkaconsumergroup.WithProcessMessageFunc(mockProcessMessage),
		kafkaconsumergroup.WithKafkaGroupName("github-group"),
		kafkaconsumergroup.WithTopic("invalid"),
	)

	assert.ErrorIs(t, err, cerrors.ErrInvalid)
	assert.Nil(t, consumer)
}

func TestNew_InvalidBrokers(t *testing.T) {
	logger := mockLog

	consumer, err := kafkaconsumergroup.New(
		kafkaconsumergroup.WithLogger(logger),
		kafkaconsumergroup.WithProcessMessageFunc(mockProcessMessage),
		kafkaconsumergroup.WithKafkaGroupName("github-group"),
		kafkaconsumergroup.WithTopic(kafkacp.KafkaTopicIdentifierGitHub.String()),
		kafkaconsumergroup.WithKafkaBrokers("invalid"),
	)

	assert.ErrorIs(t, err, cerrors.ErrInvalid)
	assert.Nil(t, consumer)
}

func TestNew_InvalidDialTimeout(t *testing.T) {
	logger := mockLog

	consumer, err := kafkaconsumergroup.New(
		kafkaconsumergroup.WithLogger(logger),
		kafkaconsumergroup.WithProcessMessageFunc(mockProcessMessage),
		kafkaconsumergroup.WithKafkaGroupName("github-group"),
		kafkaconsumergroup.WithTopic(kafkacp.KafkaTopicIdentifierGitHub.String()),
		kafkaconsumergroup.WithKafkaBrokers(kafkacp.DefaultKafkaBrokers),
		kafkaconsumergroup.WithDialTimeout(-1*time.Second),
	)

	assert.ErrorIs(t, err, cerrors.ErrInvalid)
	assert.Nil(t, consumer)
}

func TestNew_InvalidReadTimeout(t *testing.T) {
	logger := mockLog

	consumer, err := kafkaconsumergroup.New(
		kafkaconsumergroup.WithLogger(logger),
		kafkaconsumergroup.WithProcessMessageFunc(mockProcessMessage),
		kafkaconsumergroup.WithKafkaGroupName("github-group"),
		kafkaconsumergroup.WithTopic(kafkacp.KafkaTopicIdentifierGitHub.String()),
		kafkaconsumergroup.WithKafkaBrokers(kafkacp.DefaultKafkaBrokers),
		kafkaconsumergroup.WithReadTimeout(-1*time.Second),
	)

	assert.ErrorIs(t, err, cerrors.ErrInvalid)
	assert.Nil(t, consumer)
}

func TestNew_InvalidWriteTimeout(t *testing.T) {
	logger := mockLog

	consumer, err := kafkaconsumergroup.New(
		kafkaconsumergroup.WithLogger(logger),
		kafkaconsumergroup.WithProcessMessageFunc(mockProcessMessage),
		kafkaconsumergroup.WithKafkaGroupName("github-group"),
		kafkaconsumergroup.WithTopic(kafkacp.KafkaTopicIdentifierGitHub.String()),
		kafkaconsumergroup.WithKafkaBrokers(kafkacp.DefaultKafkaBrokers),
		kafkaconsumergroup.WithWriteTimeout(-1*time.Second),
	)

	assert.ErrorIs(t, err, cerrors.ErrInvalid)
	assert.Nil(t, consumer)
}

func TestNew_ZeroBackoff(t *testing.T) {
	logger := mockLog

	consumer, err := kafkaconsumergroup.New(
		kafkaconsumergroup.WithLogger(logger),
		kafkaconsumergroup.WithProcessMessageFunc(mockProcessMessage),
		kafkaconsumergroup.WithKafkaGroupName("github-group"),
		kafkaconsumergroup.WithTopic(kafkacp.KafkaTopicIdentifierGitHub.String()),
		kafkaconsumergroup.WithKafkaBrokers(kafkacp.DefaultKafkaBrokers),
		kafkaconsumergroup.WithBackoff(0),
	)

	assert.ErrorIs(t, err, cerrors.ErrValueRequired)
	assert.Nil(t, consumer)
}

func TestNew_InvalidBackoff(t *testing.T) {
	logger := mockLog

	consumer, err := kafkaconsumergroup.New(
		kafkaconsumergroup.WithLogger(logger),
		kafkaconsumergroup.WithProcessMessageFunc(mockProcessMessage),
		kafkaconsumergroup.WithKafkaGroupName("github-group"),
		kafkaconsumergroup.WithTopic(kafkacp.KafkaTopicIdentifierGitHub.String()),
		kafkaconsumergroup.WithKafkaBrokers(kafkacp.DefaultKafkaBrokers),
		kafkaconsumergroup.WithBackoff(2*time.Minute),
	)

	assert.ErrorIs(t, err, cerrors.ErrInvalid)
	assert.Nil(t, consumer)
}

func TestNew_InvalidMaxRetries(t *testing.T) {
	logger := mockLog

	consumer, err := kafkaconsumergroup.New(
		kafkaconsumergroup.WithLogger(logger),
		kafkaconsumergroup.WithProcessMessageFunc(mockProcessMessage),
		kafkaconsumergroup.WithKafkaGroupName("github-group"),
		kafkaconsumergroup.WithTopic(kafkacp.KafkaTopicIdentifierGitHub.String()),
		kafkaconsumergroup.WithKafkaBrokers(kafkacp.DefaultKafkaBrokers),
		kafkaconsumergroup.WithMaxRetries(256),
	)

	assert.ErrorIs(t, err, cerrors.ErrInvalid)
	assert.Nil(t, consumer)
}

func TestNew_InvalidKafkaVersion(t *testing.T) {
	logger := mockLog

	consumer, err := kafkaconsumergroup.New(
		kafkaconsumergroup.WithLogger(logger),
		kafkaconsumergroup.WithProcessMessageFunc(mockProcessMessage),
		kafkaconsumergroup.WithKafkaGroupName("github-group"),
		kafkaconsumergroup.WithTopic(kafkacp.KafkaTopicIdentifierGitHub.String()),
		kafkaconsumergroup.WithKafkaBrokers(kafkacp.DefaultKafkaBrokers),
		kafkaconsumergroup.WithKafkaVersion("1111"),
	)

	assert.ErrorIs(t, err, cerrors.ErrInvalid)
	assert.Nil(t, consumer)
}

func TestNew_NilSaramaConsumerGroupHandler(t *testing.T) {
	logger := mockLog

	consumer, err := kafkaconsumergroup.New(
		kafkaconsumergroup.WithLogger(logger),
		kafkaconsumergroup.WithProcessMessageFunc(mockProcessMessage),
		kafkaconsumergroup.WithKafkaGroupName("github-group"),
		kafkaconsumergroup.WithTopic(kafkacp.KafkaTopicIdentifierGitHub.String()),
		kafkaconsumergroup.WithKafkaBrokers(kafkacp.DefaultKafkaBrokers),
		kafkaconsumergroup.WithSaramaConsumerGroupHandler(nil),
	)

	assert.ErrorIs(t, err, cerrors.ErrValueRequired)
	assert.Nil(t, consumer)
}

func TestNew_NilSaramaConsumerGroupFactoryFunc(t *testing.T) {
	logger := mockLog

	consumer, err := kafkaconsumergroup.New(
		kafkaconsumergroup.WithLogger(logger),
		kafkaconsumergroup.WithProcessMessageFunc(mockProcessMessage),
		kafkaconsumergroup.WithKafkaGroupName("github-group"),
		kafkaconsumergroup.WithTopic(kafkacp.KafkaTopicIdentifierGitHub.String()),
		kafkaconsumergroup.WithKafkaBrokers(kafkacp.DefaultKafkaBrokers),
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

func TestNew_SaramaConsumerGroupFactoryFunc_Error(t *testing.T) {
	logger := mockLog

	consumerGroup := &mockConsumerGroup{}
	consumerGroupFactory := &mockConsumerGroupFactory{}
	consumerGroupFactory.On(
		"CreateConsumerGroup",
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(consumerGroup, errors.New("error")).Once()

	consumer, err := kafkaconsumergroup.New(
		kafkaconsumergroup.WithLogger(logger),
		kafkaconsumergroup.WithProcessMessageFunc(mockProcessMessage),
		kafkaconsumergroup.WithKafkaGroupName("github-group"),
		kafkaconsumergroup.WithTopic(kafkacp.KafkaTopicIdentifierGitHub.String()),
		kafkaconsumergroup.WithBackoff(100*time.Millisecond),
		kafkaconsumergroup.WithMaxRetries(1),
		kafkaconsumergroup.WithSaramaConsumerGroupFactoryFunc(consumerGroupFactory.CreateConsumerGroup),
	)

	assert.Nil(t, consumer)
	assert.Error(t, err)
	consumerGroupFactory.AssertNumberOfCalls(t, "CreateConsumerGroup", 1)
	consumerGroupFactory.AssertExpectations(t)
}

func TestNew_SaramaConsumerGroupFactoryFunc_Success(t *testing.T) {
	logger := mockLog

	consumerGroup := &mockConsumerGroup{}
	consumerGroupFactory := &mockConsumerGroupFactory{}
	consumerGroupFactory.On("CreateConsumerGroup",
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(consumerGroup, nil).Once()

	consumer, err := kafkaconsumergroup.New(
		kafkaconsumergroup.WithLogger(logger),
		kafkaconsumergroup.WithProcessMessageFunc(mockProcessMessage),
		kafkaconsumergroup.WithKafkaGroupName("github-group"),
		kafkaconsumergroup.WithTopic(kafkacp.KafkaTopicIdentifierGitHub.String()),
		kafkaconsumergroup.WithBackoff(100*time.Millisecond),
		kafkaconsumergroup.WithMaxRetries(1),
		kafkaconsumergroup.WithSaramaConsumerGroupFactoryFunc(consumerGroupFactory.CreateConsumerGroup),
	)

	assert.NotNil(t, consumer)
	assert.NoError(t, err)
	consumerGroupFactory.AssertNumberOfCalls(t, "CreateConsumerGroup", 1)
	consumerGroupFactory.AssertExpectations(t)
}

func TestNew_Consume_Success(t *testing.T) {
	logger := mockLog

	consumerGroup := &mockConsumerGroup{}
	consumerGroup.On("Errors").Return((<-chan error)(make(chan error)))

	consumerGroupFactory := &mockConsumerGroupFactory{}
	consumerGroupFactory.On("CreateConsumerGroup",
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(consumerGroup, nil)

	consumerGroup.On("Consume", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	consumerGroup.On("Close").Return(nil).Once()

	consumer, err := kafkaconsumergroup.New(
		kafkaconsumergroup.WithLogger(logger),
		kafkaconsumergroup.WithProcessMessageFunc(mockProcessMessage),
		kafkaconsumergroup.WithKafkaGroupName("github-group"),
		kafkaconsumergroup.WithTopic(kafkacp.KafkaTopicIdentifierGitHub.String()),
		kafkaconsumergroup.WithBackoff(100*time.Millisecond),
		kafkaconsumergroup.WithMaxRetries(1),
		kafkaconsumergroup.WithSaramaConsumerGroupFactoryFunc(consumerGroupFactory.CreateConsumerGroup),
	)

	assert.NotNil(t, consumer)
	assert.NoError(t, err)

	consumerGroupFactory.AssertNumberOfCalls(t, "CreateConsumerGroup", 1)
	consumerGroupFactory.AssertExpectations(t)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		err := consumer.StartConsume()
		assert.NoError(t, err)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		consumer.MessageQueue <- &sarama.ConsumerMessage{
			Topic:     kafkacp.KafkaTopicIdentifierGitHub.String(),
			Partition: 0,
			Offset:    1,
			Key:       []byte("key"),
			Value:     []byte("value"),
		}
	}()

	time.Sleep(100 * time.Millisecond)
	process, _ := os.FindProcess(syscall.Getpid())
	_ = process.Signal(os.Interrupt)
	wg.Wait()
}
