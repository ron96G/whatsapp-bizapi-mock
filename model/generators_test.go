package model

import (
	"fmt"
	"testing"
)

var (
	generators *Generators
)

func init() {
	uploadDir := "uploads/"
	contacts := []*Contact{
		{
			WaId: "49170123123",
			Profile: &Contact_Profile{
				Name: "TestUser1",
			},
		},
		{
			WaId: "49170123124",
			Profile: &Contact_Profile{
				Name: "TestUser2",
			},
		},
	}

	media := map[string]string{
		"image": "image",
		"video": "video",
		"audio": "audio",
	}

	generators = NewGenerators(uploadDir, contacts, media)
}

func Test_GenerateTextMessage(t *testing.T) {
	fmt.Println(generators.GenerateTextMessage().String())
}

func Test_GenerateImageMessage(t *testing.T) {
	fmt.Println(generators.GenerateImageMessage().String())
}

func Test_GenerateVideoMessage(t *testing.T) {
	fmt.Println(generators.GenerateVideoMessage().String())
}

func Test_GenerateAudioMessage(t *testing.T) {
	fmt.Println(generators.GenerateAudioMessage().String())
}

func Test_GenerateRndMessages(t *testing.T) {
	msgs := generators.GenerateRndMessages(100)
	for _, msg := range msgs {
		fmt.Println(msg.String())
	}
}
