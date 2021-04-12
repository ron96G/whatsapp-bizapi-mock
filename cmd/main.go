package main

import (
	"crypto/tls"
	"flag"
	"net"
	"os"
	"time"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/valyala/fasthttp/reuseport"

	"github.com/rgumi/whatsapp-mock/controller"
	"github.com/rgumi/whatsapp-mock/model"
	"github.com/rgumi/whatsapp-mock/util"
)

var (
	apiPrefix         = flag.String("apiprefix", "/v1", "the prefix for the API")
	configFile        = flag.String("configfile", "/home/app/config.json", "the application config")
	addr              = flag.String("addr", "0.0.0.0:9090", "port the webserver listens on")
	signingKey        = []byte(*flag.String("skey", "abcde", "key which is used to sign jwt"))
	webhookURL        = flag.String("webhook", "", "URL of the webhook")
	enableTLS         = flag.Bool("tls", true, "run the API with TLS (HTTPS) enabled")
	insecureTLSClient = flag.Bool("insecure", false, "validate certificate of webhook connection")
	soReuseport       = flag.Bool("reuseport", false, "(experimental) uses SO_REUSEPORT option to start TCP listener") // see https://www.nginx.com/blog/socket-sharding-nginx-release-1-9-1/
	staticAPIToken    = os.Getenv("WA_API_KEY")
)

func readConfig(path string) *model.InternalConfig {
	f, err := os.Open(path)
	if err != nil {
		util.Log.Fatal(err)
	}
	defer f.Close()

	config := new(model.InternalConfig)
	err = jsonpb.Unmarshal(f, config)
	if err != nil {
		util.Log.Fatal(err)
	}
	return config
}

func main() {
	start := time.Now()
	flag.Parse()
	util.SetupLog(4)

	config := readConfig(*configFile)

	if *webhookURL != "" {
		config.WebhookUrl = *webhookURL
	}

	if !*insecureTLSClient {
		util.DefaultClient.TLSConfig.InsecureSkipVerify = false
	}

	controller.Users = config.Users
	controller.UploadDir = config.UploadDir
	controller.SigningKey = signingKey

	contacts := make([]*model.Contact, len(config.Contacts))
	for i, c := range config.Contacts {
		contacts[i] = &model.Contact{
			WaId: c.Id,
			Profile: &model.Contact_Profile{
				Name: c.Name,
			},
		}
	}

	util.Log.Infof("Creating new webserver with prefix %v", *apiPrefix)
	server := controller.NewServer(*apiPrefix, staticAPIToken)
	generators := model.NewGenerators(config.UploadDir, contacts, config.InboundMedia)
	webhook := controller.NewWebhookConfig(config.WebhookUrl, generators)

	controller.Webhook = webhook
	errors := make(chan error, 5)
	stopWebhook := webhook.Run(errors)

	go func() {
		for {
			err := <-errors
			util.Log.Errorf(err.Error())
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
		tlsCfg, err := controller.GenerateServerTLS()
		if err != nil {
			util.Log.Fatalf("Unable to generate Server TLS config due to %v", err)
		}
		ln = tls.NewListener(ln, tlsCfg)
	}

	util.Log.Infof("Setup completed after %v", time.Since(start))
	util.Log.Infof("Starting webserver with addr %v", *addr)
	if err := server.Serve(ln); err != nil {
		util.Log.Fatalf("Server listen failed with %v\n", err)
	}

	util.Log.Info("Shutting down application")
	ln.Close()
	stopWebhook <- 1
}
