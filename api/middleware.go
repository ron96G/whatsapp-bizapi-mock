package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/ron96G/whatsapp-bizapi-mock/util"
	"github.com/uber/jaeger-client-go/config"
	"github.com/valyala/fasthttp"
	"golang.org/x/time/rate"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

var (
	requestIDHeader = "X-Request-ID"
)

func (a *API) Authorize(h fasthttp.RequestHandler) fasthttp.RequestHandler {
	return fasthttp.RequestHandler(func(ctx *fasthttp.RequestCtx) {
		token, err := parseToken(ctx)
		if err != nil {
			util.Log.Warn(err)
		} else if token.Valid && a.Tokens.Contains(token.Raw) {
			h(ctx)
			return
		}
		ctx.SetStatusCode(401)
	})
}

func (a *API) AuthorizeWithRoles(h fasthttp.RequestHandler, roles []string) fasthttp.RequestHandler {
	return fasthttp.RequestHandler(func(ctx *fasthttp.RequestCtx) {
		token, err := parseTokenWithClaims(ctx)
		if err == nil {
			if claims, ok := token.Claims.(*CustomClaims); ok &&
				token.Valid &&
				a.Tokens.Contains(token.Raw) {

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
		accessLogger := util.Log.WithFields(map[string]interface{}{
			"type":    "access",
			"client":  ctx.RemoteAddr().String(),
			"host":    string(ctx.Host()),
			"method":  string(ctx.Method()),
			"uri":     string(ctx.RequestURI()),
			"id":      util.IfEmptySetDash(string(ctx.Response.Header.Peek(requestIDHeader))),
			"agent":   string(ctx.Request.Header.UserAgent()),
			"code":    ctx.Response.StatusCode(),
			"elapsed": float64(time.Since(start)) / float64(time.Second),
		})

		accessLogger.Println()
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

func Tracer(h fasthttp.RequestHandler) fasthttp.RequestHandler {
	serviceName := "wabiz-mockserver"
	componentName := "fasthttp"

	defcfg := config.Configuration{
		ServiceName: serviceName,
		Sampler: &config.SamplerConfig{
			Type:  "const",
			Param: 1,
		},
		Reporter: &config.ReporterConfig{
			LogSpans:            true,
			BufferFlushInterval: 1 * time.Second,
		},
	}

	config, err := defcfg.FromEnv()
	if err != nil {
		panic("Could not parse Jaeger env vars: " + err.Error())
	}

	tr, _, err := config.NewTracer()
	if err != nil {
		panic("Could not initialize jaeger tracer: " + err.Error())
	}

	opentracing.SetGlobalTracer(tr)
	util.Log.Info("Successfully initialized tracer")

	return fasthttp.RequestHandler(func(ctx *fasthttp.RequestCtx) {

		req := ctx.Request
		method := string(ctx.Method())
		url := string(ctx.Path())
		opname := "HTTP " + method + " URL: " + url
		var sp opentracing.Span
		carrier := util.NewCarrier(&req.Header)

		if c, err := tr.Extract(opentracing.HTTPHeaders, carrier); err != nil {
			sp = tr.StartSpan(opname)
		} else {
			sp = tr.StartSpan(opname, ext.RPCServerOption(c))
		}

		ext.HTTPMethod.Set(sp, method)
		ext.HTTPUrl.Set(sp, url)
		ext.Component.Set(sp, componentName)

		ctx.SetUserValue("activeSpan", sp)

		h(ctx)
		status := uint16(ctx.Response.StatusCode())
		ext.HTTPStatusCode.Set(sp, status)

		if status >= http.StatusInternalServerError {
			ext.Error.Set(sp, true)
		}

		sp.Finish()
	})
}
