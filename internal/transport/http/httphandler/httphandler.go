package httphandler

import "github.com/valyala/fasthttp"

// FastHTTPHandler ...
type FastHTTPHandler interface {
	Handle(ctx *fasthttp.RequestCtx)
}
