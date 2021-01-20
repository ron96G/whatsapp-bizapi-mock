package controller

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/golang/protobuf/jsonpb"
	"github.com/rgumi/whatsapp-mock/model"
	"github.com/valyala/fasthttp"
)

func Login(ctx *fasthttp.RequestCtx) {
	username, password, err := basicAuth(ctx)
	if err != nil {
		returnError(ctx, 401, model.Error{
			Code:    401,
			Details: err.Error(),
			Title:   "Client Error",
		})
		return
	}
	if pwd, ok := Users[username]; ok {
		if pwd == password { // check if entered password is correct

			if pwd == "secret" { // check if the password has been changed, if not, it must be done now
				chPwdReq := new(model.ChangePwdRequest)
				err := jsonpb.Unmarshal(bytes.NewReader(ctx.PostBody()), chPwdReq)
				if err != nil || chPwdReq.NewPassword == "" { // check if the request contained a new password
					returnError(ctx, 400, model.Error{
						Code:    400,
						Details: "Password change required",
						Title:   "Client Error",
						Href:    "",
					})
					return
				}
				Users[username] = chPwdReq.NewPassword // change the password
			}

			// generate new token for the user
			newToken, err := generateToken(username)
			if err != nil {
				returnError(ctx, 500, model.Error{
					Code:    500,
					Details: err.Error(),
					Title:   "Server Error",
				})
			}
			returnToken(ctx, newToken)
			return
		}
	}

	returnError(ctx, 401, model.Error{
		Code:    401,
		Details: "Username or password is invalid",
		Title:   "Client Error",
		Href:    "",
	})
}

func Logout(ctx *fasthttp.RequestCtx) {
	auth := string(ctx.Request.Header.Peek("Authorization"))
	for i, token := range Tokens {
		if token == strings.TrimPrefix(auth, "Bearer ") {
			Tokens = append(Tokens[:i], Tokens[i+1:]...)
			return
		}
	}
}

func CreateUser(ctx *fasthttp.RequestCtx) {
	msg := &model.User{}
	if !unmarshalPayload(ctx, msg) {
		return
	}

	response := AcquireResponse()
	response.Reset()
	defer ReleaseResponse(response)

	response.Meta = &model.Meta{
		ApiStatus: model.Meta_stable,
		Version:   ApiVersion,
	}
	returnJSON(ctx, 200, response)
}

func DeleteUser(ctx *fasthttp.RequestCtx) {
	name := ctx.UserValue("name").(string)

	if _, ok := Users[name]; ok {
		delete(Users, name)
	} else {
		returnError(ctx, 404, model.Error{
			Code:    404,
			Details: fmt.Errorf("Could not find user with name %s", name).Error(),
			Title:   "Client Error",
			Href:    "",
		})
		return
	}
}
