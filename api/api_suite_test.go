package api_test

import (
	"bytes"
	"context"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/golang/protobuf/jsonpb"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	w_api "github.com/ron96G/whatsapp-bizapi-mock/api"
	"github.com/ron96G/whatsapp-bizapi-mock/model"
	"github.com/ron96G/whatsapp-bizapi-mock/webhook"
	"github.com/valyala/fasthttp/fasthttputil"
)

var (
	staticAPIToken = "abcdefg"
	apiPrefix      = "/v1"
	baseUrl        = "http://localhost:8080" + apiPrefix
	contacts       = []*model.Contact{}
	generators     = model.NewGenerators(w_api.Config.UploadDir, contacts, w_api.Config.InboundMedia)
	w              = webhook.NewWebhook(w_api.Config.ApplicationSettings.Webhooks.Url, w_api.Config.Version, generators)
	api            = w_api.NewAPI(apiPrefix, staticAPIToken, w_api.Config, w)

	marsheler = jsonpb.Marshaler{
		EmitDefaults: false,
		EnumsAsInts:  false,
		OrigName:     true,
		Indent:       "  ",
	}

	unmarsheler = jsonpb.Unmarshaler{
		AllowUnknownFields: true,
	}
)

func TestAPI(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "API Suite")
}

func PanicIfNotNil(err error) {
	if err != nil {
		panic(err)
	}
}

func StartNewServer() (client *http.Client) {
	ln := fasthttputil.NewInmemoryListener()

	go func() {
		if err := api.Server.Serve(ln); err != nil {
			panic(err)
		}
	}()
	client = &http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return ln.Dial()
			},
		},
		Timeout: time.Second,
	}

	return
}

var _ = Describe("Users API", func() {
	defer GinkgoRecover()

	client := StartNewServer()
	buf := bytes.NewBuffer(nil)
	password := "newPassword123!"

	w_api.Config = &model.InternalConfig{
		Users: map[string]string{
			"admin": "secret",
		},
	}

	AfterSuite(func() {
		err := api.Server.Shutdown()
		PanicIfNotNil(err)
	})

	Describe("First Login", func() {

		requestBody := model.ChangePwdRequest{
			NewPassword: password,
		}

		buf.Reset()
		marsheler.Marshal(buf, &requestBody)

		Context("Missing Authorization", func() {
			req, _ := http.NewRequest("POST", baseUrl+"/users/login", buf)
			resp, err := client.Do(req)
			PanicIfNotNil(err)

			It("Should have status code 401", func() {
				Expect(resp.StatusCode).To(Equal(401))
			})

			It("Should have an error response body", func() {
				errResp := new(model.ErrorResponse)
				PanicIfNotNil(unmarsheler.Unmarshal(resp.Body, errResp))

				Expect(errResp.Errors).To(HaveLen(1))
				Expect(errResp.Errors[0].Code).To(Equal(int32(401)))
				Expect(errResp.Errors[0].Title).To(Equal("Client Error"))
				Expect(errResp.Errors[0].Details).To(Equal("unable to find Authorization header"))
			})
		})

		buf.Reset()
		marsheler.Marshal(buf, &requestBody)

		Context("Successful first login", func() {
			req, _ := http.NewRequest("POST", baseUrl+"/users/login", buf)

			req.SetBasicAuth("admin", "secret")
			resp, err := client.Do(req)
			PanicIfNotNil(err)

			It("Should have status code 200", func() {
				Expect(resp.StatusCode).To(Equal(200))
			})

			It("Should have a bearer token in response body", func() {
				loginResp := new(model.LoginResponse)
				PanicIfNotNil(unmarsheler.Unmarshal(resp.Body, loginResp))

				Expect(loginResp.Users).To(HaveLen(1))
				Expect(loginResp.Users[0].Token).ToNot(BeEmpty())
				Expect(loginResp.Users[0].ExpiresAfter).ToNot(BeEmpty())
			})
		})
	})

	Context("Login & Logout", func() {

		buf.Reset()

		Context("Login", func() {
			req, _ := http.NewRequest("POST", baseUrl+"/users/login", buf)

			req.SetBasicAuth("admin", password)
			resp, err := client.Do(req)

			It("Should not return an error", func() {
				Expect(err).NotTo(HaveOccurred())
			})

			It("Should have status code 200", func() {
				Expect(resp.StatusCode).To(Equal(200))
			})

			It("Should have a bearer token in response body", func() {
				loginResp := new(model.LoginResponse)
				PanicIfNotNil(unmarsheler.Unmarshal(resp.Body, loginResp))

				Expect(loginResp.Users).To(HaveLen(1))
				Expect(loginResp.Users[0].Token).ToNot(BeEmpty())
				Expect(loginResp.Users[0].ExpiresAfter).ToNot(BeEmpty())
			})
		})

		buf.Reset()

		Context("Logout", func() {
			authToken, err := api.GenerateToken("admin", "ADMIN")
			PanicIfNotNil(err)

			req, _ := http.NewRequest("POST", baseUrl+"/users/logout", buf)
			req.Header.Set("Authorization", "Bearer "+authToken)

			resp, err := client.Do(req)
			PanicIfNotNil(err)

			It("Should have status code 200", func() {
				Expect(resp.StatusCode).To(Equal(200))
			})

			Context("Check that logout worked", func() {

				requestBody := model.User{
					Username: "username",
					Password: "password",
				}
				marsheler.Marshal(buf, &requestBody)

				req, _ = http.NewRequest("POST", baseUrl+"/users", buf)
				req.Header.Set("Authorization", "Bearer "+authToken)
				resp, err := client.Do(req)
				PanicIfNotNil(err)
				It("Should have status code 401", func() {
					Expect(resp.StatusCode).To(Equal(401))
				})

			})
		})
	})

})
