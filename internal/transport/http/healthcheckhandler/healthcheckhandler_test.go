package healthcheckhandler_test

import (
	"testing"

	"github.com/devchain-network/cauldron/internal/cerrors"
	"github.com/devchain-network/cauldron/internal/transport/http/healthcheckhandler"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
)

func TestNew_NoOptions(t *testing.T) {
	handler, err := healthcheckhandler.New()

	assert.ErrorIs(t, err, cerrors.ErrValueRequired)
	assert.Nil(t, handler)
}

func TestNew_EmptyVersion(t *testing.T) {
	handler, err := healthcheckhandler.New(
		healthcheckhandler.WithVersion(""),
	)

	assert.ErrorIs(t, err, cerrors.ErrValueRequired)
	assert.Nil(t, handler)
}

func TestNew_Success(t *testing.T) {
	handler, err := healthcheckhandler.New(
		healthcheckhandler.WithVersion("1.0.0"),
	)

	assert.NoError(t, err)
	assert.NotNil(t, handler)
	assert.Equal(t, "1.0.0", handler.Version)
}

func TestHandle(t *testing.T) {
	handler, err := healthcheckhandler.New(
		healthcheckhandler.WithVersion("1.0.0"),
	)

	assert.NoError(t, err)
	assert.NotNil(t, handler)

	ctx := &fasthttp.RequestCtx{}

	handler.Handle(ctx)

	assert.Equal(t, fasthttp.StatusOK, ctx.Response.StatusCode())
	assert.Equal(t, "OK 1.0.0", string(ctx.Response.Body()))
}
