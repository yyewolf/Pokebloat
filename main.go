package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"pokebloat/commands"
	"pokebloat/components"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/yyewolf/arikawa/v3/api"
	"github.com/yyewolf/arikawa/v3/api/cmdroute"
	"github.com/yyewolf/arikawa/v3/discord"
	"github.com/yyewolf/arikawa/v3/gateway"
	"github.com/yyewolf/arikawa/v3/session/shard"
	"github.com/yyewolf/arikawa/v3/state"
	"github.com/yyewolf/arikawa/v3/utils/json/option"
)

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
		s.AddIntents(gateway.IntentMessageContent)

		s.AddIntents(gateway.IntentGuilds)

		interactionsHandling := commands.NewHandler(s, m)

		s.AddInteractionHandler(interactionsHandling)
		s.AddHandler(func(e *gateway.MessageCreateEvent) {
			handleMessages(s, e)
		})

		componentHandling := components.NewHandler(s, m)

		s.AddInteractionHandler(componentHandling)

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

	_, err := shard.NewManager("Bot "+token, newShard)
	if err != nil {
		log.Fatalln("failed to create shard manager:", err)
	}

	// Block forever.
	fmt.Println("Press Ctrl+C to exit.")
	select {}
}

func overwriteCommands(s *state.State) error {
	return cmdroute.OverwriteCommands(s, commands.Commands)
}

func errorResponse(err error) *api.InteractionResponseData {
	return &api.InteractionResponseData{
		Content:         option.NewNullableString("**Error:** " + err.Error()),
		Flags:           discord.EphemeralMessage,
		AllowedMentions: &api.AllowedMentions{ /* none */ },
	}
}
