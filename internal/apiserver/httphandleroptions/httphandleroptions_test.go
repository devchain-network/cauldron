package httphandleroptions_test

import (
	"log/slog"
	"testing"

	"github.com/IBM/sarama"
	"github.com/IBM/sarama/mocks"
	"github.com/devchain-network/cauldron/internal/apiserver/httphandleroptions"
	"github.com/devchain-network/cauldron/internal/cerrors"
	"github.com/stretchr/testify/assert"
)

func getLogger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(nil, nil))
}

func TestNew_ErrorNilLogger(t *testing.T) {
	handler, err := httphandleroptions.New(
		httphandleroptions.WithKafkaProducer(&mocks.AsyncProducer{}),
		httphandleroptions.WithProducerMessageQueue(make(chan *sarama.ProducerMessage, 100)),
	)
	assert.Error(t, err)
	assert.Nil(t, handler)
	assert.ErrorIs(t, err, cerrors.ErrValueRequired)
}

func TestNew_ErrorNilKafkaProducer(t *testing.T) {
	handler, err := httphandleroptions.New(
		httphandleroptions.WithLogger(getLogger()),
		httphandleroptions.WithProducerMessageQueue(make(chan *sarama.ProducerMessage, 100)),
	)
	assert.Error(t, err)
	assert.Nil(t, handler)
	assert.ErrorIs(t, err, cerrors.ErrValueRequired)
}

func TestNew_ErrorNilMessageQueue(t *testing.T) {
	handler, err := httphandleroptions.New(
		httphandleroptions.WithLogger(getLogger()),
		httphandleroptions.WithKafkaProducer(&mocks.AsyncProducer{}),
	)
	assert.Error(t, err)
	assert.Nil(t, handler)
	assert.ErrorIs(t, err, cerrors.ErrValueRequired)
}

func TestNew_Success(t *testing.T) {
	messageQueue := make(chan *sarama.ProducerMessage, 100)
	mockProducer := mocks.NewAsyncProducer(t, nil)

	handler, err := httphandleroptions.New(
		httphandleroptions.WithLogger(getLogger()),
		httphandleroptions.WithKafkaProducer(mockProducer),
		httphandleroptions.WithProducerMessageQueue(messageQueue),
	)
	assert.NoError(t, err)
	assert.NotNil(t, handler)
	assert.Equal(t, mockProducer, handler.KafkaProducer)
	assert.Equal(t, messageQueue, handler.ProducerMessageQueue)
}
