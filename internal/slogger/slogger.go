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

	DefaultLogLevel = "INFO"
)

// JSONLogger represents custom logger in JSON format.
type JSONLogger struct {
	Level  slog.Leveler
	Writer io.Writer
}

// Option represents option function type.
type Option func(*JSONLogger) error

// WithLogLevel sets log level.
func WithLogLevel(l slog.Leveler) Option {
	return func(jl *JSONLogger) error {
		if l == nil {
			return fmt.Errorf("slogger.WithLogLevel jl.Level error: [%w]", cerrors.ErrValueRequired)
		}
		jl.Level = l

		return nil
	}
}

// WithLogLevelName sets log level from level name, such as INFO.
func WithLogLevelName(n string) Option {
	return func(jl *JSONLogger) error {
		if n == "" {
			return fmt.Errorf("slogger.WithLogLevelName error: [%w]", cerrors.ErrValueRequired)
		}

		logLevelMap := map[string]slog.Level{
			"DEBUG": LevelDebug,
			"INFO":  LevelInfo,
			"WARN":  LevelWarn,
			"ERROR": LevelError,
		}

		if level, exists := logLevelMap[n]; exists {
			jl.Level = level

			return nil
		}

		return fmt.Errorf("slogger.WithLogLevelName jl.Level '%s' error: %w", n, cerrors.ErrInvalid)
	}
}

// WithWriter sets output.
func WithWriter(w io.Writer) Option {
	return func(jl *JSONLogger) error {
		if w == nil {
			return fmt.Errorf("slogger.WithWriter jl.Writer error: [%w]", cerrors.ErrValueRequired)
		}
		jl.Writer = w

		return nil
	}
}

// New instantiates new json logger.
func New(options ...Option) (*slog.Logger, error) {
	jlogger := new(JSONLogger)

	for _, option := range options {
		if err := option(jlogger); err != nil {
			return nil, fmt.Errorf("slogger.New option error: [%w]", err)
		}
	}

	if jlogger.Level == nil {
		jlogger.Level = LevelDebug
	}
	if jlogger.Writer == nil {
		jlogger.Writer = os.Stdout
	}

	jsonHandler := slog.NewJSONHandler(
		jlogger.Writer,
		&slog.HandlerOptions{
			Level: jlogger.Level,
		})

	return slog.New(jsonHandler), nil
}
