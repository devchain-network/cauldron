package httphandleroptions

import (
	"fmt"
	"log/slog"

	"github.com/IBM/sarama"
	"github.com/devchain-network/cauldron/internal/cerrors"
)

var _ Handler = (*HTTPHandler)(nil) // compile time proof

// Handler defines http handler behaviours.
type Handler interface {
	MessageWorker(workedID int)
}

// HTTPHandler represents http handler functionality.
type HTTPHandler struct {
	Logger               *slog.Logger
	KafkaProducer        sarama.AsyncProducer
	ProducerMessageQueue chan *sarama.ProducerMessage
}

// Option represents option function type.
type Option func(*HTTPHandler) error

// MessageWorker consumes produced messages.
func (h HTTPHandler) MessageWorker(workedID int) {
	for msg := range h.ProducerMessageQueue {
		h.KafkaProducer.Input() <- msg

		select {
		case success := <-h.KafkaProducer.Successes():
			h.Logger.Info(
				"message sent",
				"worker id", workedID,
				"topic", success.Topic,
				"partition", success.Partition,
				"offset", success.Offset,
			)
		case err := <-h.KafkaProducer.Errors():
			h.Logger.Error("message send error", "error", err)
		}
	}
}

// Shutdown closes producer message queue channel.
func (h HTTPHandler) Shutdown() {
	close(h.ProducerMessageQueue)
	h.Logger.Info("shutting down, waiting for message workers to finish")
}

// WithLogger sets logger.
func WithLogger(l *slog.Logger) Option {
	return func(hh *HTTPHandler) error {
		if l == nil {
			return fmt.Errorf("httphandleroptions.WithLogger error: [%w]", cerrors.ErrValueRequired)
		}
		hh.Logger = l

		return nil
	}
}

// WithKafkaProducer sets kafka producer.
func WithKafkaProducer(kp sarama.AsyncProducer) Option {
	return func(hh *HTTPHandler) error {
		if kp == nil {
			return fmt.Errorf("httphandleroptions.WithKafkaProducer error: [%w]", cerrors.ErrValueRequired)
		}
		hh.KafkaProducer = kp

		return nil
	}
}

// WithProducerMessageQueue sets kafka producer message queue.
func WithProducerMessageQueue(mq chan *sarama.ProducerMessage) Option {
	return func(hh *HTTPHandler) error {
		if mq == nil {
			return fmt.Errorf("httphandleroptions.WithProducerMessageQueue error: [%w]", cerrors.ErrValueRequired)
		}
		hh.ProducerMessageQueue = mq

		return nil
	}
}

// New instantiates new http handler.
func New(options ...Option) (*HTTPHandler, error) {
	handler := new(HTTPHandler)

	for _, option := range options {
		if err := option(handler); err != nil {
			return nil, fmt.Errorf("httphandleroptions.New option error: [%w]", err)
		}
	}

	if handler.Logger == nil {
		return nil, fmt.Errorf("httphandleroptions.New handler.Logger error: [%w]", cerrors.ErrValueRequired)
	}
	if handler.KafkaProducer == nil {
		return nil, fmt.Errorf("httphandleroptions.New handler.KafkaProducer error: [%w]", cerrors.ErrValueRequired)
	}
	if handler.ProducerMessageQueue == nil {
		return nil, fmt.Errorf(
			"httphandleroptions.New handler.ProducerMessageQueue error: [%w]",
			cerrors.ErrValueRequired,
		)
	}

	return handler, nil
}
