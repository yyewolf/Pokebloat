package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/api/cmdroute"
	"github.com/diamondburned/arikawa/v3/api/webhook"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
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
		Type: discord.MessageCommand,
		Name: "scan",
	},
}

func main() {
	godotenv.Load()
	token := os.Getenv("BOT_TOKEN")
	if token == "" {
		log.Fatalln("No $BOT_TOKEN given.")
	}

	var (
		webhookAddr   = os.Getenv("WEBHOOK_ADDR")
		webhookPubkey = os.Getenv("WEBHOOK_PUBKEY")
	)

	if webhookAddr != "" {
		state := state.NewAPIOnlyState(token, nil)

		h := newHandler(state)

		if err := overwriteCommands(state); err != nil {
			log.Fatalln("cannot update commands:", err)
		}

		srv, err := webhook.NewInteractionServer(webhookPubkey, h)
		if err != nil {
			log.Fatalln("cannot create interaction server:", err)
		}

		log.Println("listening and serving at", webhookAddr+"/")
		log.Fatalln(http.ListenAndServe(webhookAddr, srv))
	} else {
		state := state.New("Bot " + token)
		state.AddIntents(gateway.IntentGuilds)
		state.AddHandler(func(*gateway.ReadyEvent) {
			me, _ := state.Me()
			log.Println("connected to the gateway as", me.Tag())
		})

		h := newHandler(state)
		state.AddInteractionHandler(h)

		if err := overwriteCommands(state); err != nil {
			log.Fatalln("cannot update commands:", err)
		}

		if err := h.s.Connect(context.Background()); err != nil {
			log.Fatalln("cannot connect:", err)
		}
	}
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
