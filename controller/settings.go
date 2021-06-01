package controller

import (
	"bytes"
	"net/url"

	"github.com/rgumi/whatsapp-mock/model"
	"github.com/rgumi/whatsapp-mock/util"
	"github.com/valyala/fasthttp"
	"google.golang.org/protobuf/proto"
)

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

func BackupSettings(ctx *fasthttp.RequestCtx) {
	req := &model.BackupRequest{}
	if !unmarshalPayload(ctx, req) {
		return
	}
	buf := &bytes.Buffer{}
	marsheler.Marshal(buf, Config)
	ciphertext, err := util.Encrypt(req.Password, buf)
	if err != nil {
		util.Log.Error(err)
		returnError(ctx, 500, model.Error{
			Code:    500,
			Title:   "Unable to encrypt settings",
			Details: err.Error(),
		})
	}
	resp := &model.BackupResponse{
		Settings: &model.BackupResponse_SettingsData{
			Data: ciphertext,
		},
	}

	returnJSON(ctx, 200, resp)
}

func RestoreSettings(ctx *fasthttp.RequestCtx) {
	req := &model.RestoreRequest{}
	if !unmarshalPayload(ctx, req) {
		return
	}
	buf := bytes.NewBuffer(req.Data)
	ciphertext, err := util.Decrypt(req.Password, buf)
	if err != nil {
		util.Log.Error(err)
		returnError(ctx, 500, model.Error{
			Code:    500,
			Title:   "Unable to decrypt settings",
			Details: err.Error(),
		})
	}
	buf.Reset()
	buf = bytes.NewBuffer(ciphertext)
	err = InitConfig(buf)
	if err != nil {
		util.Log.Error(err)
		returnError(ctx, 500, model.Error{
			Code:    500,
			Title:   "Unable to set settings",
			Details: err.Error(),
		})
	}
	ctx.SetStatusCode(200)
}
