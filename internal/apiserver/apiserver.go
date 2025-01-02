package apiserver

import (
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

func (s *Server) Start() error {
	panic("not implemented")
}

func (s *Server) Stop() error {
	panic("not implemented")
}
