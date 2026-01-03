package gitlabstorage_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/IBM/sarama"
	"github.com/devchain-network/cauldron/internal/cerrors"
	"github.com/devchain-network/cauldron/internal/slogger/mockslogger"
	"github.com/devchain-network/cauldron/internal/storage"
	"github.com/devchain-network/cauldron/internal/storage/gitlabstorage"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

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

	db, err := gitlabstorage.New(
		ctx,
	)

	assert.ErrorIs(t, err, cerrors.ErrValueRequired)
	assert.Nil(t, db)
}

func TestNew_NilLogger(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), storage.DefaultDBPingTimeout)
	defer cancel()

	db, err := gitlabstorage.New(
		ctx,
		gitlabstorage.WithLogger(nil),
	)

	assert.ErrorIs(t, err, cerrors.ErrValueRequired)
	assert.Nil(t, db)
}

func TestNew_NoDSN(t *testing.T) {
	logger := mockslogger.New()

	ctx, cancel := context.WithTimeout(context.Background(), storage.DefaultDBPingTimeout)
	defer cancel()

	db, err := gitlabstorage.New(
		ctx,
		gitlabstorage.WithLogger(logger),
	)

	assert.ErrorIs(t, err, cerrors.ErrValueRequired)
	assert.Nil(t, db)
}

func TestNew_EmptyDSN(t *testing.T) {
	logger := mockslogger.New()

	ctx, cancel := context.WithTimeout(context.Background(), storage.DefaultDBPingTimeout)
	defer cancel()

	db, err := gitlabstorage.New(
		ctx,
		gitlabstorage.WithLogger(logger),
		gitlabstorage.WithDatabaseDSN(""),
	)

	assert.ErrorIs(t, err, cerrors.ErrValueRequired)
	assert.Nil(t, db)
}

func TestNew_InvalidDSN(t *testing.T) {
	logger := mockslogger.New()

	ctx, cancel := context.WithTimeout(context.Background(), storage.DefaultDBPingTimeout)
	defer cancel()

	db, err := gitlabstorage.New(
		ctx,
		gitlabstorage.WithLogger(logger),
		gitlabstorage.WithDatabaseDSN("foo:://bar"),
	)

	var parseErr *pgconn.ParseConfigError
	assert.ErrorAs(t, err, &parseErr)
	assert.Nil(t, db)
}

func TestNew_Success(t *testing.T) {
	logger := mockslogger.New()

	ctx, cancel := context.WithTimeout(context.Background(), storage.DefaultDBPingTimeout)
	defer cancel()

	db, err := gitlabstorage.New(
		ctx,
		gitlabstorage.WithLogger(logger),
		gitlabstorage.WithDatabaseDSN(
			"postgres://foo:bar@localhost:5432/fake?sslmode=disable",
		),
	)

	assert.NoError(t, err)
	assert.NotNil(t, db)
}

func TestPing_Fail(t *testing.T) {
	logger := mockslogger.New()

	ctx, cancel := context.WithTimeout(context.Background(), storage.DefaultDBPingTimeout)
	defer cancel()

	fakeRetries := 3
	mockPool := new(MockPGPooler)
	mockPool.On("Ping", ctx).Return(errors.New("ping failed")).Times(fakeRetries)

	db, err := gitlabstorage.New(
		ctx,
		gitlabstorage.WithLogger(logger),
		gitlabstorage.WithDatabaseDSN(
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
	logger := mockslogger.New()

	ctx, cancel := context.WithTimeout(context.Background(), storage.DefaultDBPingTimeout)
	defer cancel()

	mockPool := new(MockPGPooler)
	mockPool.On("Ping", ctx).Return(nil)

	db, err := gitlabstorage.New(
		ctx,
		gitlabstorage.WithLogger(logger),
		gitlabstorage.WithDatabaseDSN(
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
	logger := mockslogger.New()

	ctx, cancel := context.WithTimeout(context.Background(), storage.DefaultDBPingTimeout)
	defer cancel()

	mockPool := new(MockPGPooler)

	db, err := gitlabstorage.New(
		ctx,
		gitlabstorage.WithLogger(logger),
		gitlabstorage.WithDatabaseDSN(
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

func TestStore_Fail_Message_InvalidWebhookUUID(t *testing.T) {
	logger := mockslogger.New()

	ctx, cancel := context.WithTimeout(context.Background(), storage.DefaultDBPingTimeout)
	defer cancel()

	mockPool := new(MockPGPooler)
	db, err := gitlabstorage.New(
		ctx,
		gitlabstorage.WithLogger(logger),
		gitlabstorage.WithDatabaseDSN(
			"postgres://foo:bar@localhost:5432/fake?sslmode=disable",
		),
	)

	assert.NoError(t, err)
	assert.NotNil(t, db)
	db.Pool = mockPool

	uuidKey := uuid.New()

	message := &sarama.ConsumerMessage{
		Topic: "gitlab",
		Key:   []byte(uuidKey.String()),
		Value: []byte(`{"hello": "world"}`),
		Headers: []*sarama.RecordHeader{
			{Key: []byte("webhook-uuid"), Value: []byte(`invalid-uuid`)},
			{Key: []byte("object-kind"), Value: []byte(`push`)},
		},
	}

	err = db.MessageStore(ctx, message)
	assert.Error(t, err)
}

func TestStore_Fail_Message_InvalidProjectID(t *testing.T) {
	logger := mockslogger.New()

	ctx, cancel := context.WithTimeout(context.Background(), storage.DefaultDBPingTimeout)
	defer cancel()

	mockPool := new(MockPGPooler)
	db, err := gitlabstorage.New(
		ctx,
		gitlabstorage.WithLogger(logger),
		gitlabstorage.WithDatabaseDSN(
			"postgres://foo:bar@localhost:5432/fake?sslmode=disable",
		),
	)

	assert.NoError(t, err)
	assert.NotNil(t, db)
	db.Pool = mockPool

	uuidKey := uuid.New()
	webhookUUID := uuid.New()

	message := &sarama.ConsumerMessage{
		Topic: "gitlab",
		Key:   []byte(uuidKey.String()),
		Value: []byte(`{"hello": "world"}`),
		Headers: []*sarama.RecordHeader{
			{Key: []byte("webhook-uuid"), Value: []byte(webhookUUID.String())},
			{Key: []byte("object-kind"), Value: []byte(`push`)},
			{Key: []byte("project-id"), Value: []byte(`invalid`)},
		},
	}

	err = db.MessageStore(ctx, message)
	assert.Error(t, err)
}

func TestStore_Fail_Message_InvalidUserID(t *testing.T) {
	logger := mockslogger.New()

	ctx, cancel := context.WithTimeout(context.Background(), storage.DefaultDBPingTimeout)
	defer cancel()

	mockPool := new(MockPGPooler)

	db, err := gitlabstorage.New(
		ctx,
		gitlabstorage.WithLogger(logger),
		gitlabstorage.WithDatabaseDSN(
			"postgres://foo:bar@localhost:5432/fake?sslmode=disable",
		),
	)

	assert.NoError(t, err)
	assert.NotNil(t, db)
	db.Pool = mockPool

	uuidKey := uuid.New()
	webhookUUID := uuid.New()

	message := &sarama.ConsumerMessage{
		Topic: "gitlab",
		Key:   []byte(uuidKey.String()),
		Value: []byte(`{"hello": "world"}`),
		Headers: []*sarama.RecordHeader{
			{Key: []byte("webhook-uuid"), Value: []byte(webhookUUID.String())},
			{Key: []byte("object-kind"), Value: []byte(`push`)},
			{Key: []byte("project-id"), Value: []byte(`123`)},
			{Key: []byte("project-path"), Value: []byte(`vigo/test`)},
			{Key: []byte("user-username"), Value: []byte(`vigo`)},
			{Key: []byte("user-id"), Value: []byte(`invalid`)},
		},
	}

	err = db.MessageStore(ctx, message)
	assert.Error(t, err)
}

func TestStore_Insert_Error(t *testing.T) {
	logger := mockslogger.New()

	ctx, cancel := context.WithTimeout(context.Background(), storage.DefaultDBPingTimeout)
	defer cancel()

	mockPool := new(MockPGPooler)
	mockPool.On("Exec",
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(pgconn.CommandTag{}, errors.New("insert error"))

	db, err := gitlabstorage.New(
		ctx,
		gitlabstorage.WithLogger(logger),
		gitlabstorage.WithDatabaseDSN(
			"postgres://foo:bar@localhost:5432/fake?sslmode=disable",
		),
	)

	assert.NoError(t, err)
	assert.NotNil(t, db)
	db.Pool = mockPool

	uuidKey := uuid.New()
	webhookUUID := uuid.New()

	message := &sarama.ConsumerMessage{
		Topic: "gitlab",
		Key:   []byte(uuidKey.String()),
		Value: []byte(`{"hello": "world"}`),
		Headers: []*sarama.RecordHeader{
			{Key: []byte("webhook-uuid"), Value: []byte(webhookUUID.String())},
			{Key: []byte("object-kind"), Value: []byte(`push`)},
			{Key: []byte("project-id"), Value: []byte(`123`)},
			{Key: []byte("project-path"), Value: []byte(`vigo/test`)},
			{Key: []byte("user-username"), Value: []byte(`vigo`)},
			{Key: []byte("user-id"), Value: []byte(`456`)},
		},
	}

	err = db.MessageStore(ctx, message)
	assert.Error(t, err)
	mockPool.AssertExpectations(t)
}

func TestStore_Success(t *testing.T) {
	logger := mockslogger.New()

	ctx, cancel := context.WithTimeout(context.Background(), storage.DefaultDBPingTimeout)
	defer cancel()

	mockPool := new(MockPGPooler)
	mockPool.On("Exec",
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(pgconn.CommandTag{}, nil)

	db, err := gitlabstorage.New(
		ctx,
		gitlabstorage.WithLogger(logger),
		gitlabstorage.WithDatabaseDSN(
			"postgres://foo:bar@localhost:5432/fake?sslmode=disable",
		),
	)

	assert.NoError(t, err)
	assert.NotNil(t, db)
	db.Pool = mockPool

	uuidKey := uuid.New()
	webhookUUID := uuid.New()

	message := &sarama.ConsumerMessage{
		Topic: "gitlab",
		Key:   []byte(uuidKey.String()),
		Value: []byte(`{"hello": "world"}`),
		Headers: []*sarama.RecordHeader{
			{Key: []byte("webhook-uuid"), Value: []byte(webhookUUID.String())},
			{Key: []byte("object-kind"), Value: []byte(`push`)},
			{Key: []byte("project-id"), Value: []byte(`123`)},
			{Key: []byte("project-path"), Value: []byte(`vigo/test`)},
			{Key: []byte("user-username"), Value: []byte(`vigo`)},
			{Key: []byte("user-id"), Value: []byte(`456`)},
		},
	}

	err = db.MessageStore(ctx, message)
	assert.NoError(t, err)
	mockPool.AssertExpectations(t)
}

func TestStore_Success_WithNullableFields(t *testing.T) {
	logger := mockslogger.New()

	ctx, cancel := context.WithTimeout(context.Background(), storage.DefaultDBPingTimeout)
	defer cancel()

	mockPool := new(MockPGPooler)
	mockPool.On("Exec",
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(pgconn.CommandTag{}, nil)

	db, err := gitlabstorage.New(
		ctx,
		gitlabstorage.WithLogger(logger),
		gitlabstorage.WithDatabaseDSN(
			"postgres://foo:bar@localhost:5432/fake?sslmode=disable",
		),
	)

	assert.NoError(t, err)
	assert.NotNil(t, db)
	db.Pool = mockPool

	uuidKey := uuid.New()
	webhookUUID := uuid.New()

	// Milestone event - no user, no project
	message := &sarama.ConsumerMessage{
		Topic: "gitlab",
		Key:   []byte(uuidKey.String()),
		Value: []byte(`{"object_kind": "milestone"}`),
		Headers: []*sarama.RecordHeader{
			{Key: []byte("webhook-uuid"), Value: []byte(webhookUUID.String())},
			{Key: []byte("object-kind"), Value: []byte(`milestone`)},
		},
	}

	err = db.MessageStore(ctx, message)
	assert.NoError(t, err)
	mockPool.AssertExpectations(t)
}

func TestStore_Success_SkipsEmptyProjectID(t *testing.T) {
	logger := mockslogger.New()

	ctx, cancel := context.WithTimeout(context.Background(), storage.DefaultDBPingTimeout)
	defer cancel()

	mockPool := new(MockPGPooler)
	mockPool.On("Exec",
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(pgconn.CommandTag{}, nil)

	db, err := gitlabstorage.New(
		ctx,
		gitlabstorage.WithLogger(logger),
		gitlabstorage.WithDatabaseDSN(
			"postgres://foo:bar@localhost:5432/fake?sslmode=disable",
		),
	)

	assert.NoError(t, err)
	assert.NotNil(t, db)
	db.Pool = mockPool

	uuidKey := uuid.New()
	webhookUUID := uuid.New()

	message := &sarama.ConsumerMessage{
		Topic: "gitlab",
		Key:   []byte(uuidKey.String()),
		Value: []byte(`{"hello": "world"}`),
		Headers: []*sarama.RecordHeader{
			{Key: []byte("webhook-uuid"), Value: []byte(webhookUUID.String())},
			{Key: []byte("object-kind"), Value: []byte(`push`)},
			{Key: []byte("project-id"), Value: []byte(``)}, // empty, should skip
		},
	}

	err = db.MessageStore(ctx, message)
	assert.NoError(t, err)
	mockPool.AssertExpectations(t)
}

func TestStore_Success_SkipsZeroProjectID(t *testing.T) {
	logger := mockslogger.New()

	ctx, cancel := context.WithTimeout(context.Background(), storage.DefaultDBPingTimeout)
	defer cancel()

	mockPool := new(MockPGPooler)
	mockPool.On("Exec",
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(pgconn.CommandTag{}, nil)

	db, err := gitlabstorage.New(
		ctx,
		gitlabstorage.WithLogger(logger),
		gitlabstorage.WithDatabaseDSN(
			"postgres://foo:bar@localhost:5432/fake?sslmode=disable",
		),
	)

	assert.NoError(t, err)
	assert.NotNil(t, db)
	db.Pool = mockPool

	uuidKey := uuid.New()
	webhookUUID := uuid.New()

	message := &sarama.ConsumerMessage{
		Topic: "gitlab",
		Key:   []byte(uuidKey.String()),
		Value: []byte(`{"hello": "world"}`),
		Headers: []*sarama.RecordHeader{
			{Key: []byte("webhook-uuid"), Value: []byte(webhookUUID.String())},
			{Key: []byte("object-kind"), Value: []byte(`push`)},
			{Key: []byte("project-id"), Value: []byte(`0`)}, // zero, should skip
		},
	}

	err = db.MessageStore(ctx, message)
	assert.NoError(t, err)
	mockPool.AssertExpectations(t)
}

func TestStore_Success_SkipsZeroUserID(t *testing.T) {
	logger := mockslogger.New()

	ctx, cancel := context.WithTimeout(context.Background(), storage.DefaultDBPingTimeout)
	defer cancel()

	mockPool := new(MockPGPooler)
	mockPool.On("Exec",
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(pgconn.CommandTag{}, nil)

	db, err := gitlabstorage.New(
		ctx,
		gitlabstorage.WithLogger(logger),
		gitlabstorage.WithDatabaseDSN(
			"postgres://foo:bar@localhost:5432/fake?sslmode=disable",
		),
	)

	assert.NoError(t, err)
	assert.NotNil(t, db)
	db.Pool = mockPool

	uuidKey := uuid.New()
	webhookUUID := uuid.New()

	message := &sarama.ConsumerMessage{
		Topic: "gitlab",
		Key:   []byte(uuidKey.String()),
		Value: []byte(`{"hello": "world"}`),
		Headers: []*sarama.RecordHeader{
			{Key: []byte("webhook-uuid"), Value: []byte(webhookUUID.String())},
			{Key: []byte("object-kind"), Value: []byte(`push`)},
			{Key: []byte("user-id"), Value: []byte(`0`)}, // zero, should skip
		},
	}

	err = db.MessageStore(ctx, message)
	assert.NoError(t, err)
	mockPool.AssertExpectations(t)
}
