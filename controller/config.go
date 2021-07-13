package controller

import (
	"io"

	"github.com/ron96G/whatsapp-bizapi-mock/model"
)

var (
	Config *model.InternalConfig
)

func InitConfig(r io.Reader) error {
	Config = model.NewConfig()
	return unmarsheler.Unmarshal(r, Config)
}
