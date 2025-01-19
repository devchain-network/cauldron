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
	_ storage.Pinger   = (*GitHubStorage)(nil) // compile time proof
	_ GitHubPingStorer = (*GitHubStorage)(nil) // compile time proof
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

// GitHubPingStorer ...
type GitHubPingStorer interface {
	storage.Pinger
	Store(payload *GitHub) error
}

// GitHubStorage ...
type GitHubStorage struct {
	Logger *slog.Logger
	Pool   *pgxpool.Pool
}

// Ping ...
func (s GitHubStorage) Ping(maxRetries uint8, backoff time.Duration) error {
	var pingErr error

	ctx, cancel := context.WithTimeout(context.Background(), storage.DefaultDBPingTimeout)
	defer cancel()

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

// Store ...
func (s GitHubStorage) Store(payload *GitHub) error {
	fmt.Println(payload)
	fmt.Println(gitHubStoreQuery)
	fmt.Println(s.Logger)

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

const gitHubStoreQuery = `
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
