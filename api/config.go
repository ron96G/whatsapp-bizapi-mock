package api

import (
	"io"

	"github.com/ron96G/whatsapp-bizapi-mock/model"
)

var (
	Config = &model.InternalConfig{
		Version:   Version,
		Status:    ApiStatus.String(),
		Contacts:  []*model.InternalContact{},
		UploadDir: "uploads",
		Users: map[string]string{
			"admin": "secret",
		},
		InboundMedia: map[string]string{},
		ApplicationSettings: &model.ApplicationSettings{
			Media: &model.ApplicationSettings_Media{
				AutoDownload: []string{},
			},
			Webhooks: &model.ApplicationSettings_Webhooks{
				Url:                   "",
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
