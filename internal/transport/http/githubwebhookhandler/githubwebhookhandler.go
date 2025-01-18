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
	"github.com/devchain-network/cauldron/internal/kafkaconsumer"
	"github.com/devchain-network/cauldron/internal/transport/http/httphandler"
	"github.com/valyala/fasthttp"
)

var _ GitHubWebhookHandler = (*Handler)(nil) // compile time proof

// GitHubWebhookHandler defines http handler behaviours.
type GitHubWebhookHandler interface {
	httphandler.FastHTTPHandler
}

// Handler represents http handler configuration and must satisfy GitHubWebhookHandler interface.
type Handler struct {
	MessageQueue chan *sarama.ProducerMessage
	Logger       *slog.Logger
	Topic        kafkaconsumer.KafkaTopicIdentifier
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
		h.Logger.Error("invalid hmac", "want", expectedMAC, "got", hubSignature)
		ctx.SetStatusCode(fasthttp.StatusBadRequest)

		return
	}

	githubEvent := ctx.Request.Header.Peek("X-Github-Event")
	githubDelivery := ctx.Request.Header.Peek("X-Github-Delivery")
	githubHookID := ctx.Request.Header.Peek("X-Github-Hook-Id")
	githubHookInstallationTargetID := ctx.Request.Header.Peek("X-Github-Hook-Installation-Target-Id")
	githubHookInstallationTargetType := ctx.Request.Header.Peek("X-Github-Hook-Installation-Target-Type")

	senderLogin, err := jsonparser.GetString(ctx.PostBody(), "sender", "login")
	if err != nil {
		h.Logger.Error("senderLogin jsonparser.GetString error", "error", err)
	}
	senderID, err := jsonparser.GetInt(ctx.PostBody(), "sender", "id")
	if err != nil {
		h.Logger.Error("senderID jsonparser.GetInt error", "error", err)
	}

	fmt.Println("senderLogin", senderLogin, "senderID", senderID)

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

	h.MessageQueue <- message
	h.Logger.Info(
		"kafka message queued for process",
		"key", string(githubDelivery),
		"topic", h.Topic,
	)

	ctx.SetStatusCode(fasthttp.StatusAccepted)
}

// Option represents option function type.
type Option func(*Handler) error

// WithLogger sets logger.
func WithLogger(l *slog.Logger) Option {
	return func(h *Handler) error {
		if l == nil {
			return fmt.Errorf("githubwebhookhandler.WithLogger error: [%w]", cerrors.ErrValueRequired)
		}
		h.Logger = l

		return nil
	}
}

// WithTopic sets topic name to consume.
func WithTopic(s kafkaconsumer.KafkaTopicIdentifier) Option {
	return func(h *Handler) error {
		if err := kafkaconsumer.IsKafkaTopicValid(s); err != nil {
			return fmt.Errorf("githubwebhookhandler.WithTopic h.Topic error: [%w]", err)
		}

		h.Topic = s

		return nil
	}
}

// WithWebhookSecret sets github webhook secret.
func WithWebhookSecret(s string) Option {
	return func(h *Handler) error {
		if s == "" {
			return fmt.Errorf("githubwebhookhandler.WithWebhookSecret h.Secret error: [%w]", cerrors.ErrValueRequired)
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
				"githubwebhookhandler.WithProducerGitHubMessageQueue error: [%w]",
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
			return nil, fmt.Errorf("githubwebhookhandler.New option error: [%w]", err)
		}
	}

	return handler, nil
}
