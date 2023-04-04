package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/api/cmdroute"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/session/shard"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/diamondburned/arikawa/v3/utils/json/option"
	"github.com/joho/godotenv"
)

var commands = []api.CreateCommandData{
	{
		Name:        "ping",
		Description: "ping pong!",
	},
	{
		Name:        "invite",
		Description: "Invite the bot to your server",
	},
	{
		Name:        "status",
		Description: "Get the status of the bot",
	},
	{
		Type: discord.MessageCommand,
		Name: "pokemon",
	},
	{
		Type: discord.MessageCommand,
		Name: "pokemon_acc",
	},
}

var manager *shard.Manager
var AdminID = discord.UserID(0)

func main() {
	godotenv.Load()
	token := os.Getenv("BOT_TOKEN")
	if token == "" {
		log.Fatalln("No $BOT_TOKEN given.")
	}

	aID := os.Getenv("ADMIN_ID")
	if aID == "" {
		log.Fatalln("No $ADMIN_ID given.")
	}
	i, _ := strconv.ParseInt(aID, 10, 64)
	AdminID = discord.UserID(i)

	var shardNum int
	newShard := state.NewShardFunc(func(m *shard.Manager, s *state.State) {
		// Add the needed Gateway intents.
		s.AddIntents(gateway.IntentGuildMessages)
		s.AddIntents(gateway.IntentDirectMessages)

		s.AddIntents(gateway.IntentGuilds)

		h := newHandler(s)
		s.AddInteractionHandler(h)

		if err := overwriteCommands(s); err != nil {
			log.Fatalln("cannot update commands:", err)
		}

		u, err := s.Me()
		if err != nil {
			log.Fatalln("failed to get myself:", err)
		}

		// Open the gateway connection.
		if err := s.Open(context.Background()); err != nil {
			log.Fatalln("failed to open state:", err)
		}

		s.Gateway().Send(context.Background(), &gateway.UpdatePresenceCommand{
			Status: "online",
			Activities: []discord.Activity{
				{
					Name: "I'm back.",
					Type: discord.GameActivity,
				},
			},
		})

		log.Printf("Shard %d/%d started as %s", shardNum, m.NumShards()-1, u.Tag())

		shardNum++
	})

	m, err := shard.NewManager("Bot "+token, newShard)
	if err != nil {
		log.Fatalln("failed to create shard manager:", err)
	}

	manager = m

	// Block forever.
	fmt.Println("Press Ctrl+C to exit.")
	select {}
}

func overwriteCommands(s *state.State) error {
	return cmdroute.OverwriteCommands(s, commands)
}

type handler struct {
	*cmdroute.Router
	s *state.State
}

func errorResponse(err error) *api.InteractionResponseData {
	return &api.InteractionResponseData{
		Content:         option.NewNullableString("**Error:** " + err.Error()),
		Flags:           discord.EphemeralMessage,
		AllowedMentions: &api.AllowedMentions{ /* none */ },
	}
}
