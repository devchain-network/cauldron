package httphandler

import "github.com/valyala/fasthttp"

// FastHTTPHandler defines http handler behaviours.
type FastHTTPHandler interface {
	Handle(ctx *fasthttp.RequestCtx)
}
