package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strings"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/api/cmdroute"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/diamondburned/arikawa/v3/utils/json/option"
	"github.com/diamondburned/arikawa/v3/utils/sendpart"
)

func newHandler(s *state.State) *handler {
	h := &handler{s: s}

	h.Router = cmdroute.NewRouter()
	// Automatically defer handles if they're slow.
	h.Use(cmdroute.Deferrable(s, cmdroute.DeferOpts{}))
	h.AddFunc("ping", h.cmdPing)
	h.AddFunc("scan", h.cmdScan)
	return h
}

func (h *handler) cmdPing(ctx context.Context, cmd cmdroute.CommandData) *api.InteractionResponseData {
	return &api.InteractionResponseData{
		Content: option.NewNullableString("Pong!"),
	}
}

func (h *handler) cmdScan(ctx context.Context, cmd cmdroute.CommandData) *api.InteractionResponseData {
	var message *discord.Message
	data := cmd.Event.Data.(*discord.CommandInteraction)
	if data.Resolved.Messages != nil {
		for _, msg := range data.Resolved.Messages {
			message = &msg
			break
		}
	}

	if message == nil {
		return &api.InteractionResponseData{
			Content: option.NewNullableString("No message found"),
			Flags:   discord.MessageFlags(discord.EphemeralMessage),
		}
	}

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
		return &api.InteractionResponseData{
			Content: option.NewNullableString("No image found"),
			Flags:   discord.MessageFlags(discord.EphemeralMessage),
		}
	}

	var imageExtension = strings.Split(imageURL, ".")[len(strings.Split(imageURL, "."))-1]

	// Download image
	rep, err := http.Get(imageURL)
	if err != nil {
		return &api.InteractionResponseData{
			Content: option.NewNullableString("Error downloading image"),
			Flags:   discord.MessageFlags(discord.EphemeralMessage),
		}
	}

	// Send to the API
	url := os.Getenv("API_URL")
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("upload", "image."+imageExtension)
	if err != nil {
		return &api.InteractionResponseData{
			Content: option.NewNullableString("Error uploading image"),
			Flags:   discord.MessageFlags(discord.EphemeralMessage),
		}
	}
	io.Copy(part, rep.Body)
	writer.Close()
	rep.Body.Close()

	r, err := http.NewRequest("POST", url, body)
	if err != nil {
		return &api.InteractionResponseData{
			Content: option.NewNullableString("Error uploading image"),
			Flags:   discord.MessageFlags(discord.EphemeralMessage),
		}
	}
	r.Header.Add("Content-Type", writer.FormDataContentType())
	client := &http.Client{}
	rep, err = client.Do(r)
	if err != nil {
		return &api.InteractionResponseData{
			Content: option.NewNullableString("Error uploading image"),
			Flags:   discord.MessageFlags(discord.EphemeralMessage),
		}
	}
	resp, err := io.ReadAll(rep.Body)
	rep.Body.Close()

	var result []*APIResult
	err = json.Unmarshal(resp, &result)
	if err != nil {
		return &api.InteractionResponseData{
			Content: option.NewNullableString("Error uploading image"),
			Flags:   discord.MessageFlags(discord.EphemeralMessage),
		}
	}

	return &api.InteractionResponseData{
		Files: []sendpart.File{
			{
				Name:   "result.png",
				Reader: generateImage(result),
			},
		},
		Flags: discord.MessageFlags(discord.EphemeralMessage),
	}
}
