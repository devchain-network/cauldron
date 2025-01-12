package githubhandleroptions

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/devchain-network/cauldron/internal/apiserver/httphandleroptions"
	"github.com/devchain-network/cauldron/internal/cerrors"
	"github.com/devchain-network/cauldron/internal/kafkaconsumer"
	"github.com/go-playground/webhooks/v6/github"
	"github.com/google/uuid"
)

var _ Handler = (*HTTPHandler)(nil) // compile time proof

// Handler defines http handler behaviours.
type Handler interface {
	ParseRequestHeaders(h http.Header) *RequestHeaders
}

// HTTPHandler represents github http handler functionality.
type HTTPHandler struct {
	Webhook       *github.Webhook
	CommonHandler *httphandleroptions.HTTPHandler
	Topic         kafkaconsumer.KafkaTopicIdentifier
}

// Option represents option function type.
type Option func(*HTTPHandler) error

// RequestHeaders represents required http headers for webhook transaction record.
type RequestHeaders struct {
	Event      string
	TargetType string
	DeliveryID uuid.UUID
	HookID     uint64
	TargetID   uint64
}

// ParseRequestHeaders parses incoming http headers and sets header values.
func (HTTPHandler) ParseRequestHeaders(h http.Header) *RequestHeaders {
	parsed := new(RequestHeaders)

	if val := h.Get("X-Github-Event"); val != "" {
		parsed.Event = val
	}

	if val, err := uuid.Parse(h.Get("X-Github-Delivery")); err == nil {
		parsed.DeliveryID = val
	}

	if val, err := strconv.ParseUint(h.Get("X-Github-Hook-Id"), 10, 64); err == nil {
		parsed.HookID = val
	}

	if val, err := strconv.ParseUint(h.Get("X-Github-Hook-Installation-Target-Id"), 10, 64); err == nil {
		parsed.TargetID = val
	}

	if val := h.Get("X-Github-Hook-Installation-Target-Type"); val != "" {
		parsed.TargetType = val
	}

	return parsed
}

// WithWebhook sets webhook.
func WithWebhook(wh *github.Webhook) Option {
	return func(hh *HTTPHandler) error {
		if wh == nil {
			return fmt.Errorf("githubhandleroptions.WithWebhook error: [%w]", cerrors.ErrValueRequired)
		}
		hh.Webhook = wh

		return nil
	}
}

// WithCommonHandler sets common handler.
func WithCommonHandler(h *httphandleroptions.HTTPHandler) Option {
	return func(hh *HTTPHandler) error {
		if h == nil {
			return fmt.Errorf("githubhandleroptions.WithCommonHandler error: [%w]", cerrors.ErrValueRequired)
		}
		hh.CommonHandler = h

		return nil
	}
}

// WithTopic sets topic name to consume.
func WithTopic(s kafkaconsumer.KafkaTopicIdentifier) Option {
	return func(hh *HTTPHandler) error {
		if err := kafkaconsumer.IsKafkaTopicValid(s); err != nil {
			return fmt.Errorf("githubhandleroptions.WithTopic hh.Topic error: [%w]", err)
		}

		hh.Topic = s

		return nil
	}
}

// New instantiates new http handler for github webhook.
func New(options ...Option) (*HTTPHandler, error) {
	handler := new(HTTPHandler)

	for _, option := range options {
		if err := option(handler); err != nil {
			return nil, fmt.Errorf("githubhandleroptions.New option error: [%w]", err)
		}
	}

	if handler.Webhook == nil {
		return nil, fmt.Errorf("githubhandleroptions.New handler.Webhook error: [%w]", cerrors.ErrValueRequired)
	}
	if handler.CommonHandler == nil {
		return nil, fmt.Errorf("githubhandleroptions.New handler.CommonHandler error: [%w]", cerrors.ErrValueRequired)
	}

	return handler, nil
}
