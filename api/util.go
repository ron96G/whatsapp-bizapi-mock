package api

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
	"github.com/gogo/protobuf/jsonpb"
	"github.com/gogo/protobuf/proto"
	"github.com/ron96G/whatsapp-bizapi-mock/model"
	"github.com/valyala/fasthttp"

	log "github.com/ron96G/go-common-utils/log"
)

var (
	// readLimit is the maximum number of bytes from the input used when detecting the MimeType
	readLimit uint32 = 512

	marsheler = jsonpb.Marshaler{
		EmitDefaults: false,
		EnumsAsInts:  false,
		OrigName:     true,
		Indent:       "  ",
	}
	unmarsheler = jsonpb.Unmarshaler{
		AllowUnknownFields: true,
	}

	TokenValidDuration = 7 * 24 * time.Hour
	SigningKey         = []byte("e555f49db14afa8244ab4ccf630bd0020144b124217dd56781b00a6e024cb836")

	TimeFormatTokenExpiration = "2006-01-02 15:04:05+00:00"
)

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

func unmarshalPayload(ctx *fasthttp.RequestCtx, msg Message) error {
	err := unmarsheler.Unmarshal(bytes.NewReader(ctx.PostBody()), msg)
	if err != nil {
		returnError(ctx, 400, model.Error{
			Code:    400,
			Details: err.Error(),
			Title:   "Unable to unmarshal payload",
			Href:    "",
		})
		return fmt.Errorf("unmarshal payload: %v", err)
	}
	err = validatePayload(ctx, msg)
	if err != nil {
		returnError(ctx, 400, model.Error{
			Code:    400,
			Details: err.Error(),
			Title:   "Validation of input failed",
			Href:    "",
		})
		return err
	}
	return nil
}

func validatePayload(ctx *fasthttp.RequestCtx, msg Message) error {
	if err := msg.Validate(); err != nil {
		return fmt.Errorf("validate payload: %v", err)
	}
	return nil
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

func (a *API) GenerateToken(user string, role string) (string, error) {

	// https://self-issued.info/docs/draft-ietf-oauth-json-web-token.html#rfc.section.4.1.7
	now := time.Now()
	atClaims := jwt.MapClaims{}
	atClaims["iss"] = "WhatsAppMockserver"
	atClaims["sub"] = user
	atClaims["exp"] = now.Add(TokenValidDuration).Unix()
	atClaims["iat"] = now.Unix()
	atClaims["role"] = role
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
	token, err := at.SignedString(SigningKey)
	if err != nil {
		return "", err
	}
	a.Tokens.Add(token)
	return token, nil
}

func returnToken(ctx *fasthttp.RequestCtx, token string) {
	response := AcquireLoginResponse()
	defer ReleaseLoginResponse(response)
	response.Reset()
	expires := time.Now().Add(TokenValidDuration).Format(TimeFormatTokenExpiration)
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

func savePostBody(ctx *fasthttp.RequestCtx, fpath string) (ok bool) {
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
	filePath := filepath.Join(filepath.Clean(fpath))
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

func respondWithFile(ctx *fasthttp.RequestCtx, statusCode int, fpath string) (ok bool) {
	filePath := filepath.Join(filepath.Clean(fpath))
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
		ctx.SetContentType(contentType)
		ctx.SetStatusCode(statusCode)

		w := ctx.Response.BodyWriter()
		_, err = io.Copy(w, f)
		return err == nil
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

func UpdateUnmarshaler(allowUnknownFields bool) {
	unmarsheler = jsonpb.Unmarshaler{
		AllowUnknownFields: allowUnknownFields,
	}
}

func LoggerToCtx(ctx *fasthttp.RequestCtx, logger log.Logger) {
	ctx.SetUserValue("logger", logger)
}

func (a *API) LoggerFromCtx(ctx *fasthttp.RequestCtx) log.Logger {
	logger, ok := ctx.UserValue("logger").(log.Logger)
	if !ok {
		logger = a.Log
	}
	return logger
}
