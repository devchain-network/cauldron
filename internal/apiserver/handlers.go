package apiserver

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/IBM/sarama"
	"github.com/go-playground/webhooks/v6/github"
	"github.com/google/uuid"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttpadaptor"
)

// healthCheckHandler handles "/healthz".
func healthCheckHandler(ctx *fasthttp.RequestCtx) {
	ctx.SetStatusCode(fasthttp.StatusOK)
	ctx.SetBodyString("OK")
}

// constants.
const (
	AnythingUnknown = "unknown"
)

// GitHubWebhookRequestHeaders represents important http headers to fetch.
type GitHubWebhookRequestHeaders struct {
	Event      string
	TargetType string
	DeliveryID uuid.UUID
	HookID     uint64
	TargetID   uint64
}

// ParseGitHubHTTPHeaders parses incoming http headers and returns required http headers.
func ParseGitHubHTTPHeaders(h http.Header) *GitHubWebhookRequestHeaders {
	out := &GitHubWebhookRequestHeaders{
		Event:      AnythingUnknown,
		TargetType: AnythingUnknown,
	}

	if val := h.Get("X-Github-Event"); val != "" {
		out.Event = val
	}

	if val, err := uuid.Parse(h.Get("X-Github-Delivery")); err == nil {
		out.DeliveryID = val
	}

	if val, err := strconv.ParseUint(h.Get("X-Github-Hook-Id"), 10, 64); err == nil {
		out.HookID = val
	}

	if val, err := strconv.ParseUint(h.Get("X-Github-Hook-Installation-Target-Id"), 10, 64); err == nil {
		out.TargetID = val
	}

	if val := h.Get("X-Github-Hook-Installation-Target-Type"); val != "" {
		out.TargetType = val
	}

	return out
}

func githubWebhookHandler(opts *githubHandlerOptions) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		var httpReq http.Request
		if err := fasthttpadaptor.ConvertRequest(ctx, &httpReq, true); err != nil {
			opts.logger.Error("can not convert fasthttpreq -> httpReq", "error", err)

			return
		}

		httpHeaders := ParseGitHubHTTPHeaders(httpReq.Header)
		// pretty.Print(httpHeaders)

		listenEvents := []github.Event{
			github.IssuesEvent,
			github.IssueCommentEvent,
			github.CreateEvent,
			github.DeleteEvent,
			github.PushEvent,
		}
		payload, err := opts.webhook.Parse(&httpReq, listenEvents...)
		if err != nil {
			return
		}

		opts.logger.Info(
			"webhook received",
			"event", httpHeaders.Event,
			"target type", httpHeaders.TargetType,
		)

		payloadB, errm := json.Marshal(payload)
		if errm != nil {
			opts.logger.Error("payload marshall error: [%w]", "error", errm)

			return
		}

		go func() {
			messageKey := httpHeaders.DeliveryID.String()
			message := &sarama.ProducerMessage{
				Topic: "deneme",
				Key:   sarama.StringEncoder(messageKey),
				Value: sarama.ByteEncoder(payloadB),
				Headers: []sarama.RecordHeader{
					{Key: []byte("event"), Value: []byte(httpHeaders.Event)},
					{Key: []byte("target-type"), Value: []byte(httpHeaders.TargetType)},
					{Key: []byte("hook-id"), Value: []byte(strconv.FormatUint(httpHeaders.HookID, 10))},
					{Key: []byte("content-type"), Value: []byte("application/json")},
				},
			}

			opts.kafkaProducer.Input() <- message

			select {
			case success := <-opts.kafkaProducer.Successes():
				opts.logger.Info(
					"message sent",
					"topic", success.Topic,
					"partition", success.Partition,
					"offset", success.Offset,
				)
			case err := <-opts.kafkaProducer.Errors():
				opts.logger.Error("message send error", "error", err)
			}
		}()

		opts.logger.Info("github webhook success")
		ctx.SetStatusCode(fasthttp.StatusAccepted)
	}
}
