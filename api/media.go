package api

import (
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/ron96G/whatsapp-bizapi-mock/model"
	"github.com/valyala/fasthttp"
)

// SaveMedia godoc
// @Summary Upload a media file
// @Description Upload a media file to the application
// @Tags media
// @Produce  json
// @Param file body string true "the media file"
// @Success 200 {object} model.IdResponse
// @Failure default {object} model.ErrorResponse
// @Router /media [post]
// @Security BearerAuth
func SaveMedia(ctx *fasthttp.RequestCtx) {
	fileID := uuid.New().String()

	if !savePostBody(ctx, fileID) {
		return
	}

	resp := AcquireIdResponse()
	resp.Reset()
	defer ReleaseIdResponse(resp)

	resp.Media = append(resp.Media, &model.Id{
		Id: fileID,
	})
	returnJSON(ctx, 200, resp)
}

// RetrieveMedia godoc
// @Summary Download a media file
// @Description Download the media file matching the defined id
// @Tags media
// @Success 200 {file} swagger.FileResponse The requested file
// @Failure default {object} model.ErrorResponse
// @Param fileid path string true "ID of the file to be downloaded"
// @Router /media/{fileid} [get]
// @Security BearerAuth
func (a *API) RetrieveMedia(ctx *fasthttp.RequestCtx) {
	id := ctx.UserValue("id").(string)
	filename := filepath.Base(id)
	respondWithFile(ctx, 200, filepath.Join(a.Config.UploadDir, filename))
}

// DeleteMedia godoc
// @Summary Delete a media file
// @Description Delete the file matching the defined parameter
// @Tags media
// @Success 200
// @Failure default {object} model.ErrorResponse
// @Param fileid path string true "ID of the file to be deleted"
// @Router /media/{fileid} [delete]
// @Security BearerAuth
func (a *API) DeleteMedia(ctx *fasthttp.RequestCtx) {
	id := ctx.UserValue("id").(string)
	filename := filepath.Base(id)
	err := os.Remove(a.Config.UploadDir + filename)
	if err == nil {
		ctx.SetStatusCode(200)
		return

	} else if os.IsNotExist(err) {
		ctx.SetStatusCode(404)
		return

	} else {
		returnError(ctx, 500, model.Error{
			Code:    500,
			Details: err.Error(),
			Title:   "Server Error",
			Href:    "",
		})
		return
	}
}
