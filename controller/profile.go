package controller

import (
	"github.com/google/uuid"
	"github.com/rgumi/whatsapp-mock/model"
	"github.com/valyala/fasthttp"
	"google.golang.org/protobuf/proto"
)

func SetProfileAbout(ctx *fasthttp.RequestCtx) {
	about := &model.ProfileAbout{}
	unmarshalPayload(ctx, about)
	proto.Merge(Config.ProfileAbout, about)
}

func GetProfileAbout(ctx *fasthttp.RequestCtx) {
	returnJSON(ctx, 200, Config.ProfileAbout)
}

func SetProfilePhoto(ctx *fasthttp.RequestCtx) {
	profilePhotoFilename := "pp_" + uuid.New().String()
	savePostBody(ctx, profilePhotoFilename)
	Config.ProfilePhotoFilename = profilePhotoFilename
}

func GetProfilePhoto(ctx *fasthttp.RequestCtx) {
	respondWithFile(ctx, Config.ProfilePhotoFilename)
}

func SetBusinessProfile(ctx *fasthttp.RequestCtx) {
	businessProfile := &model.BusinessProfile{}
	unmarshalPayload(ctx, businessProfile)
	proto.Merge(Config.BusinessProfile, businessProfile)
}

func GetBusinessProfile(ctx *fasthttp.RequestCtx) {
	returnJSON(ctx, 200, Config.BusinessProfile)
}
