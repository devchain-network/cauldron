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
	UnsupportedStar = "star"
)

func githubWebhookHandler(opts *githubHandlerOptions) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		var httpReq http.Request
		if err := fasthttpadaptor.ConvertRequest(ctx, &httpReq, true); err != nil {
			opts.logger.Error("can not convert fasthttpreq -> httpReq", "error", err)

			return
		}

		httpHeaders := opts.parseGitHubWebhookHTTPRequestHeaders(httpReq.Header)
		if httpHeaders.deliveryID == uuid.Nil {
			opts.logger.Error("invalid X-Github-Delivery")
			ctx.SetStatusCode(fasthttp.StatusBadRequest)

			return
		}

		listenEvents := []github.Event{
			github.CommitCommentEvent,
			github.CreateEvent,
			github.DeleteEvent,
			github.ForkEvent,
			github.GollumEvent,
			github.IssueCommentEvent,
			github.IssuesEvent,
			github.PullRequestEvent,
			github.PullRequestReviewCommentEvent,
			github.PullRequestReviewEvent,
			github.PushEvent,
			github.ReleaseEvent,
			github.StarEvent,
			github.WatchEvent,
		}
		payload, err := opts.webhook.Parse(&httpReq, listenEvents...)
		if err != nil {
			opts.logger.Info("webhook parse error", "error", err)

			return
		}

		opts.logger.Info(
			"webhook received",
			"event", httpHeaders.event,
			"target type", httpHeaders.targetType,
		)

		payloadB, errm := json.Marshal(payload)
		if errm != nil {
			opts.logger.Error("payload marshall error", "error", errm)

			return
		}

		messageKey := httpHeaders.deliveryID.String()
		message := &sarama.ProducerMessage{
			Topic: opts.topic,
			Key:   sarama.StringEncoder(messageKey),
			Value: sarama.ByteEncoder(payloadB),
			Headers: []sarama.RecordHeader{
				{Key: []byte("event"), Value: []byte(httpHeaders.event)},
				{Key: []byte("target-type"), Value: []byte(httpHeaders.targetType)},
				{Key: []byte("target-id"), Value: []byte(strconv.FormatUint(httpHeaders.targetID, 10))},
				{Key: []byte("hook-id"), Value: []byte(strconv.FormatUint(httpHeaders.hookID, 10))},
				{Key: []byte("content-type"), Value: []byte("application/json")},
			},
		}

		select {
		case opts.producerMessageQueue <- message:
			opts.logger.Info("message enqueued for processing")
		default:
			opts.logger.Warn("message queue is full, dropping", "message", message)
		}

		opts.logger.Info("github webhook received successfully")
		ctx.SetStatusCode(fasthttp.StatusAccepted)
	}
}
