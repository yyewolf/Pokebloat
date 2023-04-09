package commands

import (
	"context"
	"encoding/json"
	"pokebloat/components"
	"pokebloat/utilities"
	"strings"

	"github.com/yyewolf/arikawa/v3/api"
	"github.com/yyewolf/arikawa/v3/api/cmdroute"
	"github.com/yyewolf/arikawa/v3/discord"
	"github.com/yyewolf/arikawa/v3/utils/json/option"
	"github.com/yyewolf/arikawa/v3/utils/sendpart"
)

func (h *interactionHandler) cmdScanPokemon(ctx context.Context, cmd cmdroute.CommandData) *api.InteractionResponseData {
	return h.cmdScan(ctx, cmd, "pokemons", "")
}

func (h *interactionHandler) cmdScanPokemonBg(ctx context.Context, cmd cmdroute.CommandData) *api.InteractionResponseData {
	return h.cmdScan(ctx, cmd, "pokemons", "bg")
}

func (h *interactionHandler) cmdScanPokemonPc(ctx context.Context, cmd cmdroute.CommandData) *api.InteractionResponseData {
	return h.cmdScan(ctx, cmd, "pokemons", "pc")
}

func (h *interactionHandler) cmdScan(ctx context.Context, cmd cmdroute.CommandData, model string, forceTransform string) *api.InteractionResponseData {
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
	img, err := utilities.DownloadImage(imageURL)
	if err != nil {
		return &api.InteractionResponseData{
			Content: option.NewNullableString("The backend service is currently down. Please try again later."),
			Flags:   discord.MessageFlags(discord.EphemeralMessage),
		}
	}

	newimg, err := utilities.ApplyTransformsManual(img, forceTransform, imageExtension)
	if err != nil {
		return &api.InteractionResponseData{
			Content: option.NewNullableString("The backend service is currently down. Please try again later."),
			Flags:   discord.MessageFlags(discord.EphemeralMessage),
		}
	}

	resp, err := utilities.PredictImage(newimg, imageExtension)
	if err != nil {
		return &api.InteractionResponseData{
			Content: option.NewNullableString("The backend service is currently down. Please try again later."),
			Flags:   discord.MessageFlags(discord.EphemeralMessage),
		}
	}

	var result []*utilities.APIResult
	err = json.Unmarshal(resp, &result)
	if err != nil {
		return &api.InteractionResponseData{
			Content: option.NewNullableString("The backend service is currently down. Please try again later."),
			Flags:   discord.MessageFlags(discord.EphemeralMessage),
		}
	}

	// return &api.InteractionResponseData{
	// 	Files: []sendpart.File{
	// 		{
	// 			Name:   "result.png",
	// 			Reader: generateImage(result),
	// 		},
	// 	},
	// }

	// Embed with the result
	embed := discord.Embed{
		Title:       "Scan Result",
		Description: "The following pokemons were found in the image",
		Color:       0x00ff00,
		Image: &discord.EmbedImage{
			URL: "attachment://result.png",
		},
		Footer: &discord.EmbedFooter{
			Text: "Support : https://discord.gg/ZEAvn2M762",
		},
	}

	utilities.Cache.Set(cmd.Event.ID.String(), components.Menu{
		Data: &MoreResultData{
			results:     result,
			imageURL:    imageURL,
			interaction: cmd.Event,
		},
		Fn: moreResult,
	}, 0)
	cutResult := []*utilities.APIResult{result[0]}

	return &api.InteractionResponseData{
		Embeds: &[]discord.Embed{embed},
		Files: []sendpart.File{
			{
				Name:   "result.png",
				Reader: utilities.GenerateImage(cutResult),
			},
		},
		Components: &discord.ContainerComponents{
			&discord.ActionRowComponent{
				&discord.ButtonComponent{
					Label:    "More Results",
					Style:    discord.PrimaryButtonStyle(),
					CustomID: "more_results",
				},
			},
		},
	}
}

type MoreResultData struct {
	results     []*utilities.APIResult
	imageURL    string
	interaction *discord.InteractionEvent
}

func moreResult(ctx *components.MenuCtx) *api.InteractionResponse {
	data := ctx.Data.(*MoreResultData)
	utilities.Cache.Delete(ctx.Message.ID.String())
	return &api.InteractionResponse{
		Type: api.UpdateMessage,
		Data: &api.InteractionResponseData{
			Embeds: &[]discord.Embed{
				{
					Title: "Scan Results",
					Color: 0x00ff00,
					Image: &discord.EmbedImage{
						URL: "attachment://result2.png",
					},
					Footer: &discord.EmbedFooter{
						Text: "Support : https://discord.gg/ZEAvn2M762",
					},
				},
			},
			Components: &discord.ContainerComponents{},
			Files: []sendpart.File{
				{
					Name:   "result2.png",
					Reader: utilities.GenerateImage(data.results),
				},
			},
		},
	}
}
