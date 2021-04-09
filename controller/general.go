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
	id := uuid.New().String()
	log.Println("Generated message id " + id)
	msg.Id = id
	resp := AcquireResponse()
	resp.Reset()
	defer ReleaseResponse(resp)
	resp.Messages = append(resp.Messages, &model.Id{Id: id})
	returnJSON(ctx, 200, resp)

	stati := Webhook.Generators.GenerateSatiForMessage(msg)
	Webhook.AddStati(stati...)
}

func Contacts(ctx *fasthttp.RequestCtx) {
	notImplemented(ctx)
}

func GenerateWebhookRequests(ctx *fasthttp.RequestCtx) {

	// read the 'types' query args which is used to define the allowed
	// messages type which will be generated
	// if the query arg is not set, 'rnd' is used instead to generate rnd
	// messages
	allowedTypes, ok := getQueryArgList(ctx, "types")
	if !ok {
		allowedTypes = append(allowedTypes, "rnd")
	}

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
					Webhook.GenerateWebhookRequests(n, allowedTypes...)
				}
			}
		}()

	} else {
		messages := Webhook.GenerateWebhookRequests(n, allowedTypes...)

		resp := AcquireResponse()
		resp.Reset()
		defer ReleaseResponse(resp)

		resp.Messages = make([]*model.Id, len(messages))
		for i, msg := range messages {
			resp.Messages[i] = &model.Id{Id: msg.Id}
		}
		returnJSON(ctx, 200, resp)
	}
}

func CancelGenerateWebhookRquests(ctx *fasthttp.RequestCtx) {
	cancel <- 1
}

func PanicHandler(ctx *fasthttp.RequestCtx, in interface{}) {
	log.Printf("PANIC %v\n", in)
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
