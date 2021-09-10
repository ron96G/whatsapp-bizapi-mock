package api

import (
	"io"

	"github.com/ron96G/whatsapp-bizapi-mock/model"
)

var (
	// This is the default config if non is provided on startup
	// It will be overwritten by a config is provided on startup
	Config = &model.InternalConfig{
		Version: Version,
		Status:  ApiStatus.String(),
		Contacts: []*model.InternalContact{
			{
				Id:   "491701223123",
				Name: "Peter P.",
			},
			{
				Id:   "491701223124",
				Name: "S. Klaus",
			},
			{
				Id:   "491701223125",
				Name: "Peter L.",
			},
		},
		UploadDir: "media/",
		Users: map[string]string{
			"admin": "secret",
		},
		InboundMedia: map[string]string{
			"audio":    "audio",
			"document": "document",
			"image":    "image",
			"video":    "video",
		},
		ApplicationSettings: &model.ApplicationSettings{
			Media: &model.ApplicationSettings_Media{
				AutoDownload: []string{},
			},
			Webhooks: &model.ApplicationSettings_Webhooks{
				Url:                   "https://localhost:9000/webhook",
				MaxConcurrentRequests: 8,
			},
		},
		ProfileAbout:         &model.ProfileAbout{},
		BusinessProfile:      &model.BusinessProfile{},
		ProfilePhotoFilename: "",
		Verified:             false,
		WebhookCA:            nil,
	}
)

func InitConfig(r io.Reader) error {
	Config = model.NewConfig()
	return unmarsheler.Unmarshal(r, Config)
}
