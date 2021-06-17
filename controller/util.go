package controller

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"math/rand"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	jwt "github.com/dgrijalva/jwt-go/v4"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/rgumi/whatsapp-mock/model"
	"github.com/rgumi/whatsapp-mock/util"
	"github.com/valyala/fasthttp"
)

// readLimit is the maximum number of bytes from the input used when detecting the MimeType
var readLimit uint32 = 512

type CustomClaims struct {
	Role string `json:"role"`
	jwt.StandardClaims
}

// Extends the proto.Message interface with the Validate() to enable validation
type Message interface {
	Reset()
	String() string
	ProtoMessage()
	Validate() error
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
	response := AcquireErrorResponse()
	response.Reset()
	defer ReleaseErrorResponse(response)

	response.Meta = AcquireMeta()
	for _, err := range errors {
		response.Errors = append(response.Errors, &err)
	}
	returnJSON(ctx, statusCode, response)
}

func unmarshalPayload(ctx *fasthttp.RequestCtx, msg Message) bool {
	err := jsonpb.Unmarshal(bytes.NewReader(ctx.PostBody()), msg)
	if err != nil {
		util.Log.Warnf("Failed to unmarshal { %v } with '%v'", msg, err)
		returnError(ctx, 400, model.Error{
			Code:    400,
			Details: err.Error(),
			Title:   "Unable to unmarshal payload",
			Href:    "",
		})
		return false
	}
	return validatePayload(ctx, msg)
}

func validatePayload(ctx *fasthttp.RequestCtx, msg Message) bool {
	if err := msg.Validate(); err != nil {
		util.Log.Warnf("Failed validation of { %v } with '%v'", msg, err)
		returnError(ctx, 400, model.Error{
			Code:    400,
			Details: err.Error(),
			Title:   "Validation of input failed",
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
	response := AcquireLoginResponse()
	defer ReleaseLoginResponse(response)
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
	filePath := filepath.Join(filepath.Clean(Config.UploadDir), filepath.Clean(filename))
	f, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE, 0600)
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

func getFileContentType(r io.ReadSeeker) (string, error) {
	buffer := make([]byte, readLimit)
	_, err := r.Read(buffer)
	if err != nil {
		return "", err
	}

	// Use the net/http package's handy DectectContentType function. Always returns a valid
	// content-type by returning "application/octet-stream" if no others seemed to match.
	contentType := http.DetectContentType(buffer)

	_, err = r.Seek(0, io.SeekStart)
	return contentType, err
}

func respondWithFile(ctx *fasthttp.RequestCtx, statusCode int, filename string) (ok bool) {
	filePath := filepath.Join(filepath.Clean(Config.UploadDir), filepath.Clean(filename))
	f, err := os.OpenFile(filePath, os.O_RDONLY, 0600)
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
		util.Log.Infof("Setting Content-Type %s", contentType)
		ctx.SetContentType(contentType)
		ctx.SetStatusCode(statusCode)

		w := ctx.Response.BodyWriter()
		_, err = io.Copy(w, f)
		if err != nil {
			return false
		}
		return true
	}

	returnError(ctx, 500, model.Error{
		Code:    500,
		Details: err.Error(),
		Title:   "Server Error",
		Href:    "",
	})

	return false
}

func SaveToJSONFile(in proto.Message, path string) error {
	filePath := filepath.Clean(path)
	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	defer file.Close()
	return marsheler.Marshal(file, in)
}

func isEncodingAllowed(ctx *fasthttp.RequestCtx, encoding string) bool {
	return ctx.Request.Header.HasAcceptEncoding(encoding)
}
