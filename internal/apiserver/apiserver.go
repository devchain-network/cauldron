package apiserver

import (
	"context"
	_ "embed"
	"fmt"
	"log/slog"
	"slices"
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

func validHTTPMethods() []string {
	return []string{
		fasthttp.MethodGet,
		fasthttp.MethodHead,
		fasthttp.MethodPost,
		fasthttp.MethodPut,
		fasthttp.MethodPatch,
		fasthttp.MethodDelete,
		fasthttp.MethodConnect,
		fasthttp.MethodOptions,
		fasthttp.MethodTrace,
	}
}

// Start starts the fast http server.
func (s *Server) Start() error {
	s.Logger.Info("start listening at", "addr", s.ListenAddr, "version", ServerVersion)
	if err := s.FastHTTP.ListenAndServe(s.ListenAddr); err != nil {
		return fmt.Errorf("[apiserver.Start][ListenAndServe] error: [%w]", err)
	}

	return nil
}

// Stop stops the fast http server.
func (s *Server) Stop() error {
	s.Logger.Info("shutting down the server")
	if err := s.FastHTTP.ShutdownWithContext(context.Background()); err != nil {
		s.Logger.Error("fast http shutdown with context error", "error", err)

		return fmt.Errorf("[apiserver.Start][ShutdownWithContext] error: [%w]", err)
	}

	return nil
}

func (s Server) checkRequired() error {
	if s.Logger == nil {
		return fmt.Errorf("[apiserver.checkRequired] Logger error: [%w, 'nil' received]", cerrors.ErrValueRequired)
	}

	if s.Handlers == nil {
		return fmt.Errorf("[apiserver.checkRequired] Handlers error: [%w, 'nil' received]", cerrors.ErrValueRequired)
	}

	if !s.KafkaGitHubTopic.Valid() {
		return fmt.Errorf(
			"[apiserver.checkRequired] KafkaGitHubTopic error: [%w, '%s' received]",
			cerrors.ErrInvalid, s.KafkaGitHubTopic,
		)
	}

	return nil
}

// WithLogger sets logger.
func WithLogger(l *slog.Logger) Option {
	return func(server *Server) error {
		if l == nil {
			return fmt.Errorf(
				"[apiserver.WithLogger] error: [%w, 'nil' received]",
				cerrors.ErrValueRequired,
			)
		}
		server.Logger = l

		return nil
	}
}

// WithHTTPHandler adds given http handler to handler stack.
func WithHTTPHandler(method, path string, handler fasthttp.RequestHandler) Option {
	return func(server *Server) error {
		if method == "" {
			return fmt.Errorf(
				"[apiserver.WithHTTPHandler] method error: [%w, empty string]",
				cerrors.ErrValueRequired,
			)
		}

		if !slices.Contains(validHTTPMethods(), method) {
			return fmt.Errorf(
				"[apiserver.WithHTTPHandler] method error: ['%s' is %w]",
				method, cerrors.ErrInvalid,
			)
		}

		if path == "" {
			return fmt.Errorf(
				"[apiserver.WithHTTPHandler] path error: [%w, empty string]",
				cerrors.ErrValueRequired,
			)
		}
		if handler == nil {
			return fmt.Errorf(
				"[apiserver.WithHTTPHandler] handler error: [%w, empty string]",
				cerrors.ErrValueRequired,
			)
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
			return fmt.Errorf(
				"[apiserver.WithListenAddr] error: [%w, empty string]",
				cerrors.ErrValueRequired,
			)
		}

		if _, err := getenv.ValidateTCPNetworkAddress(addr); err != nil {
			return fmt.Errorf(
				"[apiserver.WithListenAddr] error: [%w] ['%s' %w]",
				err, addr, cerrors.ErrInvalid,
			)
		}

		server.ListenAddr = addr

		return nil
	}
}

// WithReadTimeout sets read timeout.
func WithReadTimeout(d time.Duration) Option {
	return func(server *Server) error {
		if d < 0 {
			return fmt.Errorf(
				"[apiserver.WithReadTimeout] error: [%w, '%s' received, must > 0]",
				cerrors.ErrInvalid, d,
			)
		}

		server.ReadTimeout = d

		return nil
	}
}

// WithWriteTimeout sets write timeout.
func WithWriteTimeout(d time.Duration) Option {
	return func(server *Server) error {
		if d < 0 {
			return fmt.Errorf(
				"[apiserver.WithWriteTimeout] error: [%w, '%s' received, must > 0]",
				cerrors.ErrInvalid, d,
			)
		}
		server.WriteTimeout = d

		return nil
	}
}

// WithIdleTimeout sets idle timeout.
func WithIdleTimeout(d time.Duration) Option {
	return func(server *Server) error {
		if d < 0 {
			return fmt.Errorf(
				"[apiserver.WithIdleTimeout] error: [%w, '%s' received, must > 0]",
				cerrors.ErrInvalid, d,
			)
		}
		server.IdleTimeout = d

		return nil
	}
}

// WithKafkaBrokers sets kafka brokers list.
func WithKafkaBrokers(brokers string) Option {
	return func(server *Server) error {
		var kafkaBrokers kafkacp.KafkaBrokers
		kafkaBrokers.AddFromString(brokers)

		if !kafkaBrokers.Valid() {
			return fmt.Errorf(
				"[apiserver.WithKafkaBrokers] error: [%w, '%s' received]",
				cerrors.ErrInvalid, brokers,
			)
		}

		server.KafkaBrokers = kafkaBrokers

		return nil
	}
}

// WithKafkaGitHubTopic sets kafka topic name for github webhooks.
func WithKafkaGitHubTopic(s string) Option {
	return func(server *Server) error {
		topic := kafkacp.KafkaTopicIdentifier(s)

		if !topic.Valid() {
			return fmt.Errorf(
				"[apiserver.WithKafkaGitHubTopic] error: [%w, '%s' received]",
				cerrors.ErrInvalid, s,
			)
		}
		server.KafkaGitHubTopic = topic

		return nil
	}
}

// New instantiates new api server.
func New(options ...Option) (*Server, error) {
	server := new(Server)
	var kafkaBrokers kafkacp.KafkaBrokers
	kafkaBrokers.AddFromString(kafkacp.DefaultKafkaBrokers)

	server.KafkaBrokers = kafkaBrokers
	server.ReadTimeout = ServerDefaultReadTimeout
	server.WriteTimeout = ServerDefaultWriteTimeout
	server.IdleTimeout = ServerDefaultIdleTimeout
	server.ListenAddr = ServerDefaultListenAddr

	for _, option := range options {
		if err := option(server); err != nil {
			return nil, err
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
