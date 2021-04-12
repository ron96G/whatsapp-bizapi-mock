package controller

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/rgumi/whatsapp-mock/model"
	"github.com/valyala/fasthttp"
)

func SaveMedia(ctx *fasthttp.RequestCtx) {
	fileID := uuid.New().String()

	if !savePostBody(ctx, fileID) {
		return
	}

	resp := AcquireResponse()
	resp.Reset()
	defer ReleaseResponse(resp)

	resp.Media = append(resp.Media, &model.Id{
		Id: fileID,
	})
	returnJSON(ctx, 200, resp)
}

func RetrieveMedia(ctx *fasthttp.RequestCtx) {
	id := ctx.UserValue("id").(string)
	filename := filepath.Base(id)
	respondWithFile(ctx, filename)
}

func DeleteMedia(ctx *fasthttp.RequestCtx) {
	id := ctx.UserValue("id").(string)
	filename := filepath.Base(id)
	err := os.Remove(Config.UploadDir + filename)
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

func getFileContentType(out *os.File) (string, error) {

	// Only the first 512 bytes are used to sniff the content type.
	buffer := make([]byte, 512)

	_, err := out.Read(buffer)
	if err != nil {
		return "", err
	}

	// Use the net/http package's handy DectectContentType function. Always returns a valid
	// content-type by returning "application/octet-stream" if no others seemed to match.
	contentType := http.DetectContentType(buffer)

	return contentType, nil
}
