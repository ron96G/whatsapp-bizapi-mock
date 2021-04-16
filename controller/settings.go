package controller

import (
	"net/url"

	"github.com/rgumi/whatsapp-mock/model"
	"github.com/rgumi/whatsapp-mock/util"
	"github.com/valyala/fasthttp"
	"google.golang.org/protobuf/proto"
)

// TODO also support other parameters which are given here (e.g. auto_download)
func SetApplicationSettings(ctx *fasthttp.RequestCtx) {
	appSettings := &model.ApplicationSettings{}
	if !unmarshalPayload(ctx, appSettings) {
		return
	}

	parsedUrl, err := url.Parse(appSettings.Webhooks.Url)
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

	proto.Merge(Config.ApplicationSettings, appSettings)
	Config.ApplicationSettings.Media.AutoDownload = appSettings.Media.AutoDownload

	Webhook.URL = parsedUrl.String()
	util.Log.Infof("Updated webhook URL to %s", Webhook.URL)
	returnJSON(ctx, 200, nil)
}

func GetApplicationSettings(ctx *fasthttp.RequestCtx) {
	returnJSON(ctx, 200, Config.ApplicationSettings)
}

func ResetApplicationSettings(ctx *fasthttp.RequestCtx) { notImplemented(ctx) }
