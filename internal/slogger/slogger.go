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
			return fmt.Errorf(
				"[slogger.WithLogLevel] error: [%w, 'nil' received]",
				cerrors.ErrValueRequired,
			)
		}
		jl.Level = l

		return nil
	}
}

func validLogLevels() map[string]slog.Level {
	return map[string]slog.Level{
		"DEBUG": LevelDebug,
		"INFO":  LevelInfo,
		"WARN":  LevelWarn,
		"ERROR": LevelError,
	}
}

// WithLogLevelName sets log level from level name, such as INFO.
func WithLogLevelName(s string) Option {
	return func(jl *JSONLogger) error {
		if s == "" {
			return fmt.Errorf(
				"[slogger.WithLogLevelName] error: [%w, empty string received]",
				cerrors.ErrValueRequired,
			)
		}

		if level, exists := validLogLevels()[s]; exists {
			jl.Level = level

			return nil
		}

		return fmt.Errorf(
			"[slogger.WithLogLevelName] error: [%w, '%s' received]",
			cerrors.ErrInvalid, s,
		)
	}
}

// WithWriter sets output.
func WithWriter(w io.Writer) Option {
	return func(jl *JSONLogger) error {
		if w == nil {
			return fmt.Errorf(
				"[slogger.WithWriter] error: [%w, 'nil' received]",
				cerrors.ErrValueRequired,
			)
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
			return nil, err
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
