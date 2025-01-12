package slogger_test

import (
	"bytes"
	"testing"

	"github.com/devchain-network/cauldron/internal/slogger"
	"github.com/stretchr/testify/assert"
)

func TestNew_Defaults(t *testing.T) {
	logger, err := slogger.New()

	assert.NoError(t, err)
	assert.NotNil(t, logger)
}

func TestNew_WithCustomOptions(t *testing.T) {
	var buffer bytes.Buffer

	logger, err := slogger.New(
		slogger.WithLogLevel(slogger.LevelInfo),
		slogger.WithWriter(&buffer),
	)

	assert.NoError(t, err)
	assert.NotNil(t, logger)

	logger.Info("test message", "key", "value")
	assert.Contains(t, buffer.String(), `"msg":"test message"`)
}

func TestWithLogLevel_ValidLogLevelName(t *testing.T) {
	logger, err := slogger.New(
		slogger.WithLogLevelName("ERROR"),
	)
	assert.NoError(t, err)
	assert.NotNil(t, logger)
}

func TestWithLogLevel_InvalidLogLevelName(t *testing.T) {
	logger, err := slogger.New(
		slogger.WithLogLevelName("foo"),
	)
	assert.Error(t, err)
	assert.Nil(t, logger)
}

func TestWithLogLevel_EmptyLogLevelName(t *testing.T) {
	logger, err := slogger.New(
		slogger.WithLogLevelName(""),
	)
	assert.Error(t, err)
	assert.Nil(t, logger)
}

func TestWithLogLevel_NilLogLevel(t *testing.T) {
	logger, err := slogger.New(
		slogger.WithLogLevel(nil),
	)
	assert.Error(t, err)
	assert.Nil(t, logger)
}

func TestWithLogLevel_NilWriter(t *testing.T) {
	logger, err := slogger.New(
		slogger.WithWriter(nil),
	)
	assert.Error(t, err)
	assert.Nil(t, logger)
}

func TestNew_InfoShouldNotVisibleOnErrorLevel(t *testing.T) {
	var buffer bytes.Buffer

	logger, err := slogger.New(
		slogger.WithLogLevel(slogger.LevelError),
		slogger.WithWriter(&buffer),
	)

	assert.NoError(t, err)
	assert.NotNil(t, logger)

	logger.Info("test message", "key", "value")
	assert.Contains(t, buffer.String(), "")
}
