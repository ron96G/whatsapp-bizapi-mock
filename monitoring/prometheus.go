package monitoring

import (
	"compress/gzip"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/golang/protobuf/jsonpb"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/expfmt"
	"github.com/valyala/fasthttp"
)

var (
	marsheler = jsonpb.Marshaler{
		EmitDefaults: false,
		EnumsAsInts:  false,
		OrigName:     true,
		Indent:       "  ",
	}

	gzipPool = sync.Pool{
		New: func() interface{} {
			return gzip.NewWriter(nil)
		},
	}

	registry  *prometheus.Registry
	namespace = "whatsapp_mock"

	currentRequestCount, currentAvgContentLength, currentAvgResponseTime, currentAvgWebhookResponseTime uint64

	// TotalHTTPRequests is the total amount of http requests that were received
	TotalHTTPRequests = prometheus.NewCounterFunc(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "total_http_requests",
			Help:      "the total amount of http requests that were received",
		},
		func() float64 {
			return float64(atomic.LoadUint64(&currentRequestCount))
		},
	)

	// AvgContentLength is the average content length of requests
	AvgContentLength = prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "average_content_length",
			Help:      "the average content length of requests",
		},
		func() float64 {
			return float64(currentAvgContentLength)
		},
	)

	// AvgResponseTime is the average response time of the backend
	AvgResponseTime = prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "average_response_time",
			Help:      "the average response time",
		},
		func() float64 {
			return float64(currentAvgResponseTime)
		},
	)

	HTTPResponseTime = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "http_response_time",
			Help:      "the response time of the webserver",
			Buckets:   prometheus.ExponentialBuckets(0.0001, 10, 5),
		},
	)

	TotalWebhookRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "total_webhook_messages",
			Help:      "the total amount of webhook messages generated",
		},
		[]string{},
	)

	TotalGeneratedMessages = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "total_webhook_messages",
			Help:      "the total amount of webhook messages generated",
		},
		[]string{},
	)

	WebhookQueueLength = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "webhook_message_queue_length",
			Help:      "the current length of the webhook queue",
		},
		[]string{"type"},
	)

	AvgWebhookResponseTime = prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "average_webhook_response_time",
			Help:      "the average response time of a webook request",
		},
		func() float64 {
			return float64(currentAvgWebhookResponseTime)
		},
	)
)

func init() {
	currentRequestCount = 1 // set it to 1 to avoid any dumb division by 0 errors
	registry = prometheus.NewRegistry()
	coll := prometheus.NewGoCollector()
	registry.MustRegister(coll)

	// HTTP
	registry.MustRegister(TotalHTTPRequests)
	registry.MustRegister(AvgResponseTime)
	registry.MustRegister(AvgContentLength)
	registry.MustRegister(HTTPResponseTime)

	// Webhook
	registry.MustRegister(TotalGeneratedMessages)
	registry.MustRegister(WebhookQueueLength)
	registry.MustRegister(AvgWebhookResponseTime)

}

func PrometheusHandler(ctx *fasthttp.RequestCtx) {
	format := strings.ToLower(string(ctx.QueryArgs().Peek("format")))
	data, err := registry.Gather()
	if err != nil {
		ctx.SetStatusCode(500)
		return
	}

	ctx.Response.Header.Add("Content-Encoding", "gzip")
	gz := gzipPool.Get().(*gzip.Writer)
	defer gzipPool.Put(gz)

	if format == "prometheus" {
		ctx.SetContentType("text/plain; charset=utf-8")
		gz.Reset(ctx)

		for _, entry := range data {
			expfmt.MetricFamilyToText(gz, entry) // will this ever throw an error? o.O
		}
		gz.Close()

	} else {
		ctx.SetContentType("application/json")
		gz.Reset(ctx)

		for _, entry := range data {
			marsheler.Marshal(gz, entry) // will this ever throw an error? o.O
		}
		gz.Close()
	}
}
