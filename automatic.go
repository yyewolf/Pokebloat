package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strings"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/diamondburned/arikawa/v3/utils/sendpart"
)

func handleMessages(s *state.State, e *gateway.MessageCreateEvent) {
	// Check if there's an embed
	if len(e.Message.Embeds) == 0 {
		return
	}

	// Check if the embed is a pokemon embed
	if !strings.Contains(e.Message.Embeds[0].Title, "wild pokémon has") {
		return
	}

	message := e.Message

	var imageURL string

	// Search attachments for images
	for _, attachment := range message.Attachments {
		if attachment.ContentType == "image/png" || attachment.ContentType == "image/jpeg" {
			imageURL = attachment.URL
			break
		}
	}

	// Search embeds for images
	for _, embed := range message.Embeds {
		if embed.Image != nil {
			imageURL = embed.Image.URL
			break
		}
	}

	if imageURL == "" {
		return
	}

	var imageExtension = strings.Split(imageURL, ".")[len(strings.Split(imageURL, "."))-1]

	// Download image
	rep, err := http.Get(imageURL)
	if err != nil {
		return
	}

	// Send to the API for transformation
	url := fmt.Sprintf("%s/transforms/%s/background_removal", os.Getenv("SECONDARY_API_URL"), "pokemons")
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("upload", "image."+imageExtension)
	if err != nil {
		return
	}
	io.Copy(part, rep.Body)
	writer.Close()
	rep.Body.Close()

	r, err := http.NewRequest("POST", url, body)
	if err != nil {
		return
	}
	r.Header.Set("Content-Type", writer.FormDataContentType())
	client := &http.Client{}
	rep, err = client.Do(r)
	if err != nil {
		return
	}

	// Send to the API
	url = fmt.Sprintf("%s/models/%s/predict?k=3", os.Getenv("API_URL"), "pokemons")
	body = &bytes.Buffer{}
	writer = multipart.NewWriter(body)
	part, err = writer.CreateFormFile("upload", "image."+imageExtension)
	if err != nil {
		return
	}
	io.Copy(part, rep.Body)
	writer.Close()
	rep.Body.Close()

	r, err = http.NewRequest("POST", url, body)
	if err != nil {
		return
	}
	r.Header.Add("Content-Type", writer.FormDataContentType())
	rep, err = client.Do(r)
	if err != nil {
		return
	}
	resp, err := io.ReadAll(rep.Body)
	if err != nil {
		return
	}
	rep.Body.Close()

	var result []*APIResult
	err = json.Unmarshal(resp, &result)
	if err != nil {
		return
	}

	// Embed with the result
	embed := discord.Embed{
		Title:       "Scan Results",
		Description: "The results may be affected by the April Fools event. Come check out the new support server!",
		Color:       0x00ff00,
		Image: &discord.EmbedImage{
			URL: "attachment://result.png",
		},
		Footer: &discord.EmbedFooter{
			Text: "Support : https://discord.gg/ZEAvn2M762",
		},
	}

	s.SendMessageComplex(e.ChannelID, api.SendMessageData{
		Embeds: []discord.Embed{embed},
		Files: []sendpart.File{
			{
				Name:   "result.png",
				Reader: generateImage(result),
			},
		},
	})
}
