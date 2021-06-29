package controller

import (
	"time"

	"github.com/fasthttp/router"
	"github.com/rgumi/whatsapp-mock/model"
	"github.com/rgumi/whatsapp-mock/monitoring"
	"github.com/valyala/fasthttp"

	swagger "github.com/rgumi/go-fasthttp-swagger"
	_ "github.com/rgumi/whatsapp-mock/docs"
)

var (
	ApiStatus  = model.Meta_experimental
	ApiVersion = "x.xx.x"

	TokenValidDuration = 7 * 24 * time.Hour

	SigningKey []byte
	Tokens     = []string{} // all tokens that have been signed and can be used to login

	Webhook *WebhookConfig

	cancel = make(chan int, 1)

	RequestLimit = 20
)

// @title WhatsAppMockServer API
// @version 0.1
// @description The WhatsAppMockServer offers a mock API for the WhatsApp-Business-API

// @host localhost:9090
// @schemes https
// @securityDefinitions.basic BasicAuth
// @in header
// @name Authorization
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @BasePath /v1
func NewServer(apiPrefix string, staticApiToken string) *fasthttp.Server {
	r := router.New()
	r.RedirectFixedPath = false
	r.RedirectTrailingSlash = false
	subR := r.Group(apiPrefix)

	// general resources
	subR.POST("/generate", Limiter(GenerateWebhookRequests, 2))
	subR.POST("/generate/cancel", CancelGenerateWebhookRquests)
	subR.POST("/messages", monitoring.All(Limiter(SetConnID(Authorize(SendMessages)), RequestLimit)))
	subR.POST("/contacts", monitoring.All(Limiter(SetConnID(Authorize(Contacts)), RequestLimit)))

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
	subR.POST("/settings/backup", monitoring.All(Authorize(BackupSettings)))
	subR.POST("/settings/restore", monitoring.All(Authorize(RestoreSettings)))

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

	// stats resources
	subR.ANY("/stats/{path:*}", monitoring.All(NotImplementedHandler))
	subR.GET("/metrics", monitoring.PrometheusHandler)

	r.GET("/swagger/{path:*}", swagger.SwaggerHandler())
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
