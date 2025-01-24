package githubwebhookhandler_test

import (
	"context"
	"log/slog"
	"testing"

	"github.com/IBM/sarama"
	"github.com/devchain-network/cauldron/internal/cerrors"
	"github.com/devchain-network/cauldron/internal/kafkacp"
	"github.com/devchain-network/cauldron/internal/transport/http/githubwebhookhandler"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
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

func TestNew_NoLogger(t *testing.T) {
	handler, err := githubwebhookhandler.New()

	assert.ErrorIs(t, err, cerrors.ErrValueRequired)
	assert.Nil(t, handler)
}

func TestNew_NilLogger(t *testing.T) {
	handler, err := githubwebhookhandler.New(
		githubwebhookhandler.WithLogger(nil),
	)

	assert.ErrorIs(t, err, cerrors.ErrValueRequired)
	assert.Nil(t, handler)
}

func TestNew_NoTopic(t *testing.T) {
	logger := slog.New(new(mockLogger))

	handler, err := githubwebhookhandler.New(
		githubwebhookhandler.WithLogger(logger),
	)

	assert.ErrorIs(t, err, cerrors.ErrInvalid)
	assert.Nil(t, handler)
}

func TestNew_InvalidTopic(t *testing.T) {
	logger := slog.New(new(mockLogger))

	handler, err := githubwebhookhandler.New(
		githubwebhookhandler.WithLogger(logger),
		githubwebhookhandler.WithTopic(kafkacp.KafkaTopicIdentifier("foo")),
	)

	assert.ErrorIs(t, err, cerrors.ErrInvalid)
	assert.Nil(t, handler)
}

func TestNew_NoSecret(t *testing.T) {
	logger := slog.New(new(mockLogger))

	handler, err := githubwebhookhandler.New(
		githubwebhookhandler.WithLogger(logger),
		githubwebhookhandler.WithTopic(kafkacp.KafkaTopicIdentifier("github")),
	)

	assert.ErrorIs(t, err, cerrors.ErrValueRequired)
	assert.Nil(t, handler)
}

func TestNew_EmptySecret(t *testing.T) {
	logger := slog.New(new(mockLogger))

	handler, err := githubwebhookhandler.New(
		githubwebhookhandler.WithLogger(logger),
		githubwebhookhandler.WithTopic(kafkacp.KafkaTopicIdentifier("github")),
		githubwebhookhandler.WithWebhookSecret(""),
	)

	assert.ErrorIs(t, err, cerrors.ErrValueRequired)
	assert.Nil(t, handler)
}

func TestNew_NoMessageQueue(t *testing.T) {
	logger := slog.New(new(mockLogger))

	handler, err := githubwebhookhandler.New(
		githubwebhookhandler.WithLogger(logger),
		githubwebhookhandler.WithTopic(kafkacp.KafkaTopicIdentifier("github")),
		githubwebhookhandler.WithWebhookSecret("my-secret"),
	)

	assert.ErrorIs(t, err, cerrors.ErrValueRequired)
	assert.Nil(t, handler)
}

func TestNew_NilMessageQueue(t *testing.T) {
	logger := slog.New(new(mockLogger))

	handler, err := githubwebhookhandler.New(
		githubwebhookhandler.WithLogger(logger),
		githubwebhookhandler.WithTopic(kafkacp.KafkaTopicIdentifier("github")),
		githubwebhookhandler.WithWebhookSecret("my-secret"),
		githubwebhookhandler.WithProducerGitHubMessageQueue(nil),
	)

	assert.ErrorIs(t, err, cerrors.ErrValueRequired)
	assert.Nil(t, handler)
}

func TestNew_Success(t *testing.T) {
	logger := slog.New(new(mockLogger))
	messageQueue := make(chan *sarama.ProducerMessage, 10)

	handler, err := githubwebhookhandler.New(
		githubwebhookhandler.WithLogger(logger),
		githubwebhookhandler.WithTopic(kafkacp.KafkaTopicIdentifier("github")),
		githubwebhookhandler.WithWebhookSecret("my-secret"),
		githubwebhookhandler.WithProducerGitHubMessageQueue(messageQueue),
	)

	assert.NoError(t, err)
	assert.NotNil(t, handler)
}

func TestHandle_NoBody(t *testing.T) {
	logger := slog.New(new(mockLogger))
	messageQueue := make(chan *sarama.ProducerMessage, 10)

	handler, err := githubwebhookhandler.New(
		githubwebhookhandler.WithLogger(logger),
		githubwebhookhandler.WithTopic(kafkacp.KafkaTopicIdentifier("github")),
		githubwebhookhandler.WithWebhookSecret("my-secret"),
		githubwebhookhandler.WithProducerGitHubMessageQueue(messageQueue),
	)

	assert.NoError(t, err)
	assert.NotNil(t, handler)

	ctx := &fasthttp.RequestCtx{}
	handler.Handle(ctx)

	assert.Equal(t, fasthttp.StatusBadRequest, ctx.Response.StatusCode())
}
