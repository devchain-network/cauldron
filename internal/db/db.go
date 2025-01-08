package db

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/devchain-network/cauldron/internal/cerrors"
	"github.com/jackc/pgx/v5/pgxpool"
)

var _ DB = (*PgxDatabase)(nil) // compile time proof

// DB ...
type DB interface {
	Insert() error
}

// Option represents option function type.
type Option func(*PgxDatabase) error

// PgxDatabase ...
type PgxDatabase struct {
	Logger     *slog.Logger
	Pool       *pgxpool.Pool
	DSN        string
	Backoff    time.Duration
	MaxRetries uint8
}

// Insert ...
func (d *PgxDatabase) Insert() error {
	fmt.Println(d)

	return nil
}

// WithDSN sets connection dsn.
func WithDSN(s string) Option {
	return func(pgxdb *PgxDatabase) error {
		if s == "" {
			return fmt.Errorf("db.WithDSN error: [%w]", cerrors.ErrValueRequired)
		}
		pgxdb.DSN = s

		return nil
	}
}

// WithBackoff sets backoff duration.
func WithBackoff(d time.Duration) Option {
	return func(pgxdb *PgxDatabase) error {
		if d == 0 {
			return fmt.Errorf("db.WithBackoff error: [%w]", cerrors.ErrValueRequired)
		}
		pgxdb.Backoff = d

		return nil
	}
}

// WithMaxRetries sets max retries value.
func WithMaxRetries(i int) Option {
	return func(pgxdb *PgxDatabase) error {
		if i > 255 || i < 0 {
			return fmt.Errorf("db.WithMaxRetries error: [%w]", cerrors.ErrInvalid)
		}
		pgxdb.MaxRetries = uint8(i)

		return nil
	}
}

// WithLogger sets logger.
func WithLogger(l *slog.Logger) Option {
	return func(pgxdb *PgxDatabase) error {
		if l == nil {
			return fmt.Errorf("db.WithLogger error: [%w]", cerrors.ErrValueRequired)
		}
		pgxdb.Logger = l

		return nil
	}
}

// New instantiates new pgx connection pool.
func New(options ...Option) (*PgxDatabase, error) {
	pgxDatabase := new(PgxDatabase)

	for _, option := range options {
		if err := option(pgxDatabase); err != nil {
			return nil, fmt.Errorf("db.New option error: [%w]", err)
		}
	}

	if pgxDatabase.DSN == "" {
		return nil, fmt.Errorf("db.New pgxDatabase.DSN error: [%w]", cerrors.ErrValueRequired)
	}

	config, err := pgxpool.ParseConfig(pgxDatabase.DSN)
	if err != nil {
		return nil, fmt.Errorf("db.New pgxpool.ParseConfig error: [%w]", err)
	}

	ctx := context.Background()
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("db.New pgxpool.NewWithConfig error: [%w]", err)
	}

	var pingErr error
	backoff := pgxDatabase.Backoff
	for i := range pgxDatabase.MaxRetries {
		pingErr = pool.Ping(ctx)
		if pingErr == nil {
			break
		}

		pgxDatabase.Logger.Error(
			"can not ping datavase",
			"error",
			pingErr,
			"retry",
			fmt.Sprintf("%d/%d", i, pgxDatabase.MaxRetries),
			"backoff",
			backoff.String(),
		)
		time.Sleep(backoff)
		backoff *= 2
	}

	if pingErr != nil {
		return nil, fmt.Errorf("db.New ping error: [%w]", pingErr)
	}

	return pgxDatabase, nil
}
