package controller

import (
	"crypto/tls"
	"crypto/x509"
	"net/url"
	"time"

	"github.com/rgumi/whatsapp-mock/model"
	"github.com/rgumi/whatsapp-mock/util"
	"github.com/valyala/fasthttp"
)

// TODO also support other parameters which are given here (e.g. auto_download)
func SetApplicationSettings(ctx *fasthttp.RequestCtx) {
	settings := &model.ApplicationSettings{}
	if !unmarshalPayload(ctx, settings) {
		return
	}

	parsedUrl, err := url.Parse(settings.Webhooks.Url)
	if err != nil {
		returnError(ctx, 400, model.Error{
			Code:    400,
			Title:   "Unable to parse request body",
			Details: "Failed to parse uploaded webhook url",
		})
		return
	}

	if parsedUrl.Scheme != "https" {
		returnError(ctx, 400, model.Error{
			Code:    400,
			Title:   "Unsupported scheme for webhook url",
			Details: "Webhook scheme must be https",
		})
		return
	}

	Webhook.URL = parsedUrl.String()
	util.Log.Infof("Updated webhook URL to %s", Webhook.URL)
	returnJSON(ctx, 200, nil)
}

func GetApplicationSettings(ctx *fasthttp.RequestCtx) { notImplemented(ctx) }

func ResetApplicationSettings(ctx *fasthttp.RequestCtx) { notImplemented(ctx) }

// certs
func UploadWebhookCA(ctx *fasthttp.RequestCtx) {

	uploadedCert := ctx.PostBody() //  this should be the CA
	caCertPool := x509.NewCertPool()

	if !caCertPool.AppendCertsFromPEM(uploadedCert) {
		returnError(ctx, 400, model.Error{
			Code:    400,
			Title:   "Unable to parse request body",
			Details: "Failed to parse uploaded certificate",
		})
		return
	}

	tlsConfig := &tls.Config{
		RootCAs:            caCertPool,
		InsecureSkipVerify: false,
	}

	// overwrite the current default client
	// this will be propagated to the webhook
	util.DefaultClient = &fasthttp.Client{
		NoDefaultUserAgentHeader:      true,
		DisablePathNormalizing:        false,
		DisableHeaderNamesNormalizing: false,
		ReadTimeout:                   5 * time.Second,
		WriteTimeout:                  5 * time.Second,
		TLSConfig:                     tlsConfig,
		MaxConnsPerHost:               8,
		MaxIdleConnDuration:           30 * time.Second,
		MaxConnDuration:               0, // unlimited
		MaxIdemponentCallAttempts:     2,
	}

	ctx.SetStatusCode(200)
}
