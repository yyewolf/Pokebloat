package main

import (
	"context"
	"fmt"
	"strconv"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/api/cmdroute"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/diamondburned/arikawa/v3/utils/json/option"
)

func (h *handler) cmdAnnounce(ctx context.Context, cmd cmdroute.CommandData) *api.InteractionResponseData {
	if cmd.Event.User.ID != AdminID {
		return &api.InteractionResponseData{
			Content: option.NewNullableString("You are not allowed to use this command."),
			Flags:   discord.MessageFlags(discord.EphemeralMessage),
		}
	}
	if manager == nil {
		return &api.InteractionResponseData{
			Content: option.NewNullableString("All shards are not ready yet"),
			Flags:   discord.MessageFlags(discord.EphemeralMessage),
		}
	}
	// Get amount of guilds & users.
	guildCount := 0
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
		guildCount += len(guilds)
	}

	// Create the embed.
	embed := discord.Embed{
		Title:       "I'm back.",
		Description: "Hi there, **Pocketboat** is back.\nThe AI Model is not quite yet ready but it will be soon.\nFor more informations and demo you can come to our [support server](https://discord.gg/ZEAvn2M762).",
		Color:       0x00ff00,
	}

	go func() {
		done := 0
		for s := 0; s < manager.NumShards(); s++ {
			shard := manager.Shard(s)
			state := shard.(*state.State)
			guilds, _ := state.GuildStore.Guilds()
			for _, g := range guilds {
				// Get the channel ID.
				channels, err := state.ChannelStore.Channels(g.ID)
				if err != nil {
					break
				}

				for _, c := range channels {
					// Send the embed.
					_, err = state.SendEmbeds(c.ID, embed)
					if err != nil {
						continue
					}
					break
				}
				done++
				fmt.Printf("Announced to %d/%d servers (%.2f%%)                \r", done, guildCount, float64(done)/float64(guildCount)*100)
			}
		}
		fmt.Println("Announced to", done, "servers out of", guildCount, "servers. ("+fmt.Sprintf("%.2f", float64(done)/float64(guildCount)*100)+"%)")
	}()

	return &api.InteractionResponseData{
		Content: option.NewNullableString("Announcing to **" + strconv.Itoa(guildCount) + "** servers"),
		Flags:   discord.MessageFlags(discord.EphemeralMessage),
	}
}
