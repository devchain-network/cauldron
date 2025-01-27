//nolint:all
package mockslogger

import (
	"context"
	"log/slog"
)

var _ slog.Handler = (*MockLogger)(nil) // compile time proof

type MockLogger struct{}

func (h *MockLogger) Enabled(_ context.Context, _ slog.Level) bool {
	return true
}

func (h *MockLogger) Handle(_ context.Context, _ slog.Record) error {
	return nil
}

func (h *MockLogger) WithAttrs(_ []slog.Attr) slog.Handler {
	return h
}

func (h *MockLogger) WithGroup(_ string) slog.Handler {
	return h
}

func New() *slog.Logger {
	return slog.New(new(MockLogger))
}
