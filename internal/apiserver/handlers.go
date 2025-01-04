package apiserver

import (
	"fmt"
	"net/http"

	"github.com/go-playground/webhooks/v6/github"
	"github.com/kylelemons/godebug/pretty"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttpadaptor"
)

// healthCheckHandler handles "/healthz".
func healthCheckHandler(ctx *fasthttp.RequestCtx) {
	ctx.SetStatusCode(fasthttp.StatusOK)
	ctx.SetBodyString("OK")
}

func githubWebhookHandler(opts *githubHandlerOptions) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		var httpReq http.Request
		if err := fasthttpadaptor.ConvertRequest(ctx, &httpReq, true); err != nil {
			opts.logger.Error("can not convert fasthttpreq -> httpReq", "error", err)

			return
		}

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
