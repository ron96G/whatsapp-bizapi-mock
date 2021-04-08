package controller

import (
	"log"
	"time"

	"github.com/google/uuid"

	"github.com/rgumi/whatsapp-mock/model"
	"github.com/valyala/fasthttp"
)

func SendMessages(ctx *fasthttp.RequestCtx) {
	msg := model.AcquireMessage()
	msg.Reset()
	defer model.ReleaseMessage(msg)
	if !unmarshalPayload(ctx, msg) {
		return
	}

	// validate
	if v := msg.Validate(); !v.IsValid() {
		returnError(ctx, 400, v...)
		return
	}

	// return
	id := &model.Id{
		Id: uuid.New().String(),
	}
	returnJSON(ctx, 200, id)

	stati := Webhook.Generators.GenerateSatiForMessage(msg)
	Webhook.AddStati(stati...)
}

func Contacts(ctx *fasthttp.RequestCtx) {
	notImplemented(ctx)
}

func GenerateWebhookRequests(ctx *fasthttp.RequestCtx) {
	// number of messages that are generated
	n, ok := getQueryArgInt(ctx, "volume")
	if !ok {
		return
	}
	// interval between the generation of messages
	// if 0, messages are just generated once
	r, ok := getQueryArgInt(ctx, "interval")
	if !ok {
		return
	}

	if r > 0 {
		go func() {
			for {
				select {
				case <-cancel:
					return

				case <-time.After(time.Duration(r) * time.Second):
					Webhook.GenerateWebhookRequests(n)
				}
			}
		}()

	} else {
		Webhook.GenerateWebhookRequests(n)
	}
}

func CancelGenerateWebhookRquests(ctx *fasthttp.RequestCtx) {
	cancel <- 1
}

func PanicHandler(ctx *fasthttp.RequestCtx, in interface{}) {
	log.Printf("%v\n", in)
	returnError(ctx, 500, model.Error{
		Code:    500,
		Details: "An unexpected error occured",
		Title:   "Unexpected Error",
		Href:    "",
	})
}

func HealthCheck(ctx *fasthttp.RequestCtx) {
	ctx.Write([]byte("OK"))
	ctx.SetStatusCode(200)
}
