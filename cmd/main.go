package main

import (
	"crypto/tls"
	"flag"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/valyala/fasthttp/reuseport"

	"github.com/rgumi/whatsapp-mock/controller"
	"github.com/rgumi/whatsapp-mock/docs"
	"github.com/rgumi/whatsapp-mock/model"
	"github.com/rgumi/whatsapp-mock/util"
)

var (
	apiPrefix              = flag.String("apiprefix", "/v1", "the prefix for the API")
	configFile             = flag.String("configfile", "./data/config.json", "the application config")
	addr                   = flag.String("addr", "0.0.0.0:9090", "port the webserver listens on")
	signingKey             = []byte(*flag.String("skey", "abcde", "key which is used to sign jwt"))
	webhookURL             = flag.String("webhook", "", "URL of the webhook")
	enableTLS              = flag.Bool("tls", true, "run the API with TLS (HTTPS) enabled")
	insecureSkipVerify     = flag.Bool("insecure-skip-verify", false, "skip the validation of the certificate of webhook")
	soReuseport            = flag.Bool("reuseport", false, "(experimental) uses SO_REUSEPORT option to start TCP listener") // see https://www.nginx.com/blog/socket-sharding-nginx-release-1-9-1/
	compressWebhookContent = flag.Bool("compress-webhook", false, "compress the content of the webhook requests using gzip")
	compressMinsize        = flag.Int("compress-minsize", 2048, "the minimum uncompressed sized that is required to use gzip compression")
	allowUnknownFields     = flag.Bool("allow-unknown-fields", true, "Whether to allow unknown fields in the incoming message request")

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

	if err = controller.InitConfig(f); err != nil {
		util.Log.Fatalf("Error in provided config (%v)", err)
	}
}

func main() {
	start := time.Now()
	flag.Parse()
	util.SetupLog(4)

	setupConfig(*configFile)
	util.NewClient(controller.Config.WebhookCA)

	if *webhookURL != "" {
		controller.Config.ApplicationSettings.Webhooks.Url = *webhookURL
	}

	util.DefaultClient.TLSConfig.InsecureSkipVerify = *insecureSkipVerify
	controller.SigningKey = signingKey
	controller.Compress = *compressWebhookContent
	controller.CompressMinsize = *compressMinsize
	controller.UpdateUnmarshaler(*allowUnknownFields)

	contacts := make([]*model.Contact, len(controller.Config.Contacts))
	for i, c := range controller.Config.Contacts {
		contacts[i] = &model.Contact{
			WaId: c.Id,
			Profile: &model.Contact_Profile{
				Name: c.Name,
			},
		}
	}

	util.Log.Infof("Current config: \n %v", controller.Config.String())

	// setup  swagger

	if *enableTLS {
		docs.SwaggerInfo.Schemes = []string{"https"}
	} else {
		docs.SwaggerInfo.Schemes = []string{"http"}
	}
	docs.SwaggerInfo.BasePath = *apiPrefix
	docs.SwaggerInfo.Title = "WhatsAppMockServer"

	util.Log.Infof("Creating new webserver with prefix %v\n", *apiPrefix)
	server := controller.NewServer(*apiPrefix, staticAPIToken)
	controller.ApiVersion = controller.Config.Version
	generators := model.NewGenerators(controller.Config.UploadDir, contacts, controller.Config.InboundMedia)
	webhook := controller.NewWebhookConfig(controller.Config.ApplicationSettings.Webhooks.Url, generators)

	controller.Webhook = webhook
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
		if err := server.Serve(ln); err != nil {
			util.Log.Fatalf("Server listen failed with %v\n", err)
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
	if err = controller.SaveToJSONFile(controller.Config, *configFile); err != nil {
		util.Log.Panicf("Unable to save current config (%v)", err)
	}
	util.Log.Info("Successfully shutdown application")
}
