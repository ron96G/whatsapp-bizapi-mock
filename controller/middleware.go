package controller

import (
	"time"

	"github.com/google/uuid"
	"github.com/rgumi/whatsapp-mock/util"
	"github.com/valyala/fasthttp"
	"golang.org/x/time/rate"
)

var (
	requestIDHeader = "X-Request-ID"
)

func Authorize(h fasthttp.RequestHandler) fasthttp.RequestHandler {
	return fasthttp.RequestHandler(func(ctx *fasthttp.RequestCtx) {

		token, err := verifyToken(ctx)
		if err == nil && token.Valid && contains(Tokens, token.Raw) {
			h(ctx)
			return
		}
		ctx.SetStatusCode(401)
	})
}

func AuthorizeStaticToken(h fasthttp.RequestHandler, staticToken string) fasthttp.RequestHandler {

	return fasthttp.RequestHandler(func(ctx *fasthttp.RequestCtx) {
		token := extractAuthToken(ctx, "Apikey")
		if staticToken != "" && token != staticToken {
			ctx.SetStatusCode(401)
			return
		}
		h(ctx)
	})
}

func Log(h fasthttp.RequestHandler) fasthttp.RequestHandler {
	return fasthttp.RequestHandler(func(ctx *fasthttp.RequestCtx) {
		start := time.Now()
		defer func() {
			util.Log.Infof("%s - %s - %s \"%s %s %s\" %d %v",
				string(ctx.Response.Header.Peek(requestIDHeader)),
				ctx.RemoteAddr().String(),
				string(ctx.Host()),
				string(ctx.Method()),
				string(ctx.RequestURI()),
				string(ctx.Request.Header.UserAgent()),
				ctx.Response.StatusCode(),
				time.Since(start),
			)
		}()
		h(ctx)
	})
}

func Limiter(h fasthttp.RequestHandler, concurrencyLimit int) fasthttp.RequestHandler {
	limiter := rate.NewLimiter(rate.Limit(concurrencyLimit), concurrencyLimit)
	return fasthttp.RequestHandler(func(ctx *fasthttp.RequestCtx) {
		if !limiter.Allow() {
			ctx.SetStatusCode(429)
			return
		}
		h(ctx)
	})
}

func SetConnID(h fasthttp.RequestHandler) fasthttp.RequestHandler {
	return fasthttp.RequestHandler(func(ctx *fasthttp.RequestCtx) {

		reqID := ctx.Request.Header.Peek(requestIDHeader)
		if len(reqID) == 0 {
			reqID = []byte(uuid.New().String())
			ctx.Request.Header.SetBytesV(requestIDHeader, reqID)
			ctx.Response.Header.SetBytesV(requestIDHeader, reqID)
		} else {
			ctx.Response.Header.SetBytesV(requestIDHeader, reqID)
		}

		h(ctx)
	})
}
