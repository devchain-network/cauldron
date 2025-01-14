package apiserver

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/IBM/sarama"
	"github.com/devchain-network/cauldron/internal/apiserver/githubhandleroptions"
	"github.com/go-playground/webhooks/v6/github"
	"github.com/google/uuid"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttpadaptor"
)

// HealthCheckHandler handles health check functionality.
func HealthCheckHandler(ctx *fasthttp.RequestCtx) {
	ctx.SetStatusCode(fasthttp.StatusOK)
	ctx.SetBodyString("OK")
}

// GitHubWebhookHandler handles GitHub webhooks.
func GitHubWebhookHandler(opts *githubhandleroptions.HTTPHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		var httpReq http.Request
		if err := fasthttpadaptor.ConvertRequest(ctx, &httpReq, true); err != nil {
			opts.CommonHandler.Logger.Error("can not convert fasthttpreq -> httpReq", "error", err)

			return
		}

		httpHeaders := opts.ParseRequestHeaders(httpReq.Header)
		if httpHeaders.DeliveryID == uuid.Nil {
			opts.CommonHandler.Logger.Error(
				"unsupported X-Github-Delivery",
				"value",
				httpHeaders.DeliveryID,
			)
			ctx.SetStatusCode(fasthttp.StatusAccepted)

			return
		}

		listenEvents := []github.Event{
			github.CommitCommentEvent,
			github.CreateEvent,
			github.DeleteEvent,
			github.DependabotAlertEvent,
			github.ForkEvent,
			github.GollumEvent,
			github.IssueCommentEvent,
			github.IssuesEvent,
			github.PingEvent,
			github.PullRequestEvent,
			github.PullRequestReviewCommentEvent,
			github.PullRequestReviewEvent,
			github.PushEvent,
			github.ReleaseEvent,
			github.RepositoryEvent,
			github.StarEvent,
			github.WatchEvent,
		}
		payload, err := opts.Webhook.Parse(&httpReq, listenEvents...)
		if err != nil {
			opts.CommonHandler.Logger.Error(
				"github webhook parse error",
				"error", err,
				"event", httpHeaders.Event,
			)
			ctx.SetStatusCode(fasthttp.StatusBadRequest)

			return
		}

		opts.CommonHandler.Logger.Info(
			"webhook received",
			"event", httpHeaders.Event,
			"target type", httpHeaders.TargetType,
		)

		payloadBytes, err := json.Marshal(payload)
		if err != nil {
			opts.CommonHandler.Logger.Error("github webhook payload marshall error", "error", err)
			ctx.SetStatusCode(fasthttp.StatusInternalServerError)

			return
		}

		messageKey := httpHeaders.DeliveryID.String()
		message := &sarama.ProducerMessage{
			Topic: opts.Topic.String(),
			Key:   sarama.StringEncoder(messageKey),
			Value: sarama.ByteEncoder(payloadBytes),
			Headers: []sarama.RecordHeader{
				{Key: []byte("event"), Value: []byte(httpHeaders.Event)},
				{Key: []byte("target-type"), Value: []byte(httpHeaders.TargetType)},
				{
					Key:   []byte("target-id"),
					Value: []byte(strconv.FormatUint(httpHeaders.TargetID, 10)),
				},
				{Key: []byte("hook-id"), Value: []byte(strconv.FormatUint(httpHeaders.HookID, 10))},
			},
		}

		select {
		case opts.CommonHandler.ProducerMessageQueue <- message:
			opts.CommonHandler.Logger.Info("github webhook message enqueued for processing")
		default:
			opts.CommonHandler.Logger.Warn(
				"github webhook message queue is full, dropping",
				"message",
				message,
			)
		}

		opts.CommonHandler.Logger.Info("github webhook received successfully")
		ctx.SetStatusCode(fasthttp.StatusAccepted)
	}
}
