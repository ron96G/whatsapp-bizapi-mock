package controller

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/google/uuid"

	"github.com/rgumi/whatsapp-mock/model"
	"github.com/valyala/fasthttp"
)

var (
	ApiVersion         = "3.31.5"
	UploadDir          = ""
	TokenValidDuration = 7 * 24 * time.Hour
	marsheler          = jsonpb.Marshaler{
		EmitDefaults: false,
		EnumsAsInts:  false,
		OrigName:     true,
	}
	responsePool = sync.Pool{
		New: func() interface{} {
			return new(model.APIResponse)
		},
	}

	SigningKey []byte
	Users      = map[string]string{}
	Tokens     = []string{}

	Webhook *WebhookConfig
)

func basicAuth(ctx *fasthttp.RequestCtx) (string, string, error) {
	b64 := string(ctx.Request.Header.Peek("Authorization"))
	if b64 != "" {
		auth, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(b64, "Basic "))
		if err != nil {
			return "", "", err
		}
		splittedAuth := strings.Split(string(auth), ":")
		return splittedAuth[0], splittedAuth[1], nil
	} else {
		return "", "", fmt.Errorf("Unable to find Authorization header")
	}
}

func returnError(ctx *fasthttp.RequestCtx, statusCode int, errors ...model.Error) {
	response := responsePool.Get().(*model.APIResponse)
	response.Reset()
	defer responsePool.Put(response)

	response.Meta = &model.Meta{
		ApiStatus: model.Meta_stable,
		Version:   ApiVersion,
	}
	for _, err := range errors {
		response.Errors = append(response.Errors, &err)
	}
	returnJSON(ctx, statusCode, response)
}

func unmarshalPayload(ctx *fasthttp.RequestCtx, out proto.Message) bool {
	err := jsonpb.Unmarshal(bytes.NewReader(ctx.PostBody()), out)
	if err != nil {
		returnError(ctx, 400, model.Error{
			Code:    123,
			Details: err.Error(),
			Title:   "Unable to unmarshal payload",
			Href:    "",
		})
		return false
	}
	return true
}

func returnJSON(ctx *fasthttp.RequestCtx, statusCode int, out proto.Message) {
	ctx.SetContentType("application/json")
	ctx.SetStatusCode(statusCode)
	marsheler.Marshal(ctx, out)
}

func notImplemented(ctx *fasthttp.RequestCtx) {
	returnError(ctx, 501, model.Error{
		Code:    501,
		Details: fmt.Sprintf("The resource %s is not implemented yet", ctx.Path()),
		Title:   "Not Implemented",
		Href:    "",
	})
}

func generateToken(user string) (string, error) {
	atClaims := jwt.MapClaims{}
	atClaims["authorized"] = true
	atClaims["user"] = user
	atClaims["exp"] = time.Now().Add(TokenValidDuration).Unix()
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
	token, err := at.SignedString(SigningKey)
	if err != nil {
		return "", err
	}
	return token, nil
}

func returnToken(ctx *fasthttp.RequestCtx, token string) {
	Tokens = append(Tokens, token)
	response := responsePool.Get().(*model.APIResponse)
	response.Reset()
	expires := time.Now().Add(TokenValidDuration).Format("2006-01-02 15:04:05+00:00")
	response.Users = append(response.Users,
		&model.TokenResponse{
			Token:        token,
			ExpiresAfter: expires,
		},
	)
	returnJSON(ctx, 200, response)
}

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

func SendMessages(ctx *fasthttp.RequestCtx) {
	msg := model.AcquireMessage()
	msg.Reset()
	defer model.ReleaseMessage(msg)
	if !unmarshalPayload(ctx, msg) {
		return
	}
	// validate
	id := &model.Id{
		Id: uuid.New().String(),
	}
	returnJSON(ctx, 200, id)

	stati := Webhook.Generators.GenerateSatiForMessage(msg)
	Webhook.AddStati(stati...)
}

func CreateUser(ctx *fasthttp.RequestCtx) {
	msg := &model.User{}
	if !unmarshalPayload(ctx, msg) {
		return
	}

	response := responsePool.Get().(*model.APIResponse)
	response.Reset()
	defer responsePool.Put(response)

	response.Meta = &model.Meta{
		ApiStatus: model.Meta_stable,
		Version:   ApiVersion,
	}
	returnJSON(ctx, 200, response)
	responsePool.Put(response)
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

func Contacts(ctx *fasthttp.RequestCtx) {
	notImplemented(ctx)
}

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
		returnError(ctx, 400, model.Error{
			Code:    400,
			Details: err.Error(),
			Title:   "Client Error",
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
	if err != nil {
		returnError(ctx, 400, model.Error{
			Code:    400,
			Details: err.Error(),
			Title:   "Client Error",
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
	if err != nil {
		returnError(ctx, 500, model.Error{
			Code:    500,
			Details: err.Error(),
			Title:   "Server Error",
			Href:    "",
		})
		return
	}
}

func getQueryArgInt(ctx *fasthttp.RequestCtx, key string) (n int, ok bool) {
	var err error
	queryArg := string(ctx.QueryArgs().Peek(key))
	if len(queryArg) == 0 {
		returnError(ctx, 400, model.Error{
			Code:    400,
			Details: "Unable to parse query argument",
			Title:   "Client Error",
		})
		return 0, false
	}

	if n, err = strconv.Atoi(queryArg); err != nil {
		returnError(ctx, 400, model.Error{
			Code:    400,
			Details: "Unable to parse query argument",
			Title:   "Client Error",
		})
		return 0, false
	}
	return n, true
}

var (
	cancel             = make(chan int, 1)
	maxWebhookPayload  = 100
	minWebhookInterval = 5
)

func GenerateWebhookRequests(ctx *fasthttp.RequestCtx) {
	n, ok := getQueryArgInt(ctx, "n")
	if !ok {
		return
	}
	r, ok := getQueryArgInt(ctx, "r")
	if !ok {
		return
	}

	if n > maxWebhookPayload || r < minWebhookInterval {
		err := fmt.Errorf("Interval too low or payload too large")
		returnError(ctx, 400, model.Error{
			Code:    400,
			Details: err.Error(),
			Title:   "Client Error",
		})
		return
	}

	if r > 0 {
		go func() {
			for {
				select {
				case _ = <-cancel:
					return

				case _ = <-time.After(time.Duration(r) * time.Second):
					Webhook.GenerateWebhookRequests(n)
				}
			}
		}()

	} else {
		Webhook.GenerateWebhookRequests(n)
	}
}

func CancelGenerateWebhookRquests(ctx *fasthttp.RequestCtx) {
	cancel <- 1
}

func PanicHandler(ctx *fasthttp.RequestCtx, in interface{}) {
	log.Printf("%v\n", in)
	returnError(ctx, 500, model.Error{
		Code:    500,
		Details: "An unexpected error occured",
		Title:   "Unexpected Error",
		Href:    "",
	})
	return
}
