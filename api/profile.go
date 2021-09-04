package api

import (
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/ron96G/whatsapp-bizapi-mock/model"
	"github.com/valyala/fasthttp"
	"google.golang.org/protobuf/proto"
)

func (a *API) SetProfileAbout(ctx *fasthttp.RequestCtx) {
	about := &model.ProfileAbout{}
	if !unmarshalPayload(ctx, about) {
		return
	}
	a.Config.ProfileAbout = about
}

func (a *API) GetProfileAbout(ctx *fasthttp.RequestCtx) {
	returnJSON(ctx, 200, a.Config.ProfileAbout)
}

func (a *API) SetProfilePhoto(ctx *fasthttp.RequestCtx) {
	profilePhotoFilename := "pp_" + uuid.New().String()

	if !savePostBody(ctx, profilePhotoFilename) {
		return
	}

	if a.Config.ProfilePhotoFilename != "" {
		// a profile picture already exists, delete it as not required anymore
		_ = os.Remove(a.Config.UploadDir + a.Config.ProfilePhotoFilename)
	}

	a.Config.ProfilePhotoFilename = profilePhotoFilename
	ctx.SetStatusCode(201)
}

func (a *API) GetProfilePhoto(ctx *fasthttp.RequestCtx) {
	respondWithFile(ctx, 200, filepath.Join(a.Config.UploadDir, a.Config.ProfilePhotoFilename))
}

func (a *API) SetBusinessProfile(ctx *fasthttp.RequestCtx) {
	businessProfile := &model.BusinessProfile{}
	if !unmarshalPayload(ctx, businessProfile) {
		return
	}

	proto.Merge(a.Config.BusinessProfile, businessProfile)

	// WhatsApp only allows 2 urls. Therefore, only persist the last 2 and ignore the rest
	countWebsites := len(a.Config.BusinessProfile.Websites)
	if countWebsites > 2 {
		a.Config.BusinessProfile.Websites = a.Config.BusinessProfile.Websites[countWebsites-2:]
	}
}

func (a *API) GetBusinessProfile(ctx *fasthttp.RequestCtx) {
	returnJSON(ctx, 200, a.Config.BusinessProfile)
}
