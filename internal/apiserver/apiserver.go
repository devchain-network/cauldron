package apiserver

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/devchain-network/cauldron/internal/cerrors"
	"github.com/devchain-network/cauldron/internal/slogger"
	"github.com/valyala/fasthttp"
	"github.com/vigo/getenv"
)

// constants.
const (
	serverDefaultReadTimeout  = 5 * time.Second
	serverDefaultWriteTimeout = 10 * time.Second
	serverDefaultIdleTimeout  = 15 * time.Second
	serverDefaultListenAddr   = ":8000"
)

// HTTPServer defines the basic operations for managing an HTTP server's lifecycle.
type HTTPServer interface {
	Start() error
	Stop() error
}

var _ HTTPServer = (*Server)(nil) // compile time proof

// Option represents option function type.
type Option func(*Server) error

// WithLogger sets logger.
func WithLogger(l *slog.Logger) Option {
	return func(server *Server) error {
		if l == nil {
			return fmt.Errorf("logger error: [%w]", cerrors.ErrValueRequired)
		}
		server.Logger = l

		return nil
	}
}

// WithHTTPHandler adds http handler.
func WithHTTPHandler(path string, handler fasthttp.RequestHandler) Option {
	return func(server *Server) error {
		if path == "" {
			return fmt.Errorf("path error: [%w]", cerrors.ErrValueRequired)
		}
		if handler == nil {
			return fmt.Errorf("http handler error: [%w]", cerrors.ErrValueRequired)
		}

		if server.Handlers == nil {
			server.Handlers = make(map[string]fasthttp.RequestHandler)
		}
		server.Handlers[path] = handler

		return nil
	}
}

// WithListenAddr sets listen addr.
func WithListenAddr(addr string) Option {
	return func(server *Server) error {
		if addr == "" {
			return fmt.Errorf("listen addr error: [%w]", cerrors.ErrValueRequired)
		}
		server.ListenAddr = addr

		return nil
	}
}

// WithReadTimeout sets read timeout.
func WithReadTimeout(d time.Duration) Option {
	return func(server *Server) error {
		server.ReadTimeout = d

		return nil
	}
}

// WithWriteTimeout sets write timeout.
func WithWriteTimeout(d time.Duration) Option {
	return func(server *Server) error {
		server.WriteTimeout = d

		return nil
	}
}

// WithIdleTimeout sets idle timeout.
func WithIdleTimeout(d time.Duration) Option {
	return func(server *Server) error {
		server.IdleTimeout = d

		return nil
	}
}

// Server represents server configuration. Must implements HTTPServer interface.
type Server struct {
	Logger       *slog.Logger
	FastHTTP     *fasthttp.Server
	Handlers     map[string]fasthttp.RequestHandler
	ListenAddr   string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

// Start starts the fast http server.
func (s *Server) Start() error {
	s.Logger.Info("start listening at", "addr", s.ListenAddr)
	if err := s.FastHTTP.ListenAndServe(s.ListenAddr); err != nil {
		return fmt.Errorf("listen and server error: [%w]", err)
	}

	return nil
}

// Stop stops the fast http server.
func (s *Server) Stop() error {
	s.Logger.Info("shutting down the server")
	if err := s.FastHTTP.ShutdownWithContext(context.Background()); err != nil {
		s.Logger.Error("server shutdown error", "error", err)

		return fmt.Errorf("server shutdown error: [%w]", err)
	}

	return nil
}

// New instantiates new api server.
func New(options ...Option) (*Server, error) {
	server := new(Server)
	server.ReadTimeout = serverDefaultReadTimeout
	server.WriteTimeout = serverDefaultWriteTimeout
	server.IdleTimeout = serverDefaultIdleTimeout
	server.ListenAddr = serverDefaultListenAddr

	for _, option := range options {
		if err := option(server); err != nil {
			return nil, fmt.Errorf("option error: [%w]", err)
		}
	}

	if server.Logger == nil {
		return nil, fmt.Errorf("logger error: [%w]", cerrors.ErrValueRequired)
	}

	if server.Handlers == nil {
		return nil, fmt.Errorf("http handlers error: [%w]", cerrors.ErrValueRequired)
	}

	httpRouter := func(ctx *fasthttp.RequestCtx) {
		handler, ok := server.Handlers[string(ctx.Path())]
		if !ok {
			ctx.NotFound()

			return
		}

		handler(ctx)
	}

	fastHTTPServer := &fasthttp.Server{
		Handler:         httpRouter,
		ReadTimeout:     server.ReadTimeout,
		WriteTimeout:    server.WriteTimeout,
		IdleTimeout:     server.IdleTimeout,
		CloseOnShutdown: true,
	}

	server.FastHTTP = fastHTTPServer

	return server, nil
}

// Run runs the server.
func Run() error {
	listenAddr := getenv.TCPAddr("LISTEN_ADDR", serverDefaultListenAddr)
	if err := getenv.Parse(); err != nil {
		return fmt.Errorf("run error, getenv: [%w]", err)
	}

	logger, err := slogger.New()
	if err != nil {
		return fmt.Errorf("run error, logger: [%w]", err)
	}

	server, err := New(
		WithLogger(logger),
		WithListenAddr(*listenAddr),
		WithHTTPHandler("/healthz", healthCheckHandler),
	)
	if err != nil {
		return fmt.Errorf("run error, server: [%w]", err)
	}

	ch := make(chan struct{})

	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
		<-sig

		logger.Info("shutting down the server")
		if err = server.Stop(); err != nil {
			logger.Error("server stop error: [%w]", "error", err)
		}
		close(ch)
	}()

	if err = server.Start(); err != nil {
		return fmt.Errorf("run error, server start: [%w]", err)
	}

	<-ch
	logger.Info("all clear, goodbye")

	return nil
}

func healthCheckHandler(ctx *fasthttp.RequestCtx) {
	ctx.SetStatusCode(fasthttp.StatusOK)
	ctx.SetBodyString("OK")
}
