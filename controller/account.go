package controller

import (
	"math/rand"
	"time"

	"github.com/rgumi/whatsapp-mock/model"
	"github.com/rgumi/whatsapp-mock/util"
	"github.com/valyala/fasthttp"
)

var (
	// Verified can be used to identify if the instance has been verified yet
	Verified = false

	// random 6-digit code which is used to verify the instance
	expectedVerifyCode = ""
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// RegisterAccount mocks the registration of  an account for this instance
func RegisterAccount(ctx *fasthttp.RequestCtx) {

	req := new(model.RegistrationRequest)
	if !unmarshalPayload(ctx, req) {
		return
	}

	if v := req.Validate(); !v.IsValid() {
		returnError(ctx, 400, v...)
		return
	}

	// in reality only 1 code can be requested at a time
	if expectedVerifyCode == "" {
		expectedVerifyCode = generateRandomCode(6)
		util.Log.Warnf("GENERATED VERIFY CODE %s", expectedVerifyCode)
	}

	returnJSON(ctx, 202, &model.Meta{
		ApiStatus: model.Meta_stable,
		Version:   ApiVersion,
	})
}

// VerifyAccount mocks the verification to finish the registration of an account
func VerifyAccount(ctx *fasthttp.RequestCtx) {

	req := new(model.VerifyRequest)
	if !unmarshalPayload(ctx, req) {
		return
	}

	if req.Code != expectedVerifyCode {
		returnError(ctx, 400, model.Error{
			Code:    400,
			Details: "Wrong verification code",
			Title:   "Client Error",
			Href:    "",
		})
		return
	}

	util.Log.Print("Successfully verified account")
	Verified = true
	expectedVerifyCode = "" // reset code
}
