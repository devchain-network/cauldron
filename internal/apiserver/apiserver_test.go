package apiserver_test

// import (
// 	"context"
// 	"log/slog"
// 	"testing"
// 	"time"
//
// 	"github.com/devchain-network/cauldron/internal/apiserver"
// 	"github.com/devchain-network/cauldron/internal/cerrors"
// 	"github.com/devchain-network/cauldron/internal/kafkaconsumer"
// 	"github.com/stretchr/testify/assert"
// 	"github.com/valyala/fasthttp"
// )
//
// type MockJSONLogHandler struct{}
//
// func (h *MockJSONLogHandler) Enabled(_ context.Context, _ slog.Level) bool {
// 	return true
// }
//
// func (h *MockJSONLogHandler) Handle(_ context.Context, record slog.Record) error {
// 	return nil
// }
//
// func (h *MockJSONLogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
// 	return h
// }
//
// func (h *MockJSONLogHandler) WithGroup(name string) slog.Handler {
// 	return h
// }
//
// func getLogger() *slog.Logger {
// 	return slog.New(new(MockJSONLogHandler))
// }
//
// func TestApiServer_NoLogger(t *testing.T) {
// 	server, err := apiserver.New()
//
// 	assert.Error(t, err)
// 	assert.Nil(t, server)
// 	assert.ErrorIs(t, err, cerrors.ErrValueRequired)
// }
//
// func TestApiServer_NilLogger(t *testing.T) {
// 	server, err := apiserver.New(
// 		apiserver.WithLogger(nil),
// 	)
//
// 	assert.Error(t, err)
// 	assert.Nil(t, server)
// 	assert.ErrorIs(t, err, cerrors.ErrValueRequired)
// }
//
// func TestApiServer_NilHandlers(t *testing.T) {
// 	server, err := apiserver.New(
// 		apiserver.WithLogger(getLogger()),
// 	)
//
// 	assert.Error(t, err)
// 	assert.Nil(t, server)
// 	assert.ErrorIs(t, err, cerrors.ErrValueRequired)
// }
//
// func TestApiServer_HandlersWithEmptyMethod(t *testing.T) {
// 	server, err := apiserver.New(
// 		apiserver.WithLogger(getLogger()),
// 		apiserver.WithHTTPHandler("", "/foo", nil),
// 	)
//
// 	assert.Error(t, err)
// 	assert.Nil(t, server)
// 	assert.ErrorIs(t, err, cerrors.ErrValueRequired)
// }
//
// func TestApiServer_HandlersWithEmptyPath(t *testing.T) {
// 	server, err := apiserver.New(
// 		apiserver.WithLogger(getLogger()),
// 		apiserver.WithHTTPHandler(fasthttp.MethodGet, "", nil),
// 	)
//
// 	assert.Error(t, err)
// 	assert.Nil(t, server)
// 	assert.ErrorIs(t, err, cerrors.ErrValueRequired)
// }
//
// func TestApiServer_HandlersWithNilHandler(t *testing.T) {
// 	server, err := apiserver.New(
// 		apiserver.WithLogger(getLogger()),
// 		apiserver.WithHTTPHandler(fasthttp.MethodGet, "/foo", nil),
// 	)
//
// 	assert.Error(t, err)
// 	assert.Nil(t, server)
// 	assert.ErrorIs(t, err, cerrors.ErrValueRequired)
// }
//
// func getMockFastHTTPHandlerOK() fasthttp.RequestHandler {
// 	return func(ctx *fasthttp.RequestCtx) {
// 		ctx.SetStatusCode(fasthttp.StatusOK)
// 	}
// }
//
// func getBrokers() []string {
// 	return kafkaconsumer.TCPAddrs(kafkaconsumer.DefaultKafkaBrokers).List()
// }
//
// func TestApiServer_HandlersWithDefaultListenAddr(t *testing.T) {
// 	server, err := apiserver.New(
// 		apiserver.WithLogger(getLogger()),
// 		apiserver.WithHTTPHandler(fasthttp.MethodGet, "/foo", getMockFastHTTPHandlerOK()),
// 		apiserver.WithKafkaBrokers(getBrokers()),
// 	)
//
// 	assert.NoError(t, err)
// 	assert.NotNil(t, server)
// }
//
// func TestApiServer_HandlersWithEmptyListenAddr(t *testing.T) {
// 	server, err := apiserver.New(
// 		apiserver.WithLogger(getLogger()),
// 		apiserver.WithHTTPHandler(fasthttp.MethodGet, "/foo", getMockFastHTTPHandlerOK()),
// 		apiserver.WithKafkaBrokers(getBrokers()),
// 		apiserver.WithListenAddr(""),
// 	)
//
// 	assert.Error(t, err)
// 	assert.Nil(t, server)
// 	assert.ErrorIs(t, err, cerrors.ErrValueRequired)
// }
//
// func TestApiServer_HandlersWithInvalidListenAddr(t *testing.T) {
// 	server, err := apiserver.New(
// 		apiserver.WithLogger(getLogger()),
// 		apiserver.WithHTTPHandler(fasthttp.MethodGet, "/foo", getMockFastHTTPHandlerOK()),
// 		apiserver.WithKafkaBrokers(getBrokers()),
// 		apiserver.WithListenAddr("invalid"),
// 	)
//
// 	assert.Error(t, err)
// 	assert.Nil(t, server)
// 	assert.ErrorIs(t, err, cerrors.ErrInvalid)
// }
//
// func TestApiServer_HandlersWithDefaultReadTimeout(t *testing.T) {
// 	server, err := apiserver.New(
// 		apiserver.WithLogger(getLogger()),
// 		apiserver.WithHTTPHandler(fasthttp.MethodGet, "/foo", getMockFastHTTPHandlerOK()),
// 		apiserver.WithKafkaBrokers(getBrokers()),
// 	)
//
// 	assert.NoError(t, err)
// 	assert.NotNil(t, server)
// 	assert.Equal(t, server.ReadTimeout, 5*time.Second)
// }
//
// func TestApiServer_HandlersWithDefaultWriteTimeout(t *testing.T) {
// 	server, err := apiserver.New(
// 		apiserver.WithLogger(getLogger()),
// 		apiserver.WithHTTPHandler(fasthttp.MethodGet, "/foo", getMockFastHTTPHandlerOK()),
// 		apiserver.WithKafkaBrokers(getBrokers()),
// 	)
//
// 	assert.NoError(t, err)
// 	assert.NotNil(t, server)
// 	assert.Equal(t, server.WriteTimeout, 10*time.Second)
// }
//
// func TestApiServer_HandlersWithDefaultIdleTimeout(t *testing.T) {
// 	server, err := apiserver.New(
// 		apiserver.WithLogger(getLogger()),
// 		apiserver.WithHTTPHandler(fasthttp.MethodGet, "/foo", getMockFastHTTPHandlerOK()),
// 		apiserver.WithKafkaBrokers(getBrokers()),
// 	)
//
// 	assert.NoError(t, err)
// 	assert.NotNil(t, server)
// 	assert.Equal(t, server.IdleTimeout, 15*time.Second)
// }
//
// func TestApiServer_HandlersWithReadTimeoutInvalid(t *testing.T) {
// 	server, err := apiserver.New(
// 		apiserver.WithLogger(getLogger()),
// 		apiserver.WithHTTPHandler(fasthttp.MethodGet, "/foo", getMockFastHTTPHandlerOK()),
// 		apiserver.WithReadTimeout(-1*time.Second),
// 		apiserver.WithKafkaBrokers(getBrokers()),
// 	)
//
// 	assert.Error(t, err)
// 	assert.Nil(t, server)
// 	assert.ErrorIs(t, err, cerrors.ErrInvalid)
// }
//
// func TestApiServer_HandlersWithWriteTimeoutInvalid(t *testing.T) {
// 	server, err := apiserver.New(
// 		apiserver.WithLogger(getLogger()),
// 		apiserver.WithHTTPHandler(fasthttp.MethodGet, "/foo", getMockFastHTTPHandlerOK()),
// 		apiserver.WithWriteTimeout(-1*time.Second),
// 		apiserver.WithKafkaBrokers(getBrokers()),
// 	)
//
// 	assert.Error(t, err)
// 	assert.Nil(t, server)
// 	assert.ErrorIs(t, err, cerrors.ErrInvalid)
// }
//
// func TestApiServer_HandlersWithIdleTimeoutTimeoutInvalid(t *testing.T) {
// 	server, err := apiserver.New(
// 		apiserver.WithLogger(getLogger()),
// 		apiserver.WithHTTPHandler(fasthttp.MethodGet, "/foo", getMockFastHTTPHandlerOK()),
// 		apiserver.WithIdleTimeout(-1*time.Second),
// 		apiserver.WithKafkaBrokers(getBrokers()),
// 	)
//
// 	assert.Error(t, err)
// 	assert.Nil(t, server)
// 	assert.ErrorIs(t, err, cerrors.ErrInvalid)
// }
//
// func TestApiServer_HandlersWithKafkaBrokersInvalid(t *testing.T) {
// 	server, err := apiserver.New(
// 		apiserver.WithLogger(getLogger()),
// 		apiserver.WithHTTPHandler(fasthttp.MethodGet, "/foo", getMockFastHTTPHandlerOK()),
// 		apiserver.WithKafkaBrokers([]string{"foo"}),
// 	)
//
// 	assert.Error(t, err)
// 	assert.Nil(t, server)
// 	assert.ErrorIs(t, err, cerrors.ErrInvalid)
// }
//
// func TestApiServer_HandlersWithKafkaBrokersEmpty(t *testing.T) {
// 	server, err := apiserver.New(
// 		apiserver.WithLogger(getLogger()),
// 		apiserver.WithHTTPHandler(fasthttp.MethodGet, "/foo", getMockFastHTTPHandlerOK()),
// 		apiserver.WithKafkaBrokers([]string{""}),
// 	)
//
// 	assert.Error(t, err)
// 	assert.Nil(t, server)
// 	assert.ErrorIs(t, err, cerrors.ErrValueRequired)
// }
//
// func TestApiServer_HandlersWithKafkaBrokersNil(t *testing.T) {
// 	server, err := apiserver.New(
// 		apiserver.WithLogger(getLogger()),
// 		apiserver.WithHTTPHandler(fasthttp.MethodGet, "/foo", getMockFastHTTPHandlerOK()),
// 		apiserver.WithKafkaBrokers(nil),
// 	)
//
// 	assert.Error(t, err)
// 	assert.Nil(t, server)
// 	assert.ErrorIs(t, err, cerrors.ErrValueRequired)
// }
//
// func TestApiServer_HandlersWithKafkaGitHubTopicEmpty(t *testing.T) {
// 	server, err := apiserver.New(
// 		apiserver.WithLogger(getLogger()),
// 		apiserver.WithHTTPHandler(fasthttp.MethodGet, "/foo", getMockFastHTTPHandlerOK()),
// 		apiserver.WithKafkaBrokers(getBrokers()),
// 		apiserver.WithKafkaGitHubTopic(kafkaconsumer.KafkaTopicIdentifier("")),
// 	)
//
// 	assert.Error(t, err)
// 	assert.Nil(t, server)
// 	assert.ErrorIs(t, err, cerrors.ErrValueRequired)
// }
//
// func TestApiServer_HandlersWithKafkaGitHubTopicInvalid(t *testing.T) {
// 	server, err := apiserver.New(
// 		apiserver.WithLogger(getLogger()),
// 		apiserver.WithHTTPHandler(fasthttp.MethodGet, "/foo", getMockFastHTTPHandlerOK()),
// 		apiserver.WithKafkaBrokers(getBrokers()),
// 		apiserver.WithKafkaGitHubTopic(kafkaconsumer.KafkaTopicIdentifier("foo")),
// 	)
//
// 	assert.Error(t, err)
// 	assert.Nil(t, server)
// 	assert.ErrorIs(t, err, cerrors.ErrInvalid)
// }
//
// func TestApiServer_HandlersWithKafkaGitHubTopicValid(t *testing.T) {
// 	server, err := apiserver.New(
// 		apiserver.WithLogger(getLogger()),
// 		apiserver.WithListenAddr(":9000"),
// 		apiserver.WithHTTPHandler(fasthttp.MethodGet, "/foo", getMockFastHTTPHandlerOK()),
// 		apiserver.WithKafkaBrokers(getBrokers()),
// 		apiserver.WithKafkaGitHubTopic(kafkaconsumer.KafkaTopicIdentifierGitHub),
// 		apiserver.WithReadTimeout(5*time.Second),
// 		apiserver.WithWriteTimeout(5*time.Second),
// 		apiserver.WithIdleTimeout(5*time.Second),
// 	)
//
// 	assert.NoError(t, err)
// 	assert.NotNil(t, server)
// }
//
// func TestApiServer_StatusNotFound(t *testing.T) {
// 	server, err := apiserver.New(
// 		apiserver.WithLogger(getLogger()),
// 		apiserver.WithListenAddr(":8000"),
// 		apiserver.WithHTTPHandler(fasthttp.MethodGet, "/healthz", apiserver.HealthCheckHandler),
// 		apiserver.WithKafkaBrokers(getBrokers()),
// 		apiserver.WithKafkaGitHubTopic(kafkaconsumer.KafkaTopicIdentifierGitHub),
// 	)
// 	assert.NoError(t, err)
// 	assert.NotNil(t, server)
//
// 	ctx := &fasthttp.RequestCtx{}
// 	ctx.Request.SetRequestURI("/invalid-path")
// 	ctx.Request.Header.SetMethod(fasthttp.MethodGet)
//
// 	server.FastHTTP.Handler(ctx)
// 	assert.Equal(t, fasthttp.StatusNotFound, ctx.Response.StatusCode())
// }
//
// func TestApiServer_MethodNotAllowed(t *testing.T) {
// 	server, err := apiserver.New(
// 		apiserver.WithLogger(getLogger()),
// 		apiserver.WithListenAddr(":8000"),
// 		apiserver.WithHTTPHandler(fasthttp.MethodGet, "/healthz", apiserver.HealthCheckHandler),
// 		apiserver.WithKafkaBrokers(getBrokers()),
// 		apiserver.WithKafkaGitHubTopic(kafkaconsumer.KafkaTopicIdentifierGitHub),
// 	)
// 	assert.NoError(t, err)
// 	assert.NotNil(t, server)
//
// 	ctx := &fasthttp.RequestCtx{}
// 	ctx.Request.SetRequestURI("/healthz")
// 	ctx.Request.Header.SetMethod(fasthttp.MethodPut)
//
// 	server.FastHTTP.Handler(ctx)
// 	assert.Equal(t, fasthttp.StatusMethodNotAllowed, ctx.Response.StatusCode())
// }
//
// func TestApiServer_HealthCheckHandler(t *testing.T) {
// 	server, err := apiserver.New(
// 		apiserver.WithLogger(getLogger()),
// 		apiserver.WithListenAddr(":8000"),
// 		apiserver.WithHTTPHandler(fasthttp.MethodGet, "/healthz", apiserver.HealthCheckHandler),
// 		apiserver.WithKafkaBrokers(getBrokers()),
// 		apiserver.WithKafkaGitHubTopic(kafkaconsumer.KafkaTopicIdentifierGitHub),
// 	)
// 	assert.NoError(t, err)
// 	assert.NotNil(t, server)
//
// 	ctx := &fasthttp.RequestCtx{}
// 	ctx.Request.SetRequestURI("/healthz")
// 	ctx.Request.Header.SetMethod(fasthttp.MethodGet)
//
// 	server.FastHTTP.Handler(ctx)
// 	assert.Equal(t, fasthttp.StatusOK, ctx.Response.StatusCode())
// }
//
// func TestApiServer_NoKafkaBrokers(t *testing.T) {
// 	server, err := apiserver.New(
// 		apiserver.WithLogger(getLogger()),
// 		apiserver.WithListenAddr(":8000"),
// 		apiserver.WithHTTPHandler(fasthttp.MethodGet, "/healthz", apiserver.HealthCheckHandler),
// 		apiserver.WithKafkaGitHubTopic(kafkaconsumer.KafkaTopicIdentifierGitHub),
// 	)
// 	assert.Error(t, err)
// 	assert.Nil(t, server)
// 	assert.ErrorIs(t, err, cerrors.ErrValueRequired)
// }
