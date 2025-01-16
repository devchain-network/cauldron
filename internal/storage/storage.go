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

var (
	_ Storer       = (*Manager)(nil) // compile time proof
	_ GitHubStorer = (*Manager)(nil) // compile time proof
)

// GitHubStorer defines storage behaviours for github webhook.
type GitHubStorer interface {
	GitHubStore(payload *GitHubWebhookData) error
}

// Storer defines storage behaviours for different webhooks.
type Storer interface {
	Ping() error
	GitHubStorer
}

// Manager represents database manager.
type Manager struct {
	Logger     *slog.Logger
	Pool       *pgxpool.Pool
	DSN        string
	BackOff    time.Duration
	MaxRetries uint8
}

// Option represents option function type.
type Option func(*Manager) error

// GitHubStore stores given github webhook data with extras to database.
func (m Manager) GitHubStore(data *GitHubWebhookData) error {
	_, err := m.Pool.Exec(
		context.Background(),
		gitHubStoreQuery,
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

// Ping implements a retry and backoff mechanism to check whether a database
// connection can be established.
func (m Manager) Ping() error {
	var pingErr error
	backOff := m.BackOff
	ctx := context.Background()

	for i := range m.MaxRetries {
		pingErr = m.Pool.Ping(ctx)
		if pingErr == nil {
			m.Logger.Info("connected to database")

			break
		}
		m.Logger.Error(
			"can not ping database",
			"error", pingErr,
			"retry", fmt.Sprintf("%d/%d", i, m.MaxRetries),
			"backoff", backOff.String(),
		)
		time.Sleep(backOff)
		backOff *= 2
	}

	if pingErr != nil {
		return fmt.Errorf("storage.ping error: [%w]", pingErr)
	}

	return nil
}

// WithDSN sets connection dsn.
func WithDSN(s string) Option {
	return func(m *Manager) error {
		if s == "" {
			return fmt.Errorf("storage.WithDSN error: [%w]", cerrors.ErrValueRequired)
		}
		m.DSN = s

		return nil
	}
}

// WithBackoff sets backoff duration.
func WithBackoff(d time.Duration) Option {
	return func(m *Manager) error {
		if d == 0 {
			return fmt.Errorf("storage.WithBackoff error: [%w]", cerrors.ErrValueRequired)
		}
		m.BackOff = d

		return nil
	}
}

// WithMaxRetries sets max retries value.
func WithMaxRetries(i int) Option {
	return func(m *Manager) error {
		if i > 255 || i < 0 {
			return fmt.Errorf("storage.WithMaxRetries error: [%w]", cerrors.ErrInvalid)
		}
		m.MaxRetries = uint8(i)

		return nil
	}
}

// WithLogger sets logger.
func WithLogger(l *slog.Logger) Option {
	return func(m *Manager) error {
		if l == nil {
			return fmt.Errorf("storage.WithLogger error: [%w]", cerrors.ErrValueRequired)
		}
		m.Logger = l

		return nil
	}
}

// New instantiates new pgx connection pool.
func New(options ...Option) (*Manager, error) {
	manager := new(Manager)
	manager.BackOff = DefaultDBPingBackoff
	manager.MaxRetries = DefaultDBPingMaxRetries

	for _, option := range options {
		if err := option(manager); err != nil {
			return nil, fmt.Errorf("db.New option error: [%w]", err)
		}
	}

	if manager.DSN == "" {
		return nil, fmt.Errorf("storage.New manager.DSN error: [%w]", cerrors.ErrValueRequired)
	}

	if manager.Logger == nil {
		return nil, fmt.Errorf("storage.New manager.Logger error: [%w]", cerrors.ErrValueRequired)
	}

	config, err := pgxpool.ParseConfig(manager.DSN)
	if err != nil {
		return nil, fmt.Errorf("storage.New pgxpool.ParseConfig error: [%w]", err)
	}

	ctx := context.Background()
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("storage.New pgxpool.NewWithConfig error: [%w]", err)
	}

	manager.Pool = pool

	return manager, nil
}
