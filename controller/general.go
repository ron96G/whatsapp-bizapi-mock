package controller

import (
	"regexp"
	"time"

	"github.com/google/uuid"

	"github.com/rgumi/whatsapp-mock/model"
	"github.com/rgumi/whatsapp-mock/util"
	"github.com/valyala/fasthttp"
)

var (
	regexPhoneNumber = regexp.MustCompile(`^\+(?:[0-9-\(\)] ?){6,14}[0-9]$`)
	cleanUp          = regexp.MustCompile(`[- \(\)]`)
)

func SendMessages(ctx *fasthttp.RequestCtx) {
	msg := model.AcquireMessage()
	msg.Reset()
	defer model.ReleaseMessage(msg)
	if !unmarshalPayload(ctx, msg) {
		return
	}

	// return
	id := uuid.New().String()
	util.Log.Infof("Generated message id " + id)
	msg.Id = id
	resp := AcquireIdResponse()
	resp.Reset()
	defer ReleaseIdResponse(resp)
	resp.Messages = append(resp.Messages, &model.Id{Id: id})
	returnJSON(ctx, 200, resp)

	stati := Webhook.Generators.GenerateSatiForMessage(msg)
	Webhook.AddStati(stati...)
}

func Contacts(ctx *fasthttp.RequestCtx) {
	msg := new(model.ContactRequest)
	msg.Reset()
	if !unmarshalPayload(ctx, msg) {
		return
	}

	resp := new(model.ContactResponse)
	resp.Contacts = make([]*model.Contact, len(msg.Contacts))

	for i, phoneNumber := range msg.Contacts {
		resp.Contacts[i] = new(model.Contact)

		if !regexPhoneNumber.MatchString(phoneNumber) { // TODO check if the contact exists
			resp.Contacts[i].Input = phoneNumber
			resp.Contacts[i].Status = model.Contact_invalid
			continue
		}
		resp.Contacts[i].Input = phoneNumber
		resp.Contacts[i].WaId = cleanUp.ReplaceAllString(phoneNumber, "")
		resp.Contacts[i].Status = model.Contact_valid
		// TODO add the contact to a cache to check wether is can be used to send outbound messages to
	}

	returnJSON(ctx, 200, resp)
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
		dur := time.Duration(r)
		go func() {
			for {
				select {
				case <-cancel:
					return

				case <-time.After(dur * time.Second):
					go Webhook.GenerateWebhookRequests(n, allowedTypes...)
				}
			}
		}()

	} else {
		messages := Webhook.GenerateWebhookRequests(n, allowedTypes...)

		resp := AcquireIdResponse()
		resp.Reset()
		defer ReleaseIdResponse(resp)

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
	util.Log.Errorf("%v\n", in)
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
