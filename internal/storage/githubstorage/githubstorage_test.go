package githubstorage_test

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/devchain-network/cauldron/internal/cerrors"
	"github.com/devchain-network/cauldron/internal/storage/githubstorage"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockPool struct {
	mock.Mock
}

func (m *MockPool) Ping(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockPool) Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
	args := m.Called(ctx, sql, arguments)
	return args.Get(0).(pgconn.CommandTag), args.Error(1)
}

func (m *MockPool) Close() {
	m.Called()
}

func (m *MockPool) Acquire(ctx context.Context) (*pgxpool.Conn, error) {
	args := m.Called(ctx)
	return args.Get(0).(*pgxpool.Conn), args.Error(1)
}

func (m *MockPool) AcquireAllIdle(ctx context.Context) ([]*pgxpool.Conn, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*pgxpool.Conn), args.Error(1)
}

func (m *MockPool) AcquireFunc(ctx context.Context, f func(*pgxpool.Conn) error) error {
	args := m.Called(ctx, f)
	return args.Error(0)
}

func TestNew(t *testing.T) {
	t.Run("success with valid options", func(t *testing.T) {
		dsn := "postgresql://user:pass@localhost/db"
		ctx := context.Background()
		logger := slog.Default()

		storage, err := githubstorage.New(ctx,
			githubstorage.WithDatabaseDSN(dsn),
			githubstorage.WithLogger(logger),
		)

		assert.NoError(t, err)
		assert.NotNil(t, storage)
		assert.Equal(t, dsn, storage.DatabaseDSN)
		assert.NotNil(t, storage.Logger)
		assert.NotNil(t, storage.Pool)
	})

	t.Run("missing required fields", func(t *testing.T) {
		ctx := context.Background()

		storage, err := githubstorage.New(ctx)
		assert.Error(t, err)
		assert.Nil(t, storage)
		assert.ErrorIs(t, err, cerrors.ErrValueRequired)
	})

	t.Run("missing required fields - dsn", func(t *testing.T) {
		ctx := context.Background()

		storage, err := githubstorage.New(ctx,
			githubstorage.WithLogger(slog.Default()),
		)
		assert.Error(t, err)
		assert.Nil(t, storage)
		assert.ErrorIs(t, err, cerrors.ErrValueRequired)
	})

	t.Run("invalid DSN format", func(t *testing.T) {
		ctx := context.Background()

		storage, err := githubstorage.New(ctx,
			githubstorage.WithDatabaseDSN("invalid-dsn"),
			githubstorage.WithLogger(slog.Default()),
		)

		assert.Error(t, err)
		assert.Nil(t, storage)
		var parseErr *pgconn.ParseConfigError
		assert.ErrorAs(t, err, &parseErr)
	})

	t.Run("nil logger", func(t *testing.T) {
		ctx := context.Background()
		dsn := "postgresql://user:pass@localhost/db"
		storage, err := githubstorage.New(ctx,
			githubstorage.WithDatabaseDSN(dsn),
			githubstorage.WithLogger(nil),
		)

		assert.Error(t, err)
		assert.ErrorIs(t, err, cerrors.ErrValueRequired)
		assert.Nil(t, storage)
	})

	t.Run("empty dsn", func(t *testing.T) {
		ctx := context.Background()
		storage, err := githubstorage.New(ctx,
			githubstorage.WithLogger(slog.Default()),
			githubstorage.WithDatabaseDSN(""),
		)

		assert.Error(t, err)
		assert.ErrorIs(t, err, cerrors.ErrValueRequired)
		assert.Nil(t, storage)
	})
}

func TestPing(t *testing.T) {
	t.Run("successful ping", func(t *testing.T) {
		mockPool := new(MockPool)
		mockPool.On("Ping", mock.Anything).Return(nil).Once()

		storage := githubstorage.GitHubStorage{
			Logger: slog.Default(),
			Pool:   mockPool,
		}

		ctx := context.Background()
		err := storage.Ping(ctx, 3, time.Second)
		assert.NoError(t, err)
		mockPool.AssertExpectations(t)
	})

	t.Run("ping failure with retries", func(t *testing.T) {
		fakeRetries := 1
		mockPool := new(MockPool)
		mockPool.On("Ping", mock.Anything).Return(errors.New("ping failed")).Times(fakeRetries)

		storage := githubstorage.GitHubStorage{
			Logger: slog.Default(),
			Pool:   mockPool,
		}

		ctx := context.Background()
		err := storage.Ping(ctx, uint8(fakeRetries), time.Second)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ping failed")
		mockPool.AssertExpectations(t)
	})
}

func TestStore(t *testing.T) {
	t.Run("successful store", func(t *testing.T) {
		mockPool := new(MockPool)
		mockPool.On("Exec", mock.Anything, githubstorage.GitHubStoreQuery, mock.Anything).
			Return(pgconn.CommandTag{}, nil).Once()

		storage := githubstorage.GitHubStorage{
			Logger: slog.Default(),
			Pool:   mockPool,
		}

		payload := &githubstorage.GitHub{
			DeliveryID:     uuid.New(),
			Event:          "push",
			TargetType:     "repo",
			UserLogin:      "test_user",
			TargetID:       12345,
			HookID:         67890,
			UserID:         112233,
			KafkaOffset:    0,
			KafkaPartition: 1,
			Payload:        "example payload",
		}

		ctx := context.Background()
		err := storage.Store(ctx, payload)
		assert.NoError(t, err)
		mockPool.AssertExpectations(t)
	})

	t.Run("store failure", func(t *testing.T) {
		mockPool := new(MockPool)
		mockPool.On("Exec", mock.Anything, githubstorage.GitHubStoreQuery, mock.Anything).
			Return(pgconn.CommandTag{}, errors.New("insert error")).Once()

		storage := githubstorage.GitHubStorage{
			Logger: slog.Default(),
			Pool:   mockPool,
		}

		payload := &githubstorage.GitHub{
			DeliveryID:     uuid.New(),
			Event:          "push",
			TargetType:     "repo",
			UserLogin:      "test_user",
			TargetID:       12345,
			HookID:         67890,
			UserID:         112233,
			KafkaOffset:    0,
			KafkaPartition: 1,
			Payload:        "example payload",
		}

		ctx := context.Background()
		err := storage.Store(ctx, payload)

		assert.Error(t, err)
		mockPool.AssertExpectations(t)
	})
}
