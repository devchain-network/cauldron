package slogger

import (
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/devchain-network/cauldron/internal/cerrors"
)

// log levels.
const (
	LevelDebug = slog.LevelDebug
	LevelInfo  = slog.LevelInfo
	LevelError = slog.LevelError
	LevelWarn  = slog.LevelWarn
)

type jsonLogger struct {
	level  slog.Leveler
	writer io.Writer
}

// Option represents option function type.
type Option func(*jsonLogger) error

// WithLogLevel sets log level.
func WithLogLevel(l slog.Leveler) Option {
	return func(jl *jsonLogger) error {
		if l == nil {
			return fmt.Errorf("log level error: [%w]", cerrors.ErrValueRequired)
		}
		jl.level = l

		return nil
	}
}

// WithLogLevelName sets log level from level name, such as INFO.
func WithLogLevelName(n string) Option {
	return func(jl *jsonLogger) error {
		if n == "" {
			return fmt.Errorf("log level name error: [%w]", cerrors.ErrValueRequired)
		}

		logLevelMap := map[string]slog.Level{
			"DEBUG": LevelDebug,
			"INFO":  LevelInfo,
			"WARN":  LevelWarn,
			"ERROR": LevelError,
		}

		if level, exists := logLevelMap[n]; exists {
			jl.level = level

			return nil
		}

		return fmt.Errorf("log level name %s error: %w", n, cerrors.ErrInvalid)
	}
}

// WithWriter sets output.
func WithWriter(w io.Writer) Option {
	return func(jl *jsonLogger) error {
		if w == nil {
			return fmt.Errorf("log writer error: [%w]", cerrors.ErrValueRequired)
		}
		jl.writer = w

		return nil
	}
}

// New instantiates new json logger.
func New(options ...Option) (*slog.Logger, error) {
	jlogger := new(jsonLogger)

	for _, option := range options {
		if err := option(jlogger); err != nil {
			return nil, fmt.Errorf("option error: [%w]", err)
		}
	}

	if jlogger.level == nil {
		jlogger.level = LevelDebug
	}
	if jlogger.writer == nil {
		jlogger.writer = os.Stdout
	}

	jsonHandler := slog.NewJSONHandler(
		jlogger.writer,
		&slog.HandlerOptions{
			Level: jlogger.level,
		})

	return slog.New(jsonHandler), nil
}
