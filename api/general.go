package api

import (
	"regexp"
	"time"

	"github.com/google/uuid"

	"github.com/ron96G/whatsapp-bizapi-mock/model"
	"github.com/valyala/fasthttp"
)

var (
	regexPhoneNumber = regexp.MustCompile(`^\+?(?:[0-9-\(\)] ?){6,14}[0-9]$`)
	cleanUp          = regexp.MustCompile(`[^\+0-9]`)
)

func (a *API) SendMessages(ctx *fasthttp.RequestCtx) {
	msg := model.AcquireMessage()
	msg.Reset()
	defer model.ReleaseMessage(msg)
	if !unmarshalPayload(ctx, msg) {
		return
	}
	logger := a.LoggerFromCtx(ctx)

	// return
	id := uuid.New().String()
	logger.Info("Generated message ", "msg_id", id)
	msg.Id = id
	resp := AcquireIdResponse()
	resp.Reset()
	defer ReleaseIdResponse(resp)
	resp.Messages = append(resp.Messages, &model.Id{Id: id})
	returnJSON(ctx, 200, resp)

	stati := a.Webhook.Generators.GenerateSatiForMessage(msg)
	a.Webhook.AddStati(stati...)
}

func Contacts(ctx *fasthttp.RequestCtx) {
	msg := new(model.ContactRequest)
	msg.Reset()
	if !unmarshalPayload(ctx, msg) {
		return
	}

	resp := AcquireContactResponse()
	resp.Reset()
	defer ReleaseContactResponse(resp)
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

func (a *API) GenerateWebhookRequests(ctx *fasthttp.RequestCtx) {

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
				case <-a.cancel:
					return

				case <-time.After(dur * time.Second):
					go a.Webhook.GenerateWebhookRequests(n, allowedTypes...)
				}
			}
		}()

	} else {
		messages := a.Webhook.GenerateWebhookRequests(n, allowedTypes...)

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

func (a *API) CancelGenerateWebhookRquests(ctx *fasthttp.RequestCtx) {
	a.cancel <- 1
}

func (a *API) PanicHandler(ctx *fasthttp.RequestCtx, in interface{}) {
	a.Log.Crit("Panic handler catched error", "error", in)
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
