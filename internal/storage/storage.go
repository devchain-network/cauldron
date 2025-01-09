package storage

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/devchain-network/cauldron/internal/cerrors"
	"github.com/jackc/pgx/v5/pgxpool"
)

// constants.
const (
	DefaultDBPingBackoff    = 2 * time.Second
	DefaultDBPingMaxRetries = 10
)

var _ Storer = (*Manager)(nil) // compile time proof

// GitHubStorer defines storage behaviours for github webhook.
type GitHubStorer interface {
	GitHubStore(payload *GitHubWebhookData) error
}

// Storer defines storage behaviours for different webhooks.
type Storer interface {
	GitHubStorer
}

// Manager represents database manager.
type Manager struct {
	logger     *slog.Logger
	Pool       *pgxpool.Pool
	dsn        string
	backOff    time.Duration
	maxRetries uint8
}

// Option represents option function type.
type Option func(*Manager) error

// GitHubStore stores given github webhook data with extras to database.
func (m *Manager) GitHubStore(data *GitHubWebhookData) error {
	_, err := m.Pool.Exec(
		context.Background(),
		githubWebhookQuery,
		data.DeliveryID,
		data.Event,
		data.Target,
		data.TargetID,
		data.HookID,
		data.UserLogin,
		data.UserID,
		data.Offset,
		data.Partition,
		data.Payload,
	)
	if err != nil {
		return fmt.Errorf("storage.GitHubStore error: [%w]", err)
	}

	return nil
}

// WithDSN sets connection dsn.
func WithDSN(s string) Option {
	return func(m *Manager) error {
		if s == "" {
			return fmt.Errorf("storage.WithDSN error: [%w]", cerrors.ErrValueRequired)
		}
		m.dsn = s

		return nil
	}
}

// WithBackoff sets backoff duration.
func WithBackoff(d time.Duration) Option {
	return func(m *Manager) error {
		if d == 0 {
			return fmt.Errorf("storage.WithBackoff error: [%w]", cerrors.ErrValueRequired)
		}
		m.backOff = d

		return nil
	}
}

// WithMaxRetries sets max retries value.
func WithMaxRetries(i int) Option {
	return func(m *Manager) error {
		if i > 255 || i < 0 {
			return fmt.Errorf("storage.WithMaxRetries error: [%w]", cerrors.ErrInvalid)
		}
		m.maxRetries = uint8(i)

		return nil
	}
}

// WithLogger sets logger.
func WithLogger(l *slog.Logger) Option {
	return func(m *Manager) error {
		if l == nil {
			return fmt.Errorf("storage.WithLogger error: [%w]", cerrors.ErrValueRequired)
		}
		m.logger = l

		return nil
	}
}

// New instantiates new pgx connection pool.
func New(options ...Option) (*Manager, error) {
	manager := new(Manager)
	manager.backOff = DefaultDBPingBackoff
	manager.maxRetries = DefaultDBPingMaxRetries

	for _, option := range options {
		if err := option(manager); err != nil {
			return nil, fmt.Errorf("db.New option error: [%w]", err)
		}
	}

	if manager.dsn == "" {
		return nil, fmt.Errorf("storage.New manager.dsn error: [%w]", cerrors.ErrValueRequired)
	}

	if manager.logger == nil {
		return nil, fmt.Errorf("storage.New manager.logger error: [%w]", cerrors.ErrValueRequired)
	}

	config, err := pgxpool.ParseConfig(manager.dsn)
	if err != nil {
		return nil, fmt.Errorf("storage.New pgxpool.ParseConfig error: [%w]", err)
	}

	ctx := context.Background()
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("storage.New pgxpool.NewWithConfig error: [%w]", err)
	}

	var pingErr error
	backOff := manager.backOff
	for i := range manager.maxRetries {
		pingErr = pool.Ping(ctx)
		if pingErr == nil {
			manager.logger.Info("connected to database")

			break
		}

		manager.logger.Error(
			"can not ping database",
			"error", pingErr,
			"retry", fmt.Sprintf("%d/%d", i, manager.maxRetries),
			"backoff", backOff.String(),
		)
		time.Sleep(backOff)
		backOff *= 2
	}

	if pingErr != nil {
		return nil, fmt.Errorf("storage.New ping error: [%w]", pingErr)
	}

	manager.Pool = pool

	return manager, nil
}
