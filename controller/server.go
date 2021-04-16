package controller

import (
	"sync"
	"time"

	"github.com/fasthttp/router"
	"github.com/golang/protobuf/jsonpb"
	"github.com/rgumi/whatsapp-mock/model"
	"github.com/rgumi/whatsapp-mock/monitoring"
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
	r.RedirectFixedPath = false
	r.RedirectTrailingSlash = false
	subR := r.Group(apiPrefix)

	// general resources
	subR.POST("/generate", Limiter(GenerateWebhookRequests, 2))
	subR.POST("/generate/cancel", CancelGenerateWebhookRquests)
	subR.POST("/messages", monitoring.All(Limiter(SetConnID(Authorize(SendMessages)), 20)))
	subR.POST("/contacts", monitoring.All(Limiter(SetConnID(Authorize(Contacts)), 20)))

	subR.GET("/health", Limiter(AuthorizeStaticToken(HealthCheck, staticApiToken), 5))

	// User resources
	subR.POST("/users/login", monitoring.All(Login))
	subR.POST("/users/logout", monitoring.All(Authorize(Logout)))
	subR.POST("/users", monitoring.All(AuthorizeWithRoles(CreateUser, []string{"ADMIN"})))
	subR.DELETE("/users/{name}", monitoring.All(AuthorizeWithRoles(DeleteUser, []string{"ADMIN"})))

	// Media resources
	subR.POST("/media", monitoring.All(Authorize(SaveMedia)))
	subR.GET("/media/{id}", monitoring.All(Authorize(RetrieveMedia)))
	subR.DELETE("/media/{id}", monitoring.All(Authorize(DeleteMedia)))

	// settings resources
	subR.PATCH("/settings/application", monitoring.All(Authorize(SetApplicationSettings)))
	subR.GET("/settings/application", monitoring.All(Authorize(GetApplicationSettings)))
	subR.DELETE("/settings/application", monitoring.All(Authorize(ResetApplicationSettings)))
	subR.POST("/certificates/webhooks/ca", monitoring.All(Authorize(UploadWebhookCA)))

	// registration resources
	subR.POST("/account/verify", monitoring.All(Authorize(VerifyAccount)))
	subR.POST("/account", monitoring.All(Authorize(RegisterAccount)))

	// profile resources
	subR.PATCH("/settings/profile/about", monitoring.All(Authorize(SetProfileAbout)))
	subR.GET("/settings/profile/about", monitoring.All(Authorize(GetProfileAbout)))
	subR.POST("/settings/profile/photo", monitoring.All(Authorize(SetProfilePhoto)))
	subR.GET("/settings/profile/photo", monitoring.All(Authorize(GetProfilePhoto)))
	subR.POST("/settings/business/profile", monitoring.All(Authorize(SetBusinessProfile)))
	subR.GET("/settings/business/profile", monitoring.All(Authorize(GetBusinessProfile)))

	// stickerpacks resources
	subR.ANY("/stickerpacks/{path:*}", monitoring.All(NotImplementedHandler))

	// groups resources
	subR.ANY("/groups/{path:*}", monitoring.All(NotImplementedHandler))

	// stats resources
	subR.ANY("/stats/{path:*}", monitoring.All(NotImplementedHandler))

	r.GET("/metrics", monitoring.PrometheusHandler)
	r.PanicHandler = PanicHandler
	server := &fasthttp.Server{
		Handler:                       Log(r.Handler),
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
