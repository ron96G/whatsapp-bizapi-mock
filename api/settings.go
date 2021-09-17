package api

import (
	"bytes"
	"net/url"

	"github.com/ron96G/whatsapp-bizapi-mock/model"
	"github.com/ron96G/whatsapp-bizapi-mock/util"
	"github.com/valyala/fasthttp"
	"google.golang.org/protobuf/proto"
)

func (a *API) SetApplicationSettings(ctx *fasthttp.RequestCtx) {
	appSettings := &model.ApplicationSettings{}
	logger := a.LoggerFromCtx(ctx)
	if err := unmarshalPayload(ctx, appSettings); err != nil {
		logger.Warn("Unable to set application settings", "error", err)
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

	proto.Merge(a.Config.ApplicationSettings, appSettings)
	a.Config.ApplicationSettings.Media.AutoDownload = appSettings.Media.AutoDownload

	a.Webhook.URL = parsedUrl.String()
	a.Log.Info("Updated webhook URL", "url", a.Webhook.URL)
	returnJSON(ctx, 200, nil)
}

func (a *API) GetApplicationSettings(ctx *fasthttp.RequestCtx) {
	returnJSON(ctx, 200, a.Config.ApplicationSettings)
}

func ResetApplicationSettings(ctx *fasthttp.RequestCtx) { notImplemented(ctx) }

func (a *API) BackupSettings(ctx *fasthttp.RequestCtx) {
	req := &model.BackupRequest{}
	logger := a.LoggerFromCtx(ctx)
	if err := unmarshalPayload(ctx, req); err != nil {
		logger.Warn("Unable to backup application settings", "error", err)
		return
	}
	buf := &bytes.Buffer{}
	marsheler.Marshal(buf, a.Config)
	ciphertext, err := util.Encrypt(req.Password, buf)
	if err != nil {
		a.Log.Error("Failed to encrypt settings", "error", err)
		returnError(ctx, 500, model.Error{
			Code:    500,
			Title:   "Unable to encrypt settings",
			Details: err.Error(),
		})
		return
	}
	resp := &model.BackupResponse{
		Settings: &model.BackupResponse_SettingsData{
			Data: ciphertext,
		},
	}

	returnJSON(ctx, 200, resp)
}

func (a *API) RestoreSettings(ctx *fasthttp.RequestCtx) {
	req := &model.RestoreRequest{}
	logger := a.LoggerFromCtx(ctx)
	if err := unmarshalPayload(ctx, req); err != nil {
		logger.Warn("Unable to restore application settings", "error", err)
		return
	}
	buf := bytes.NewBuffer(req.Data)
	ciphertext, err := util.Decrypt(req.Password, buf)
	if err != nil {
		a.Log.Error("Failed to decrypt settings", "error", err)
		returnError(ctx, 500, model.Error{
			Code:    500,
			Title:   "Unable to decrypt settings",
			Details: err.Error(),
		})
		return
	}
	buf.Reset()
	buf = bytes.NewBuffer(ciphertext)
	cfg := new(model.InternalConfig)
	err = unmarsheler.Unmarshal(buf, cfg)
	if err != nil {
		a.Log.Error("Failed to update settings", "error", err)
		returnError(ctx, 500, model.Error{
			Code:    500,
			Title:   "Unable to set settings",
			Details: err.Error(),
		})
		return
	}
	a.Config = cfg
	ctx.SetStatusCode(200)
}
