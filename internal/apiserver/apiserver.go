package apiserver

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/IBM/sarama"
	"github.com/devchain-network/cauldron/internal/cerrors"
	"github.com/go-playground/webhooks/v6/github"
	"github.com/valyala/fasthttp"
)

// constants.
const (
	serverDefaultReadTimeout  = 5 * time.Second
	serverDefaultWriteTimeout = 10 * time.Second
	serverDefaultIdleTimeout  = 15 * time.Second
	serverDefaultListenAddr   = ":8000"

	kkDefaultGitHubTopic = "github"

	kkDefaultQueueSize = 100
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
			return fmt.Errorf("apiserver.WithLogger 'l' logger error: [%w]", cerrors.ErrValueRequired)
		}
		server.Logger = l

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

		if server.Handlers == nil {
			server.Handlers = make(map[string]methodHandler)
		}
		server.Handlers[path] = methodHandler{method: handler}

		return nil
	}
}

// WithListenAddr sets listen addr.
func WithListenAddr(addr string) Option {
	return func(server *Server) error {
		if addr == "" {
			return fmt.Errorf("apiserver.WithListenAddr listen 'addr' error: [%w]", cerrors.ErrValueRequired)
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

// WithKafkaBrokers sets kafka brokers list.
func WithKafkaBrokers(brokers []string) Option {
	return func(server *Server) error {
		if brokers == nil {
			return fmt.Errorf("apiserver.WithKafkaBrokers 'brokers' error: [%w]", cerrors.ErrValueRequired)
		}

		server.KafkaBrokers = make([]string, len(brokers))
		copy(server.KafkaBrokers, brokers)

		return nil
	}
}

// WithKafkaGitHubTopic sets kafka topic name for github webhooks.
func WithKafkaGitHubTopic(s string) Option {
	return func(server *Server) error {
		if s == "" {
			return fmt.Errorf("apiserver.WithKafkaGitHubTopic 's' github topic error: [%w]", cerrors.ErrValueRequired)
		}
		server.KafkaGitHubTopic = s

		return nil
	}
}

type methodHandler map[string]fasthttp.RequestHandler

// Server represents server configuration. Must implements HTTPServer interface.
type Server struct {
	Logger           *slog.Logger
	FastHTTP         *fasthttp.Server
	Handlers         map[string]methodHandler
	ListenAddr       string
	KafkaGitHubTopic string
	KafkaBrokers     []string
	ReadTimeout      time.Duration
	WriteTimeout     time.Duration
	IdleTimeout      time.Duration
}

type httpHandlerOptions struct {
	logger               *slog.Logger
	kafkaProducer        sarama.AsyncProducer
	producerMessageQueue chan *sarama.ProducerMessage
}

func (h httpHandlerOptions) messageWorker(workedID int) {
	for msg := range h.producerMessageQueue {
		h.kafkaProducer.Input() <- msg

		select {
		case success := <-h.kafkaProducer.Successes():
			h.logger.Info(
				"message sent",
				"worker id", workedID,
				"topic", success.Topic,
				"partition", success.Partition,
				"offset", success.Offset,
			)
		case err := <-h.kafkaProducer.Errors():
			h.logger.Error("message send error", "error", err)
		}
	}
}

func (h httpHandlerOptions) shutdown() {
	close(h.producerMessageQueue)
	h.logger.Info("waiting for message workers to finish")
}

type githubHandlerOptions struct {
	webhook *github.Webhook
	httpHandlerOptions
	topic string
}

// Start starts the fast http server.
func (s *Server) Start() error {
	s.Logger.Info("start listening at", "addr", s.ListenAddr)
	if err := s.FastHTTP.ListenAndServe(s.ListenAddr); err != nil {
		return fmt.Errorf("apiserver.Server.Start FastHTTP.ListenAndServe error: [%w]", err)
	}

	return nil
}

// Stop stops the fast http server.
func (s *Server) Stop() error {
	s.Logger.Info("shutting down the server")
	if err := s.FastHTTP.ShutdownWithContext(context.Background()); err != nil {
		s.Logger.Error("server shutdown error", "error", err)

		return fmt.Errorf("apiserver.Server.Stop FastHTTP.ShutdownWithContext error: [%w]", err)
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
			return nil, fmt.Errorf("apiserver.New option error: [%w]", err)
		}
	}

	if server.Logger == nil {
		return nil, fmt.Errorf("apiserver.New server.Logger error: [%w]", cerrors.ErrValueRequired)
	}

	if server.Handlers == nil {
		return nil, fmt.Errorf("apiserver.New server.Handlers error: [%w]", cerrors.ErrValueRequired)
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
