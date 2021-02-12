package controller

import (
	"crypto/tls"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/rgumi/whatsapp-mock/model"
	"github.com/valyala/fasthttp"
)

var (
	webhookReqPool = sync.Pool{
		New: func() interface{} {
			return new(model.WebhookRequest)
		},
	}
)

type WebhookConfig struct {
	URL          string
	Generators   *model.Generators
	StatusQueue  []*model.Status
	MessageQueue []*model.Message
	Queue        chan *model.WebhookRequest
	client       *fasthttp.Client
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
		client: &fasthttp.Client{
			NoDefaultUserAgentHeader:      true,
			DisablePathNormalizing:        false,
			DisableHeaderNamesNormalizing: false,
			ReadTimeout:                   5 * time.Second,
			WriteTimeout:                  5 * time.Second,
			TLSConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
			MaxConnsPerHost:           8,
			MaxIdleConnDuration:       30 * time.Second,
			MaxConnDuration:           0, // unlimited
			MaxIdemponentCallAttempts: 2,
		},
	}
}

func (wc *WebhookConfig) Send(req *fasthttp.Request) (*fasthttp.Response, error) {
	resp := fasthttp.AcquireResponse()
	err := wc.client.Do(req, resp)
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
			case _ = <-stop:
				return
			case _ = <-time.After(10 * time.Second):
				if len(wc.StatusQueue) == 0 {
					continue
				}

				wc.mux.Lock()
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

func (wc *WebhookConfig) GenerateWebhookRequests(numberOfEntries int) {
	wc.mux.Lock()
	defer wc.mux.Unlock()

	whReq := webhookReqPool.Get().(*model.WebhookRequest)
	whReq.Reset()
	messages := wc.Generators.GenerateRndMessages(numberOfEntries)
	whReq.Messages = append(whReq.Messages, messages...)
	whReq.Contacts = append(whReq.Contacts, wc.Generators.Contacts...)
	whReq.Errors = append(whReq.Errors, nil)
	whReq.Statuses = wc.StatusQueue
	wc.StatusQueue = []*model.Status{}
	wc.Queue <- whReq
}

func (wc *WebhookConfig) Run(errors chan error) (stop chan int) {
	stop = make(chan int, 1)
	stopStatus := wc.statusRunner()

	go func() {
		for {
			select {
			case _ = <-stop:
				stopStatus <- 1
				return

			case whReq := <-wc.Queue:
				time.Sleep(wc.WaitInterval)
				req := fasthttp.AcquireRequest()
				marsheler.Marshal(req.BodyWriter(), whReq)
				req.SetRequestURI(wc.URL)
				req.Header.Set("User-Agent", "WhatsApp Mockserver")
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
					errors <- fmt.Errorf("Webook-request to %s failed with status %d", wc.URL, resp.StatusCode())
					wc.Queue <- whReq
					continue
				}
				wc.WaitInterval = 0
				webhookReqPool.Put(whReq)
				log.Printf("Webook-request to %s successfully returned status 2xx\n", wc.URL)

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
