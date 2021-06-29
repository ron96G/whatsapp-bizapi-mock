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

	currentRequestCount, currentAvgContentLength uint64

	// AvgContentLength is the average content length of requests
	AvgContentLength = prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "average_content_length",
			Help:      "the average content length of the response body",
		},
		func() float64 {
			return float64(currentAvgContentLength)
		},
	)

	HTTPResponseTime = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "http_response_time",
			Help:      "the response time of the webserver in milliseconds",
			Buckets:   []float64{1, 5, 10, 100, 500},
		},
	)

	TotalGeneratedMessages = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "webhook_total_messages",
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

	AvgWebhookResponseTime = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "webhook_response_time",
			Help:      "the response time of a webook request in milliseconds",
			Buckets:   []float64{1, 5, 10, 100, 500},
		},
	)
)

func init() {

	registry = prometheus.NewRegistry()
	coll := collectors.NewGoCollector()
	registry.MustRegister(coll)

	// HTTP
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
