package main

import (
	"crypto/tls"
	"flag"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/valyala/fasthttp/reuseport"

	"github.com/ron96G/whatsapp-bizapi-mock/api"
	"github.com/ron96G/whatsapp-bizapi-mock/docs"
	"github.com/ron96G/whatsapp-bizapi-mock/model"
	"github.com/ron96G/whatsapp-bizapi-mock/util"
	"github.com/ron96G/whatsapp-bizapi-mock/webhook"
)

var (
	apiPrefix              = flag.String("apiprefix", "/v1", "the prefix for the API")
	configFile             = flag.String("configfile", "", "the application config")
	addr                   = flag.String("addr", "0.0.0.0:9090", "port the webserver listens on")
	webhookURL             = flag.String("webhook", "", "URL of the webhook")
	enableTLS              = flag.Bool("tls", true, "run the API with TLS (HTTPS) enabled")
	insecureSkipVerify     = flag.Bool("insecureSkipVerify", false, "skip the validation of the certificate of webhook")
	soReuseport            = flag.Bool("reuseport", false, "(experimental) uses SO_REUSEPORT option to start TCP listener") // see https://www.nginx.com/blog/socket-sharding-nginx-release-1-9-1/
	compressWebhookContent = flag.Bool("compressWebhook", false, "compress the content of the webhook requests using gzip")
	compressMinsize        = flag.Int("compressMinsize", 2048, "the minimum uncompressed sized that is required to use gzip compression")
	allowUnknownFields     = flag.Bool("allowUnknownFields", true, "Whether to allow unknown fields in the incoming message request")
	logLevel               = flag.Uint("loglevel", 4, "set the loglevel for the app (4=INFO, 5=DEBUG)")
	logFormatter           = flag.String("logformat", "json", "set the logformatter of the application (either 'text' or 'json')")

	staticAPIToken = os.Getenv("WA_API_KEY")

	signalLog = "Received %s. Shutting down"
)

func setupConfig(path string) {
	filePath := filepath.Clean(path)
	f, err := os.Open(filePath)
	if err != nil {
		util.Log.Fatal(err)
	}
	defer f.Close()

	if err = api.InitConfig(f); err != nil {
		util.Log.Fatalf("Error in provided config (%v)", err)
	}
}

func main() {
	start := time.Now()
	flag.Parse()
	util.SetupLog(*logLevel, strings.ToLower(*logFormatter))

	if *configFile != "" {
		setupConfig(*configFile)
	} else {
		util.Log.Infof("No configfile set. Using default config")
	}

	util.NewClient(api.Config.WebhookCA)

	if *webhookURL != "" {
		api.Config.ApplicationSettings.Webhooks.Url = *webhookURL
	}

	util.DefaultClient.TLSConfig.InsecureSkipVerify = *insecureSkipVerify
	webhook.Compress = *compressWebhookContent
	webhook.CompressMinsize = *compressMinsize
	api.UpdateUnmarshaler(*allowUnknownFields)

	contacts := make([]*model.Contact, len(api.Config.Contacts))
	for i, c := range api.Config.Contacts {
		contacts[i] = &model.Contact{
			WaId: c.Id,
			Profile: &model.Contact_Profile{
				Name: c.Name,
			},
		}
	}

	// setup  swagger

	if *enableTLS {
		docs.SwaggerInfo.Schemes = []string{"https"}
	} else {
		docs.SwaggerInfo.Schemes = []string{"http"}
	}
	docs.SwaggerInfo.BasePath = *apiPrefix
	docs.SwaggerInfo.Title = "WhatsAppMockServer"

	util.Log.Infof("Creating new webserver with prefix %v", *apiPrefix)

	generators := model.NewGenerators(api.Config.UploadDir, contacts, api.Config.InboundMedia)
	webhook := webhook.NewWebhook(api.Config.ApplicationSettings.Webhooks.Url, api.Config.Version, generators)
	apiServer := api.NewAPI(*apiPrefix, staticAPIToken, api.Config, webhook)

	errors := make(chan error, 5)
	stopWebhook := webhook.Run(errors)

	go func() {
		for {
			err := <-errors
			util.Log.Error("Async error occured: " + err.Error())
		}
	}()
	var ln net.Listener
	var err error

	if *soReuseport {
		ln, err = reuseport.Listen("tcp4", *addr)
		if err != nil {
			util.Log.Fatalf("Reuseport listener failed with %v", err)
		}
	} else {
		ln, err = net.Listen("tcp", *addr)
		if err != nil {
			util.Log.Fatalf("Listener failed with %v", err)
		}
	}

	if *enableTLS {
		util.Log.Debugf("Creating new Server TLS config as TLS is enabled")
		tlsCfg, err := util.GenerateServerTLS()
		if err != nil {
			util.Log.Fatalf("Unable to generate Server TLS config due to %v", err)
		}
		ln = tls.NewListener(ln, tlsCfg)
	}

	go func() {
		util.Log.Infof("Setup completed after %v", time.Since(start))
		util.Log.Infof("Starting webserver with addr %v", *addr)
		if err := apiServer.Server.Serve(ln); err != nil {
			util.Log.Fatalf("Server listen failed with %v", err)
		}
	}()

	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)

	sig := <-signalChannel
	switch sig {
	case os.Interrupt:
		util.Log.Warnf(signalLog, sig)

	case syscall.SIGTERM:
		util.Log.Warnf(signalLog, sig)
	}

	util.Log.Info("Shutting down application")
	ln.Close()
	stopWebhook <- 1
	if err = api.SaveToJSONFile(apiServer.Config, *configFile); err != nil {
		util.Log.Panicf("Unable to save current config (%v)", err)
	}
	util.Log.Info("Successfully shutdown application")
}
