package webhook

import (
	"sync"

	"github.com/ron96G/whatsapp-bizapi-mock/model"
)

var (
	webhookReqPool = sync.Pool{
		New: func() interface{} {
			return new(model.WebhookRequest)
		},
	}
)

func AcquireWebhookRequest() *model.WebhookRequest {
	return webhookReqPool.Get().(*model.WebhookRequest)
}

func ReleaseWebhookRequest(s *model.WebhookRequest) {
	webhookReqPool.Put(s)
}
