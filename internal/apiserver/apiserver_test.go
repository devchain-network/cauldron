package apiserver_test

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/devchain-network/cauldron/internal/apiserver"
	"github.com/devchain-network/cauldron/internal/cerrors"
	"github.com/devchain-network/cauldron/internal/kafkacp"
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

func TestNew_NoParams(t *testing.T) {
	server, err := apiserver.New()

	assert.ErrorIs(t, err, cerrors.ErrValueRequired)
	assert.Nil(t, server)
}

func TestNew_NilLogger(t *testing.T) {
	server, err := apiserver.New(
		apiserver.WithLogger(nil),
	)

	assert.ErrorIs(t, err, cerrors.ErrValueRequired)
	assert.Nil(t, server)
}

func TestNew_EmptyListenAddr(t *testing.T) {
	logger := slog.New(new(mockLogger))

	server, err := apiserver.New(
		apiserver.WithLogger(logger),
		apiserver.WithListenAddr(""),
	)

	assert.ErrorIs(t, err, cerrors.ErrValueRequired)
	assert.Nil(t, server)
}

func TestNew_InvalidListenAddr(t *testing.T) {
	logger := slog.New(new(mockLogger))

	server, err := apiserver.New(
		apiserver.WithLogger(logger),
		apiserver.WithListenAddr("invalid"),
	)

	assert.ErrorIs(t, err, cerrors.ErrInvalid)
	assert.Nil(t, server)
}

func TestNew_InvalidKafkaTopic(t *testing.T) {
	logger := slog.New(new(mockLogger))

	server, err := apiserver.New(
		apiserver.WithLogger(logger),
		apiserver.WithKafkaGitHubTopic(kafkacp.KafkaTopicIdentifier("foo")),
	)

	assert.ErrorIs(t, err, cerrors.ErrInvalid)
	assert.Nil(t, server)
}

func TestNew_InvalidBrokers(t *testing.T) {
	logger := slog.New(new(mockLogger))

	var kafkaBrokers kafkacp.KafkaBrokers
	kafkaBrokers.AddFromString("foo")

	server, err := apiserver.New(
		apiserver.WithLogger(logger),
		apiserver.WithKafkaBrokers(kafkaBrokers),
	)

	assert.ErrorIs(t, err, cerrors.ErrInvalid)
	assert.Nil(t, server)
}

func TestNew_InvalidReadTimeout(t *testing.T) {
	logger := slog.New(new(mockLogger))

	server, err := apiserver.New(
		apiserver.WithLogger(logger),
		apiserver.WithReadTimeout(-1*time.Second),
	)

	assert.ErrorIs(t, err, cerrors.ErrInvalid)
	assert.Nil(t, server)
}

func TestNew_InvalidWriteTimeout(t *testing.T) {
	logger := slog.New(new(mockLogger))

	server, err := apiserver.New(
		apiserver.WithLogger(logger),
		apiserver.WithWriteTimeout(-1*time.Second),
	)

	assert.ErrorIs(t, err, cerrors.ErrInvalid)
	assert.Nil(t, server)
}

func TestNew_InvalidIdleTimeout(t *testing.T) {
	logger := slog.New(new(mockLogger))

	server, err := apiserver.New(
		apiserver.WithLogger(logger),
		apiserver.WithIdleTimeout(-1*time.Second),
	)

	assert.ErrorIs(t, err, cerrors.ErrInvalid)
	assert.Nil(t, server)
}

func TestNew_NilHTTPHandler(t *testing.T) {
	logger := slog.New(new(mockLogger))

	var kafkaBrokers kafkacp.KafkaBrokers
	kafkaBrokers.AddFromString("localhost:9194")

	server, err := apiserver.New(
		apiserver.WithLogger(logger),
		apiserver.WithListenAddr(":9000"),
		apiserver.WithReadTimeout(5*time.Second),
		apiserver.WithWriteTimeout(5*time.Second),
		apiserver.WithIdleTimeout(5*time.Second),
		apiserver.WithKafkaBrokers(kafkaBrokers),
		apiserver.WithKafkaGitHubTopic(kafkacp.KafkaTopicIdentifierGitHub),
	)

	assert.ErrorIs(t, err, cerrors.ErrValueRequired)
	assert.Nil(t, server)
}

func TestNew_InvalidKafkaTopic_check(t *testing.T) {
	logger := slog.New(new(mockLogger))

	var kafkaBrokers kafkacp.KafkaBrokers
	kafkaBrokers.AddFromString("localhost:9194")

	server, err := apiserver.New(
		apiserver.WithLogger(logger),
		apiserver.WithListenAddr(":9000"),
		apiserver.WithReadTimeout(5*time.Second),
		apiserver.WithWriteTimeout(5*time.Second),
		apiserver.WithIdleTimeout(5*time.Second),
		apiserver.WithKafkaBrokers(kafkaBrokers),
		apiserver.WithHTTPHandler(
			fasthttp.MethodGet,
			"/test",
			func(ctx *fasthttp.RequestCtx) { ctx.SetStatusCode(fasthttp.StatusOK) },
		),
	)

	assert.ErrorIs(t, err, cerrors.ErrInvalid)
	assert.Nil(t, server)
}

func TestNew_MissingArgsHTTPHandler_method(t *testing.T) {
	logger := slog.New(new(mockLogger))

	server, err := apiserver.New(
		apiserver.WithLogger(logger),
		apiserver.WithListenAddr(":9000"),
		apiserver.WithReadTimeout(5*time.Second),
		apiserver.WithWriteTimeout(5*time.Second),
		apiserver.WithIdleTimeout(5*time.Second),
		apiserver.WithKafkaGitHubTopic(kafkacp.KafkaTopicIdentifierGitHub),
		apiserver.WithHTTPHandler(
			"",
			"/test",
			func(ctx *fasthttp.RequestCtx) { ctx.SetStatusCode(fasthttp.StatusOK) },
		),
	)

	assert.ErrorIs(t, err, cerrors.ErrValueRequired)
	assert.Nil(t, server)
}

func TestNew_MissingArgsHTTPHandler_path(t *testing.T) {
	logger := slog.New(new(mockLogger))

	server, err := apiserver.New(
		apiserver.WithLogger(logger),
		apiserver.WithListenAddr(":9000"),
		apiserver.WithReadTimeout(5*time.Second),
		apiserver.WithWriteTimeout(5*time.Second),
		apiserver.WithIdleTimeout(5*time.Second),
		apiserver.WithKafkaGitHubTopic(kafkacp.KafkaTopicIdentifierGitHub),
		apiserver.WithHTTPHandler(
			fasthttp.MethodGet,
			"",
			func(ctx *fasthttp.RequestCtx) { ctx.SetStatusCode(fasthttp.StatusOK) },
		),
	)

	assert.ErrorIs(t, err, cerrors.ErrValueRequired)
	assert.Nil(t, server)
}

func TestNew_MissingArgsHTTPHandler_handler(t *testing.T) {
	logger := slog.New(new(mockLogger))

	server, err := apiserver.New(
		apiserver.WithLogger(logger),
		apiserver.WithListenAddr(":9000"),
		apiserver.WithReadTimeout(5*time.Second),
		apiserver.WithWriteTimeout(5*time.Second),
		apiserver.WithIdleTimeout(5*time.Second),
		apiserver.WithKafkaGitHubTopic(kafkacp.KafkaTopicIdentifierGitHub),
		apiserver.WithHTTPHandler(
			fasthttp.MethodGet,
			"/test",
			nil,
		),
	)

	assert.ErrorIs(t, err, cerrors.ErrValueRequired)
	assert.Nil(t, server)
}

func TestHttpRouter_NotFound(t *testing.T) {
	logger := slog.New(new(mockLogger))

	server, err := apiserver.New(
		apiserver.WithLogger(logger),
		apiserver.WithKafkaGitHubTopic(kafkacp.KafkaTopicIdentifierGitHub),
		apiserver.WithHTTPHandler(
			fasthttp.MethodGet,
			"/existing-path",
			func(ctx *fasthttp.RequestCtx) { ctx.SetStatusCode(fasthttp.StatusOK) },
		),
	)
	assert.NoError(t, err)

	ctx := &fasthttp.RequestCtx{}
	ctx.Request.SetRequestURI("/non-existent-path")
	ctx.Request.Header.SetMethod(fasthttp.MethodGet)

	server.FastHTTP.Handler(ctx)

	assert.Equal(t, fasthttp.StatusNotFound, ctx.Response.StatusCode())
}

func TestHttpRouter_MethodNotAllowed(t *testing.T) {
	logger := slog.New(new(mockLogger))
	server, err := apiserver.New(
		apiserver.WithLogger(logger),
		apiserver.WithKafkaGitHubTopic(kafkacp.KafkaTopicIdentifierGitHub),
		apiserver.WithHTTPHandler(
			fasthttp.MethodGet,
			"/existing-path",
			func(ctx *fasthttp.RequestCtx) { ctx.SetStatusCode(fasthttp.StatusOK) },
		),
	)
	assert.NoError(t, err)

	ctx := &fasthttp.RequestCtx{}
	ctx.Request.SetRequestURI("/existing-path")
	ctx.Request.Header.SetMethod(fasthttp.MethodPost)

	server.FastHTTP.Handler(ctx)

	assert.Equal(t, fasthttp.StatusMethodNotAllowed, ctx.Response.StatusCode())
}

func TestHttpRouter_ValidRouteAndMethod(t *testing.T) {
	logger := slog.New(new(mockLogger))
	server, err := apiserver.New(
		apiserver.WithLogger(logger),
		apiserver.WithKafkaGitHubTopic(kafkacp.KafkaTopicIdentifierGitHub),
		apiserver.WithHTTPHandler(
			fasthttp.MethodGet,
			"/existing-path",
			func(ctx *fasthttp.RequestCtx) {
				ctx.SetStatusCode(fasthttp.StatusOK)
				ctx.SetBody([]byte("success"))
			},
		),
	)
	assert.NoError(t, err)

	ctx := &fasthttp.RequestCtx{}
	ctx.Request.SetRequestURI("/existing-path")
	ctx.Request.Header.SetMethod(fasthttp.MethodGet)

	server.FastHTTP.Handler(ctx)

	assert.Equal(t, fasthttp.StatusOK, ctx.Response.StatusCode())
	assert.Equal(t, "success", string(ctx.Response.Body()))
}
