package apiserver

import "github.com/valyala/fasthttp"

// healthCheckHandler handles "/healthz".
func healthCheckHandler(ctx *fasthttp.RequestCtx) {
	ctx.SetStatusCode(fasthttp.StatusOK)
	ctx.SetBodyString("OK")
}
