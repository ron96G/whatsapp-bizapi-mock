package model

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	sync "sync"
	"time"

	"github.com/google/uuid"

	log "github.com/ron96G/go-common-utils/log"
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
	Log       log.Logger
	Types     []MessageType
	Sha256    map[string]string
}

func init() {
	rand.Seed(time.Now().Unix())
}

func NewGenerators(uploadDir string, contacts []*Contact, media map[string]string) (*Generators, error) {

	absUploadDir, err := filepath.Abs(uploadDir)
	if err != nil {
		return nil, err
	}
	absUploadDir += "/"

	if uploadDir == "" {
		return nil, fmt.Errorf("uploaddir cannot be empty")
	}
	if len(contacts) == 0 {
		return nil, fmt.Errorf("contacts cannot be empty")
	}
	if len(media) == 0 {
		return nil, fmt.Errorf("media cannot be empty")
	}

	g := &Generators{
		UploadDir: absUploadDir,
		Contacts:  contacts,
		Log:       log.New("generators_logger", "component", "generators"),
		Media:     media,
		Types: []MessageType{
			MessageType_audio, MessageType_image, MessageType_text, MessageType_document, MessageType_video,
		},
		Sha256: map[string]string{},
	}
	for k, f := range g.Media {
		g.Sha256[k], err = g.generateSha256(g.UploadDir + f)
		if err != nil {
			g.Log.Error("Unable to generate sha256 from media file", "file", f, "error", err)
		}
	}
	return g, nil
}

func (g *Generators) generateSha256(path string) (string, error) {
	filePath := filepath.Clean(path)
	f, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return hex.EncodeToString(h.Sum(nil)), nil
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
	g.Log.Info("Generating new media file", "media_type", t, "media_id", id)
	if err := os.Symlink(g.UploadDir+g.Media[t], g.UploadDir+id); err != nil {
		panic(err)
	}
	return id
}

func (g *Generators) generateBaseMessage() *Message {
	contact := g.selectRndContact()
	msg := AcquireMessage()
	msg.Reset()
	msg.From = contact.GetWaId()
	msg.Id = uuid.New().String()
	msg.Timestamp = time.Now().Unix()
	return msg
}

func (g *Generators) GenerateMessages(n int, types ...string) []*Message {
	out := make([]*Message, n)

	for i := 0; i < n; i++ {
		typ := types[rand.Intn(len(types))]

		switch typ {
		case "text":
			out[i] = g.GenerateTextMessage()

		case "image":
			out[i] = g.GenerateImageMessage()

		case "audio":
			out[i] = g.GenerateAudioMessage()

		case "video":
			out[i] = g.GenerateVideoMessage()

		case "document":
			out[i] = g.GenerateDocumentMessage()
		default:
			g.Log.Warn("Unsupported message type", "message_type", typ)
		}
	}
	return out
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
		File:     "mockImagefile",
		Sha256:   g.Sha256[MessageType_image.String()],
	}
	return msg
}

func (g *Generators) GenerateVideoMessage() *Message {
	msg := g.generateBaseMessage()
	msg.Type = MessageType_video
	msg.Video = &VideoMessage{
		Id:       g.generateMedia(MessageType_video.String()),
		MimeType: "video/mp4",
		File:     "mockVideofile",
		Sha256:   g.Sha256[MessageType_video.String()],
	}
	return msg
}

func (g *Generators) GenerateAudioMessage() *Message {
	msg := g.generateBaseMessage()
	msg.Type = MessageType_audio
	msg.Audio = &AudioMessage{
		Id:       g.generateMedia(MessageType_audio.String()),
		MimeType: "audio/mp4",
		File:     "mockAudiofile",
		Sha256:   g.Sha256[MessageType_audio.String()],
	}
	return msg
}

func (g *Generators) GenerateDocumentMessage() *Message {
	msg := g.generateBaseMessage()
	msg.Type = MessageType_document
	msg.Document = &DocumentMessage{
		Id:       g.generateMedia(MessageType_document.String()),
		MimeType: "application/pdf",
		File:     "mockDocumentfile",
		Sha256:   g.Sha256[MessageType_document.String()],
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
	stat.Reset()
	stat.Id = msgID
	stat.RecipientId = recipient
	stat.Timestamp = time.Now().Unix()
	stat.Status = Status_StatusEnum(Status_StatusEnum_value[status])
	return stat
}
