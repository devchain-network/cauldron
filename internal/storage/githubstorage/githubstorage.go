package githubstorage

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/devchain-network/cauldron/internal/cerrors"
	"github.com/devchain-network/cauldron/internal/storage"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	_ storage.Pinger     = (*GitHubStorage)(nil) // compile time proof
	_ storage.Storer     = (*GitHubStorage)(nil) // compile time proof
	_ storage.PingStorer = (*GitHubStorage)(nil) // compile time proof
)

// queries.
const (
	GitHubStoreQuery = `
	INSERT INTO github (
		delivery_id, 
		event, 
		target, 
		target_id, 
		hook_id, 
		user_login, 
		user_id, 
		"offset", 
		partition, 
		payload
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`
)

// GitHub represents `github` table model fields.
type GitHub struct {
	Payload    any
	Event      string
	Target     string
	UserLogin  string
	DeliveryID uuid.UUID
	TargetID   uint64
	HookID     uint64
	UserID     int64
	Offset     int64
	Partition  int32
}

// GitHubStorage implements GitHubPingStorer interface.
type GitHubStorage struct {
	Logger      *slog.Logger
	Pool        storage.PGPooler
	DatabaseDSN string
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
		return fmt.Errorf("storage.GitHubStorage.Ping error: [%w]", pingErr)
	}

	return nil
}

// Store stores given github webhook data with extras to database.
func (s GitHubStorage) Store(ctx context.Context, payload any) error {
	githubPayload, ok := payload.(*GitHub)
	if !ok {
		return fmt.Errorf("githubstorage.GitHubStorage.Store payload error: [%w]", cerrors.ErrInvalid)
	}
	_, err := s.Pool.Exec(
		ctx,
		GitHubStoreQuery,
		githubPayload.DeliveryID,
		githubPayload.Event,
		githubPayload.Target,
		githubPayload.TargetID,
		githubPayload.HookID,
		githubPayload.UserLogin,
		githubPayload.UserID,
		githubPayload.Offset,
		githubPayload.Partition,
		githubPayload.Payload,
	)
	if err != nil {
		return fmt.Errorf("githubstorage.GitHubStorage.Store error: [%w]", err)
	}

	return nil
}

func (s GitHubStorage) checkRequired() error {
	if s.Logger == nil {
		return fmt.Errorf("githubstorage.New Logger error: [%w]", cerrors.ErrValueRequired)
	}

	if s.DatabaseDSN == "" {
		return fmt.Errorf("githubstorage.New DatabaseDSN error: [%w]", cerrors.ErrValueRequired)
	}

	return nil
}

// Option represents option function type.
type Option func(*GitHubStorage) error

// WithLogger sets logger.
func WithLogger(l *slog.Logger) Option {
	return func(s *GitHubStorage) error {
		if l == nil {
			return fmt.Errorf("githubstorage.WithLogger error: [%w]", cerrors.ErrValueRequired)
		}
		s.Logger = l

		return nil
	}
}

// WithDatabaseDSN sets database data source.
func WithDatabaseDSN(dsn string) Option {
	return func(s *GitHubStorage) error {
		if dsn == "" {
			return fmt.Errorf("githubstorage.WithDatabaseDSN error: [%w]", cerrors.ErrValueRequired)
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
			return nil, fmt.Errorf("githubstorage.New option error: [%w]", err)
		}
	}

	if err := githubStorage.checkRequired(); err != nil {
		return nil, err
	}

	config, err := pgxpool.ParseConfig(githubStorage.DatabaseDSN)
	if err != nil {
		return nil, fmt.Errorf("githubstorage.New pgxpool.ParseConfig error: [%w]", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("githubstorage.New pgxpool.NewWithConfig error: [%w]", err)
	}

	githubStorage.Pool = pool

	return githubStorage, nil
}
