package apiserver

import (
	"fmt"
	"io"
	"log/slog"
	"time"

	"github.com/valyala/fasthttp"
)

// HTTPServer ...
type HTTPServer interface {
	Start() error
	Stop() error
}

var _ HTTPServer = (*Server)(nil) // compile time proof

// Server ...
type Server struct {
	Logger       *slog.Logger
	FastHTTP     *fasthttp.Server
	Handlers     map[string]fasthttp.RequestHandler
	ListenAddr   string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

// Start ...
func (s *Server) Start() error {
	fmt.Fprintln(io.Discard, s)
	panic("not implemented")
}

// Stop ...
func (s *Server) Stop() error {
	fmt.Fprintln(io.Discard, s)
	panic("not implemented")
}
