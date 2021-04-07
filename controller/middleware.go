package controller

import (
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/valyala/fasthttp"
	"golang.org/x/time/rate"
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

func Log(h fasthttp.RequestHandler) fasthttp.RequestHandler {
	return fasthttp.RequestHandler(func(ctx *fasthttp.RequestCtx) {
		start := time.Now()
		defer func() {
			log.Printf("%s - %s - %s \"%s %s %s\" %d %v",
				string(ctx.Response.Header.Peek("X-Request-ID")),
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
		ctx.Response.Header.Set("X-Request-ID", uuid.New().String())
		h(ctx)
	})
}
