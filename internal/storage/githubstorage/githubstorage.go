package githubstorage

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
	_ storage.Pinger        = (*GitHubStorage)(nil) // compile time proof
	_ storage.MessageStorer = (*GitHubStorage)(nil) // compile time proof
	_ storage.PingStorer    = (*GitHubStorage)(nil) // compile time proof
)

// queries.
const (
	GitHubStoreQuery = `
	INSERT INTO github (
		delivery_id, 
		event, 
		target_type, 
		target_id, 
		hook_id, 
		user_login, 
		user_id, 
		kafka_offset, 
		kafka_partition, 
		payload
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`
)

// GitHub represents `github` table model fields.
type GitHub struct {
	Payload        any
	Event          string
	TargetType     string
	UserLogin      string
	DeliveryID     uuid.UUID
	TargetID       uint64
	HookID         uint64
	UserID         int64
	KafkaOffset    int64
	KafkaPartition int32
}

// GitHubStorage implements GitHubPingStorer interface.
type GitHubStorage struct {
	Logger      *slog.Logger
	Pool        storage.PGPooler
	DatabaseDSN string
}

func (GitHubStorage) prepareGitHubPayload(message *sarama.ConsumerMessage) (*GitHub, error) {
	githubStorage := new(GitHub)

	deliveryID, err := uuid.Parse(string(message.Key))
	if err != nil {
		return nil, fmt.Errorf("githubstorage prepareGitHubPayload deliveryID error: [%w]", err)
	}
	githubStorage.DeliveryID = deliveryID

	var targetID uint64
	var targetIDErr error

	var hookID uint64
	var hookIDErr error

	var userID int64
	var userIDErr error

	for _, header := range message.Headers {
		key := string(header.Key)
		value := string(header.Value)

		switch key {
		case "event":
			githubStorage.Event = value
		case "target-type":
			githubStorage.TargetType = value
		case "target-id":
			targetID, targetIDErr = strconv.ParseUint(value, 10, 64)
			if targetIDErr != nil {
				return nil, fmt.Errorf("githubstorage prepareGitHubPayload targetID error: [%w]", targetIDErr)
			}
			githubStorage.TargetID = targetID
		case "hook-id":
			hookID, hookIDErr = strconv.ParseUint(value, 10, 64)
			if hookIDErr != nil {
				return nil, fmt.Errorf("githubstorage prepareGitHubPayload hookID error: [%w]", hookIDErr)
			}
			githubStorage.HookID = hookID
		case "sender-login":
			githubStorage.UserLogin = value
		case "sender-id":
			userID, userIDErr = strconv.ParseInt(value, 10, 64)
			if userIDErr != nil {
				return nil, fmt.Errorf("githubstorage prepareGitHubPayload userID error: [%w]", userIDErr)
			}
			githubStorage.UserID = userID
		}
	}

	githubStorage.Payload = message.Value

	return githubStorage, nil
}

// Ping pings database and makes sure db communication is ok.
func (s GitHubStorage) Ping(ctx context.Context, maxRetries uint8, backoff time.Duration) error {
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
		return fmt.Errorf("githubstorage Ping error: [%w]", pingErr)
	}

	return nil
}

// MessageStore stores received kafka message to database.
func (s GitHubStorage) MessageStore(ctx context.Context, message *sarama.ConsumerMessage) error {
	payload, err := s.prepareGitHubPayload(message)
	if err != nil {
		return fmt.Errorf("githubstorage Store payload error: [%w]", err)
	}

	_, err = s.Pool.Exec(
		ctx,
		GitHubStoreQuery,
		payload.DeliveryID,
		payload.Event,
		payload.TargetType,
		payload.TargetID,
		payload.HookID,
		payload.UserLogin,
		payload.UserID,
		payload.KafkaOffset,
		payload.KafkaPartition,
		payload.Payload,
	)
	if err != nil {
		return fmt.Errorf("githubstorage Store Pool.Exec error: [%w]", err)
	}

	return nil
}

func (s GitHubStorage) checkRequired() error {
	if s.Logger == nil {
		return fmt.Errorf("githubstorage check required, Logger error: [%w]", cerrors.ErrValueRequired)
	}

	if s.DatabaseDSN == "" {
		return fmt.Errorf("githubstorage check required, DatabaseDSN error: [%w]", cerrors.ErrValueRequired)
	}

	return nil
}

// Option represents option function type.
type Option func(*GitHubStorage) error

// WithLogger sets logger.
func WithLogger(l *slog.Logger) Option {
	return func(s *GitHubStorage) error {
		if l == nil {
			return fmt.Errorf("githubstorage WithLogger error: [%w]", cerrors.ErrValueRequired)
		}
		s.Logger = l

		return nil
	}
}

// WithDatabaseDSN sets database data source.
func WithDatabaseDSN(dsn string) Option {
	return func(s *GitHubStorage) error {
		if dsn == "" {
			return fmt.Errorf("githubstorage WithDatabaseDSN error: [%w]", cerrors.ErrValueRequired)
		}
		s.DatabaseDSN = dsn

		return nil
	}
}

// New instantiates new github storage.
func New(ctx context.Context, options ...Option) (*GitHubStorage, error) {
	githubStorage := new(GitHubStorage)

	for _, option := range options {
		if err := option(githubStorage); err != nil {
			return nil, fmt.Errorf("githubstorage option error: [%w]", err)
		}
	}

	if err := githubStorage.checkRequired(); err != nil {
		return nil, err
	}

	config, err := pgxpool.ParseConfig(githubStorage.DatabaseDSN)
	if err != nil {
		return nil, fmt.Errorf("githubstorage pgxpool.ParseConfig error: [%w]", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("githubstorage pgxpool.NewWithConfig error: [%w]", err)
	}

	githubStorage.Pool = pool

	return githubStorage, nil
}
