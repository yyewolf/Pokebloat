package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"runtime"
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
	h.AddFunc("pokemon_acc", h.cmdScanPokemonBg)
	h.AddFunc("pokemon", h.cmdScanPokemon)
	h.AddFunc("invite", h.cmdInvite)
	h.AddFunc("status", h.cmdStatus)
	return h
}

func (h *handler) cmdInvite(ctx context.Context, cmd cmdroute.CommandData) *api.InteractionResponseData {
	return &api.InteractionResponseData{
		Content: option.NewNullableString("https://discord.com/api/oauth2/authorize?client_id=837040988378759249&permissions=8&scope=applications.commands%20bot"),
		Flags:   discord.MessageFlags(discord.EphemeralMessage),
	}
}

func (h *handler) cmdStatus(ctx context.Context, cmd cmdroute.CommandData) *api.InteractionResponseData {
	if manager == nil {
		return &api.InteractionResponseData{
			Content: option.NewNullableString("All shards are not ready yet"),
			Flags:   discord.MessageFlags(discord.EphemeralMessage),
		}
	}
	guildCount := 0
	userCount := 0
	for s := 0; s < manager.NumShards(); s++ {
		shard := manager.Shard(s)
		state := shard.(*state.State)
		guilds, err := state.GuildStore.Guilds()
		if err != nil {
			return &api.InteractionResponseData{
				Content: option.NewNullableString("Error getting guilds"),
				Flags:   discord.MessageFlags(discord.EphemeralMessage),
			}
		}
		for _, g := range guilds {
			users, _ := state.MemberStore.Members(g.ID)
			userCount += len(users)
		}
		guildCount += len(guilds)
	}

	shard := h.s.Ready().Shard.ShardID()
	shardCount := h.s.Ready().Shard.NumShards()
	// tasks is the number of goroutines running
	tasks := runtime.NumGoroutine()

	embed := &discord.Embed{
		Title:       "Status : Alive",
		Description: fmt.Sprintf("**Servers**: %d\n**Current Shard**: %d/%d\n**Users**: %d\n**Tasks**: tasks", guildCount, shard+1, shardCount, userCount, tasks),
		Color:       0x00ff00,
	}

	return &api.InteractionResponseData{
		Embeds: &[]discord.Embed{*embed},
		Flags:  discord.MessageFlags(discord.EphemeralMessage),
	}
}

func (h *handler) cmdPing(ctx context.Context, cmd cmdroute.CommandData) *api.InteractionResponseData {
	return &api.InteractionResponseData{
		Content: option.NewNullableString("Pong!"),
		Flags:   discord.MessageFlags(discord.EphemeralMessage),
	}
}

func (h *handler) cmdScanPokemon(ctx context.Context, cmd cmdroute.CommandData) *api.InteractionResponseData {
	return h.cmdScan(ctx, cmd, "pokemons", false)
}

func (h *handler) cmdScanPokemonBg(ctx context.Context, cmd cmdroute.CommandData) *api.InteractionResponseData {
	return h.cmdScan(ctx, cmd, "pokemons", true)
}

func (h *handler) cmdScan(ctx context.Context, cmd cmdroute.CommandData, model string, transform bool) *api.InteractionResponseData {
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
			Content: option.NewNullableString("Could not find an image in the message"),
			Flags:   discord.MessageFlags(discord.EphemeralMessage),
		}
	}

	var imageExtension = strings.Split(imageURL, ".")[len(strings.Split(imageURL, "."))-1]

	// Download image
	rep, err := http.Get(imageURL)
	if err != nil {
		return &api.InteractionResponseData{
			Content: option.NewNullableString("Could not download image from message"),
			Flags:   discord.MessageFlags(discord.EphemeralMessage),
		}
	}

	if transform {
		// Send to the API for transformation
		url := fmt.Sprintf("%s/transforms/%s/background_removal", os.Getenv("API_URL"), model)
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		part, err := writer.CreateFormFile("upload", "image."+imageExtension)
		if err != nil {
			return &api.InteractionResponseData{
				Content: option.NewNullableString("The backend service is currently down. Please try again later."),
				Flags:   discord.MessageFlags(discord.EphemeralMessage),
			}
		}
		io.Copy(part, rep.Body)
		writer.Close()
		rep.Body.Close()

		r, err := http.NewRequest("POST", url, body)
		if err != nil {
			return &api.InteractionResponseData{
				Content: option.NewNullableString("The backend service is currently down. Please try again later."),
				Flags:   discord.MessageFlags(discord.EphemeralMessage),
			}
		}
		r.Header.Set("Content-Type", writer.FormDataContentType())
		client := &http.Client{}
		rep, err = client.Do(r)
		if err != nil {
			return &api.InteractionResponseData{
				Content: option.NewNullableString("The backend service is currently down. Please try again later."),
				Flags:   discord.MessageFlags(discord.EphemeralMessage),
			}
		}
	}

	// Send to the API
	url := fmt.Sprintf("%s/models/%s/predict?k=3", os.Getenv("API_URL"), model)
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("upload", "image."+imageExtension)
	if err != nil {
		return &api.InteractionResponseData{
			Content: option.NewNullableString("The backend service is currently down. Please try again later."),
			Flags:   discord.MessageFlags(discord.EphemeralMessage),
		}
	}
	io.Copy(part, rep.Body)
	writer.Close()
	rep.Body.Close()

	r, err := http.NewRequest("POST", url, body)
	if err != nil {
		return &api.InteractionResponseData{
			Content: option.NewNullableString("The backend service is currently down. Please try again later."),
			Flags:   discord.MessageFlags(discord.EphemeralMessage),
		}
	}
	r.Header.Add("Content-Type", writer.FormDataContentType())
	client := &http.Client{}
	rep, err = client.Do(r)
	if err != nil {
		return &api.InteractionResponseData{
			Content: option.NewNullableString("The backend service is currently down. Please try again later."),
			Flags:   discord.MessageFlags(discord.EphemeralMessage),
		}
	}
	resp, err := io.ReadAll(rep.Body)
	if err != nil {
		return &api.InteractionResponseData{
			Content: option.NewNullableString("The backend service is currently down. Please try again later."),
			Flags:   discord.MessageFlags(discord.EphemeralMessage),
		}
	}
	rep.Body.Close()

	var result []*APIResult
	err = json.Unmarshal(resp, &result)
	if err != nil {
		return &api.InteractionResponseData{
			Content: option.NewNullableString("The backend service is currently down. Please try again later."),
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
	}
}
