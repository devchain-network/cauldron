package githubwebhookhandler

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/IBM/sarama"
	"github.com/buger/jsonparser"
	"github.com/devchain-network/cauldron/internal/cerrors"
	"github.com/devchain-network/cauldron/internal/kafkacp"
	"github.com/devchain-network/cauldron/internal/transport/http/httphandler"
	"github.com/valyala/fasthttp"
)

var _ GitHubWebhookHandler = (*Handler)(nil) // compile time proof

const (
	missingHTTPHandlerHeaderText   = "header"
	missingHTTPHandlerErrorMessage = "missing http header"
)

// GitHubWebhookHandler defines http handler behaviours.
type GitHubWebhookHandler interface {
	httphandler.FastHTTPHandler
}

// Handler represents http handler configuration and must satisfy GitHubWebhookHandler interface.
type Handler struct {
	MessageQueue chan *sarama.ProducerMessage
	Logger       *slog.Logger
	Topic        kafkacp.KafkaTopicIdentifier
	Secret       string
}

// Handle is a fasthttp handler function.
func (h Handler) Handle(ctx *fasthttp.RequestCtx) {
	if len(ctx.PostBody()) == 0 {
		h.Logger.Error("empty post body", "length", len(ctx.PostBody()))
		ctx.SetStatusCode(fasthttp.StatusBadRequest)

		return
	}

	hubSignature := string(ctx.Request.Header.Peek("X-Hub-Signature-256"))
	hubSignature = strings.TrimPrefix(hubSignature, "sha256=")

	mac := hmac.New(sha256.New, []byte(h.Secret))
	_, _ = mac.Write(ctx.PostBody())
	expectedMAC := hex.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(hubSignature), []byte(expectedMAC)) {
		h.Logger.Error("invalid github hmac/secret")
		ctx.SetStatusCode(fasthttp.StatusBadRequest)

		return
	}

	githubEvent := ctx.Request.Header.Peek("X-Github-Event")
	if len(githubEvent) == 0 {
		h.Logger.Error(missingHTTPHandlerErrorMessage, missingHTTPHandlerHeaderText, "X-Github-Event")
		ctx.SetStatusCode(fasthttp.StatusBadRequest)

		return
	}

	githubDelivery := ctx.Request.Header.Peek("X-Github-Delivery")
	if len(githubDelivery) == 0 {
		h.Logger.Error(missingHTTPHandlerErrorMessage, missingHTTPHandlerHeaderText, "X-Github-Delivery")
		ctx.SetStatusCode(fasthttp.StatusBadRequest)

		return
	}

	githubHookID := ctx.Request.Header.Peek("X-Github-Hook-Id")
	if len(githubHookID) == 0 {
		h.Logger.Error(missingHTTPHandlerErrorMessage, missingHTTPHandlerHeaderText, "X-Github-Hook-Id")
		ctx.SetStatusCode(fasthttp.StatusBadRequest)

		return
	}

	githubHookInstallationTargetID := ctx.Request.Header.Peek("X-Github-Hook-Installation-Target-Id")
	if len(githubHookInstallationTargetID) == 0 {
		h.Logger.Error(
			missingHTTPHandlerErrorMessage,
			missingHTTPHandlerHeaderText,
			"X-Github-Hook-Installation-Target-Id",
		)
		ctx.SetStatusCode(fasthttp.StatusBadRequest)

		return
	}

	githubHookInstallationTargetType := ctx.Request.Header.Peek("X-Github-Hook-Installation-Target-Type")
	if len(githubHookInstallationTargetType) == 0 {
		h.Logger.Error(
			missingHTTPHandlerErrorMessage,
			missingHTTPHandlerHeaderText,
			"X-Github-Hook-Installation-Target-Type",
		)
		ctx.SetStatusCode(fasthttp.StatusBadRequest)

		return
	}

	senderLogin, err := jsonparser.GetString(ctx.PostBody(), "sender", "login")
	if err != nil {
		h.Logger.Error("senderLogin jsonparser.GetString error", "error", err)
		ctx.SetStatusCode(fasthttp.StatusBadRequest)

		return
	}
	senderID, err := jsonparser.GetInt(ctx.PostBody(), "sender", "id")
	if err != nil {
		h.Logger.Error("senderID jsonparser.GetInt error", "error", err)
		ctx.SetStatusCode(fasthttp.StatusBadRequest)

		return
	}

	h.Logger.Info("recevied github webhook", "senderLogin", senderLogin, "senderID", senderID)

	message := &sarama.ProducerMessage{
		Topic: h.Topic.String(),
		Key:   sarama.StringEncoder(githubDelivery),
		Value: sarama.ByteEncoder(ctx.PostBody()),
		Headers: []sarama.RecordHeader{
			{Key: []byte("event"), Value: githubEvent},
			{Key: []byte("target-type"), Value: githubHookInstallationTargetType},
			{Key: []byte("target-id"), Value: githubHookInstallationTargetID},
			{Key: []byte("hook-id"), Value: githubHookID},
			{Key: []byte("sender-login"), Value: []byte(senderLogin)},
			{Key: []byte("sender-id"), Value: []byte(strconv.FormatInt(senderID, 10))},
		},
	}

	go func() {
		select {
		case h.MessageQueue <- message:
			h.Logger.Info(
				"kafka message queued for processing",
				"key", string(githubDelivery),
				"topic", h.Topic,
			)
		case <-ctx.Done():
			h.Logger.Info("received context Done")
		default:
			h.Logger.Error(
				"kafka message queue full, dropping message",
				"key", string(githubDelivery),
				"topic", h.Topic,
			)
		}
	}()

	h.Logger.Info("github webhook received successfully")
	ctx.SetStatusCode(fasthttp.StatusAccepted)
}

func (h Handler) checkRequired() error {
	if h.Logger == nil {
		return fmt.Errorf(
			"[githubwebhookhandler.checkRequired] Logger error: [%w, 'nil' received]",
			cerrors.ErrValueRequired,
		)
	}
	if !h.Topic.Valid() {
		return fmt.Errorf(
			"[githubwebhookhandler.checkRequired] Topic error: [%w, '%s' received]",
			cerrors.ErrInvalid, h.Topic,
		)
	}
	if h.Secret == "" {
		return fmt.Errorf(
			"[githubwebhookhandler.checkRequired] Secret error: [%w, empty string received]",
			cerrors.ErrValueRequired,
		)
	}
	if h.MessageQueue == nil {
		return fmt.Errorf(
			"[githubwebhookhandler.checkRequired] MessageQueue error: [%w, 'nil' received]",
			cerrors.ErrValueRequired,
		)
	}

	return nil
}

// Option represents option function type.
type Option func(*Handler) error

// WithLogger sets logger.
func WithLogger(l *slog.Logger) Option {
	return func(h *Handler) error {
		if l == nil {
			return fmt.Errorf(
				"[githubwebhookhandler.WithLogger] error: [%w, 'nil' received]",
				cerrors.ErrValueRequired,
			)
		}
		h.Logger = l

		return nil
	}
}

// WithTopic sets topic name to consume.
func WithTopic(s string) Option {
	return func(h *Handler) error {
		topic := kafkacp.KafkaTopicIdentifier(s)
		if !topic.Valid() {
			return fmt.Errorf(
				"[githubwebhookhandler.WithTopic] error: [%w, '%s' received]",
				cerrors.ErrInvalid, s,
			)
		}
		h.Topic = topic

		return nil
	}
}

// WithWebhookSecret sets github webhook secret.
func WithWebhookSecret(s string) Option {
	return func(h *Handler) error {
		if s == "" {
			return fmt.Errorf(
				"[githubwebhookhandler.WithWebhookSecret] error: [%w, empty string received]",
				cerrors.ErrValueRequired,
			)
		}

		h.Secret = s

		return nil
	}
}

// WithProducerGitHubMessageQueue sets kafka producer message queue for github webhooks.
func WithProducerGitHubMessageQueue(mq chan *sarama.ProducerMessage) Option {
	return func(h *Handler) error {
		if mq == nil {
			return fmt.Errorf(
				"[githubwebhookhandler.WithProducerGitHubMessageQueue] error: [%w, 'nil' received]",
				cerrors.ErrValueRequired,
			)
		}
		h.MessageQueue = mq

		return nil
	}
}

// New instantiates new handler instance.
func New(options ...Option) (*Handler, error) {
	handler := new(Handler)

	for _, option := range options {
		if err := option(handler); err != nil {
			return nil, err
		}
	}

	if err := handler.checkRequired(); err != nil {
		return nil, err
	}

	return handler, nil
}
