package healthcheckhandler

import (
	"fmt"

	"github.com/devchain-network/cauldron/internal/cerrors"
	"github.com/devchain-network/cauldron/internal/transport/http/httphandler"
	"github.com/valyala/fasthttp"
)

var _ HealthCheckHandler = (*Handler)(nil) // compile time proof

// HealthCheckHandler defines http handler behaviours.
type HealthCheckHandler interface {
	httphandler.FastHTTPHandler
}

// Handler represents http handler configuration and must satisfy HealthCheckHandler interface.
type Handler struct {
	Version string
}

// Handle is a fasthttp handler function.
func (h Handler) Handle(ctx *fasthttp.RequestCtx) {
	ctx.SetStatusCode(fasthttp.StatusOK)
	ctx.SetBodyString("OK " + h.Version)
}

func (h Handler) checkRequired() error {
	if h.Version == "" {
		return fmt.Errorf("health check handler check required, Version error: [%w]", cerrors.ErrValueRequired)
	}

	return nil
}

// Option represents option function type.
type Option func(*Handler) error

// WithVersion sets version information.
func WithVersion(s string) Option {
	return func(h *Handler) error {
		if s == "" {
			return fmt.Errorf("health checkhandler WithVersion error: [%w]", cerrors.ErrValueRequired)
		}
		h.Version = s

		return nil
	}
}

// New instantiates new handler instance.
func New(options ...Option) (*Handler, error) {
	handler := new(Handler)

	for _, option := range options {
		if err := option(handler); err != nil {
			return nil, fmt.Errorf("health check handler option error: [%w]", err)
		}
	}

	if err := handler.checkRequired(); err != nil {
		return nil, err
	}

	return handler, nil
}
