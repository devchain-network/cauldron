package apiserver

import (
	"context"
	_ "embed"
	"fmt"
	"log/slog"
	"time"

	"github.com/devchain-network/cauldron/internal/cerrors"
	"github.com/valyala/fasthttp"
)

//go:embed VERSION
var serverVersion string

// constants.
const (
	serverDefaultReadTimeout  = 5 * time.Second
	serverDefaultWriteTimeout = 10 * time.Second
	serverDefaultIdleTimeout  = 15 * time.Second
	serverDefaultListenAddr   = ":8000"

	kkDefaultGitHubTopic = "github"

	kkDefaultQueueSize = 100
)

var _ HTTPServer = (*Server)(nil) // compile time proof

// HTTPServer defines the basic operations for managing an HTTP server's lifecycle.
type HTTPServer interface {
	Start() error
	Stop() error
}

type methodHandler map[string]fasthttp.RequestHandler

// Option represents option function type.
type Option func(*Server) error

// Server represents server configuration. Must implements HTTPServer interface.
type Server struct {
	logger           *slog.Logger
	fastHTTP         *fasthttp.Server
	handlers         map[string]methodHandler
	listenAddr       string
	kafkaGitHubTopic string
	kafkaBrokers     []string
	readTimeout      time.Duration
	writeTimeout     time.Duration
	idleTimeout      time.Duration
}

// Start starts the fast http server.
func (s *Server) Start() error {
	s.logger.Info("start listening at", "addr", s.listenAddr, "version", serverVersion)
	if err := s.fastHTTP.ListenAndServe(s.listenAddr); err != nil {
		return fmt.Errorf("apiserver.Server.Start FastHTTP.ListenAndServe error: [%w]", err)
	}

	return nil
}

// Stop stops the fast http server.
func (s *Server) Stop() error {
	s.logger.Info("shutting down the server")
	if err := s.fastHTTP.ShutdownWithContext(context.Background()); err != nil {
		s.logger.Error("server shutdown error", "error", err)

		return fmt.Errorf("apiserver.Server.Stop FastHTTP.ShutdownWithContext error: [%w]", err)
	}

	return nil
}

// WithLogger sets logger.
func WithLogger(l *slog.Logger) Option {
	return func(server *Server) error {
		if l == nil {
			return fmt.Errorf("apiserver.WithLogger 'l' logger error: [%w]", cerrors.ErrValueRequired)
		}
		server.logger = l

		return nil
	}
}

// WithHTTPHandler adds http handler.
func WithHTTPHandler(method, path string, handler fasthttp.RequestHandler) Option {
	return func(server *Server) error {
		if method == "" {
			return fmt.Errorf("apiserver.WithHTTPHandler 'method' error: [%w]", cerrors.ErrValueRequired)
		}
		if path == "" {
			return fmt.Errorf("apiserver.WithHTTPHandler 'path' error: [%w]", cerrors.ErrValueRequired)
		}
		if handler == nil {
			return fmt.Errorf("apiserver.WithHTTPHandler 'http' handler error: [%w]", cerrors.ErrValueRequired)
		}

		if server.handlers == nil {
			server.handlers = make(map[string]methodHandler)
		}
		server.handlers[path] = methodHandler{method: handler}

		return nil
	}
}

// WithListenAddr sets listen addr.
func WithListenAddr(addr string) Option {
	return func(server *Server) error {
		if addr == "" {
			return fmt.Errorf("apiserver.WithListenAddr listen 'addr' error: [%w]", cerrors.ErrValueRequired)
		}
		server.listenAddr = addr

		return nil
	}
}

// WithReadTimeout sets read timeout.
func WithReadTimeout(d time.Duration) Option {
	return func(server *Server) error {
		server.readTimeout = d

		return nil
	}
}

// WithWriteTimeout sets write timeout.
func WithWriteTimeout(d time.Duration) Option {
	return func(server *Server) error {
		server.writeTimeout = d

		return nil
	}
}

// WithIdleTimeout sets idle timeout.
func WithIdleTimeout(d time.Duration) Option {
	return func(server *Server) error {
		server.idleTimeout = d

		return nil
	}
}

// WithKafkaBrokers sets kafka brokers list.
func WithKafkaBrokers(brokers []string) Option {
	return func(server *Server) error {
		if brokers == nil {
			return fmt.Errorf("apiserver.WithKafkaBrokers 'brokers' error: [%w]", cerrors.ErrValueRequired)
		}

		server.kafkaBrokers = make([]string, len(brokers))
		copy(server.kafkaBrokers, brokers)

		return nil
	}
}

// WithKafkaGitHubTopic sets kafka topic name for github webhooks.
func WithKafkaGitHubTopic(s string) Option {
	return func(server *Server) error {
		if s == "" {
			return fmt.Errorf("apiserver.WithKafkaGitHubTopic 's' github topic error: [%w]", cerrors.ErrValueRequired)
		}
		server.kafkaGitHubTopic = s

		return nil
	}
}

// New instantiates new api server.
func New(options ...Option) (*Server, error) {
	server := new(Server)
	server.readTimeout = serverDefaultReadTimeout
	server.writeTimeout = serverDefaultWriteTimeout
	server.idleTimeout = serverDefaultIdleTimeout
	server.listenAddr = serverDefaultListenAddr

	for _, option := range options {
		if err := option(server); err != nil {
			return nil, fmt.Errorf("apiserver.New option error: [%w]", err)
		}
	}

	if server.logger == nil {
		return nil, fmt.Errorf("apiserver.New server.Logger error: [%w]", cerrors.ErrValueRequired)
	}

	if server.handlers == nil {
		return nil, fmt.Errorf("apiserver.New server.Handlers error: [%w]", cerrors.ErrValueRequired)
	}

	httpRouter := func(ctx *fasthttp.RequestCtx) {
		methodsHandlers, ok := server.handlers[string(ctx.Path())]
		if !ok {
			ctx.NotFound()

			return
		}
		var methodMatched bool
		for method, handler := range methodsHandlers {
			if method == string(ctx.Method()) {
				methodMatched = true
				handler(ctx)
			}
		}
		if !methodMatched {
			ctx.SetStatusCode(fasthttp.StatusMethodNotAllowed)
		}
	}

	server.fastHTTP = &fasthttp.Server{
		Handler:         httpRouter,
		ReadTimeout:     server.readTimeout,
		WriteTimeout:    server.writeTimeout,
		IdleTimeout:     server.idleTimeout,
		CloseOnShutdown: true,
	}

	return server, nil
}
