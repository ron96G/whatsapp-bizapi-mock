package api

import (
	"os"
	"time"

	"github.com/fasthttp/router"
	"github.com/ron96G/whatsapp-bizapi-mock/model"
	"github.com/ron96G/whatsapp-bizapi-mock/monitoring"
	"github.com/ron96G/whatsapp-bizapi-mock/util"
	"github.com/ron96G/whatsapp-bizapi-mock/webhook"
	"github.com/valyala/fasthttp"

	swagger "github.com/ron96G/go-fasthttp-swagger"
	_ "github.com/ron96G/whatsapp-bizapi-mock/docs"

	fh_mw "github.com/ron96G/go-common-utils/fasthttp"
	log "github.com/ron96G/go-common-utils/log"
)

const (
	ApiStatus       = model.Meta_experimental
	Version         = "2.35"
	Servername      = "WhatsAppMockserver/v" + Version
	EnableAccessLog = true
	EnableTracing   = true
)

var (
	loggerConfig = fh_mw.LoggerConfig{
		Output:     os.Stdout,
		TimeFormat: log.TimeFormat,
		Format: `{"time":"${time}","hostname":"${env:HOSTNAME}","type":"access","pod_node":"${env:POD_NODE}","id":"${id}",` +
			`"remote_ip":"${remote_ip}","method":"${method}","user_agent":"${user_agent}",` +
			`"status_code":${status},"elapsed_time":${latency},"elapsed_time_human":"${latency_human}"` +
			`,"request_length":${bytes_in}}` + "\n",
	}
)

type API struct {
	Server       *fasthttp.Server
	Status       string
	Config       *model.InternalConfig
	Tokens       *util.Set
	Webhook      *webhook.Webhook
	RequestLimit uint
	Log          log.Logger
	cancel       chan int
}

func NewAPI(apiPrefix, staticApiToken string, requestLimit uint, cfg *model.InternalConfig, webhook *webhook.Webhook) *API {
	api := &API{
		Status:       model.Meta_experimental.String(),
		Config:       cfg,
		Tokens:       util.NewSet(),
		Webhook:      webhook,
		RequestLimit: requestLimit,
		Log:          log.New("api_logger", "component", "api"),
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
	subR.POST("/messages", monitoring.All(Limiter(a.Authorize(a.SendMessages), a.RequestLimit)))
	subR.POST("/contacts", monitoring.All(Limiter(a.Authorize(a.Contacts), a.RequestLimit)))

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
	subR.POST("/settings/restore", monitoring.All(a.Authorize(a.RestoreSettings)))

	// registration resources
	subR.POST("/account/verify", monitoring.All(a.Authorize(a.VerifyAccount)))
	subR.POST("/account", monitoring.All(a.Authorize(a.RegisterAccount)))

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

	r.PanicHandler = a.PanicHandler

	handler := a.SetConnID(r.Handler)
	if EnableTracing {
		handler = Tracer(handler)
	}
	if EnableAccessLog {
		handler = fh_mw.LoggerWithConfig(handler, loggerConfig)
	}

	server := &fasthttp.Server{
		Handler:                       handler,
		Name:                          Servername,
		Concurrency:                   256 * 1024,
		DisableKeepalive:              false,
		ReadTimeout:                   5 * time.Second,
		WriteTimeout:                  5 * time.Second,
		IdleTimeout:                   15 * time.Second,
		MaxConnsPerIP:                 0,
		MaxRequestsPerConn:            0,
		TCPKeepalive:                  false,
		DisableHeaderNamesNormalizing: false,
		NoDefaultServerHeader:         false,
	}

	a.Server = server
}
