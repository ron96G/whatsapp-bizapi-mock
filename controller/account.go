package controller

import (
	"log"
	"math/rand"
	"time"

	"github.com/rgumi/whatsapp-mock/model"
	"github.com/valyala/fasthttp"
)

// random 6-digit code
var expectedVerifyCode = ""

func init() {
	rand.Seed(time.Now().UnixNano())
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

	} else {
		log.Print("Successfully verified account")
	}
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
	}

	returnJSON(ctx, 202, &model.Meta{
		ApiStatus: model.Meta_stable,
		Version:   ApiVersion,
	})
}
