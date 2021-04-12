package controller

import (
	"sync"
	"time"

	"github.com/fasthttp/router"
	"github.com/golang/protobuf/jsonpb"
	"github.com/rgumi/whatsapp-mock/model"
	"github.com/valyala/fasthttp"
)

var (
	ApiVersion         = "3.31.5"
	TokenValidDuration = 7 * 24 * time.Hour
	marsheler          = jsonpb.Marshaler{
		EmitDefaults: false,
		EnumsAsInts:  false,
		OrigName:     true,
		Indent:       "  ",
	}
	responsePool = sync.Pool{
		New: func() interface{} {
			return new(model.APIResponse)
		},
	}

	SigningKey []byte
	Tokens     = []string{} // all tokens that have been signed and can be used to login

	Webhook *WebhookConfig

	cancel = make(chan int, 1)
)

func AcquireResponse() *model.APIResponse {
	return responsePool.Get().(*model.APIResponse)
}

func ReleaseResponse(s *model.APIResponse) {
	responsePool.Put(s)
}

func NewServer(apiPrefix string, staticApiToken string) *fasthttp.Server {
	r := router.New()

	// general resources
	r.POST(apiPrefix+"/generate", Limiter(Log(GenerateWebhookRequests), 2))
	r.POST(apiPrefix+"/generate/cancel", Log(CancelGenerateWebhookRquests))
	r.POST(apiPrefix+"/messages", Limiter(SetConnID(Log(Authorize(SendMessages))), 20))
	r.POST(apiPrefix+"/contacts", Limiter(SetConnID(Log(Authorize(Contacts))), 20))
	r.GET(apiPrefix+"/health", Limiter(Log(AuthorizeStaticToken(HealthCheck, staticApiToken)), 20))

	// User resources
	r.POST(apiPrefix+"/users/login", Log(Login))
	r.POST(apiPrefix+"/users/logout", Log(Authorize(Logout)))
	r.POST(apiPrefix+"/users", Log(Authorize(CreateUser)))
	r.DELETE(apiPrefix+"/users/{name}", Log(Authorize(DeleteUser)))

	// Media resources
	r.POST(apiPrefix+"/media", Log(Authorize(SaveMedia)))
	r.GET(apiPrefix+"/media/{id}", Log(Authorize(RetrieveMedia)))
	r.DELETE(apiPrefix+"/media/{id}", Log(Authorize(DeleteMedia)))

	// settings resources
	r.PATCH(apiPrefix+"/settings/application", Log(Authorize(SetApplicationSettings)))
	r.GET(apiPrefix+"/settings/application", Log(Authorize(GetApplicationSettings)))
	r.DELETE(apiPrefix+"/settings/application", Log(Authorize(ResetApplicationSettings)))
	r.POST(apiPrefix+"/certificates/webhooks/ca", Log(Authorize(UploadWebhookCA)))

	// registration resources
	r.POST(apiPrefix+"/account/verify", Log(Authorize(VerifyAccount)))
	r.POST(apiPrefix+"/account", Log(Authorize(RegisterAccount)))

	// profile resources
	r.PATCH(apiPrefix+"/settings/profile/about", Log(Authorize(SetProfileAbout)))
	r.GET(apiPrefix+"/settings/profile/about", Log(Authorize(GetProfileAbout)))
	r.POST(apiPrefix+"/settings/profile/photo", Log(Authorize(SetProfilePhoto)))
	r.GET(apiPrefix+"/settings/profile/photo", Log(Authorize(GetProfilePhoto)))
	r.POST(apiPrefix+"/settings/business/profile", Log(Authorize(SetBusinessProfile)))
	r.GET(apiPrefix+"/settings/business/profile", Log(Authorize(GetBusinessProfile)))

	// stickerpacks resources
	r.ANY(apiPrefix+"/stickerpacks/{path:*}", Log(NotImplementedHandler))

	// groups resources
	r.ANY(apiPrefix+"/groups/{path:*}", Log(NotImplementedHandler))

	// stats resources
	r.ANY(apiPrefix+"/stats/{path:*}", Log(NotImplementedHandler))

	r.PanicHandler = PanicHandler
	server := &fasthttp.Server{
		Handler:                       r.Handler,
		Name:                          "WhatsAppMockserver",
		Concurrency:                   256 * 1024,
		DisableKeepalive:              false,
		ReadTimeout:                   5 * time.Second,
		WriteTimeout:                  5 * time.Second,
		IdleTimeout:                   30 * time.Second,
		MaxConnsPerIP:                 0,
		MaxRequestsPerConn:            0,
		TCPKeepalive:                  false,
		DisableHeaderNamesNormalizing: false,
		NoDefaultServerHeader:         false,
	}

	return server
}
