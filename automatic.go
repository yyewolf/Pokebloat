package main

import (
	"encoding/json"
	"fmt"
	"log"
	"pokebloat/components"
	"pokebloat/utilities"
	"strings"

	"github.com/yyewolf/arikawa/v3/api"
	"github.com/yyewolf/arikawa/v3/discord"
	"github.com/yyewolf/arikawa/v3/gateway"
	"github.com/yyewolf/arikawa/v3/state"
	"github.com/yyewolf/arikawa/v3/utils/json/option"
	"github.com/yyewolf/arikawa/v3/utils/sendpart"
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
	img, err := utilities.DownloadImage(imageURL)

	newimg, err := utilities.ApplyTransforms(img, e.Author.ID, imageExtension)
	if err != nil {
		return
	}

	resp, err := utilities.PredictImage(newimg, imageExtension)
	if err != nil {
		return
	}

	var result []*utilities.APIResult
	err = json.Unmarshal(resp, &result)
	if err != nil {
		return
	}

	// Embed with the result
	embed := discord.Embed{
		Title: "Scan Results",
		Color: 0x00ff00,
		Image: &discord.EmbedImage{
			URL: "attachment://result.png",
		},
		Footer: &discord.EmbedFooter{
			Text: "Support : https://discord.gg/ZEAvn2M762",
		},
	}

	cutResult := []*utilities.APIResult{result[0]}

	m, err := s.SendMessageComplex(e.ChannelID, api.SendMessageData{
		Embeds: []discord.Embed{embed},
		Files: []sendpart.File{
			{
				Name:   "result.png",
				Reader: utilities.GenerateImage(cutResult),
			},
		},
		Components: []discord.ContainerComponent{
			&discord.ActionRowComponent{
				&discord.ButtonComponent{
					Label:    "More Results",
					Style:    discord.PrimaryButtonStyle(),
					CustomID: "more_results",
				},
			},
		},
	})
	if err != nil {
		log.Println(err)
		return
	}

	utilities.Cache.Set(m.ID.String(), components.Menu{
		Data: &MoreResultData{
			results:  result,
			imageURL: imageURL,
		},
		Fn: moreResult,
	}, 0)
}

type MoreResultData struct {
	results  []*utilities.APIResult
	imageURL string
}

func moreResult(ctx *components.MenuCtx) *api.InteractionResponse {
	data := ctx.Data.(*MoreResultData)
	utilities.Cache.Set(ctx.Message.ID.String(), components.Menu{
		Data: data,
		Fn:   report,
	}, 0)
	return &api.InteractionResponse{
		Type: api.UpdateMessage,
		Data: &api.InteractionResponseData{
			Embeds: &[]discord.Embed{
				{
					Title: "Scan Results",
					Color: 0x00ff00,
					Image: &discord.EmbedImage{
						URL: "attachment://result.png",
					},
					Footer: &discord.EmbedFooter{
						Text: "Support : https://discord.gg/ZEAvn2M762",
					},
				},
			},
			Components: &discord.ContainerComponents{
				&discord.ActionRowComponent{
					&discord.ButtonComponent{
						Label:    "Report",
						Style:    discord.DangerButtonStyle(),
						CustomID: "report",
					},
				},
			},
			Files: []sendpart.File{
				{
					Name:   "result.png",
					Reader: utilities.GenerateImage(data.results),
				},
			},
		},
	}
}

func report(ctx *components.MenuCtx) *api.InteractionResponse {
	// Send the message and the result to the report channel
	reportID := discord.ChannelID(1093802419546238976)
	data := ctx.Data.(*MoreResultData)
	_, err := ctx.S.SendMessageComplex(reportID, api.SendMessageData{
		Content: fmt.Sprintf("Report from %s", ctx.InteractionEvent.Member.User.Username),
		Embeds: []discord.Embed{
			{
				Title:       "Scan Results",
				Description: data.imageURL,
				Color:       0x00ff00,
				Image: &discord.EmbedImage{
					URL: "attachment://result.png",
				},
				Footer: &discord.EmbedFooter{
					Text: "Support : https://discord.gg/ZEAvn2M762",
				},
			},
		},
		Files: []sendpart.File{
			{
				Name:   "result.png",
				Reader: utilities.GenerateImage(data.results),
			},
		},
	})
	if err != nil {
		return &api.InteractionResponse{
			Type: api.MessageInteractionWithSource,
			Data: &api.InteractionResponseData{
				Content: option.NewNullableString("An error occured while sending the report"),
				Flags:   discord.EphemeralMessage,
			},
		}
	}
	utilities.Cache.Delete(ctx.Message.ID.String())
	return &api.InteractionResponse{
		Type: api.MessageInteractionWithSource,
		Data: &api.InteractionResponseData{
			Content: option.NewNullableString("Report sent"),
			Flags:   discord.EphemeralMessage,
		},
	}
}
