package apiserver

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/IBM/sarama"
	"github.com/devchain-network/cauldron/internal/cerrors"
	"github.com/go-playground/webhooks/v6/github"
	"github.com/google/uuid"
)

type githubHTTPRequestHeaders struct {
	event      string
	targetType string
	deliveryID uuid.UUID
	hookID     uint64
	targetID   uint64
}

type httpHandlerOptions struct {
	logger               *slog.Logger
	kafkaProducer        sarama.AsyncProducer
	producerMessageQueue chan *sarama.ProducerMessage
}

type githubHandlerOptions struct {
	webhook *github.Webhook
	httpHandlerOptions
	topic string
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
	h.logger.Info("shutting down, waiting for message workers to finish")
}

func (githubHandlerOptions) parseGitHubWebhookHTTPRequestHeaders(h http.Header) *githubHTTPRequestHeaders {
	parsed := &githubHTTPRequestHeaders{
		event:      AnythingUnknown,
		targetType: AnythingUnknown,
	}

	if val := h.Get("X-Github-Event"); val != "" {
		parsed.event = val
	}

	if val, err := uuid.Parse(h.Get("X-Github-Delivery")); err == nil {
		parsed.deliveryID = val
	}

	if val, err := strconv.ParseUint(h.Get("X-Github-Hook-Id"), 10, 64); err == nil {
		parsed.hookID = val
	}

	if val, err := strconv.ParseUint(h.Get("X-Github-Hook-Installation-Target-Id"), 10, 64); err == nil {
		parsed.targetID = val
	}

	if val := h.Get("X-Github-Hook-Installation-Target-Type"); val != "" {
		parsed.targetType = val
	}

	return parsed
}

type handlerOption func(*httpHandlerOptions) error

func handlerOptionsWithLogger(l *slog.Logger) handlerOption {
	return func(ho *httpHandlerOptions) error {
		if l == nil {
			return fmt.Errorf("apiserver.handlerOptionsWithLogger error: [%w]", cerrors.ErrValueRequired)
		}
		ho.logger = l

		return nil
	}
}

func handlerOptionsWithKafkaProducer(kp sarama.AsyncProducer) handlerOption {
	return func(ho *httpHandlerOptions) error {
		if kp == nil {
			return fmt.Errorf("apiserver.handlerOptionsWithKafkaProducer error: [%w]", cerrors.ErrValueRequired)
		}
		ho.kafkaProducer = kp

		return nil
	}
}

func handlerOptionsWithProducerMessageQueue(mq chan *sarama.ProducerMessage) handlerOption {
	return func(ho *httpHandlerOptions) error {
		if mq == nil {
			return fmt.Errorf("apiserver.handlerOptionsWithProducerMessageQueue error: [%w]", cerrors.ErrValueRequired)
		}
		ho.producerMessageQueue = mq

		return nil
	}
}

func newHandlerOptions(options ...handlerOption) (*httpHandlerOptions, error) {
	hopts := new(httpHandlerOptions)

	for _, option := range options {
		if err := option(hopts); err != nil {
			return nil, fmt.Errorf("apiserver.newHandlerOptions option error: [%w]", err)
		}
	}

	if hopts.logger == nil {
		return nil, fmt.Errorf("apiserver.newHandlerOptions hopts.logger error: [%w]", cerrors.ErrValueRequired)
	}

	return hopts, nil
}
