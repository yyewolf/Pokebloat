package commands

import (
	"context"
	"fmt"
	"runtime"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/api/cmdroute"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/diamondburned/arikawa/v3/utils/json/option"
)

func (h *interactionHandler) cmdStatus(ctx context.Context, cmd cmdroute.CommandData) *api.InteractionResponseData {
	if h.m == nil {
		return &api.InteractionResponseData{
			Content: option.NewNullableString("All shards are not ready yet"),
			Flags:   discord.MessageFlags(discord.EphemeralMessage),
		}
	}
	guildCount := 0
	userCount := 0
	for s := 0; s < h.m.NumShards(); s++ {
		shard := h.m.Shard(s)
		if shard == nil {
			continue
		}
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
		Description: fmt.Sprintf("**Servers**: %d\n**Current Shard**: %d/%d\n**Users**: %d\n**Tasks**: %d", guildCount, shard+1, shardCount, userCount, tasks),
		Color:       0x00ff00,
	}

	return &api.InteractionResponseData{
		Embeds: &[]discord.Embed{*embed},
		Flags:  discord.MessageFlags(discord.EphemeralMessage),
	}
}
