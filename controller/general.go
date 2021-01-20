package controller

import (
	"fmt"
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
	n, ok := getQueryArgInt(ctx, "n")
	if !ok {
		return
	}
	r, ok := getQueryArgInt(ctx, "r")
	if !ok {
		return
	}

	if n > MaxWebhookPayload || r < MinWebhookInterval {
		err := fmt.Errorf("Interval too low or payload too large")
		returnError(ctx, 400, model.Error{
			Code:    400,
			Details: err.Error(),
			Title:   "Client Error",
		})
		return
	}

	if r > 0 {
		go func() {
			for {
				select {
				case _ = <-cancel:
					return

				case _ = <-time.After(time.Duration(r) * time.Second):
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
	return
}
