package main

import (
	"context"
	"crypto/tls"
	"crypto/x509/pkix"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/valyala/fasthttp/reuseport"
	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/ron96G/whatsapp-bizapi-mock/api"
	"github.com/ron96G/whatsapp-bizapi-mock/docs"
	"github.com/ron96G/whatsapp-bizapi-mock/model"
	"github.com/ron96G/whatsapp-bizapi-mock/util"
	"github.com/ron96G/whatsapp-bizapi-mock/webhook"

	cert "github.com/ron96G/go-common-utils/certificate"
	log "github.com/ron96G/go-common-utils/log"
)

var (
	app = kingpin.New("wabiz-api-mock", "A WhatsApp Business API Mockserver")

	// required
	configfile = app.Flag("configfile", "the configuration of the application").OverrideDefaultFromEnvar("WA_CONFIGFILE").Required().String()

	// optional
	apiPrefix              = app.Flag("apiprefix", "the prefix for the API").Default("/v1").OverrideDefaultFromEnvar("WA_API_PREFIX").String()
	addr                   = app.Flag("addr", "the address the API will listen on").Default("0.0.0.0:9090").OverrideDefaultFromEnvar("WA_ADDR").String()
	webhookURL             = app.Flag("webhook", "the default webhook url").Default("https://localhost:9000/webhook").OverrideDefaultFromEnvar("WA_WEBHOOK").String()
	disableTLS             = app.Flag("disableTLS", "run the API with tls disabled").OverrideDefaultFromEnvar("WA_TLS_ENABLED").Bool()
	insecureSkipVerify     = app.Flag("insecureSkipVerify", "do not validate the certificate of the webhook").OverrideDefaultFromEnvar("WA_INSECURE_SKIP_VERIFY").Bool()
	soReuseport            = app.Flag("reuseport", "(experimental) uses SO_REUSEPORT option to start TCP listener").Bool() // see https://www.nginx.com/blog/socket-sharding-nginx-release-1-9-1/
	compressWebhookContent = app.Flag("compress", "compress the content of the webhook requests using gzip").Bool()
	compressMinsize        = app.Flag("compressMinsize", "the minimum uncompressed size that is required to use gzip compression").Default("2048").Int()
	allowUnknownFields     = app.Flag("allowUnknownFields", "Whether to allow unknown fields in the incoming message request").Bool()
	loglevel               = app.Flag("loglevel", "set the loglevel for the application").Default("info").OverrideDefaultFromEnvar("WA_LOGLEVEL").String()
	logformat              = app.Flag("logformat", "set the logformat for the application").Default("json").String()
	graceperiod            = app.Flag("graceperiod", "duration to wait for the api to shutdown").Default("5s").Duration()
	requestLimit           = app.Flag("requestlimit", "set a requestlimit (req/s) for specific endpoints").Default("20").Uint()
	maxStatiPerWebhook     = app.Flag("maxStatiPerWebhook", "set the maximum amout of stati that will be sent in a single webhook").Default("1000").Int()

	staticAPIToken = os.Getenv("WA_API_KEY")
)

func setupConfig(path string) error {
	filePath := filepath.Clean(path)
	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	return api.InitConfig(f)
}

func main() {
	start := time.Now()
	kingpin.MustParse(app.Parse(os.Args[1:]))

	log.Configure(*loglevel, *logformat, os.Stdout)
	mainLogger := log.New("main_logger")

	if *configfile != "" {
		mainLogger.Info("Trying to setup config", "configfile", *configfile)
		if err := setupConfig(*configfile); err != nil {
			mainLogger.Crit("Failed to setup config", "configfile", *configfile, "error", err)
			os.Exit(1)
		}
	}

	util.NewClient(api.Config.WebhookCA)

	if *webhookURL != "" {
		api.Config.ApplicationSettings.Webhooks.Url = *webhookURL
	}

	util.DefaultClient.TLSConfig.InsecureSkipVerify = *insecureSkipVerify
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

	if *disableTLS {
		docs.SwaggerInfo.Schemes = []string{"http"}
	} else {
		docs.SwaggerInfo.Schemes = []string{"https"}
	}
	docs.SwaggerInfo.BasePath = *apiPrefix
	docs.SwaggerInfo.Title = "WhatsAppMockServer"

	mainLogger.Info("Creating new webserver", "prefix", *apiPrefix)

	generators, err := model.NewGenerators(api.Config.UploadDir, contacts, api.Config.InboundMedia)
	if err != nil {
		mainLogger.Crit("Failed to create generators", "error", err)
		os.Exit(1)
	}
	wh := webhook.NewWebhook(api.Config.ApplicationSettings.Webhooks.Url, api.Config.Version, generators)
	wh.Compress = *compressWebhookContent
	wh.CompressMinsize = *compressMinsize
	wh.MaxStatiPerWebhookRequest = *maxStatiPerWebhook

	apiServer := api.NewAPI(*apiPrefix, staticAPIToken, *requestLimit, api.Config, wh)

	errors := make(chan error, 5)
	stopWebhook := wh.Run(errors)

	go func() {
		for {
			err := <-errors
			mainLogger.Error("Async error occured", "error", err)
		}
	}()
	var ln net.Listener

	if *soReuseport {
		ln, err = reuseport.Listen("tcp4", *addr)
		if err != nil {
			mainLogger.Crit("Failed to create reuseport tcp listener", "addr", *addr, "error", err)
			os.Exit(1)
		}
	} else {
		ln, err = net.Listen("tcp", *addr)
		if err != nil {
			mainLogger.Crit("Failed to create tcp listener", "addr", *addr, "error", err)
			os.Exit(1)
		}
	}
	defer ln.Close()

	if !*disableTLS {
		mainLogger.Debug("Creating new Server TLS config as TLS is enabled")
		tlsCfg, err := cert.GenerateServerTLS(cert.Options{
			Subject: pkix.Name{
				Organization: []string{"WhatsApp Mockserver Fake Certificate"},
				Country:      []string{"DE"},
				Province:     []string{"NRW"},
				Locality:     []string{"Bonn"},
			},
		})
		if err != nil {
			mainLogger.Crit("Unable to generate Server TLS config", "error", err)
			os.Exit(1)
		}
		ln = tls.NewListener(ln, tlsCfg)
	}

	stopChan := SetupSignalHandler()

	go func() {
		mainLogger.Info("Setup completed", "elapsed_time", time.Since(start).Milliseconds())
		mainLogger.Info("Starting webserver", "addr", *addr)
		if err := apiServer.Server.Serve(ln); err != nil && err != http.ErrServerClosed {
			mainLogger.Crit("Failed to start server with listener", "addr", *addr, "error", err)
			os.Exit(1)
		}
	}()

	<-stopChan
	mainLogger.Info("Shutting down application", "graceperiod", *graceperiod)

	shutdownCtx, cancel := context.WithTimeout(context.Background(), *graceperiod)

	go func() {
		stopWebhook <- 1
		apiServer.Server.Shutdown()
		cancel()
	}()

	<-shutdownCtx.Done()

	if err = api.SaveToJSONFile(apiServer.Config, *configfile); err != nil {
		mainLogger.Crit("Unable to save current config", "error", err)
	}
	mainLogger.Info("Successfully shutdown application")
}

func SetupSignalHandler() (stopCh <-chan struct{}) {
	stop := make(chan struct{})
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		close(stop)
		<-c
		os.Exit(143) // second signal. Exit directly.
	}()

	return stop
}
