package apiserver

import (
	"context"
	_ "embed"
	"fmt"
	"log/slog"
	"time"

	"github.com/devchain-network/cauldron/internal/cerrors"
	"github.com/devchain-network/cauldron/internal/kafkacp"
	"github.com/valyala/fasthttp"
	"github.com/vigo/getenv"
)

// ServerVersion holds servers release version.
//
//go:embed VERSION
var ServerVersion string

// default values.
const (
	ServerDefaultReadTimeout  = 5 * time.Second
	ServerDefaultWriteTimeout = 10 * time.Second
	ServerDefaultIdleTimeout  = 15 * time.Second
	ServerDefaultListenAddr   = ":8000"
)

var _ HTTPServer = (*Server)(nil) // compile time proof

// HTTPServer defines the basic operations for managing an HTTP server's lifecycle.
type HTTPServer interface {
	Start() error
	Stop() error
}

// MethodHandler holds http method and http handler function information.
type MethodHandler map[string]fasthttp.RequestHandler

// Option represents option function type.
type Option func(*Server) error

// Server represents server configuration. Must implements HTTPServer interface.
type Server struct {
	Logger           *slog.Logger
	FastHTTP         *fasthttp.Server
	Handlers         map[string]MethodHandler
	ListenAddr       string
	KafkaGitHubTopic kafkacp.KafkaTopicIdentifier
	KafkaBrokers     kafkacp.KafkaBrokers
	ReadTimeout      time.Duration
	WriteTimeout     time.Duration
	IdleTimeout      time.Duration
}

// Start starts the fast http server.
func (s *Server) Start() error {
	s.Logger.Info("start listening at", "addr", s.ListenAddr, "version", ServerVersion)
	if err := s.FastHTTP.ListenAndServe(s.ListenAddr); err != nil {
		return fmt.Errorf("fast http listen and serve error: [%w]", err)
	}

	return nil
}

// Stop stops the fast http server.
func (s *Server) Stop() error {
	s.Logger.Info("shutting down the server")
	if err := s.FastHTTP.ShutdownWithContext(context.Background()); err != nil {
		s.Logger.Error("fast http shutdown with context error", "error", err)

		return fmt.Errorf("fast http shutdown with context error: [%w]", err)
	}

	return nil
}

func (s Server) checkRequired() error {
	if s.Logger == nil {
		return fmt.Errorf("api server check required, Logger error: [%w]", cerrors.ErrValueRequired)
	}

	if s.Handlers == nil {
		return fmt.Errorf("api server check required, Handlers error: [%w]", cerrors.ErrValueRequired)
	}

	if s.KafkaBrokers == nil {
		return fmt.Errorf("api server check required, KafkaBrokers error: [%w]", cerrors.ErrValueRequired)
	}

	if s.ListenAddr == "" {
		return fmt.Errorf("api server check required, ListenAddr error: [%w]", cerrors.ErrValueRequired)
	}

	if _, err := getenv.ValidateTCPNetworkAddress(s.ListenAddr); err != nil {
		return fmt.Errorf(
			"api server check required, ListenAddr error: [%w] [%w]", err, cerrors.ErrInvalid,
		)
	}

	if !s.KafkaGitHubTopic.Valid() {
		return fmt.Errorf("api server check required, KafkaGitHubTopic error: [%w]", cerrors.ErrInvalid)
	}

	return nil
}

// WithLogger sets logger.
func WithLogger(l *slog.Logger) Option {
	return func(server *Server) error {
		if l == nil {
			return fmt.Errorf("api server WithLogger error: [%w]", cerrors.ErrValueRequired)
		}
		server.Logger = l

		return nil
	}
}

// WithHTTPHandler adds given http handler to handler stack.
func WithHTTPHandler(method, path string, handler fasthttp.RequestHandler) Option {
	return func(server *Server) error {
		if method == "" {
			return fmt.Errorf("api server WithHTTPHandler method error: [%w]", cerrors.ErrValueRequired)
		}
		if path == "" {
			return fmt.Errorf("api server WithHTTPHandler path error: [%w]", cerrors.ErrValueRequired)
		}
		if handler == nil {
			return fmt.Errorf("api server WithHTTPHandler http handler error: [%w]", cerrors.ErrValueRequired)
		}

		if server.Handlers == nil {
			server.Handlers = make(map[string]MethodHandler)
		}
		server.Handlers[path] = MethodHandler{method: handler}

		return nil
	}
}

// WithListenAddr sets listen addr.
func WithListenAddr(addr string) Option {
	return func(server *Server) error {
		if addr == "" {
			return fmt.Errorf("api server WithListenAddr addr error: [%w]", cerrors.ErrValueRequired)
		}

		if _, err := getenv.ValidateTCPNetworkAddress(addr); err != nil {
			return fmt.Errorf("api server WithListenAddr tcp addr error: [%w] [%w]", err, cerrors.ErrInvalid)
		}

		server.ListenAddr = addr

		return nil
	}
}

// WithReadTimeout sets read timeout.
func WithReadTimeout(d time.Duration) Option {
	return func(server *Server) error {
		if d < 0 {
			return fmt.Errorf("api server WithReadTimeout error: [%w]", cerrors.ErrInvalid)
		}

		server.ReadTimeout = d

		return nil
	}
}

// WithWriteTimeout sets write timeout.
func WithWriteTimeout(d time.Duration) Option {
	return func(server *Server) error {
		if d < 0 {
			return fmt.Errorf("api server WithWriteTimeout error: [%w]", cerrors.ErrInvalid)
		}
		server.WriteTimeout = d

		return nil
	}
}

// WithIdleTimeout sets idle timeout.
func WithIdleTimeout(d time.Duration) Option {
	return func(server *Server) error {
		if d < 0 {
			return fmt.Errorf("api server WithIdleTimeout error: [%w]", cerrors.ErrInvalid)
		}
		server.IdleTimeout = d

		return nil
	}
}

// WithKafkaBrokers sets kafka brokers list.
func WithKafkaBrokers(brokers kafkacp.KafkaBrokers) Option {
	return func(server *Server) error {
		if !brokers.Valid() {
			return fmt.Errorf("api server WithKafkaBrokers error: [%w]", cerrors.ErrInvalid)
		}

		server.KafkaBrokers = brokers

		return nil
	}
}

// WithKafkaGitHubTopic sets kafka topic name for github webhooks.
func WithKafkaGitHubTopic(s kafkacp.KafkaTopicIdentifier) Option {
	return func(server *Server) error {
		if !s.Valid() {
			return fmt.Errorf("api server WithKafkaGitHubTopic error: [%w]", cerrors.ErrInvalid)
		}
		server.KafkaGitHubTopic = s

		return nil
	}
}

// New instantiates new api server.
func New(options ...Option) (*Server, error) {
	server := new(Server)
	server.ReadTimeout = ServerDefaultReadTimeout
	server.WriteTimeout = ServerDefaultWriteTimeout
	server.IdleTimeout = ServerDefaultIdleTimeout
	server.ListenAddr = ServerDefaultListenAddr

	for _, option := range options {
		if err := option(server); err != nil {
			return nil, fmt.Errorf("api server option error: [%w]", err)
		}
	}

	if err := server.checkRequired(); err != nil {
		return nil, err
	}

	httpRouter := func(ctx *fasthttp.RequestCtx) {
		methodsHandlers, ok := server.Handlers[string(ctx.Path())]
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

	server.FastHTTP = &fasthttp.Server{
		Handler:         httpRouter,
		ReadTimeout:     server.ReadTimeout,
		WriteTimeout:    server.WriteTimeout,
		IdleTimeout:     server.IdleTimeout,
		CloseOnShutdown: true,
	}

	return server, nil
}
