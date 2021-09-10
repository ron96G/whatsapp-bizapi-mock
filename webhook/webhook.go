package webhook

import (
	"fmt"
	"io"
	"strconv"
	"sync"
	"time"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/ron96G/whatsapp-bizapi-mock/model"
	"github.com/ron96G/whatsapp-bizapi-mock/monitoring"
	"github.com/ron96G/whatsapp-bizapi-mock/util"
	"github.com/valyala/fasthttp"

	log "github.com/ron96G/go-common-utils/log"
)

var (
	Compress        = false
	CompressMinsize = 2048

	// StatusMergeInterval determines the duration in which the status queue of outbound messages
	// is checked (>0) and merged into a new webhook-request, which is then added to the webhook queue.
	// This should never be set below 2 seconds to avoid starvation of the webhook queue
	//
	// 1. A shorter duration means that more webhook requests are created with just status-information rather than
	// generated inbound messages aswell. However, status-information will be quicker to be sent to the webhook.
	//
	// 2. A longer duration means that more status-information will be merged with generated inbound messages.
	// However, with little generated inbound messages, the status-information will take longer to be sent to the webhook
	// (Max is the here defined duration).
	StatusMergeInterval = 3 * time.Second

	marsheler = jsonpb.Marshaler{
		EmitDefaults: false,
		EnumsAsInts:  false,
		OrigName:     true,
		Indent:       "  ",
	}
)

type Webhook struct {
	URL          string
	Generators   *model.Generators
	StatusQueue  []*model.Status
	MessageQueue []*model.Message
	Queue        chan *model.WebhookRequest
	Log          log.Logger
	WaitInterval time.Duration
	userAgent    string
	mux          sync.Mutex
}

func NewWebhook(url, version string, g *model.Generators) *Webhook {
	return &Webhook{
		URL:          url,
		Generators:   g,
		Queue:        make(chan *model.WebhookRequest, 100),
		Log:          log.New("webhook_logger", "component", "webhook"),
		userAgent:    "WhatsappMockserver/" + version,
		StatusQueue:  []*model.Status{},
		WaitInterval: 0 * time.Second,
	}
}

func (w *Webhook) Send(req *fasthttp.Request) (*fasthttp.Response, error) {
	start := time.Now()
	urlStr := string(req.URI().Path())
	resp := fasthttp.AcquireResponse()
	err := util.DefaultClient.Do(req, resp)
	delta := float64(time.Since(start)) / float64(time.Second)
	if err != nil {
		monitoring.WebhookRequestDuration.WithLabelValues("failed", urlStr).Observe(delta)
		return nil, err
	}

	statusStr := strconv.Itoa(resp.StatusCode())
	monitoring.WebhookRequestDuration.WithLabelValues(statusStr, urlStr).Observe(delta)
	return resp, err
}

func (w *Webhook) AddStati(stati ...*model.Status) {
	w.mux.Lock()
	w.StatusQueue = append(w.StatusQueue, stati...)
	w.mux.Unlock()
	amount := float64(len(stati))
	monitoring.WebhookGeneratedMessages.With(prometheus.Labels{"type": "status"}).Add(amount)
	monitoring.WebhookQueueLength.With(prometheus.Labels{"type": "status"}).Add(amount)
}

// collect all stati of outbound messages and send them to webhook
func (w *Webhook) statusRunner() (stop chan int) {
	stop = make(chan int, 1)
	go func() {
		for {
			select {
			case <-stop:
				return
			case <-time.After(StatusMergeInterval):
				w.mux.Lock()

				if len(w.StatusQueue) == 0 {
					w.mux.Unlock()
					continue
				}

				whReq := AcquireWebhookRequest()
				whReq.Reset()
				whReq.Statuses = w.StatusQueue
				w.StatusQueue = []*model.Status{}
				w.Queue <- whReq
				w.mux.Unlock()
			}
		}
	}()
	return
}

func (w *Webhook) GenerateWebhookRequests(numberOfEntries int, types ...string) []*model.Message {
	w.mux.Lock()
	defer w.mux.Unlock()

	whReq := webhookReqPool.Get().(*model.WebhookRequest)
	whReq.Reset()
	var messages []*model.Message

	if types[0] == "rnd" {
		messages = w.Generators.GenerateRndMessages(numberOfEntries)
	} else {
		messages = w.Generators.GenerateMessages(numberOfEntries, types...)
	}
	whReq.Messages = append(whReq.Messages, messages...)
	whReq.Contacts = append(whReq.Contacts, w.Generators.Contacts...)
	whReq.Errors = nil // Set the errors array to nil to skip it in marshalling
	whReq.Statuses = w.StatusQueue
	w.StatusQueue = []*model.Status{}
	w.Queue <- whReq

	amount := float64(numberOfEntries)
	monitoring.WebhookQueueLength.With(prometheus.Labels{"type": "message"}).Add(amount)
	monitoring.WebhookGeneratedMessages.With(prometheus.Labels{"type": "message"}).Add(amount)
	return messages
}

func (w *Webhook) Run(errors chan error) (stop chan int) {
	stop = make(chan int, 1)
	stopStatus := w.statusRunner()

	go func() {

		for {
			select {
			case <-stop:
				stopStatus <- 1
				return

			case whReq := <-w.Queue:
				var err error
				msgCount := len(whReq.Messages)
				staCount := len(whReq.Statuses)

				time.Sleep(w.WaitInterval)
				req := fasthttp.AcquireRequest()
				defer fasthttp.ReleaseRequest(req)

				writer := req.BodyWriter()

				buf := util.AcquireBuffer()
				buf.Reset()
				defer util.ReleaseBuffer(buf)

				if err := marsheler.Marshal(buf, whReq); err != nil {
					w.WaitInterval = w.WaitInterval + 3*time.Second
					errors <- err
					w.Queue <- whReq
					continue
				}

				if Compress && buf.Len() > CompressMinsize {
					gz := util.AcquireGzip()
					defer util.ReleaseGzip(gz)

					gz.Reset(writer)
					_, err = io.Copy(gz, buf)
					gz.Close()

					if err != nil {
						errors <- err
					} else {
						req.Header.Add("Content-Encoding", "gzip")
						goto send
					}

				}

				_, err = io.Copy(writer, buf)
				if err != nil {
					w.WaitInterval = w.WaitInterval + 3*time.Second
					errors <- err
					w.Queue <- whReq
					continue
				}

			send:
				req.SetRequestURI(w.URL)
				req.Header.Set("User-Agent", w.userAgent)
				req.Header.Set("Content-Type", "application/json")
				req.Header.SetMethod("POST")

				resp, err := w.Send(req)
				if err != nil {
					w.WaitInterval = w.WaitInterval + 3*time.Second
					errors <- err
					w.Queue <- whReq
					continue
				}
				defer fasthttp.ReleaseResponse(resp)
				code := resp.StatusCode()

				if code >= 300 || code < 200 {
					w.WaitInterval = w.WaitInterval + 3*time.Second
					errors <- fmt.Errorf("webook to %s failed with status %d", w.URL, code)
					w.Queue <- whReq
					continue
				}

				monitoring.WebhookQueueLength.With(prometheus.Labels{"type": "message"}).Sub(float64(msgCount))
				monitoring.WebhookQueueLength.With(prometheus.Labels{"type": "status"}).Sub(float64(staCount))

				w.WaitInterval = 2
				ReleaseWebhookRequest(whReq)
				w.Log.Info("Webhook succeeded", "url", w.URL, "status_code", code)

				for _, msg := range whReq.Messages {
					model.ReleaseMessage(msg)
				}
				for _, s := range whReq.Statuses {
					model.ReleaseStatus(s)
				}
			}
		}
	}()
	return
}
