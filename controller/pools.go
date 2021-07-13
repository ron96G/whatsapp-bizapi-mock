package controller

import (
	"bytes"
	"compress/gzip"
	"sync"

	"github.com/ron96G/whatsapp-bizapi-mock/model"
)

/*
	Pools are used for objects that may be generated very often
	To increase memory efficacy objects are reused

*/
var (
	loginResponsePool = sync.Pool{
		New: func() interface{} {
			return new(model.LoginResponse)
		},
	}

	contactResponsePool = sync.Pool{
		New: func() interface{} {
			return new(model.ContactResponse)
		},
	}

	idResponsePool = sync.Pool{
		New: func() interface{} {
			return new(model.IdResponse)
		},
	}

	errorResponsePool = sync.Pool{
		New: func() interface{} {
			return new(model.ErrorResponse)
		},
	}

	gzipPool = sync.Pool{
		New: func() interface{} {
			return gzip.NewWriter(nil)
		},
	}

	metaPool = sync.Pool{
		New: func() interface{} {
			return &model.Meta{
				ApiStatus: ApiStatus,
				Version:   ApiVersion,
			}
		},
	}

	bufferPool = sync.Pool{
		New: func() interface{} {
			return bytes.NewBuffer(nil)
		},
	}

	webhookReqPool = sync.Pool{
		New: func() interface{} {
			return new(model.WebhookRequest)
		},
	}
)

func AcquireLoginResponse() *model.LoginResponse {
	return loginResponsePool.Get().(*model.LoginResponse)
}

func ReleaseLoginResponse(s *model.LoginResponse) {
	loginResponsePool.Put(s)
}

func AcquireContactResponse() *model.ContactResponse {
	return contactResponsePool.Get().(*model.ContactResponse)
}

func ReleaseContactResponse(s *model.ContactResponse) {
	contactResponsePool.Put(s)
}

func AcquireIdResponse() *model.IdResponse {
	return idResponsePool.Get().(*model.IdResponse)
}

func ReleaseIdResponse(s *model.IdResponse) {
	idResponsePool.Put(s)
}

func AcquireErrorResponse() *model.ErrorResponse {
	return errorResponsePool.Get().(*model.ErrorResponse)
}

func ReleaseErrorResponse(s *model.ErrorResponse) {
	errorResponsePool.Put(s)
}

func AcquireMeta() *model.Meta {
	return metaPool.Get().(*model.Meta)
}

func ReleaseMeta(s *model.Meta) {
	metaPool.Put(s)
}

func AcquireGzip() *gzip.Writer {
	return gzipPool.Get().(*gzip.Writer)
}

func ReleaseGzip(s *gzip.Writer) {
	gzipPool.Put(s)
}

func AcquireBuffer() *bytes.Buffer {
	return bufferPool.Get().(*bytes.Buffer)
}

func ReleaseBuffer(s *bytes.Buffer) {
	bufferPool.Put(s)
}

func AcquireWebhookRequest() *model.WebhookRequest {
	return webhookReqPool.Get().(*model.WebhookRequest)
}

func ReleaseWebhookRequest(s *model.WebhookRequest) {
	webhookReqPool.Put(s)
}
