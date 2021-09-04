package api

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/ron96G/whatsapp-bizapi-mock/model"
	"github.com/valyala/fasthttp"
)

// Login godoc
// @Summary Login into the application
// @Description Login into the application using basic auth
// @Tags users
// @Produce json
// @Success 200 {object} model.LoginResponse
// @Failure default {object} model.ErrorResponse
// @Router /users/login [post]
// @Security BasicAuth
func (a *API) Login(ctx *fasthttp.RequestCtx) {
	username, password, err := basicAuth(ctx)
	if err != nil {
		returnError(ctx, 401, model.Error{
			Code:    401,
			Details: err.Error(),
			Title:   "Client Error",
		})
		return
	}
	if pwd, ok := a.Config.Users[username]; ok {
		if pwd == password { // check if entered password is correct
			if pwd == "secret" { // check if the password has been changed, if not, it must be done now
				chPwdReq := new(model.ChangePwdRequest)
				err := unmarsheler.Unmarshal(bytes.NewReader(ctx.PostBody()), chPwdReq)
				if err != nil || chPwdReq.NewPassword == "" { // check if the request contained a new password
					returnError(ctx, 400, model.Error{
						Code:    400,
						Details: "Password change required",
						Title:   "Client Error",
						Href:    "",
					})
					return
				}
				a.Config.Users[username] = chPwdReq.NewPassword // change the password
			}

			role := "USER"
			if username == "admin" {
				role = "ADMIN"
			}
			// generate new token for the user
			newToken, err := a.GenerateToken(username, role)
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

// Logout godoc
// @Summary Log the user out
// @Description Logout by supplying a bearer token of the user that should be logged out
// @Tags users
// @Produce json
// @Success 200
// @Failure default {object} model.ErrorResponse
// @Router /users/logout [post]
// @Security BearerAuth
func (a *API) Logout(ctx *fasthttp.RequestCtx) {
	auth := string(ctx.Request.Header.Peek("Authorization"))
	token := strings.TrimPrefix(auth, "Bearer ")
	a.Tokens.Del(token)
}

// CreateUser godoc
// @Summary create a new user in the application
// @Description An admin may use this endpoint to create a new user
// @Tags users
// @Consume json
// @Produce json
// @Param body body model.User true "username and password of the user"
// @Success 200 {object} model.LoginResponse
// @Failure default {object} model.ErrorResponse
// @Router /users [post]
// @Security BearerAuth
func (a *API) CreateUser(ctx *fasthttp.RequestCtx) {
	user := &model.User{}
	if !unmarshalPayload(ctx, user) {
		return
	}
	response := AcquireLoginResponse()
	response.Reset()
	defer ReleaseLoginResponse(response)

	response.Meta = AcquireMeta()
	defer ReleaseMeta(response.Meta)

	if _, exists := a.Config.Users[user.Username]; exists {
		returnError(ctx, 400, model.Error{
			Code:    400,
			Title:   "User  already  exists",
			Details: fmt.Sprintf("The requested user %s already exists", user.Username),
		})
		return
	}
	a.Config.Users[user.Username] = user.Password
	returnJSON(ctx, 201, response)
}

// DeleteUser godoc
// @Summary delete an existing user in the application
// @Description An admin may use this endpoint to delete an existing user
// @Tags users
// @Param username path string true "Name of the user that should be deleted"
// @Produce json
// @Success 200
// @Failure default {object} model.ErrorResponse
// @Router /users/{username} [delete]
// @Security BearerAuth
func (a *API) DeleteUser(ctx *fasthttp.RequestCtx) {
	name := ctx.UserValue("name").(string)

	if name == "admin" {
		returnError(ctx, 400, model.Error{
			Code:    400,
			Details: fmt.Errorf("The user %s cannot be deleted", name).Error(),
			Title:   "Client Error",
			Href:    "",
		})
		return
	}

	if _, ok := a.Config.Users[name]; ok {
		delete(a.Config.Users, name)
	} else {
		returnError(ctx, 404, model.Error{
			Code:    404,
			Details: fmt.Errorf("could not find user with name %s", name).Error(),
			Title:   "Client Error",
			Href:    "",
		})
		return
	}
}
