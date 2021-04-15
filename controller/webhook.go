package controller

import (
	"fmt"
	"sync"
	"time"

	"github.com/rgumi/whatsapp-mock/model"
	"github.com/rgumi/whatsapp-mock/util"
	"github.com/valyala/fasthttp"
)

var (
	webhookReqPool = sync.Pool{
		New: func() interface{} {
			return new(model.WebhookRequest)
		},
	}

	userAgent = "WhatsAppMockserver/" + ApiVersion

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
)

type WebhookConfig struct {
	URL          string
	Generators   *model.Generators
	StatusQueue  []*model.Status
	MessageQueue []*model.Message
	Queue        chan *model.WebhookRequest
	WaitInterval time.Duration
	mux          sync.Mutex
}

func NewWebhookConfig(url string, g *model.Generators) *WebhookConfig {
	return &WebhookConfig{
		URL:          url,
		Generators:   g,
		Queue:        make(chan *model.WebhookRequest, 100),
		StatusQueue:  []*model.Status{},
		WaitInterval: 0 * time.Second,
	}
}

func (wc *WebhookConfig) Send(req *fasthttp.Request) (*fasthttp.Response, error) {
	resp := fasthttp.AcquireResponse()
	err := util.DefaultClient.Do(req, resp)
	return resp, err
}

func (wc *WebhookConfig) AddStati(stati ...*model.Status) {
	wc.mux.Lock()
	wc.StatusQueue = append(wc.StatusQueue, stati...)
	wc.mux.Unlock()
}

// collect all stati of outbound messages and send them to webhook
func (wc *WebhookConfig) statusRunner() (stop chan int) {
	stop = make(chan int, 1)
	go func() {
		for {
			select {
			case <-stop:
				return
			case <-time.After(StatusMergeInterval):
				wc.mux.Lock()

				if len(wc.StatusQueue) == 0 {
					wc.mux.Unlock()
					continue
				}

				whReq := webhookReqPool.Get().(*model.WebhookRequest)
				whReq.Reset()
				whReq.Statuses = wc.StatusQueue
				wc.StatusQueue = []*model.Status{}
				wc.Queue <- whReq
				wc.mux.Unlock()
			}
		}
	}()
	return
}

func (wc *WebhookConfig) GenerateWebhookRequests(numberOfEntries int, types ...string) []*model.Message {
	wc.mux.Lock()
	defer wc.mux.Unlock()

	whReq := webhookReqPool.Get().(*model.WebhookRequest)
	whReq.Reset()
	var messages []*model.Message

	if types[0] == "rnd" {
		messages = wc.Generators.GenerateRndMessages(numberOfEntries)
	} else {
		messages = wc.Generators.GenerateMessages(numberOfEntries, types...)
	}
	whReq.Messages = append(whReq.Messages, messages...)
	whReq.Contacts = append(whReq.Contacts, wc.Generators.Contacts...)
	whReq.Errors = append(whReq.Errors, nil)
	whReq.Statuses = wc.StatusQueue
	wc.StatusQueue = []*model.Status{}
	wc.Queue <- whReq
	return messages
}

func (wc *WebhookConfig) Run(errors chan error) (stop chan int) {
	stop = make(chan int, 1)
	stopStatus := wc.statusRunner()

	go func() {
		for {
			select {
			case <-stop:
				stopStatus <- 1
				return

			case whReq := <-wc.Queue:
				time.Sleep(wc.WaitInterval)
				req := fasthttp.AcquireRequest()
				marsheler.Marshal(req.BodyWriter(), whReq)
				req.SetRequestURI(wc.URL)
				req.Header.Set("User-Agent", userAgent)
				req.Header.Set("Content-Type", "application/json")
				req.Header.SetMethod("POST")
				resp, err := wc.Send(req)
				defer fasthttp.ReleaseRequest(req)
				defer fasthttp.ReleaseResponse(resp)

				if err != nil {
					wc.WaitInterval = wc.WaitInterval + 3*time.Second
					errors <- err
					wc.Queue <- whReq
					continue
				}
				if resp.StatusCode() >= 300 || resp.StatusCode() < 200 {
					wc.WaitInterval = wc.WaitInterval + 3*time.Second
					errors <- fmt.Errorf("webook-request to %s failed with status %d", wc.URL, resp.StatusCode())
					wc.Queue <- whReq
					continue
				}
				wc.WaitInterval = 0
				webhookReqPool.Put(whReq)
				util.Log.Infof("Webook-request to %s successfully returned status 2xx\n", wc.URL)

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
