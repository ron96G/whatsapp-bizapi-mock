package controller

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"io"
	"math/rand"
	"mime"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/rgumi/whatsapp-mock/model"
	"github.com/valyala/fasthttp"
)

var (
	gzipPool = sync.Pool{
		New: func() interface{} {
			return gzip.NewWriter(nil)
		},
	}
)

type CustomClaims struct {
	Role string `json:"role"`
	jwt.StandardClaims
}

func extractAuthToken(ctx *fasthttp.RequestCtx, key string) (val string, ok bool) {
	auth := strings.TrimPrefix(
		string(ctx.Request.Header.Peek("Authorization")),
		strings.TrimSpace(key)+" ",
	)
	return auth, auth != ""
}

func parseToken(ctx *fasthttp.RequestCtx) (*jwt.Token, error) {
	tokenString, ok := extractAuthToken(ctx, "Bearer ")
	if !ok {
		return nil, fmt.Errorf("unable to find bearer token in request")
	}
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return SigningKey, nil
	})
	return token, err
}

func parseTokenWithClaims(ctx *fasthttp.RequestCtx) (*jwt.Token, error) {
	tokenString, ok := extractAuthToken(ctx, "Bearer ")
	if !ok {
		return nil, fmt.Errorf("unable to find bearer token in request")
	}
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return SigningKey, nil
	})
	return token, err
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

func notImplemented(ctx *fasthttp.RequestCtx) {
	returnError(ctx, 501, model.Error{
		Code:    501,
		Details: fmt.Sprintf("The resource %s is not implemented yet", ctx.Path()),
		Title:   "Not Implemented",
		Href:    "",
	})
}

func NotImplementedHandler(ctx *fasthttp.RequestCtx) {
	notImplemented(ctx)
}

func generateToken(user string, role string) (string, error) {

	// https://self-issued.info/docs/draft-ietf-oauth-json-web-token.html#rfc.section.4.1.7
	atClaims := jwt.MapClaims{}
	atClaims["iss"] = "WhatsAppMockserver"
	atClaims["sub"] = user
	atClaims["exp"] = time.Now().Add(TokenValidDuration).Unix()
	atClaims["iat"] = time.Now().Unix()
	atClaims["role"] = role
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

func savePostBody(ctx *fasthttp.RequestCtx, filename string) (ok bool) {
	_, err := mime.ExtensionsByType(string(ctx.Request.Header.ContentType()))
	if err != nil {
		returnError(ctx, 400, model.Error{
			Code:    400,
			Details: err.Error(),
			Title:   "Client Error",
			Href:    "",
		})
		return false
	}
	f, err := os.OpenFile(Config.UploadDir+filename, os.O_WRONLY|os.O_CREATE, 0777)
	if err != nil {
		returnError(ctx, 500, model.Error{
			Code:    500,
			Details: err.Error(),
			Title:   "Server Error",
			Href:    "",
		})
		return false
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
		return false
	}
	return true
}

func respondWithFile(ctx *fasthttp.RequestCtx, statusCode int, filename string, compress bool) (ok bool) {
	f, err := os.OpenFile(Config.UploadDir+filename, os.O_RDONLY, 0777)
	if err != nil && os.IsNotExist(err) {
		ctx.SetStatusCode(404)
		return false

	} else if err != nil {
		returnError(ctx, 500, model.Error{
			Code:    500,
			Details: err.Error(),
			Title:   "Server Error",
			Href:    "",
		})

		return false
	}

	defer f.Close()

	contentType, err := getFileContentType(f)
	if err == nil {
		_, err := f.Seek(0, io.SeekStart)
		if err == nil {
			fmt.Println("A")
			ctx.SetContentType(contentType)
			ctx.SetStatusCode(statusCode)

			w := ctx.Response.BodyWriter()

			if compress {
				gz := gzipPool.Get().(*gzip.Writer)

				defer gzipPool.Put(gz)
				gz.Reset(w)
				io.Copy(gz, f)
				gz.Close()
				ctx.Response.Header.Add("Content-Encoding", "gzip")

			} else {
				io.Copy(w, f)
			}
			return true
		}
	}

	returnError(ctx, 500, model.Error{
		Code:    500,
		Details: err.Error(),
		Title:   "Server Error",
		Href:    "",
	})

	return false
}

func SaveToJSONFile(in proto.Message, filepath string) error {
	file, err := os.OpenFile(filepath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	defer file.Close()
	return marsheler.Marshal(file, in)
}

func isEncodingAllowed(ctx *fasthttp.RequestCtx, encoding string) bool {
	return ctx.Request.Header.HasAcceptEncoding(encoding)
}
