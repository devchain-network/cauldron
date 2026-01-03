package gitlabwebhookhandler_test

import (
	"testing"
	"time"

	"github.com/IBM/sarama"
	"github.com/devchain-network/cauldron/internal/cerrors"
	"github.com/devchain-network/cauldron/internal/kafkacp"
	"github.com/devchain-network/cauldron/internal/slogger/mockslogger"
	"github.com/devchain-network/cauldron/internal/transport/http/gitlabwebhookhandler"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
)

func TestNew_NoLogger(t *testing.T) {
	handler, err := gitlabwebhookhandler.New()

	assert.ErrorIs(t, err, cerrors.ErrValueRequired)
	assert.Nil(t, handler)
}

func TestNew_NilLogger(t *testing.T) {
	handler, err := gitlabwebhookhandler.New(
		gitlabwebhookhandler.WithLogger(nil),
	)

	assert.ErrorIs(t, err, cerrors.ErrValueRequired)
	assert.Nil(t, handler)
}

func TestNew_NoTopic(t *testing.T) {
	logger := mockslogger.New()

	handler, err := gitlabwebhookhandler.New(
		gitlabwebhookhandler.WithLogger(logger),
	)

	assert.ErrorIs(t, err, cerrors.ErrInvalid)
	assert.Nil(t, handler)
}

func TestNew_InvalidTopic(t *testing.T) {
	logger := mockslogger.New()

	handler, err := gitlabwebhookhandler.New(
		gitlabwebhookhandler.WithLogger(logger),
		gitlabwebhookhandler.WithTopic("foo"),
	)

	assert.ErrorIs(t, err, cerrors.ErrInvalid)
	assert.Nil(t, handler)
}

func TestNew_NoSecret(t *testing.T) {
	logger := mockslogger.New()

	handler, err := gitlabwebhookhandler.New(
		gitlabwebhookhandler.WithLogger(logger),
		gitlabwebhookhandler.WithTopic(kafkacp.KafkaTopicIdentifierGitLab.String()),
	)

	assert.ErrorIs(t, err, cerrors.ErrValueRequired)
	assert.Nil(t, handler)
}

func TestNew_EmptySecret(t *testing.T) {
	logger := mockslogger.New()

	handler, err := gitlabwebhookhandler.New(
		gitlabwebhookhandler.WithLogger(logger),
		gitlabwebhookhandler.WithTopic(kafkacp.KafkaTopicIdentifierGitLab.String()),
		gitlabwebhookhandler.WithWebhookSecret(""),
	)

	assert.ErrorIs(t, err, cerrors.ErrValueRequired)
	assert.Nil(t, handler)
}

func TestNew_NoMessageQueue(t *testing.T) {
	logger := mockslogger.New()

	handler, err := gitlabwebhookhandler.New(
		gitlabwebhookhandler.WithLogger(logger),
		gitlabwebhookhandler.WithTopic(kafkacp.KafkaTopicIdentifierGitLab.String()),
		gitlabwebhookhandler.WithWebhookSecret("my-secret"),
	)

	assert.ErrorIs(t, err, cerrors.ErrValueRequired)
	assert.Nil(t, handler)
}

func TestNew_NilMessageQueue(t *testing.T) {
	logger := mockslogger.New()

	handler, err := gitlabwebhookhandler.New(
		gitlabwebhookhandler.WithLogger(logger),
		gitlabwebhookhandler.WithTopic(kafkacp.KafkaTopicIdentifierGitLab.String()),
		gitlabwebhookhandler.WithWebhookSecret("my-secret"),
		gitlabwebhookhandler.WithProducerGitLabMessageQueue(nil),
	)

	assert.ErrorIs(t, err, cerrors.ErrValueRequired)
	assert.Nil(t, handler)
}

func TestNew_Success(t *testing.T) {
	logger := mockslogger.New()
	messageQueue := make(chan *sarama.ProducerMessage, 10)

	handler, err := gitlabwebhookhandler.New(
		gitlabwebhookhandler.WithLogger(logger),
		gitlabwebhookhandler.WithTopic(kafkacp.KafkaTopicIdentifierGitLab.String()),
		gitlabwebhookhandler.WithWebhookSecret("my-secret"),
		gitlabwebhookhandler.WithProducerGitLabMessageQueue(messageQueue),
	)

	assert.NoError(t, err)
	assert.NotNil(t, handler)
}

func TestHandle_NoBody(t *testing.T) {
	logger := mockslogger.New()
	messageQueue := make(chan *sarama.ProducerMessage, 10)

	handler, err := gitlabwebhookhandler.New(
		gitlabwebhookhandler.WithLogger(logger),
		gitlabwebhookhandler.WithTopic(kafkacp.KafkaTopicIdentifierGitLab.String()),
		gitlabwebhookhandler.WithWebhookSecret("my-secret"),
		gitlabwebhookhandler.WithProducerGitLabMessageQueue(messageQueue),
	)

	assert.NoError(t, err)
	assert.NotNil(t, handler)

	ctx := &fasthttp.RequestCtx{}
	handler.Handle(ctx)

	assert.Equal(t, fasthttp.StatusBadRequest, ctx.Response.StatusCode())
}

func TestHandle_NoToken(t *testing.T) {
	logger := mockslogger.New()
	messageQueue := make(chan *sarama.ProducerMessage, 10)

	handler, err := gitlabwebhookhandler.New(
		gitlabwebhookhandler.WithLogger(logger),
		gitlabwebhookhandler.WithTopic(kafkacp.KafkaTopicIdentifierGitLab.String()),
		gitlabwebhookhandler.WithWebhookSecret("my-secret"),
		gitlabwebhookhandler.WithProducerGitLabMessageQueue(messageQueue),
	)

	assert.NoError(t, err)
	assert.NotNil(t, handler)

	ctx := &fasthttp.RequestCtx{}
	ctx.Request.SetBodyString(`{"object_kind": "push"}`)
	handler.Handle(ctx)

	assert.Equal(t, fasthttp.StatusBadRequest, ctx.Response.StatusCode())
}

func TestHandle_InvalidToken(t *testing.T) {
	logger := mockslogger.New()
	messageQueue := make(chan *sarama.ProducerMessage, 10)

	handler, err := gitlabwebhookhandler.New(
		gitlabwebhookhandler.WithLogger(logger),
		gitlabwebhookhandler.WithTopic(kafkacp.KafkaTopicIdentifierGitLab.String()),
		gitlabwebhookhandler.WithWebhookSecret("my-secret"),
		gitlabwebhookhandler.WithProducerGitLabMessageQueue(messageQueue),
	)

	assert.NoError(t, err)
	assert.NotNil(t, handler)

	ctx := &fasthttp.RequestCtx{}
	ctx.Request.SetBodyString(`{"object_kind": "push"}`)
	ctx.Request.Header.Set("X-Gitlab-Token", "wrong-secret")
	handler.Handle(ctx)

	assert.Equal(t, fasthttp.StatusBadRequest, ctx.Response.StatusCode())
}

func newMockRequestCtx() *fasthttp.RequestCtx {
	var ctx fasthttp.RequestCtx
	ctx.Init(&fasthttp.Request{}, nil, nil)

	return &ctx
}

func TestHandle_NoEventUUID(t *testing.T) {
	logger := mockslogger.New()
	messageQueue := make(chan *sarama.ProducerMessage, 10)

	handler, err := gitlabwebhookhandler.New(
		gitlabwebhookhandler.WithLogger(logger),
		gitlabwebhookhandler.WithTopic(kafkacp.KafkaTopicIdentifierGitLab.String()),
		gitlabwebhookhandler.WithWebhookSecret("my-secret"),
		gitlabwebhookhandler.WithProducerGitLabMessageQueue(messageQueue),
	)

	assert.NoError(t, err)
	assert.NotNil(t, handler)

	ctx := newMockRequestCtx()
	ctx.Request.SetBodyString(`{"object_kind": "push"}`)
	ctx.Request.Header.Set("X-Gitlab-Token", "my-secret")

	done := make(chan bool)
	go func() {
		select {
		case msg := <-messageQueue:
			t.Errorf("unexpected message in queue: %v", msg)
		case <-time.After(100 * time.Millisecond):
			done <- true
		}
	}()

	handler.Handle(ctx)

	assert.Equal(t, fasthttp.StatusBadRequest, ctx.Response.StatusCode())
	<-done
}

func TestHandle_NoWebhookUUID(t *testing.T) {
	logger := mockslogger.New()
	messageQueue := make(chan *sarama.ProducerMessage, 10)

	handler, err := gitlabwebhookhandler.New(
		gitlabwebhookhandler.WithLogger(logger),
		gitlabwebhookhandler.WithTopic(kafkacp.KafkaTopicIdentifierGitLab.String()),
		gitlabwebhookhandler.WithWebhookSecret("my-secret"),
		gitlabwebhookhandler.WithProducerGitLabMessageQueue(messageQueue),
	)

	assert.NoError(t, err)
	assert.NotNil(t, handler)

	ctx := newMockRequestCtx()
	ctx.Request.SetBodyString(`{"object_kind": "push"}`)
	ctx.Request.Header.Set("X-Gitlab-Token", "my-secret")
	ctx.Request.Header.Set("X-Gitlab-Event-Uuid", uuid.New().String())

	done := make(chan bool)
	go func() {
		select {
		case msg := <-messageQueue:
			t.Errorf("unexpected message in queue: %v", msg)
		case <-time.After(100 * time.Millisecond):
			done <- true
		}
	}()

	handler.Handle(ctx)

	assert.Equal(t, fasthttp.StatusBadRequest, ctx.Response.StatusCode())
	<-done
}

func TestHandle_NoObjectKindOrEventName(t *testing.T) {
	logger := mockslogger.New()
	messageQueue := make(chan *sarama.ProducerMessage, 10)

	handler, err := gitlabwebhookhandler.New(
		gitlabwebhookhandler.WithLogger(logger),
		gitlabwebhookhandler.WithTopic(kafkacp.KafkaTopicIdentifierGitLab.String()),
		gitlabwebhookhandler.WithWebhookSecret("my-secret"),
		gitlabwebhookhandler.WithProducerGitLabMessageQueue(messageQueue),
	)

	assert.NoError(t, err)
	assert.NotNil(t, handler)

	ctx := newMockRequestCtx()
	ctx.Request.SetBodyString(`{"foo": "bar"}`)
	ctx.Request.Header.Set("X-Gitlab-Token", "my-secret")
	ctx.Request.Header.Set("X-Gitlab-Event-Uuid", uuid.New().String())
	ctx.Request.Header.Set("X-Gitlab-Webhook-Uuid", uuid.New().String())

	done := make(chan bool)
	go func() {
		select {
		case msg := <-messageQueue:
			t.Errorf("unexpected message in queue: %v", msg)
		case <-time.After(100 * time.Millisecond):
			done <- true
		}
	}()

	handler.Handle(ctx)

	assert.Equal(t, fasthttp.StatusBadRequest, ctx.Response.StatusCode())
	<-done
}

func TestHandle_Success_ObjectKind(t *testing.T) {
	logger := mockslogger.New()
	messageQueue := make(chan *sarama.ProducerMessage, 10)

	handler, err := gitlabwebhookhandler.New(
		gitlabwebhookhandler.WithLogger(logger),
		gitlabwebhookhandler.WithTopic(kafkacp.KafkaTopicIdentifierGitLab.String()),
		gitlabwebhookhandler.WithWebhookSecret("my-secret"),
		gitlabwebhookhandler.WithProducerGitLabMessageQueue(messageQueue),
	)

	assert.NoError(t, err)
	assert.NotNil(t, handler)

	body := `{
		"object_kind": "merge_request",
		"user": {"id": 123, "username": "vigo"},
		"project": {"id": 456, "path_with_namespace": "devchain/cauldron"}
	}`

	ctx := newMockRequestCtx()
	ctx.Request.SetBodyString(body)
	ctx.Request.Header.Set("X-Gitlab-Token", "my-secret")
	ctx.Request.Header.Set("X-Gitlab-Event-Uuid", uuid.New().String())
	ctx.Request.Header.Set("X-Gitlab-Webhook-Uuid", uuid.New().String())

	handler.Handle(ctx)

	assert.Equal(t, fasthttp.StatusAccepted, ctx.Response.StatusCode())
	assert.NotEmpty(t, <-messageQueue)
}

func TestHandle_Success_EventName(t *testing.T) {
	logger := mockslogger.New()
	messageQueue := make(chan *sarama.ProducerMessage, 10)

	handler, err := gitlabwebhookhandler.New(
		gitlabwebhookhandler.WithLogger(logger),
		gitlabwebhookhandler.WithTopic(kafkacp.KafkaTopicIdentifierGitLab.String()),
		gitlabwebhookhandler.WithWebhookSecret("my-secret"),
		gitlabwebhookhandler.WithProducerGitLabMessageQueue(messageQueue),
	)

	assert.NoError(t, err)
	assert.NotNil(t, handler)

	// Premium event with event_name instead of object_kind
	body := `{
		"event_name": "project_create",
		"project_id": 22,
		"path_with_namespace": "group1/project1"
	}`

	ctx := newMockRequestCtx()
	ctx.Request.SetBodyString(body)
	ctx.Request.Header.Set("X-Gitlab-Token", "my-secret")
	ctx.Request.Header.Set("X-Gitlab-Event-Uuid", uuid.New().String())
	ctx.Request.Header.Set("X-Gitlab-Webhook-Uuid", uuid.New().String())

	handler.Handle(ctx)

	assert.Equal(t, fasthttp.StatusAccepted, ctx.Response.StatusCode())
	assert.NotEmpty(t, <-messageQueue)
}

func TestHandle_Success_PushEvent_FlatUser(t *testing.T) {
	logger := mockslogger.New()
	messageQueue := make(chan *sarama.ProducerMessage, 10)

	handler, err := gitlabwebhookhandler.New(
		gitlabwebhookhandler.WithLogger(logger),
		gitlabwebhookhandler.WithTopic(kafkacp.KafkaTopicIdentifierGitLab.String()),
		gitlabwebhookhandler.WithWebhookSecret("my-secret"),
		gitlabwebhookhandler.WithProducerGitLabMessageQueue(messageQueue),
	)

	assert.NoError(t, err)
	assert.NotNil(t, handler)

	// Push event has flat user fields
	body := `{
		"object_kind": "push",
		"user_id": 383362,
		"user_username": "vigo",
		"project_id": 77447212,
		"project": {"id": 77447212, "path_with_namespace": "vigo/webhook-tests-repo"}
	}`

	ctx := newMockRequestCtx()
	ctx.Request.SetBodyString(body)
	ctx.Request.Header.Set("X-Gitlab-Token", "my-secret")
	ctx.Request.Header.Set("X-Gitlab-Event-Uuid", uuid.New().String())
	ctx.Request.Header.Set("X-Gitlab-Webhook-Uuid", uuid.New().String())

	handler.Handle(ctx)

	assert.Equal(t, fasthttp.StatusAccepted, ctx.Response.StatusCode())
	msg := <-messageQueue
	assert.NotNil(t, msg)
	assert.Equal(t, string(msg.Value.(sarama.ByteEncoder)), body)
}

func TestHandle_Success_UserAddToGroup(t *testing.T) {
	logger := mockslogger.New()
	messageQueue := make(chan *sarama.ProducerMessage, 10)

	handler, err := gitlabwebhookhandler.New(
		gitlabwebhookhandler.WithLogger(logger),
		gitlabwebhookhandler.WithTopic(kafkacp.KafkaTopicIdentifierGitLab.String()),
		gitlabwebhookhandler.WithWebhookSecret("my-secret"),
		gitlabwebhookhandler.WithProducerGitLabMessageQueue(messageQueue),
	)

	assert.NoError(t, err)
	assert.NotNil(t, handler)

	// Premium event: user_add_to_group
	body := `{
		"event_name": "user_add_to_group",
		"user_username": "test_user",
		"user_id": 64,
		"group_id": 100,
		"group_path": "webhook-test"
	}`

	ctx := newMockRequestCtx()
	ctx.Request.SetBodyString(body)
	ctx.Request.Header.Set("X-Gitlab-Token", "my-secret")
	ctx.Request.Header.Set("X-Gitlab-Event-Uuid", uuid.New().String())
	ctx.Request.Header.Set("X-Gitlab-Webhook-Uuid", uuid.New().String())

	handler.Handle(ctx)

	assert.Equal(t, fasthttp.StatusAccepted, ctx.Response.StatusCode())
	assert.NotEmpty(t, <-messageQueue)
}

func TestHandle_Success_SubgroupCreate(t *testing.T) {
	logger := mockslogger.New()
	messageQueue := make(chan *sarama.ProducerMessage, 10)

	handler, err := gitlabwebhookhandler.New(
		gitlabwebhookhandler.WithLogger(logger),
		gitlabwebhookhandler.WithTopic(kafkacp.KafkaTopicIdentifierGitLab.String()),
		gitlabwebhookhandler.WithWebhookSecret("my-secret"),
		gitlabwebhookhandler.WithProducerGitLabMessageQueue(messageQueue),
	)

	assert.NoError(t, err)
	assert.NotNil(t, handler)

	// Premium event: subgroup_create
	body := `{
		"event_name": "subgroup_create",
		"group_id": 10,
		"full_path": "group1/subgroup1"
	}`

	ctx := newMockRequestCtx()
	ctx.Request.SetBodyString(body)
	ctx.Request.Header.Set("X-Gitlab-Token", "my-secret")
	ctx.Request.Header.Set("X-Gitlab-Event-Uuid", uuid.New().String())
	ctx.Request.Header.Set("X-Gitlab-Webhook-Uuid", uuid.New().String())

	handler.Handle(ctx)

	assert.Equal(t, fasthttp.StatusAccepted, ctx.Response.StatusCode())
	assert.NotEmpty(t, <-messageQueue)
}

func TestMessageQueue_Scenarios(t *testing.T) {
	logger := mockslogger.New()
	messageQueue := make(chan *sarama.ProducerMessage, 1)

	handler, err := gitlabwebhookhandler.New(
		gitlabwebhookhandler.WithLogger(logger),
		gitlabwebhookhandler.WithTopic(kafkacp.KafkaTopicIdentifierGitLab.String()),
		gitlabwebhookhandler.WithWebhookSecret("my-secret"),
		gitlabwebhookhandler.WithProducerGitLabMessageQueue(messageQueue),
	)

	assert.NoError(t, err)
	assert.NotNil(t, handler)

	body := `{
		"object_kind": "push",
		"user": {"id": 123, "username": "vigo"},
		"project": {"id": 456, "path_with_namespace": "devchain/cauldron"}
	}`

	ctx := newMockRequestCtx()
	ctx.Request.SetBodyString(body)
	ctx.Request.Header.Set("X-Gitlab-Token", "my-secret")
	ctx.Request.Header.Set("X-Gitlab-Event-Uuid", uuid.New().String())
	ctx.Request.Header.Set("X-Gitlab-Webhook-Uuid", uuid.New().String())

	handler.Handle(ctx)

	go func() {
		msg := <-messageQueue
		assert.NotNil(t, msg)
		assert.Equal(t, string(msg.Value.(sarama.ByteEncoder)), body)
	}()

	assert.Equal(t, fasthttp.StatusAccepted, ctx.Response.StatusCode())
	assert.NotEmpty(t, <-messageQueue)
}
