package model

import (
	"math/rand"
	"os"
	sync "sync"
	"time"

	"github.com/google/uuid"
)

var (
	messagePool = sync.Pool{
		New: func() interface{} {
			return new(Message)
		},
	}

	statusPool = sync.Pool{
		New: func() interface{} {
			return new(Status)
		},
	}
)

type Generators struct {
	UploadDir string
	Contacts  []*Contact
	Media     map[string]string
	Types     []MessageType
}

func init() {
	rand.Seed(time.Now().Unix())
}

func NewGenerators(uploadDir string, c []*Contact, m map[string]string) *Generators {
	return &Generators{
		UploadDir: uploadDir,
		Contacts:  c,
		Media:     m,
		Types: []MessageType{
			MessageType_audio, MessageType_image, MessageType_text, MessageType_document, MessageType_video,
		},
	}
}

func AcquireStatus() *Status {
	return statusPool.Get().(*Status)
}

func ReleaseStatus(s *Status) {
	statusPool.Put(s)
}

func AcquireMessage() *Message {
	return messagePool.Get().(*Message)
}

func ReleaseMessage(msg *Message) {
	messagePool.Put(msg)
}

func (g *Generators) selectRndContact() *Contact {
	return g.Contacts[rand.Intn(len(g.Contacts))]
}

func (g *Generators) generateMedia(t string) string {
	id := uuid.New().String()
	if err := os.Symlink(g.UploadDir+g.Media[t], g.UploadDir+id); err != nil {
		panic(err)
	}
	return id
}

func (g *Generators) generateBaseMessage() *Message {
	contact := g.selectRndContact()
	msg := AcquireMessage()
	msg = &Message{
		From: contact.GetWaId(),
		Id:   uuid.New().String(),
	}
	return msg
}

func (g *Generators) GenerateRndMessages(n int) []*Message {
	out := make([]*Message, n)

	for i := 0; i < n; i++ {
		typ := g.Types[rand.Intn(len(g.Types))]

		switch typ {
		case MessageType_text:
			out[i] = g.GenerateTextMessage()

		case MessageType_image:
			out[i] = g.GenerateImageMessage()

		case MessageType_audio:
			out[i] = g.GenerateAudioMessage()

		case MessageType_video:
			out[i] = g.GenerateVideoMessage()

		case MessageType_document:
			out[i] = g.GenerateDocumentMessage()
		}
	}
	return out
}

func (g *Generators) AppendContextToMessage(msg *Message) {
	contact := g.selectRndContact()
	msg.Context = &Context{
		Id:        uuid.New().String(),
		From:      contact.GetWaId(),
		Forwarded: true,
	}
}

func (g *Generators) GenerateTextMessage() *Message {
	msg := g.generateBaseMessage()
	msg.Type = MessageType_text
	msg.Text = &TextMessage{
		Body: "Textbody",
	}
	return msg
}

func (g *Generators) GenerateImageMessage() *Message {
	msg := g.generateBaseMessage()
	msg.Type = MessageType_image
	msg.Image = &ImageMessage{
		Caption:  "Hello World!",
		Id:       g.generateMedia(MessageType_image.String()),
		MimeType: "image/png",
	}
	return msg
}

func (g *Generators) GenerateVideoMessage() *Message {
	msg := g.generateBaseMessage()
	msg.Type = MessageType_video
	msg.Video = &VideoMessage{
		Id:       g.generateMedia(MessageType_video.String()),
		MimeType: "video/mp4",
	}
	return msg
}

func (g *Generators) GenerateAudioMessage() *Message {
	msg := g.generateBaseMessage()
	msg.Type = MessageType_audio
	msg.Audio = &AudioMessage{
		Id:       g.generateMedia(MessageType_audio.String()),
		MimeType: "audio/mp4",
	}
	return msg
}

func (g *Generators) GenerateDocumentMessage() *Message {
	msg := g.generateBaseMessage()
	msg.Type = MessageType_document
	msg.Document = &DocumentMessage{
		Id:       g.generateMedia(MessageType_document.String()),
		MimeType: "application/pdf",
	}
	return msg
}

func (g *Generators) GenerateSatiForMessage(msg *Message) []*Status {
	stati := []*Status{}
	stati = append(
		stati,
		g.generateStatus(msg.To, msg.Id, "sent"),
		g.generateStatus(msg.To, msg.Id, "delivered"),
		g.generateStatus(msg.To, msg.Id, "read"),
	)
	return stati
}

func (g *Generators) generateStatus(recipient string, msgID string, status string) *Status {
	stat := AcquireStatus()
	stat = &Status{
		Id:          msgID,
		RecipientId: recipient,
		Timestamp:   time.Now().Unix(),
		Status:      Status_StatusEnum(Status_StatusEnum_value[status]),
	}
	return stat
}
