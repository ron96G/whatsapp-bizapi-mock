package controller

import (
	"github.com/google/uuid"
	"github.com/rgumi/whatsapp-mock/model"
	"github.com/valyala/fasthttp"
)

var (
	currentAbout           *model.ProfileAbout
	currentBusinessProfile *model.BusinessProfile
	profilePhotoFilename   = ""
)

func SetProfileAbout(ctx *fasthttp.RequestCtx) {
	currentAbout = &model.ProfileAbout{}
	unmarshalPayload(ctx, currentAbout)
}

func GetProfileAbout(ctx *fasthttp.RequestCtx) {
	returnJSON(ctx, 200, currentAbout)
}

func SetProfilePhoto(ctx *fasthttp.RequestCtx) {
	profilePhotoFilename = "pp_" + uuid.New().String()
	savePostBody(ctx, profilePhotoFilename)
}

func GetProfilePhoto(ctx *fasthttp.RequestCtx) {
	respondWithFile(ctx, profilePhotoFilename)
}

func SetBusinessProfile(ctx *fasthttp.RequestCtx) {
	currentBusinessProfile = &model.BusinessProfile{}
	unmarshalPayload(ctx, currentBusinessProfile)
}

func GetBusinessProfile(ctx *fasthttp.RequestCtx) {
	returnJSON(ctx, 200, currentBusinessProfile)
}
