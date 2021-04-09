package controller

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/form3tech-oss/jwt-go"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/rgumi/whatsapp-mock/model"
	"github.com/valyala/fasthttp"
)

func extractAuthToken(ctx *fasthttp.RequestCtx, key string) string {
	auth := string(ctx.Request.Header.Peek("Authorization"))
	return strings.TrimPrefix(auth, key+" ")
}

func verifyToken(ctx *fasthttp.RequestCtx) (*jwt.Token, error) {
	tokenString := extractAuthToken(ctx, "Bearer")
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return SigningKey, nil
	})
	if err != nil {
		return nil, err
	}
	return token, nil
}

func contains(slice []string, item string) bool {
	for _, element := range slice {
		if element == item {
			return true
		}
	}
	return false
}

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
		return "", "", fmt.Errorf("unable to find Authorization header")
	}
}

func returnError(ctx *fasthttp.RequestCtx, statusCode int, errors ...model.Error) {
	response := AcquireResponse()
	response.Reset()
	defer ReleaseResponse(response)

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

func returnBytes(ctx *fasthttp.RequestCtx, statusCode int, content []byte) {
	ctx.SetContentType("application/json")
	ctx.SetStatusCode(statusCode)
	ctx.Write(content)
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

func getQueryArgList(ctx *fasthttp.RequestCtx, key string) (l []string, ok bool) {

	queryArg := ctx.QueryArgs().PeekMulti(key)
	for _, i := range queryArg {
		l = append(l, string(i))
	}
	if len(l) == 0 {
		return l, false
	}
	return l, true
}

func generateRandomCode(n int) (numbers string) {
	for i := 0; i < n; i++ {
		numbers += fmt.Sprintf("%d", rand.Intn(9))
	}
	return
}
