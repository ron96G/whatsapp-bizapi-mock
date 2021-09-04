package api

import (
	"crypto/tls"
	"crypto/x509"
	"time"

	"github.com/ron96G/whatsapp-bizapi-mock/model"
	"github.com/ron96G/whatsapp-bizapi-mock/util"
	"github.com/valyala/fasthttp"
)

func (a *API) UploadWebhookCA(ctx *fasthttp.RequestCtx) {

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
		MinVersion:         tls.VersionTLS12,
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
	a.Config.WebhookCA = uploadedCert

	ctx.SetStatusCode(200)
}
