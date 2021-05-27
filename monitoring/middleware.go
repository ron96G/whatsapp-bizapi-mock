package monitoring

import (
	"sync/atomic"
	"time"

	"github.com/valyala/fasthttp"
)

func All(h fasthttp.RequestHandler) fasthttp.RequestHandler {
	return ResponseTime(ContentLength(CountRequest(h)))
}

func CountRequest(h fasthttp.RequestHandler) fasthttp.RequestHandler {
	return fasthttp.RequestHandler(func(ctx *fasthttp.RequestCtx) {
		h(ctx)
		atomic.AddUint64(&currentRequestCount, 1)
	})
}

func ContentLength(h fasthttp.RequestHandler) fasthttp.RequestHandler {
	return fasthttp.RequestHandler(func(ctx *fasthttp.RequestCtx) {
		h(ctx)
		atomic.StoreUint64(
			&currentAvgContentLength,
			floatingAverage(currentAvgContentLength, uint64(ctx.Request.Header.ContentLength()), currentRequestCount),
		)
	})
}

func ResponseTime(h fasthttp.RequestHandler) fasthttp.RequestHandler {
	return fasthttp.RequestHandler(func(ctx *fasthttp.RequestCtx) {
		start := ctx.Time()
		h(ctx)
		HTTPResponseTime.Observe(time.Since(start).Seconds())

		/*
			atomic.StoreUint64(
				&currentAvgResponseTime,
				floatingAverage(currentAvgResponseTime, uint64(time.Since(start).Milliseconds()), currentRequestCount),
			)
		*/
	})
}

// https://math.stackexchange.com/questions/106700/incremental-averageing
func floatingAverage(a, x, k uint64) uint64 {
	if a == 0 {
		return x
	}
	return a + (x-a)/k
}
