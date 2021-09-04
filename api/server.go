package api

import (
	"time"

	"github.com/fasthttp/router"
	"github.com/ron96G/whatsapp-bizapi-mock/model"
	"github.com/ron96G/whatsapp-bizapi-mock/monitoring"
	"github.com/ron96G/whatsapp-bizapi-mock/util"
	"github.com/ron96G/whatsapp-bizapi-mock/webhook"
	"github.com/valyala/fasthttp"

	swagger "github.com/ron96G/go-fasthttp-swagger"
	_ "github.com/ron96G/whatsapp-bizapi-mock/docs"
)

var (
	ApiStatus = model.Meta_experimental
	Version   = "2.35"
)

type API struct {
	Server       *fasthttp.Server
	Status       string
	Config       *model.InternalConfig
	Tokens       *util.LockedList
	Webhook      *webhook.Webhook
	RequestLimit int
	cancel       chan int
}

func NewAPI(apiPrefix, staticApiToken string, cfg *model.InternalConfig, webhook *webhook.Webhook) *API {
	api := &API{
		Status:       model.Meta_experimental.String(),
		Config:       cfg,
		Tokens:       util.NewLockedList(),
		Webhook:      webhook,
		RequestLimit: 20,
		cancel:       make(chan int, 1),
	}
	api.NewServer(apiPrefix, staticApiToken)
	return api
}

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
func (a *API) NewServer(apiPrefix string, staticApiToken string) {
	r := router.New()
	r.RedirectFixedPath = false
	r.RedirectTrailingSlash = false
	subR := r.Group(apiPrefix)

	// general resources
	subR.POST("/generate", Limiter(a.GenerateWebhookRequests, 2))
	subR.POST("/generate/cancel", a.CancelGenerateWebhookRquests)
	subR.POST("/messages", monitoring.All(Limiter(SetConnID(a.Authorize(a.SendMessages)), a.RequestLimit)))
	subR.POST("/contacts", monitoring.All(Limiter(SetConnID(a.Authorize(Contacts)), a.RequestLimit)))

	subR.GET("/health", Limiter(AuthorizeStaticToken(HealthCheck, staticApiToken), 5))

	// User resources
	subR.POST("/users/login", monitoring.All(a.Login))
	subR.POST("/users/logout", monitoring.All(a.Authorize(a.Logout)))
	subR.POST("/users", monitoring.All(a.AuthorizeWithRoles(a.CreateUser, []string{"ADMIN"})))
	subR.DELETE("/users/{name}", monitoring.All(a.AuthorizeWithRoles(a.DeleteUser, []string{"ADMIN"})))

	// Media resources
	subR.POST("/media", monitoring.All(a.Authorize(SaveMedia)))
	subR.GET("/media/{id}", monitoring.All(a.Authorize(a.RetrieveMedia)))
	subR.DELETE("/media/{id}", monitoring.All(a.Authorize(a.DeleteMedia)))

	// settings resources
	subR.PATCH("/settings/application", monitoring.All(a.Authorize(a.SetApplicationSettings)))
	subR.GET("/settings/application", monitoring.All(a.Authorize(a.GetApplicationSettings)))
	subR.DELETE("/settings/application", monitoring.All(a.Authorize(ResetApplicationSettings)))
	subR.POST("/certificates/webhooks/ca", monitoring.All(a.Authorize(a.UploadWebhookCA)))
	subR.POST("/settings/backup", monitoring.All(a.Authorize(a.BackupSettings)))
	subR.POST("/settings/restore", monitoring.All(a.Authorize(RestoreSettings)))

	// registration resources
	subR.POST("/account/verify", monitoring.All(a.Authorize(VerifyAccount)))
	subR.POST("/account", monitoring.All(a.Authorize(RegisterAccount)))

	// profile resources
	subR.PATCH("/settings/profile/about", monitoring.All(a.Authorize(a.SetProfileAbout)))
	subR.GET("/settings/profile/about", monitoring.All(a.Authorize(a.GetProfileAbout)))
	subR.POST("/settings/profile/photo", monitoring.All(a.Authorize(a.SetProfilePhoto)))
	subR.GET("/settings/profile/photo", monitoring.All(a.Authorize(a.GetProfilePhoto)))
	subR.POST("/settings/business/profile", monitoring.All(a.Authorize(a.SetBusinessProfile)))
	subR.GET("/settings/business/profile", monitoring.All(a.Authorize(a.GetBusinessProfile)))

	// stickerpacks resources
	subR.ANY("/stickerpacks/{path:*}", monitoring.All(NotImplementedHandler))

	// stats resources
	subR.ANY("/stats/{path:*}", monitoring.All(NotImplementedHandler))
	subR.GET("/metrics", monitoring.All(monitoring.PrometheusHandler))

	r.GET("/swagger/{path:*}", swagger.SwaggerHandler())
	r.GET("/metrics", monitoring.All(monitoring.PrometheusHandler))

	r.PanicHandler = PanicHandler
	server := &fasthttp.Server{
		Handler:                       Log(Tracer(r.Handler)),
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

	a.Server = server
}
