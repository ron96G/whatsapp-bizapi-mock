package main

import (
	"flag"
	"log"
	"os"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/valyala/fasthttp/reuseport"

	"github.com/rgumi/whatsapp-mock/controller"
	c "github.com/rgumi/whatsapp-mock/controller"
	"github.com/rgumi/whatsapp-mock/model"
)

var (
	apiPrefix  = "/v1"
	configFile = flag.String("configfile", "/home/app/config.json", "the application config")
	addr       = flag.String("addr", "0.0.0.0:8080", "port the webserver listens on")
	signingKey = []byte(*flag.String("skey", "abcde", "key which is used to sign jwt"))
	webhookURL = flag.String("webhook", "", "URL of the webhook")
)

func readConfig(path string) *model.InternalConfig {
	f, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	config := new(model.InternalConfig)
	err = jsonpb.Unmarshal(f, config)
	if err != nil {
		log.Fatal(err)
	}
	return config
}

func main() {
	flag.Parse()

	config := readConfig(*configFile)

	if *webhookURL != "" {
		config.WebhookUrl = *webhookURL
	}

	c.Users = config.Users
	c.UploadDir = config.UploadDir
	c.SigningKey = signingKey

	contacts := make([]*model.Contact, len(config.Contacts))
	for i, c := range config.Contacts {
		contacts[i] = &model.Contact{
			WaId: c.Id,
			Profile: &model.Contact_Profile{
				Name: c.Name,
			},
		}
	}

	server := c.NewServer(apiPrefix)
	generators := model.NewGenerators(config.UploadDir, contacts, config.InboundMedia)
	webhook := controller.NewWebhookConfig(config.WebhookUrl, generators)

	controller.Webhook = webhook
	errors := make(chan error, 5)
	stopWebhook := webhook.Run(errors)

	go func() {
		for {
			err := <-errors
			log.Printf(err.Error())
		}
	}()

	log.Printf("Starting webserver with addr %v\n", *addr)
	ln, err := reuseport.Listen("tcp4", *addr)
	if err != nil {
		log.Fatalf("Reuseport listener failed with %v\n", err)
	}

	if err := server.Serve(ln); err != nil {
		log.Fatalf("Server listen failed with %v\n", err)
	}

	log.Printf("Shutting down application\n")
	ln.Close()
	stopWebhook <- 1
}
