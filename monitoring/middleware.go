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
		delta := float64(time.Since(start).Milliseconds())
		statusStr := strconv.Itoa(ctx.Response.StatusCode())
		methodStr := string(ctx.Request.Header.Method())
		urlStr := ctx.Request.URI().String()

		ApiRequestDuration.WithLabelValues(statusStr, methodStr, urlStr).Observe(delta)
	})
}

// https://math.stackexchange.com/questions/106700/incremental-averageing
func floatingAverage(a, x, k uint64) uint64 {
	if a == 0 {
		return x
	}
	return a + (x-a)/k
}
