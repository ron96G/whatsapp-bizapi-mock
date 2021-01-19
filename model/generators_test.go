package model

import (
	"fmt"
	"testing"
)

func init() {
	UploadDir = "/mnt/d/PROJECTS/whatsapp-mock/uploads/"
	Contacts = []*Contact{
		&Contact{
			WaId: "49170123123",
			Profile: &Contact_Profile{
				Name: "TestUser1",
			},
		},
		&Contact{
			WaId: "49170123124",
			Profile: &Contact_Profile{
				Name: "TestUser2",
			},
		},
	}

	Media = map[string]string{
		"image": "image",
		"video": "video",
		"audio": "audio",
	}
}

func Test_GenerateTextMessage(t *testing.T) {
	fmt.Println(GenerateTextMessage().String())
}

func Test_GenerateImageMessage(t *testing.T) {
	fmt.Println(GenerateImageMessage().String())
}

func Test_GenerateVideoMessage(t *testing.T) {
	fmt.Println(GenerateVideoMessage().String())
}

func Test_GenerateAudioMessage(t *testing.T) {
	fmt.Println(GenerateAudioMessage().String())
}

func Test_GenerateRndMessages(t *testing.T) {
	msgs := GenerateRndMessages(100)
	for _, msg := range msgs {
		fmt.Println(msg.String())
	}
}
