package apiserver

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-playground/webhooks/v6/github"
	"github.com/google/uuid"
	"github.com/kylelemons/godebug/pretty"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttpadaptor"
)

// healthCheckHandler handles "/healthz".
func healthCheckHandler(ctx *fasthttp.RequestCtx) {
	ctx.SetStatusCode(fasthttp.StatusOK)
	ctx.SetBodyString("OK")
}

// AnythingUnknown ...
const AnythingUnknown = "unknown"

// GitHubWebhookRequestHeaders ...
type GitHubWebhookRequestHeaders struct {
	Event      string
	TargetType string
	DeliveryID uuid.UUID
	HookID     uint64
	TargetID   uint64
}

// ParseGitHubHTTPHeaders ...
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
		pretty.Print(httpHeaders)

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

		switch payload := payload.(type) {
		case github.IssueCommentPayload:
			fmt.Println("IssueCommentPayload")
			pretty.Print(payload)
		case github.IssuesPayload:
			fmt.Println("IssuesPayload")
			pretty.Print(payload)
		case github.CreatePayload:
			fmt.Println("CreatePayload")
			pretty.Print(payload)
		case github.DeletePayload:
			fmt.Println("DeletePayload")
			pretty.Print(payload)
		case github.PushPayload:
			fmt.Println("PushPayload")
			pretty.Print(payload)
		default:
			fmt.Printf("payload type: %T\n", payload)
			pretty.Print(payload)
		}

		opts.logger.Info("github webhook success")
		ctx.SetStatusCode(fasthttp.StatusAccepted)
	}
}
