package githubstorage_test

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/IBM/sarama"
	"github.com/devchain-network/cauldron/internal/cerrors"
	"github.com/devchain-network/cauldron/internal/storage"
	"github.com/devchain-network/cauldron/internal/storage/githubstorage"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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

type MockPGPooler struct {
	mock.Mock
}

func (m *MockPGPooler) Ping(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockPGPooler) Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
	args := m.Called(ctx, sql, arguments)
	return args.Get(0).(pgconn.CommandTag), args.Error(1)
}

func (m *MockPGPooler) Close() {
	m.Called()
}

func (m *MockPGPooler) Acquire(ctx context.Context) (*pgxpool.Conn, error) {
	args := m.Called(ctx)
	return args.Get(0).(*pgxpool.Conn), args.Error(1)
}

func (m *MockPGPooler) AcquireAllIdle(ctx context.Context) ([]*pgxpool.Conn, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*pgxpool.Conn), args.Error(1)
}

func (m *MockPGPooler) AcquireFunc(ctx context.Context, f func(*pgxpool.Conn) error) error {
	args := m.Called(ctx, f)
	return args.Error(0)
}

func TestNew_NoLogger(t *testing.T) {
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

func TestNew_Success(t *testing.T) {
	logger := slog.New(new(mockLogger))

	ctx, cancel := context.WithTimeout(context.Background(), storage.DefaultDBPingTimeout)
	defer cancel()

	db, err := githubstorage.New(
		ctx,
		githubstorage.WithLogger(logger),
		githubstorage.WithDatabaseDSN(
			"postgres://foo:bar@localhost:5432/fake?sslmode=disable",
		),
	)

	assert.NoError(t, err)
	assert.NotNil(t, db)
}

func TestPing_Fail(t *testing.T) {
	logger := slog.New(new(mockLogger))

	ctx, cancel := context.WithTimeout(context.Background(), storage.DefaultDBPingTimeout)
	defer cancel()

	fakeRetries := 3
	mockPool := new(MockPGPooler)
	mockPool.On("Ping", ctx).Return(errors.New("ping failed")).Times(fakeRetries)

	db, err := githubstorage.New(
		ctx,
		githubstorage.WithLogger(logger),
		githubstorage.WithDatabaseDSN(
			"postgres://foo:bar@localhost:5432/fake?sslmode=disable",
		),
	)

	assert.NoError(t, err)
	assert.NotNil(t, db)
	db.Pool = mockPool

	err = db.Ping(ctx, uint8(fakeRetries), time.Millisecond*10)
	assert.Error(t, err)
	mockPool.AssertExpectations(t)
}

func TestPing_Success(t *testing.T) {
	logger := slog.New(new(mockLogger))

	ctx, cancel := context.WithTimeout(context.Background(), storage.DefaultDBPingTimeout)
	defer cancel()

	mockPool := new(MockPGPooler)
	mockPool.On("Ping", ctx).Return(nil)

	db, err := githubstorage.New(
		ctx,
		githubstorage.WithLogger(logger),
		githubstorage.WithDatabaseDSN(
			"postgres://foo:bar@localhost:5432/fake?sslmode=disable",
		),
	)

	assert.NoError(t, err)
	assert.NotNil(t, db)
	db.Pool = mockPool

	err = db.Ping(ctx, 3, time.Millisecond*10)
	assert.NoError(t, err)
	mockPool.AssertExpectations(t)
}

func TestStore_Fail_EmptyMessage(t *testing.T) {
	logger := slog.New(new(mockLogger))

	ctx, cancel := context.WithTimeout(context.Background(), storage.DefaultDBPingTimeout)
	defer cancel()

	mockPool := new(MockPGPooler)

	db, err := githubstorage.New(
		ctx,
		githubstorage.WithLogger(logger),
		githubstorage.WithDatabaseDSN(
			"postgres://foo:bar@localhost:5432/fake?sslmode=disable",
		),
	)

	assert.NoError(t, err)
	assert.NotNil(t, db)
	db.Pool = mockPool

	message := &sarama.ConsumerMessage{}

	err = db.MessageStore(ctx, message)
	assert.Error(t, err)
}

func TestStore_Fail_Message_InvalidTargetID(t *testing.T) {
	logger := slog.New(new(mockLogger))

	ctx, cancel := context.WithTimeout(context.Background(), storage.DefaultDBPingTimeout)
	defer cancel()

	mockPool := new(MockPGPooler)
	db, err := githubstorage.New(
		ctx,
		githubstorage.WithLogger(logger),
		githubstorage.WithDatabaseDSN(
			"postgres://foo:bar@localhost:5432/fake?sslmode=disable",
		),
	)

	assert.NoError(t, err)
	assert.NotNil(t, db)
	db.Pool = mockPool

	uuidKey := uuid.New()

	message := &sarama.ConsumerMessage{
		Topic: "github",
		Key:   []byte(uuidKey.String()),
		Value: []byte(`{"hello": "world"}`),
		Headers: []*sarama.RecordHeader{
			{Key: []byte("event"), Value: []byte(`push`)},
			{Key: []byte("target-type"), Value: []byte(`repository`)},
			{Key: []byte("target-id"), Value: []byte(`invalid`)},
		},
	}

	err = db.MessageStore(ctx, message)
	assert.Error(t, err)
}

func TestStore_Fail_Message_InvalidHookID(t *testing.T) {
	logger := slog.New(new(mockLogger))

	ctx, cancel := context.WithTimeout(context.Background(), storage.DefaultDBPingTimeout)
	defer cancel()

	mockPool := new(MockPGPooler)
	db, err := githubstorage.New(
		ctx,
		githubstorage.WithLogger(logger),
		githubstorage.WithDatabaseDSN(
			"postgres://foo:bar@localhost:5432/fake?sslmode=disable",
		),
	)

	assert.NoError(t, err)
	assert.NotNil(t, db)
	db.Pool = mockPool

	uuidKey := uuid.New()

	message := &sarama.ConsumerMessage{
		Topic: "github",
		Key:   []byte(uuidKey.String()),
		Value: []byte(`{"hello": "world"}`),
		Headers: []*sarama.RecordHeader{
			{Key: []byte("event"), Value: []byte(`push`)},
			{Key: []byte("target-type"), Value: []byte(`repository`)},
			{Key: []byte("target-id"), Value: []byte(`1`)},
			{Key: []byte("hook-id"), Value: []byte(`invalid`)},
		},
	}

	err = db.MessageStore(ctx, message)
	assert.Error(t, err)
}

func TestStore_Fail_Message_InvalidSenderID(t *testing.T) {
	logger := slog.New(new(mockLogger))

	ctx, cancel := context.WithTimeout(context.Background(), storage.DefaultDBPingTimeout)
	defer cancel()

	mockPool := new(MockPGPooler)

	db, err := githubstorage.New(
		ctx,
		githubstorage.WithLogger(logger),
		githubstorage.WithDatabaseDSN(
			"postgres://foo:bar@localhost:5432/fake?sslmode=disable",
		),
	)

	assert.NoError(t, err)
	assert.NotNil(t, db)
	db.Pool = mockPool

	uuidKey := uuid.New()

	message := &sarama.ConsumerMessage{
		Topic: "github",
		Key:   []byte(uuidKey.String()),
		Value: []byte(`{"hello": "world"}`),
		Headers: []*sarama.RecordHeader{
			{Key: []byte("event"), Value: []byte(`push`)},
			{Key: []byte("target-type"), Value: []byte(`repository`)},
			{Key: []byte("target-id"), Value: []byte(`1`)},
			{Key: []byte("hook-id"), Value: []byte(`2`)},
			{Key: []byte("sender-login"), Value: []byte(`vigo`)},
			{Key: []byte("sender-id"), Value: []byte(`invalid`)},
		},
	}

	err = db.MessageStore(ctx, message)
	assert.Error(t, err)
}

func TestStore_Insert_Error(t *testing.T) {
	logger := slog.New(new(mockLogger))

	ctx, cancel := context.WithTimeout(context.Background(), storage.DefaultDBPingTimeout)
	defer cancel()

	mockPool := new(MockPGPooler)
	mockPool.On("Exec",
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(pgconn.CommandTag{}, errors.New("insert error"))

	db, err := githubstorage.New(
		ctx,
		githubstorage.WithLogger(logger),
		githubstorage.WithDatabaseDSN(
			"postgres://foo:bar@localhost:5432/fake?sslmode=disable",
		),
	)

	assert.NoError(t, err)
	assert.NotNil(t, db)
	db.Pool = mockPool

	uuidKey := uuid.New()

	message := &sarama.ConsumerMessage{
		Topic: "github",
		Key:   []byte(uuidKey.String()),
		Value: []byte(`{"hello": "world"}`),
		Headers: []*sarama.RecordHeader{
			{Key: []byte("event"), Value: []byte(`push`)},
			{Key: []byte("target-type"), Value: []byte(`repository`)},
			{Key: []byte("target-id"), Value: []byte(`1`)},
			{Key: []byte("hook-id"), Value: []byte(`2`)},
			{Key: []byte("sender-login"), Value: []byte(`vigo`)},
			{Key: []byte("sender-id"), Value: []byte(`3`)},
		},
	}

	err = db.MessageStore(ctx, message)
	assert.Error(t, err)
	mockPool.AssertExpectations(t)
}

func TestStore_Success(t *testing.T) {
	logger := slog.New(new(mockLogger))

	ctx, cancel := context.WithTimeout(context.Background(), storage.DefaultDBPingTimeout)
	defer cancel()

	mockPool := new(MockPGPooler)
	mockPool.On("Exec",
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(pgconn.CommandTag{}, nil)

	db, err := githubstorage.New(
		ctx,
		githubstorage.WithLogger(logger),
		githubstorage.WithDatabaseDSN(
			"postgres://foo:bar@localhost:5432/fake?sslmode=disable",
		),
	)

	assert.NoError(t, err)
	assert.NotNil(t, db)
	db.Pool = mockPool

	uuidKey := uuid.New()

	message := &sarama.ConsumerMessage{
		Topic: "github",
		Key:   []byte(uuidKey.String()),
		Value: []byte(`{"hello": "world"}`),
		Headers: []*sarama.RecordHeader{
			{Key: []byte("event"), Value: []byte(`push`)},
			{Key: []byte("target-type"), Value: []byte(`repository`)},
			{Key: []byte("target-id"), Value: []byte(`1`)},
			{Key: []byte("hook-id"), Value: []byte(`2`)},
			{Key: []byte("sender-login"), Value: []byte(`vigo`)},
			{Key: []byte("sender-id"), Value: []byte(`3`)},
		},
	}

	err = db.MessageStore(ctx, message)
	assert.NoError(t, err)
	mockPool.AssertExpectations(t)
}
