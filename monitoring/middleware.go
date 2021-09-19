package monitoring

import (
	"strconv"
	"time"

	"github.com/valyala/fasthttp"
)

func All(h fasthttp.RequestHandler) fasthttp.RequestHandler {
	return ResponseTime(h)
}

func ResponseTime(h fasthttp.RequestHandler) fasthttp.RequestHandler {
	return fasthttp.RequestHandler(func(ctx *fasthttp.RequestCtx) {
		start := ctx.Time()
		h(ctx)
		delta := float64(time.Since(start)) / float64(time.Second)
		statusStr := strconv.Itoa(ctx.Response.StatusCode())
		methodStr := string(ctx.Request.Header.Method())
		urlStr := string(ctx.Request.URI().Path())
		ApiRequestDuration.WithLabelValues(statusStr, methodStr, urlStr).Observe(delta)
	})
}
