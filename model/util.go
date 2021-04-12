package model

import (
	"bytes"

	"github.com/gogo/protobuf/jsonpb"
)

var (
	marsheler = jsonpb.Marshaler{
		EmitDefaults: false,
		EnumsAsInts:  false,
		OrigName:     true,
	}
)

func NewConfig() *InternalConfig {
	return &InternalConfig{
		ApplicationSettings: &ApplicationSettings{
			Media: &ApplicationSettings_Media{
				AutoDownload: []string{},
			},
			Webhooks: &ApplicationSettings_Webhooks{
				Url:                   "",
				MaxConcurrentRequests: 8,
			},
		},
		InboundMedia:    map[string]string{},
		Contacts:        []*InternalContact{},
		Users:           map[string]string{},
		BusinessProfile: &BusinessProfile{},
		ProfileAbout:    &ProfileAbout{},
	}
}

func (x *InternalConfig) DeepCopy() (*InternalConfig, error) {
	buf, err := x.Marshal()
	if err != nil {
		return nil, err
	}
	out := NewConfig()
	err = jsonpb.Unmarshal(buf, out)
	return out, err
}

func (x *InternalConfig) Marshal() (*bytes.Buffer, error) {
	buf := bytes.NewBuffer(nil)
	err := marsheler.Marshal(buf, x)
	return buf, err
}
