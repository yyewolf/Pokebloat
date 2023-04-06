package commands

import (
	"context"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/api/cmdroute"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/utils/json/option"
)

func (h *interactionHandler) cmdInvite(ctx context.Context, cmd cmdroute.CommandData) *api.InteractionResponseData {
	return &api.InteractionResponseData{
		Content: option.NewNullableString("Invite me : https://discord.com/api/oauth2/authorize?client_id=837040988378759249&permissions=8&scope=applications.commands%20bot\nSupport: https://discord.gg/ZEAvn2M762"),
		Flags:   discord.MessageFlags(discord.EphemeralMessage),
	}
}
