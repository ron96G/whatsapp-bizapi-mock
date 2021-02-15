package controller

import (
	"bytes"
	"io"
	"io/ioutil"
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
	b, err := ioutil.ReadAll(f)
	if err != nil {
		returnError(ctx, 500, model.Error{
			Code:    500,
			Details: err.Error(),
			Title:   "Server Error",
			Href:    "",
		})
		return
	}
	contentType := http.DetectContentType(b)

	ctx.SetContentType(contentType)
	ctx.SetStatusCode(200)
	ctx.Write(b)
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
