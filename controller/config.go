package controller

import (
	"io"

	"github.com/rgumi/whatsapp-mock/model"
)

var (
	Config *model.InternalConfig
)

func InitConfig(r io.Reader) error {
	Config = model.NewConfig()
	return unmarsheler.Unmarshal(r, Config)
}
