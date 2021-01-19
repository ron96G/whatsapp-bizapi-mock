package controller

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/form3tech-oss/jwt-go"
	"github.com/google/uuid"
	"github.com/valyala/fasthttp"
	"golang.org/x/time/rate"
)

func extractToken(ctx *fasthttp.RequestCtx) string {
	auth := string(ctx.Request.Header.Peek("Authorization"))
	return strings.TrimPrefix(auth, "Bearer ")
}

func verifyToken(ctx *fasthttp.RequestCtx) (*jwt.Token, error) {
	tokenString := extractToken(ctx)
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return SigningKey, nil
	})
	if err != nil {
		return nil, err
	}
	return token, nil
}

func contains(slice []string, item string) bool {
	for _, element := range slice {
		if element == item {
			return true
		}
	}
	return false
}

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
		if limiter.Allow() == false {
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
