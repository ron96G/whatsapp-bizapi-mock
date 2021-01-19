package controller

import (
	"time"

	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
)

func NewServer(apiPrefix string) *fasthttp.Server {
	r := router.New()

	r.POST(apiPrefix+"/generate", Limiter(Log(GenerateWebhookRequests), 2))
	r.POST(apiPrefix+"/generate/cancel", Log(CancelGenerateWebhookRquests))
	r.POST(apiPrefix+"/messages", Limiter(SetConnID(Log(Authorize(SendMessages))), 20))
	r.POST(apiPrefix+"/contacts", Limiter(SetConnID(Log(Authorize(Contacts))), 20))

	// User resources
	r.POST(apiPrefix+"/users/login", Log(Login))
	r.POST(apiPrefix+"/users/logout", Log(Authorize(Logout)))
	r.POST(apiPrefix+"/users", Log(Authorize(CreateUser)))
	r.DELETE(apiPrefix+"/users/{name}", Log(Authorize(DeleteUser)))

	// Media resources
	r.POST(apiPrefix+"/media", Log(Authorize(SaveMedia)))
	r.GET(apiPrefix+"/media/{id}", Log(Authorize(RetrieveMedia)))
	r.DELETE(apiPrefix+"/media/{id}", Log(Authorize(DeleteMedia)))
	r.PanicHandler = PanicHandler

	server := &fasthttp.Server{
		Handler:                       r.Handler,
		Name:                          "WhatsApp Mockserver",
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
