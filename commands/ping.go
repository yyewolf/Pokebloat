package commands

import (
	"context"

	"github.com/yyewolf/arikawa/v3/api"
	"github.com/yyewolf/arikawa/v3/api/cmdroute"
	"github.com/yyewolf/arikawa/v3/discord"
	"github.com/yyewolf/arikawa/v3/utils/json/option"
)

func (h *interactionHandler) cmdPing(ctx context.Context, cmd cmdroute.CommandData) *api.InteractionResponseData {
	return &api.InteractionResponseData{
		Content: option.NewNullableString("Pong!"),
		Flags:   discord.MessageFlags(discord.EphemeralMessage),
	}
}
