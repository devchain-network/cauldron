package githubwebhookhandler_test

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"testing"
	"time"

	"github.com/IBM/sarama"
	"github.com/devchain-network/cauldron/internal/cerrors"
	"github.com/devchain-network/cauldron/internal/kafkacp"
	"github.com/devchain-network/cauldron/internal/slogger/mockslogger"
	"github.com/devchain-network/cauldron/internal/transport/http/githubwebhookhandler"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
)

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
	logger := mockslogger.New()

	handler, err := githubwebhookhandler.New(
		githubwebhookhandler.WithLogger(logger),
	)

	assert.ErrorIs(t, err, cerrors.ErrInvalid)
	assert.Nil(t, handler)
}

func TestNew_InvalidTopic(t *testing.T) {
	logger := mockslogger.New()

	handler, err := githubwebhookhandler.New(
		githubwebhookhandler.WithLogger(logger),
		githubwebhookhandler.WithTopic("foo"),
	)

	assert.ErrorIs(t, err, cerrors.ErrInvalid)
	assert.Nil(t, handler)
}

func TestNew_NoSecret(t *testing.T) {
	logger := mockslogger.New()

	handler, err := githubwebhookhandler.New(
		githubwebhookhandler.WithLogger(logger),
		githubwebhookhandler.WithTopic(kafkacp.KafkaTopicIdentifierGitHub.String()),
	)

	assert.ErrorIs(t, err, cerrors.ErrValueRequired)
	assert.Nil(t, handler)
}

func TestNew_EmptySecret(t *testing.T) {
	logger := mockslogger.New()

	handler, err := githubwebhookhandler.New(
		githubwebhookhandler.WithLogger(logger),
		githubwebhookhandler.WithTopic(kafkacp.KafkaTopicIdentifierGitHub.String()),
		githubwebhookhandler.WithWebhookSecret(""),
	)

	assert.ErrorIs(t, err, cerrors.ErrValueRequired)
	assert.Nil(t, handler)
}

func TestNew_NoMessageQueue(t *testing.T) {
	logger := mockslogger.New()

	handler, err := githubwebhookhandler.New(
		githubwebhookhandler.WithLogger(logger),
		githubwebhookhandler.WithTopic(kafkacp.KafkaTopicIdentifierGitHub.String()),
		githubwebhookhandler.WithWebhookSecret("my-secret"),
	)

	assert.ErrorIs(t, err, cerrors.ErrValueRequired)
	assert.Nil(t, handler)
}

func TestNew_NilMessageQueue(t *testing.T) {
	logger := mockslogger.New()

	handler, err := githubwebhookhandler.New(
		githubwebhookhandler.WithLogger(logger),
		githubwebhookhandler.WithTopic(kafkacp.KafkaTopicIdentifierGitHub.String()),
		githubwebhookhandler.WithWebhookSecret("my-secret"),
		githubwebhookhandler.WithProducerGitHubMessageQueue(nil),
	)

	assert.ErrorIs(t, err, cerrors.ErrValueRequired)
	assert.Nil(t, handler)
}

func TestNew_Success(t *testing.T) {
	logger := mockslogger.New()
	messageQueue := make(chan *sarama.ProducerMessage, 10)

	handler, err := githubwebhookhandler.New(
		githubwebhookhandler.WithLogger(logger),
		githubwebhookhandler.WithTopic(kafkacp.KafkaTopicIdentifierGitHub.String()),
		githubwebhookhandler.WithWebhookSecret("my-secret"),
		githubwebhookhandler.WithProducerGitHubMessageQueue(messageQueue),
	)

	assert.NoError(t, err)
	assert.NotNil(t, handler)
}

func TestHandle_NoBody(t *testing.T) {
	logger := mockslogger.New()
	messageQueue := make(chan *sarama.ProducerMessage, 10)

	handler, err := githubwebhookhandler.New(
		githubwebhookhandler.WithLogger(logger),
		githubwebhookhandler.WithTopic(kafkacp.KafkaTopicIdentifierGitHub.String()),
		githubwebhookhandler.WithWebhookSecret("my-secret"),
		githubwebhookhandler.WithProducerGitHubMessageQueue(messageQueue),
	)

	assert.NoError(t, err)
	assert.NotNil(t, handler)

	ctx := &fasthttp.RequestCtx{}
	handler.Handle(ctx)

	assert.Equal(t, fasthttp.StatusBadRequest, ctx.Response.StatusCode())
}

func TestHandle_NoHMAC(t *testing.T) {
	logger := mockslogger.New()
	messageQueue := make(chan *sarama.ProducerMessage, 10)

	handler, err := githubwebhookhandler.New(
		githubwebhookhandler.WithLogger(logger),
		githubwebhookhandler.WithTopic(kafkacp.KafkaTopicIdentifierGitHub.String()),
		githubwebhookhandler.WithWebhookSecret("my-secret"),
		githubwebhookhandler.WithProducerGitHubMessageQueue(messageQueue),
	)

	assert.NoError(t, err)
	assert.NotNil(t, handler)

	ctx := &fasthttp.RequestCtx{}
	ctx.Request.SetBodyString(`{"sender": {"login": "test", "id": 123}}`)
	handler.Handle(ctx)

	assert.Equal(t, fasthttp.StatusBadRequest, ctx.Response.StatusCode())
}

func TestHandle_InvalidHMAC(t *testing.T) {
	logger := mockslogger.New()
	messageQueue := make(chan *sarama.ProducerMessage, 10)

	handler, err := githubwebhookhandler.New(
		githubwebhookhandler.WithLogger(logger),
		githubwebhookhandler.WithTopic(kafkacp.KafkaTopicIdentifierGitHub.String()),
		githubwebhookhandler.WithWebhookSecret("my-secret"),
		githubwebhookhandler.WithProducerGitHubMessageQueue(messageQueue),
	)

	assert.NoError(t, err)
	assert.NotNil(t, handler)

	ctx := &fasthttp.RequestCtx{}
	ctx.Request.SetBodyString(`{"sender": {"login": "test", "id": 123}}`)
	ctx.Request.Header.Set("X-Hub-Signature-256", "sha256=invalidsignature")
	handler.Handle(ctx)

	assert.Equal(t, fasthttp.StatusBadRequest, ctx.Response.StatusCode())
}

func newMockRequestCtx() *fasthttp.RequestCtx {
	var ctx fasthttp.RequestCtx
	ctx.Init(&fasthttp.Request{}, nil, nil)

	return &ctx
}

func TestHandle_NoXGithubEvent(t *testing.T) {
	logger := mockslogger.New()
	messageQueue := make(chan *sarama.ProducerMessage, 10)

	handler, err := githubwebhookhandler.New(
		githubwebhookhandler.WithLogger(logger),
		githubwebhookhandler.WithTopic(kafkacp.KafkaTopicIdentifierGitHub.String()),
		githubwebhookhandler.WithWebhookSecret("my-secret"),
		githubwebhookhandler.WithProducerGitHubMessageQueue(messageQueue),
	)

	assert.NoError(t, err)
	assert.NotNil(t, handler)

	secret := "my-secret"
	body := `{"sender": {"login": "test", "id": 123}}`
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(body))
	signature := "sha256=" + hex.EncodeToString(mac.Sum(nil))

	ctx := newMockRequestCtx()
	ctx.Request.SetBodyString(body)
	ctx.Request.Header.Set("X-Hub-Signature-256", signature)

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

func TestHandle_NoXGithubDeliveryID(t *testing.T) {
	logger := mockslogger.New()
	messageQueue := make(chan *sarama.ProducerMessage, 10)

	handler, err := githubwebhookhandler.New(
		githubwebhookhandler.WithLogger(logger),
		githubwebhookhandler.WithTopic(kafkacp.KafkaTopicIdentifierGitHub.String()),
		githubwebhookhandler.WithWebhookSecret("my-secret"),
		githubwebhookhandler.WithProducerGitHubMessageQueue(messageQueue),
	)

	assert.NoError(t, err)
	assert.NotNil(t, handler)

	secret := "my-secret"
	body := `{"sender": {"login": "test", "id": 123}}`
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(body))
	signature := "sha256=" + hex.EncodeToString(mac.Sum(nil))

	ctx := newMockRequestCtx()
	ctx.Request.SetBodyString(body)
	ctx.Request.Header.Set("X-Hub-Signature-256", signature)
	ctx.Request.Header.Set("X-Github-Event", "push")

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

func TestHandle_NoXGithubHookID(t *testing.T) {
	logger := mockslogger.New()
	messageQueue := make(chan *sarama.ProducerMessage, 10)

	handler, err := githubwebhookhandler.New(
		githubwebhookhandler.WithLogger(logger),
		githubwebhookhandler.WithTopic(kafkacp.KafkaTopicIdentifierGitHub.String()),
		githubwebhookhandler.WithWebhookSecret("my-secret"),
		githubwebhookhandler.WithProducerGitHubMessageQueue(messageQueue),
	)

	assert.NoError(t, err)
	assert.NotNil(t, handler)

	secret := "my-secret"
	body := `{"sender": {"login": "test", "id": 123}}`
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(body))
	signature := "sha256=" + hex.EncodeToString(mac.Sum(nil))

	ctx := newMockRequestCtx()
	ctx.Request.SetBodyString(body)
	ctx.Request.Header.Set("X-Hub-Signature-256", signature)
	ctx.Request.Header.Set("X-Github-Event", "push")
	ctx.Request.Header.Set("X-Github-Delivery", uuid.New().String())

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

func TestHandle_NoXGithubInstallationTargetID(t *testing.T) {
	logger := mockslogger.New()
	messageQueue := make(chan *sarama.ProducerMessage, 10)

	handler, err := githubwebhookhandler.New(
		githubwebhookhandler.WithLogger(logger),
		githubwebhookhandler.WithTopic(kafkacp.KafkaTopicIdentifierGitHub.String()),
		githubwebhookhandler.WithWebhookSecret("my-secret"),
		githubwebhookhandler.WithProducerGitHubMessageQueue(messageQueue),
	)

	assert.NoError(t, err)
	assert.NotNil(t, handler)

	secret := "my-secret"
	body := `{"sender": {"login": "test", "id": 123}}`
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(body))
	signature := "sha256=" + hex.EncodeToString(mac.Sum(nil))

	ctx := newMockRequestCtx()
	ctx.Request.SetBodyString(body)
	ctx.Request.Header.Set("X-Hub-Signature-256", signature)
	ctx.Request.Header.Set("X-Github-Event", "push")
	ctx.Request.Header.Set("X-Github-Delivery", uuid.New().String())
	ctx.Request.Header.Set("X-Github-Hook-Id", "123")

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

func TestHandle_NoXGithubInstallationTargetType(t *testing.T) {
	logger := mockslogger.New()
	messageQueue := make(chan *sarama.ProducerMessage, 10)

	handler, err := githubwebhookhandler.New(
		githubwebhookhandler.WithLogger(logger),
		githubwebhookhandler.WithTopic(kafkacp.KafkaTopicIdentifierGitHub.String()),
		githubwebhookhandler.WithWebhookSecret("my-secret"),
		githubwebhookhandler.WithProducerGitHubMessageQueue(messageQueue),
	)

	assert.NoError(t, err)
	assert.NotNil(t, handler)

	secret := "my-secret"
	body := `{"sender": {"login": "test", "id": 123}}`
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(body))
	signature := "sha256=" + hex.EncodeToString(mac.Sum(nil))

	ctx := newMockRequestCtx()
	ctx.Request.SetBodyString(body)
	ctx.Request.Header.Set("X-Hub-Signature-256", signature)
	ctx.Request.Header.Set("X-Github-Event", "push")
	ctx.Request.Header.Set("X-Github-Delivery", uuid.New().String())
	ctx.Request.Header.Set("X-Github-Hook-Id", "123")
	ctx.Request.Header.Set("X-Github-Hook-Installation-Target-Id", "456")

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

func TestHandle_NoSenderLogin(t *testing.T) {
	logger := mockslogger.New()
	messageQueue := make(chan *sarama.ProducerMessage, 10)

	handler, err := githubwebhookhandler.New(
		githubwebhookhandler.WithLogger(logger),
		githubwebhookhandler.WithTopic(kafkacp.KafkaTopicIdentifierGitHub.String()),
		githubwebhookhandler.WithWebhookSecret("my-secret"),
		githubwebhookhandler.WithProducerGitHubMessageQueue(messageQueue),
	)

	assert.NoError(t, err)
	assert.NotNil(t, handler)

	secret := "my-secret"
	body := `{"sender": {"loginx": "test", "id": 123}}`
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(body))
	signature := "sha256=" + hex.EncodeToString(mac.Sum(nil))

	ctx := newMockRequestCtx()
	ctx.Request.SetBodyString(body)
	ctx.Request.Header.Set("X-Hub-Signature-256", signature)
	ctx.Request.Header.Set("X-Github-Event", "push")
	ctx.Request.Header.Set("X-Github-Delivery", uuid.New().String())
	ctx.Request.Header.Set("X-Github-Hook-Id", "123")
	ctx.Request.Header.Set("X-Github-Hook-Installation-Target-Id", "456")
	ctx.Request.Header.Set("X-Github-Hook-Installation-Target-Type", "repository")

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

func TestHandle_NoSenderID(t *testing.T) {
	logger := mockslogger.New()
	messageQueue := make(chan *sarama.ProducerMessage, 10)

	handler, err := githubwebhookhandler.New(
		githubwebhookhandler.WithLogger(logger),
		githubwebhookhandler.WithTopic(kafkacp.KafkaTopicIdentifierGitHub.String()),
		githubwebhookhandler.WithWebhookSecret("my-secret"),
		githubwebhookhandler.WithProducerGitHubMessageQueue(messageQueue),
	)

	assert.NoError(t, err)
	assert.NotNil(t, handler)

	secret := "my-secret"
	body := `{"sender": {"login": "test", "idx": 123}}`
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(body))
	signature := "sha256=" + hex.EncodeToString(mac.Sum(nil))

	ctx := newMockRequestCtx()
	ctx.Request.SetBodyString(body)
	ctx.Request.Header.Set("X-Hub-Signature-256", signature)
	ctx.Request.Header.Set("X-Github-Event", "push")
	ctx.Request.Header.Set("X-Github-Delivery", uuid.New().String())
	ctx.Request.Header.Set("X-Github-Hook-Id", "123")
	ctx.Request.Header.Set("X-Github-Hook-Installation-Target-Id", "456")
	ctx.Request.Header.Set("X-Github-Hook-Installation-Target-Type", "repository")

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

func TestMessageQueue_Scenarios(t *testing.T) {
	logger := mockslogger.New()
	messageQueue := make(chan *sarama.ProducerMessage, 1)

	handler, err := githubwebhookhandler.New(
		githubwebhookhandler.WithLogger(logger),
		githubwebhookhandler.WithTopic(kafkacp.KafkaTopicIdentifierGitHub.String()),
		githubwebhookhandler.WithWebhookSecret("my-secret"),
		githubwebhookhandler.WithProducerGitHubMessageQueue(messageQueue),
	)

	assert.NoError(t, err)
	assert.NotNil(t, handler)

	secret := "my-secret"
	body := `{"sender": {"login": "test", "id": 123, "payload": "OK"}}`
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(body))
	signature := "sha256=" + hex.EncodeToString(mac.Sum(nil))

	ctx := newMockRequestCtx()
	ctx.Request.SetBodyString(body)
	ctx.Request.Header.Set("X-Hub-Signature-256", signature)
	ctx.Request.Header.Set("X-Github-Event", "push")
	ctx.Request.Header.Set("X-Github-Delivery", uuid.New().String())
	ctx.Request.Header.Set("X-Github-Hook-Id", "123")
	ctx.Request.Header.Set("X-Github-Hook-Installation-Target-Id", "456")
	ctx.Request.Header.Set("X-Github-Hook-Installation-Target-Type", "repository")

	handler.Handle(ctx)

	go func() {
		msg := <-messageQueue
		assert.NotNil(t, msg)
		assert.Equal(t, string(msg.Value.(sarama.ByteEncoder)), body)
	}()

	assert.Equal(t, fasthttp.StatusAccepted, ctx.Response.StatusCode())
	assert.NotEmpty(t, <-messageQueue)
}

func TestHandle_Success(t *testing.T) {
	logger := mockslogger.New()
	messageQueue := make(chan *sarama.ProducerMessage, 10)

	handler, err := githubwebhookhandler.New(
		githubwebhookhandler.WithLogger(logger),
		githubwebhookhandler.WithTopic(kafkacp.KafkaTopicIdentifierGitHub.String()),
		githubwebhookhandler.WithWebhookSecret("my-secret"),
		githubwebhookhandler.WithProducerGitHubMessageQueue(messageQueue),
	)

	assert.NoError(t, err)
	assert.NotNil(t, handler)

	secret := "my-secret"
	body := `{"sender": {"login": "test", "id": 123}}`
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(body))
	signature := "sha256=" + hex.EncodeToString(mac.Sum(nil))

	ctx := newMockRequestCtx()
	ctx.Request.SetBodyString(body)
	ctx.Request.Header.Set("X-Hub-Signature-256", signature)
	ctx.Request.Header.Set("X-Github-Event", "push")
	ctx.Request.Header.Set("X-Github-Delivery", uuid.New().String())
	ctx.Request.Header.Set("X-Github-Hook-Id", "123")
	ctx.Request.Header.Set("X-Github-Hook-Installation-Target-Id", "456")
	ctx.Request.Header.Set("X-Github-Hook-Installation-Target-Type", "repository")

	handler.Handle(ctx)

	assert.Equal(t, fasthttp.StatusAccepted, ctx.Response.StatusCode())
	assert.NotEmpty(t, <-messageQueue)
}
