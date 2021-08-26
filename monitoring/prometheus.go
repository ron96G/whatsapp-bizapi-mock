package monitoring

import (
	"compress/gzip"
	"strings"
	"sync"

	"github.com/golang/protobuf/jsonpb"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
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

	ApiRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "request_duration_milliseconds",
			Help:      "The HTTP request latencies in milliseconds.",
			Buckets:   []float64{10, 50, 100, 500, 1000},
		},
		[]string{"code", "method", "url"},
	)

	WebhookGeneratedMessages = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "webhook_generated",
			Help:      "The amount of generated objects by the webhook",
		},
		[]string{"type"},
	)
	WebhookQueueLength = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "webhook_queue_length",
			Help:      "The current length of the webhook queue",
		},
		[]string{"type"},
	)

	WebhookRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "webhook_duration_milliseconds",
			Help:      "The HTTP request latencies of the webhook in milliseconds.",
			Buckets:   []float64{10, 50, 100, 500, 1000},
		},
		[]string{"status", "url"},
	)
)

func init() {

	registry = prometheus.NewRegistry()
	coll := collectors.NewGoCollector()
	registry.MustRegister(coll)

	// API
	registry.MustRegister(ApiRequestDuration)

	// Webhook
	registry.MustRegister(WebhookQueueLength)
	registry.MustRegister(WebhookRequestDuration)
	registry.MustRegister(WebhookGeneratedMessages)
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
