package components

import (
	"pokebloat/utilities"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/session/shard"
	"github.com/diamondburned/arikawa/v3/state"
)

type componentHandler struct {
	s *state.State
	m *shard.Manager
}

type Menu struct {
	Data interface{}
	Fn   func(ctx *MenuCtx) *api.InteractionResponse
}

type MenuCtx struct {
	*componentHandler
	*discord.InteractionEvent
	MenuID discord.MessageID
	Data   interface{}
}

func NewHandler(s *state.State, m *shard.Manager) *componentHandler {
	h := &componentHandler{
		s: s,
		m: m,
	}

	return h
}

func (h *componentHandler) HandleInteraction(e *discord.InteractionEvent) *api.InteractionResponse {
	menuid := e.Message.ID

	menu, found := utilities.Cache.Get(menuid.String())
	if !found {
		// Just ack to discord
		return &api.InteractionResponse{
			Type: api.DeferredMessageUpdate,
		}
	}

	if menu, ok := menu.(Menu); ok {
		ctx := &MenuCtx{
			componentHandler: h,
			InteractionEvent: e,
			MenuID:           menuid,
			Data:             menu.Data,
		}
		return menu.Fn(ctx)
	}

	return &api.InteractionResponse{
		Type: api.DeferredMessageUpdate,
	}
}