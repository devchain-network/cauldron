package githubhandleroptions_test

import (
	"net/http"
	"testing"

	"github.com/devchain-network/cauldron/internal/apiserver/githubhandleroptions"
	"github.com/devchain-network/cauldron/internal/apiserver/httphandleroptions"
	"github.com/devchain-network/cauldron/internal/cerrors"
	"github.com/devchain-network/cauldron/internal/kafkaconsumer"
	"github.com/go-playground/webhooks/v6/github"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNew_ErrorNilWebhook(t *testing.T) {
	handler, err := githubhandleroptions.New(
		githubhandleroptions.WithCommonHandler(&httphandleroptions.HTTPHandler{}),
		githubhandleroptions.WithTopic(kafkaconsumer.KafkaTopicIdentifierGitHub),
	)
	assert.Error(t, err)
	assert.Nil(t, handler)
	assert.ErrorIs(t, err, cerrors.ErrValueRequired)
}

func TestNew_ErrorNilCommonHandler(t *testing.T) {
	webhook := &github.Webhook{}
	handler, err := githubhandleroptions.New(
		githubhandleroptions.WithWebhook(webhook),
		githubhandleroptions.WithTopic(kafkaconsumer.KafkaTopicIdentifierGitHub),
	)
	assert.Error(t, err)
	assert.Nil(t, handler)
	assert.ErrorIs(t, err, cerrors.ErrValueRequired)
}

func TestNew_ErrorInvalidTopic(t *testing.T) {
	webhook := &github.Webhook{}
	handler, err := githubhandleroptions.New(
		githubhandleroptions.WithWebhook(webhook),
		githubhandleroptions.WithCommonHandler(&httphandleroptions.HTTPHandler{}),
		githubhandleroptions.WithTopic("invalid-topic"),
	)
	assert.Error(t, err)
	assert.Nil(t, handler)
	assert.ErrorIs(t, err, cerrors.ErrInvalid)
}

func TestNew_Success(t *testing.T) {
	webhook := &github.Webhook{}
	commonHandler := &httphandleroptions.HTTPHandler{}
	handler, err := githubhandleroptions.New(
		githubhandleroptions.WithWebhook(webhook),
		githubhandleroptions.WithCommonHandler(commonHandler),
		githubhandleroptions.WithTopic(kafkaconsumer.KafkaTopicIdentifierGitHub),
	)
	assert.NoError(t, err)
	assert.NotNil(t, handler)
	assert.Equal(t, webhook, handler.Webhook)
	assert.Equal(t, commonHandler, handler.CommonHandler)
	assert.Equal(t, kafkaconsumer.KafkaTopicIdentifierGitHub, handler.Topic)
}

func TestParseRequestHeaders(t *testing.T) {
	headers := http.Header{}
	headers.Set("X-Github-Event", "push")
	headers.Set("X-Github-Delivery", uuid.New().String())
	headers.Set("X-Github-Hook-Id", "12345")
	headers.Set("X-Github-Hook-Installation-Target-Id", "67890")
	headers.Set("X-Github-Hook-Installation-Target-Type", "organization")

	handler := &githubhandleroptions.HTTPHandler{}
	parsed := handler.ParseRequestHeaders(headers)

	assert.Equal(t, "push", parsed.Event)
	assert.NotEqual(t, uuid.Nil, parsed.DeliveryID)
	assert.Equal(t, uint64(12345), parsed.HookID)
	assert.Equal(t, uint64(67890), parsed.TargetID)
	assert.Equal(t, "organization", parsed.TargetType)
}

func TestParseRequestHeaders_MissingHeaders(t *testing.T) {
	headers := http.Header{}

	handler := &githubhandleroptions.HTTPHandler{}
	parsed := handler.ParseRequestHeaders(headers)

	assert.Empty(t, parsed.Event)
	assert.Equal(t, uuid.Nil, parsed.DeliveryID)
	assert.Equal(t, uint64(0), parsed.HookID)
	assert.Equal(t, uint64(0), parsed.TargetID)
	assert.Empty(t, parsed.TargetType)
}
