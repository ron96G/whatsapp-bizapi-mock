package api_test

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/gogo/protobuf/jsonpb"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	w_api "github.com/ron96G/whatsapp-bizapi-mock/api"
	"github.com/ron96G/whatsapp-bizapi-mock/model"
	"github.com/ron96G/whatsapp-bizapi-mock/webhook"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttputil"

	log "github.com/ron96G/go-common-utils/log"
)

var (
	staticAPIToken = "abcdefg"
	apiPrefix      = "/v1"
	baseUrl        = "http://localhost:8080" + apiPrefix
	contacts       = []*model.Contact{}
	generators, _  = model.NewGenerators(w_api.Config.UploadDir, contacts, w_api.Config.InboundMedia)
	w              = webhook.NewWebhook(w_api.Config.ApplicationSettings.Webhooks.Url, w_api.Config.Version, generators)
	api            = w_api.NewAPI(apiPrefix, staticAPIToken, uint(20), w_api.Config, w)
	client         = StartNewServer(api.Server)

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

func init() {
	log.Configure("debug", "json", os.Stdout)
}

func TestAPI(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "API Suite")
}

func PanicIfNotNil(err error) {
	if err != nil {
		panic(err)
	}
}

func StartNewServer(s *fasthttp.Server) (client *http.Client) {
	ln := fasthttputil.NewInmemoryListener()

	go func() {
		if err := s.Serve(ln); err != nil {
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

		Context("Missing Authorization Header", func() {
			marsheler.Marshal(buf, &requestBody)
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
				Expect(errResp.Errors[0].Details).To(Equal("Missing Authorization"))
			})
		})

		buf.Reset()

		Context("Incorrect Authorization", func() {
			marsheler.Marshal(buf, &requestBody)
			req, _ := http.NewRequest("POST", baseUrl+"/users/login", buf)
			req.SetBasicAuth("admin", "wrongPassword")

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
				Expect(errResp.Errors[0].Details).To(Equal("Username or password is invalid"))
			})
		})

		buf.Reset()

		Context("Missing Password Change", func() {
			req, _ := http.NewRequest("POST", baseUrl+"/users/login", buf)
			req.SetBasicAuth("admin", "secret")

			resp, err := client.Do(req)
			PanicIfNotNil(err)

			It("Should have status code 400", func() {
				Expect(resp.StatusCode).To(Equal(400))
			})

			It("Should have an error response body", func() {
				errResp := new(model.ErrorResponse)
				PanicIfNotNil(unmarsheler.Unmarshal(resp.Body, errResp))

				Expect(errResp.Errors).To(HaveLen(1))
				Expect(errResp.Errors[0].Code).To(Equal(int32(400)))
				Expect(errResp.Errors[0].Title).To(Equal("Client Error"))
				Expect(errResp.Errors[0].Details).To(Equal("Password change required"))
			})
		})

		buf.Reset()

		Context("Successful First Login", func() {
			marsheler.Marshal(buf, &requestBody)
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

	buf.Reset()

	Context("Login & Logout", func() {

		Context("Login", func() {
			req, _ := http.NewRequest("POST", baseUrl+"/users/login", buf)

			req.SetBasicAuth("admin", password)
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

			Context("Verify that logout worked", func() {

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

	Context("Creating and Deleting Users", func() {

		adminAuthToken, err := api.GenerateToken("admin", "ADMIN")
		PanicIfNotNil(err)

		username := "username"
		requestBody := &model.User{
			Username: username,
			Password: "password",
		}

		buf.Reset()

		Context("Creating a new User", func() {

			marsheler.Marshal(buf, requestBody)

			req, _ := http.NewRequest("POST", baseUrl+"/users", buf)
			req.Header.Set("Authorization", "Bearer "+adminAuthToken)

			resp, err := client.Do(req)
			PanicIfNotNil(err)

			It("Should have status code 201", func() {
				Expect(resp.StatusCode).To(Equal(201))
			})
		})

		buf.Reset()

		Context("Creating an already existing User", func() {

			marsheler.Marshal(buf, requestBody)

			req, _ := http.NewRequest("POST", baseUrl+"/users", buf)
			req.Header.Set("Authorization", "Bearer "+adminAuthToken)

			resp, err := client.Do(req)
			PanicIfNotNil(err)

			It("Should have status code 400", func() {
				Expect(resp.StatusCode).To(Equal(400))
			})

			It("Should have an error response body", func() {
				errResp := new(model.ErrorResponse)
				PanicIfNotNil(unmarsheler.Unmarshal(resp.Body, errResp))

				Expect(errResp.Errors).To(HaveLen(1))
				Expect(errResp.Errors[0].Code).To(Equal(int32(400)))
				Expect(errResp.Errors[0].Title).To(Equal("User already exists"))
				Expect(errResp.Errors[0].Details).To(Equal(fmt.Sprintf("The requested user %s already exists", username)))
			})
		})

		buf.Reset()

		Context("Deleting a User", func() {
			req, _ := http.NewRequest("DELETE", baseUrl+"/users/"+username, buf)
			req.Header.Set("Authorization", "Bearer "+adminAuthToken)

			resp, err := client.Do(req)
			PanicIfNotNil(err)

			It("Should have status code 200", func() {
				Expect(resp.StatusCode).To(Equal(200))
			})
		})

		buf.Reset()

		Context("Deleting nonexistent User", func() {
			req, _ := http.NewRequest("DELETE", baseUrl+"/users/"+username, buf)
			req.Header.Set("Authorization", "Bearer "+adminAuthToken)

			resp, err := client.Do(req)
			PanicIfNotNil(err)

			It("Should have status code 404", func() {
				Expect(resp.StatusCode).To(Equal(404))
			})

			It("Should have an error response body", func() {
				errResp := new(model.ErrorResponse)
				PanicIfNotNil(unmarsheler.Unmarshal(resp.Body, errResp))

				Expect(errResp.Errors).To(HaveLen(1))
				Expect(errResp.Errors[0].Code).To(Equal(int32(404)))
				Expect(errResp.Errors[0].Title).To(Equal("Client Error"))
				Expect(errResp.Errors[0].Details).To(Equal(fmt.Sprintf("Could not find user with name %s", username)))
			})
		})
	})

}) // Users API
