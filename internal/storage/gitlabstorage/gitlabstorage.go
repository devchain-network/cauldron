package gitlabstorage

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/IBM/sarama"
	"github.com/devchain-network/cauldron/internal/cerrors"
	"github.com/devchain-network/cauldron/internal/storage"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	_ storage.Pinger        = (*GitLabStorage)(nil) // compile time proof
	_ storage.MessageStorer = (*GitLabStorage)(nil) // compile time proof
	_ storage.PingStorer    = (*GitLabStorage)(nil) // compile time proof
)

// queries.
const (
	GitLabStoreQuery = `
	INSERT INTO gitlab (
		event_uuid,
		webhook_uuid,
		object_kind,
		project_id,
		project_path,
		user_username,
		user_id,
		kafka_offset,
		kafka_partition,
		payload
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`
)

// GitLab represents `gitlab` table model fields.
type GitLab struct {
	Payload        any
	ObjectKind     string
	ProjectPath    *string
	UserUsername   *string
	EventUUID      uuid.UUID
	WebhookUUID    uuid.UUID
	ProjectID      *int64
	UserID         *int64
	KafkaOffset    int64
	KafkaPartition int32
}

// GitLabStorage implements GitLabPingStorer interface.
type GitLabStorage struct {
	Logger      *slog.Logger
	Pool        storage.PGPooler
	DatabaseDSN string
}

func (GitLabStorage) prepareGitLabPayload(message *sarama.ConsumerMessage) (*GitLab, error) {
	gitlabStorage := new(GitLab)
	gitlabStorage.KafkaPartition = message.Partition
	gitlabStorage.KafkaOffset = message.Offset

	messageKey := string(message.Key)
	eventUUID, err := uuid.Parse(messageKey)
	if err != nil {
		return nil, fmt.Errorf(
			"[gitlabstorage.prepareGitLabPayload] eventUUID error: ['%s' received, %w]",
			messageKey, err,
		)
	}
	gitlabStorage.EventUUID = eventUUID

	for _, header := range message.Headers {
		key := string(header.Key)
		value := string(header.Value)

		switch key {
		case "webhook-uuid":
			webhookUUID, uuidErr := uuid.Parse(value)
			if uuidErr != nil {
				return nil, fmt.Errorf(
					"[gitlabstorage.prepareGitLabPayload] webhookUUID error: ['%s' received, %w]",
					value, uuidErr,
				)
			}
			gitlabStorage.WebhookUUID = webhookUUID
		case "object-kind":
			gitlabStorage.ObjectKind = value
		case "project-id":
			if value == "" || value == "0" {
				break
			}
			projectID, parseErr := strconv.ParseInt(value, 10, 64)
			if parseErr != nil {
				return nil, fmt.Errorf(
					"[gitlabstorage.prepareGitLabPayload] projectID error: ['%s' received, %w]",
					value, parseErr,
				)
			}
			gitlabStorage.ProjectID = &projectID
		case "project-path":
			if value != "" {
				gitlabStorage.ProjectPath = &value
			}
		case "user-username":
			if value != "" {
				gitlabStorage.UserUsername = &value
			}
		case "user-id":
			if value == "" || value == "0" {
				break
			}
			userID, parseErr := strconv.ParseInt(value, 10, 64)
			if parseErr != nil {
				return nil, fmt.Errorf(
					"[gitlabstorage.prepareGitLabPayload] userID error: ['%s' received, %w]",
					value, parseErr,
				)
			}
			gitlabStorage.UserID = &userID
		}
	}

	gitlabStorage.Payload = message.Value

	return gitlabStorage, nil
}

// Ping pings database and makes sure db communication is ok.
func (s GitLabStorage) Ping(ctx context.Context, maxRetries uint8, backoff time.Duration) error {
	var pingErr error

	for i := range maxRetries {
		pingErr = s.Pool.Ping(ctx)
		if pingErr == nil {
			s.Logger.Info("successfully pinged the database")

			break
		}

		s.Logger.Error(
			"can not ping the database",
			"error", pingErr,
			"retry", fmt.Sprintf("%d/%d", i, maxRetries),
			"backoff", backoff.String(),
		)
		time.Sleep(backoff)
		backoff *= 2
	}

	if pingErr != nil {
		return fmt.Errorf("[gitlabstorage.Ping] error: [%w]", pingErr)
	}

	return nil
}

// MessageStore stores received kafka message to database.
func (s GitLabStorage) MessageStore(ctx context.Context, message *sarama.ConsumerMessage) error {
	payload, err := s.prepareGitLabPayload(message)
	if err != nil {
		return fmt.Errorf("[gitlabstorage.MessageStore] payload error: [%w]", err)
	}

	_, err = s.Pool.Exec(
		ctx,
		GitLabStoreQuery,
		payload.EventUUID,
		payload.WebhookUUID,
		payload.ObjectKind,
		payload.ProjectID,
		payload.ProjectPath,
		payload.UserUsername,
		payload.UserID,
		payload.KafkaOffset,
		payload.KafkaPartition,
		payload.Payload,
	)
	if err != nil {
		return fmt.Errorf("[gitlabstorage.MessageStore][Pool.Exec] error: [%w]", err)
	}

	return nil
}

func (s GitLabStorage) checkRequired() error {
	if s.Logger == nil {
		return fmt.Errorf(
			"[gitlabstorage.checkRequired] Logger error: [%w, 'nil' received]",
			cerrors.ErrValueRequired,
		)
	}

	if s.DatabaseDSN == "" {
		return fmt.Errorf(
			"[gitlabstorage.checkRequired] DatabaseDSN error: [%w, empty string received]",
			cerrors.ErrValueRequired,
		)
	}

	return nil
}

// Option represents option function type.
type Option func(*GitLabStorage) error

// WithLogger sets logger.
func WithLogger(l *slog.Logger) Option {
	return func(s *GitLabStorage) error {
		if l == nil {
			return fmt.Errorf(
				"[gitlabstorage.WithLogger] error: [%w, 'nil' received]",
				cerrors.ErrValueRequired,
			)
		}
		s.Logger = l

		return nil
	}
}

// WithDatabaseDSN sets database data source.
func WithDatabaseDSN(dsn string) Option {
	return func(s *GitLabStorage) error {
		if dsn == "" {
			return fmt.Errorf(
				"[gitlabstorage.WithDatabaseDSN] error: [%w, empty string received]",
				cerrors.ErrValueRequired,
			)
		}
		s.DatabaseDSN = dsn

		return nil
	}
}

// New instantiates new gitlab storage.
func New(ctx context.Context, options ...Option) (*GitLabStorage, error) {
	gitlabStorage := new(GitLabStorage)

	for _, option := range options {
		if err := option(gitlabStorage); err != nil {
			return nil, err
		}
	}

	if err := gitlabStorage.checkRequired(); err != nil {
		return nil, err
	}

	config, err := pgxpool.ParseConfig(gitlabStorage.DatabaseDSN)
	if err != nil {
		return nil, fmt.Errorf("[gitlabstorage.New][pgxpool.ParseConfig] error: [%w]", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("[gitlabstorage.New][pgxpool.NewWithConfig] error: [%w]", err)
	}

	gitlabStorage.Pool = pool

	return gitlabStorage, nil
}
