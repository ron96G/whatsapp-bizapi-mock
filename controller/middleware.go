package controller

import (
	"fmt"
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
		token, err := parseToken(ctx)
		if err != nil {
			util.Log.Warn(err)

		} else if token.Valid && contains(Tokens, token.Raw) {
			h(ctx)
			return
		}
		ctx.SetStatusCode(401)
	})
}

func AuthorizeWithRoles(h fasthttp.RequestHandler, roles []string) fasthttp.RequestHandler {
	return fasthttp.RequestHandler(func(ctx *fasthttp.RequestCtx) {
		token, err := parseTokenWithClaims(ctx)
		if err == nil {
			if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid && contains(Tokens, token.Raw) {
				if contains(roles, claims.Role) {
					h(ctx)
					return
				}
			}
			err = fmt.Errorf("invalid role or token")

		}
		util.Log.Warn(err)

		ctx.SetStatusCode(401)
	})
}

func AuthorizeStaticToken(h fasthttp.RequestHandler, staticToken string) fasthttp.RequestHandler {
	return fasthttp.RequestHandler(func(ctx *fasthttp.RequestCtx) {
		token, _ := extractAuthToken(ctx, "Apikey")
		if staticToken != "" && token != staticToken {
			ctx.SetStatusCode(401)
			return
		}
		h(ctx)
	})
}

func Log(h fasthttp.RequestHandler) fasthttp.RequestHandler {
	return fasthttp.RequestHandler(func(ctx *fasthttp.RequestCtx) {
		start := ctx.Time()
		h(ctx)

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
