package controller

import (
	"os"

	"github.com/google/uuid"
	"github.com/rgumi/whatsapp-mock/model"
	"github.com/valyala/fasthttp"
	"google.golang.org/protobuf/proto"
)

func SetProfileAbout(ctx *fasthttp.RequestCtx) {
	about := &model.ProfileAbout{}
	if !unmarshalPayload(ctx, about) {
		return
	}
	proto.Merge(Config.ProfileAbout, about)
}

func GetProfileAbout(ctx *fasthttp.RequestCtx) {
	returnJSON(ctx, 200, Config.ProfileAbout)
}

func SetProfilePhoto(ctx *fasthttp.RequestCtx) {
	profilePhotoFilename := "pp_" + uuid.New().String()

	if !savePostBody(ctx, profilePhotoFilename) {
		return
	}

	if Config.ProfilePhotoFilename != "" {
		// a profile picture already exists, delete it as not required anymore
		_ = os.Remove(Config.UploadDir + Config.ProfilePhotoFilename)
	}

	Config.ProfilePhotoFilename = profilePhotoFilename
	ctx.SetStatusCode(201)
}

func GetProfilePhoto(ctx *fasthttp.RequestCtx) {
	respondWithFile(ctx, 200, Config.ProfilePhotoFilename)
}

func SetBusinessProfile(ctx *fasthttp.RequestCtx) {
	businessProfile := &model.BusinessProfile{}
	if !unmarshalPayload(ctx, businessProfile) {
		return
	}
	proto.Merge(Config.BusinessProfile, businessProfile)
}

func GetBusinessProfile(ctx *fasthttp.RequestCtx) {
	returnJSON(ctx, 200, Config.BusinessProfile)
}
