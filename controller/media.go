package controller

import (
	"bytes"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/rgumi/whatsapp-mock/model"
	"github.com/valyala/fasthttp"
)

func SaveMedia(ctx *fasthttp.RequestCtx) {
	fileID := uuid.New().String()
	_, err := mime.ExtensionsByType(string(ctx.Request.Header.ContentType()))
	if err != nil {
		returnError(ctx, 400, model.Error{
			Code:    400,
			Details: err.Error(),
			Title:   "Client Error",
			Href:    "",
		})
		return
	}
	filename := fileID
	f, err := os.OpenFile(UploadDir+filename, os.O_WRONLY|os.O_CREATE, 0777)
	if err != nil {
		returnError(ctx, 500, model.Error{
			Code:    500,
			Details: err.Error(),
			Title:   "Server Error",
			Href:    "",
		})
		return
	}
	defer f.Close()

	r := bytes.NewReader(ctx.PostBody())
	_, err = io.Copy(f, r)
	if err != nil {
		returnError(ctx, 500, model.Error{
			Code:    500,
			Details: err.Error(),
			Title:   "Server Error",
			Href:    "",
		})
		return
	}

	id := &model.Id{
		Id: fileID,
	}
	returnJSON(ctx, 200, id)
}

func RetrieveMedia(ctx *fasthttp.RequestCtx) {
	id := ctx.UserValue("id").(string)
	filename := filepath.Base(id)
	f, err := os.OpenFile(UploadDir+filename, os.O_RDONLY, 0777)
	if err != nil && os.IsNotExist(err) {
		ctx.SetStatusCode(404)
		return

	} else if err != nil {
		returnError(ctx, 500, model.Error{
			Code:    500,
			Details: err.Error(),
			Title:   "Server Error",
			Href:    "",
		})
		return
	}

	defer f.Close()
	contentType, err := getFileContentType(f)
	if err == nil {
		_, err := f.Seek(0, io.SeekStart)
		if err == nil {
			ctx.SetContentType(contentType)
			ctx.SetStatusCode(200)
			io.Copy(ctx, f)
			return
		}
	}

	returnError(ctx, 500, model.Error{
		Code:    500,
		Details: err.Error(),
		Title:   "Server Error",
		Href:    "",
	})
}

func DeleteMedia(ctx *fasthttp.RequestCtx) {
	id := ctx.UserValue("id").(string)
	filename := filepath.Base(id)
	err := os.Remove(UploadDir + filename)
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
