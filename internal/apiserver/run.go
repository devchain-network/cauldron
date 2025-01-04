package apiserver

import (
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/devchain-network/cauldron/internal/slogger"
	"github.com/go-playground/webhooks/v6/github"
	"github.com/valyala/fasthttp"
	"github.com/vigo/getenv"
)

type githubHandlerOptions struct {
	logger  *slog.Logger
	webhook *github.Webhook
}

// Run runs the server.
func Run() error {
	listenAddr := getenv.TCPAddr("LISTEN_ADDR", serverDefaultListenAddr)
	logLevel := getenv.String("LOG_LEVEL", loggerDefaultLevel)
	githubHMACSecret := getenv.String("GITHUB_HMAC_SECRET", "notset")

	if err := getenv.Parse(); err != nil {
		return fmt.Errorf("run error, getenv: [%w]", err)
	}

	githubWebhook, err := github.New(github.Options.Secret(*githubHMACSecret))
	if err != nil {
		return fmt.Errorf("run error, githubWebhook: [%w]", err)
	}

	logger, err := slogger.New(
		slogger.WithLogLevelName(*logLevel),
	)
	if err != nil {
		return fmt.Errorf("run error, logger: [%w]", err)
	}

	githubHandlerOpts := githubHandlerOptions{
		logger:  logger,
		webhook: githubWebhook,
	}

	server, err := New(
		WithLogger(logger),
		WithListenAddr(*listenAddr),
		WithHTTPHandler(fasthttp.MethodGet, "/healthz", healthCheckHandler),
		WithHTTPHandler(fasthttp.MethodPost, "/v1/webhook/github", githubWebhookHandler(&githubHandlerOpts)),
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
		if errStop := server.Stop(); err != nil {
			logger.Error("server stop error: [%w]", "error", errStop)
		}
		close(ch)
	}()

	if errStop := server.Start(); err != nil {
		return fmt.Errorf("run error, server start: [%w]", errStop)
	}

	<-ch
	logger.Info("all clear, goodbye")

	return nil
}
