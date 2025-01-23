package githubstorage_test

import (
	"context"
	"log/slog"
	"strconv"
	"testing"

	"github.com/devchain-network/cauldron/internal/cerrors"
	"github.com/devchain-network/cauldron/internal/storage"
	"github.com/devchain-network/cauldron/internal/storage/githubstorage"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
)

type mockLogger struct{}

func (h *mockLogger) Enabled(_ context.Context, _ slog.Level) bool {
	return true
}

func (h *mockLogger) Handle(_ context.Context, record slog.Record) error {
	return nil
}

func (h *mockLogger) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h
}

func (h *mockLogger) WithGroup(name string) slog.Handler {
	return h
}

func TestNew_NoLogger(t *testing.T) {
	// logger := slog.New(new(mockLogger))

	ctx, cancel := context.WithTimeout(context.Background(), storage.DefaultDBPingTimeout)
	defer cancel()

	db, err := githubstorage.New(
		ctx,
	)

	assert.ErrorIs(t, err, cerrors.ErrValueRequired)
	assert.Nil(t, db)
}

func TestNew_NilLogger(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), storage.DefaultDBPingTimeout)
	defer cancel()

	db, err := githubstorage.New(
		ctx,
		githubstorage.WithLogger(nil),
	)

	assert.ErrorIs(t, err, cerrors.ErrValueRequired)
	assert.Nil(t, db)
}

func TestNew_NoDSN(t *testing.T) {
	logger := slog.New(new(mockLogger))

	ctx, cancel := context.WithTimeout(context.Background(), storage.DefaultDBPingTimeout)
	defer cancel()

	db, err := githubstorage.New(
		ctx,
		githubstorage.WithLogger(logger),
	)

	assert.ErrorIs(t, err, cerrors.ErrValueRequired)
	assert.Nil(t, db)
}

func TestNew_EmptyDSN(t *testing.T) {
	logger := slog.New(new(mockLogger))

	ctx, cancel := context.WithTimeout(context.Background(), storage.DefaultDBPingTimeout)
	defer cancel()

	db, err := githubstorage.New(
		ctx,
		githubstorage.WithLogger(logger),
		githubstorage.WithDatabaseDSN(""),
	)

	assert.ErrorIs(t, err, cerrors.ErrValueRequired)
	assert.Nil(t, db)
}

func TestNew_InvalidDSN(t *testing.T) {
	logger := slog.New(new(mockLogger))

	ctx, cancel := context.WithTimeout(context.Background(), storage.DefaultDBPingTimeout)
	defer cancel()

	db, err := githubstorage.New(
		ctx,
		githubstorage.WithLogger(logger),
		githubstorage.WithDatabaseDSN("foo:://bar"),
	)

	var parseErr *pgconn.ParseConfigError
	assert.ErrorAs(t, err, &parseErr)
	assert.Nil(t, db)
}

func TestNew_InvalidPoolConfig(t *testing.T) {
	logger := slog.New(new(mockLogger))

	ctx, cancel := context.WithTimeout(context.Background(), storage.DefaultDBPingTimeout)
	defer cancel()

	db, err := githubstorage.New(
		ctx,
		githubstorage.WithLogger(logger),
		githubstorage.WithDatabaseDSN(
			"postgres://foo:bar@localhost:5432/fake?sslmode=disable&pool_min_conns=abc",
		),
	)

	var parseErr *strconv.NumError
	assert.ErrorAs(t, err, &parseErr)

	assert.Nil(t, db)
}
