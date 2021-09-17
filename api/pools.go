package api

import (
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

	metaPool = sync.Pool{
		New: func() interface{} {
			return &model.Meta{
				ApiStatus: ApiStatus,
				Version:   Version,
			}
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
