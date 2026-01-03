package gitlabwebhookhandler

import (
	"fmt"
	"log/slog"
	"strconv"

	"github.com/IBM/sarama"
	"github.com/buger/jsonparser"
	"github.com/devchain-network/cauldron/internal/cerrors"
	"github.com/devchain-network/cauldron/internal/kafkacp"
	"github.com/devchain-network/cauldron/internal/transport/http/httphandler"
	"github.com/valyala/fasthttp"
)

var _ GitLabWebhookHandler = (*Handler)(nil) // compile time proof

const (
	missingHTTPHandlerHeaderText   = "header"
	missingHTTPHandlerErrorMessage = "missing http header"
)

// GitLabWebhookHandler defines http handler behaviours.
type GitLabWebhookHandler interface {
	httphandler.FastHTTPHandler
}

// Handler represents http handler configuration and must satisfy GitLabWebhookHandler interface.
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

	// GitLab uses plain text token comparison (not HMAC)
	gitlabToken := string(ctx.Request.Header.Peek("X-Gitlab-Token"))
	if gitlabToken != h.Secret {
		h.Logger.Error("invalid gitlab token")
		ctx.SetStatusCode(fasthttp.StatusBadRequest)

		return
	}

	gitlabEventUUID := ctx.Request.Header.Peek("X-Gitlab-Event-Uuid")
	if len(gitlabEventUUID) == 0 {
		h.Logger.Error(missingHTTPHandlerErrorMessage, missingHTTPHandlerHeaderText, "X-Gitlab-Event-Uuid")
		ctx.SetStatusCode(fasthttp.StatusBadRequest)

		return
	}

	gitlabWebhookUUID := ctx.Request.Header.Peek("X-Gitlab-Webhook-Uuid")
	if len(gitlabWebhookUUID) == 0 {
		h.Logger.Error(missingHTTPHandlerErrorMessage, missingHTTPHandlerHeaderText, "X-Gitlab-Webhook-Uuid")
		ctx.SetStatusCode(fasthttp.StatusBadRequest)

		return
	}

	// object_kind or event_name - one of them is required
	objectKind, err := jsonparser.GetString(ctx.PostBody(), "object_kind")
	if err != nil {
		// fallback to event_name for premium/ultimate hooks (group, project, subgroup events)
		objectKind, err = jsonparser.GetString(ctx.PostBody(), "event_name")
		if err != nil {
			h.Logger.Error("objectKind/eventName jsonparser.GetString error", "error", err)
			ctx.SetStatusCode(fasthttp.StatusBadRequest)

			return
		}
	}

	// project.id - try nested first, fallback to top-level (for premium events like project_create)
	projectID, err := jsonparser.GetInt(ctx.PostBody(), "project", "id")
	if err != nil {
		projectID, _ = jsonparser.GetInt(ctx.PostBody(), "project_id")
	}

	// project.path_with_namespace - try nested first, fallback for premium events
	projectPath, err := jsonparser.GetString(ctx.PostBody(), "project", "path_with_namespace")
	if err != nil {
		// try top-level path_with_namespace (project_create)
		projectPath, err = jsonparser.GetString(ctx.PostBody(), "path_with_namespace")
		if err != nil {
			// try full_path (subgroup_create)
			projectPath, err = jsonparser.GetString(ctx.PostBody(), "full_path")
			if err != nil {
				// try group_path (user_add_to_group)
				projectPath, _ = jsonparser.GetString(ctx.PostBody(), "group_path")
			}
		}
	}

	// user info - try nested first (user.id, user.username), fallback to flat (user_id, user_username)
	userID, err := jsonparser.GetInt(ctx.PostBody(), "user", "id")
	if err != nil {
		userID, _ = jsonparser.GetInt(ctx.PostBody(), "user_id")
	}

	userUsername, err := jsonparser.GetString(ctx.PostBody(), "user", "username")
	if err != nil {
		userUsername, _ = jsonparser.GetString(ctx.PostBody(), "user_username")
	}

	h.Logger.Info("received gitlab webhook",
		"objectKind", objectKind,
		"userUsername", userUsername,
		"userID", userID,
		"projectPath", projectPath,
	)

	message := &sarama.ProducerMessage{
		Topic: h.Topic.String(),
		Key:   sarama.StringEncoder(gitlabEventUUID),
		Value: sarama.ByteEncoder(ctx.PostBody()),
		Headers: []sarama.RecordHeader{
			{Key: []byte("event-uuid"), Value: gitlabEventUUID},
			{Key: []byte("webhook-uuid"), Value: gitlabWebhookUUID},
			{Key: []byte("object-kind"), Value: []byte(objectKind)},
			{Key: []byte("project-id"), Value: []byte(strconv.FormatInt(projectID, 10))},
			{Key: []byte("project-path"), Value: []byte(projectPath)},
			{Key: []byte("user-username"), Value: []byte(userUsername)},
			{Key: []byte("user-id"), Value: []byte(strconv.FormatInt(userID, 10))},
		},
	}

	go func() {
		select {
		case h.MessageQueue <- message:
			h.Logger.Info(
				"kafka message queued for processing",
				"key", string(gitlabEventUUID),
				"topic", h.Topic,
			)
		case <-ctx.Done():
			h.Logger.Info("received context Done")
		default:
			h.Logger.Error(
				"kafka message queue full, dropping message",
				"key", string(gitlabEventUUID),
				"topic", h.Topic,
			)
		}
	}()

	h.Logger.Info("gitlab webhook received successfully")
	ctx.SetStatusCode(fasthttp.StatusOK)
}

func (h Handler) checkRequired() error {
	if h.Logger == nil {
		return fmt.Errorf(
			"[gitlabwebhookhandler.checkRequired] Logger error: [%w, 'nil' received]",
			cerrors.ErrValueRequired,
		)
	}
	if !h.Topic.Valid() {
		return fmt.Errorf(
			"[gitlabwebhookhandler.checkRequired] Topic error: [%w, '%s' received]",
			cerrors.ErrInvalid, h.Topic,
		)
	}
	if h.Secret == "" {
		return fmt.Errorf(
			"[gitlabwebhookhandler.checkRequired] Secret error: [%w, empty string received]",
			cerrors.ErrValueRequired,
		)
	}
	if h.MessageQueue == nil {
		return fmt.Errorf(
			"[gitlabwebhookhandler.checkRequired] MessageQueue error: [%w, 'nil' received]",
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
				"[gitlabwebhookhandler.WithLogger] error: [%w, 'nil' received]",
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
				"[gitlabwebhookhandler.WithTopic] error: [%w, '%s' received]",
				cerrors.ErrInvalid, s,
			)
		}
		h.Topic = topic

		return nil
	}
}

// WithWebhookSecret sets gitlab webhook secret.
func WithWebhookSecret(s string) Option {
	return func(h *Handler) error {
		if s == "" {
			return fmt.Errorf(
				"[gitlabwebhookhandler.WithWebhookSecret] error: [%w, empty string received]",
				cerrors.ErrValueRequired,
			)
		}

		h.Secret = s

		return nil
	}
}

// WithProducerGitLabMessageQueue sets kafka producer message queue for gitlab webhooks.
func WithProducerGitLabMessageQueue(mq chan *sarama.ProducerMessage) Option {
	return func(h *Handler) error {
		if mq == nil {
			return fmt.Errorf(
				"[gitlabwebhookhandler.WithProducerGitLabMessageQueue] error: [%w, 'nil' received]",
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
