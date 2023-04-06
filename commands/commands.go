package commands

import (
	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/api/cmdroute"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/session/shard"
	"github.com/diamondburned/arikawa/v3/state"
)

type interactionHandler struct {
	*cmdroute.Router
	s *state.State
	m *shard.Manager
}

func NewHandler(s *state.State, m *shard.Manager) *interactionHandler {
	h := &interactionHandler{
		s: s,
		m: m,
	}

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

var Commands = []api.CreateCommandData{
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
