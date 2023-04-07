package components

import (
	"pokebloat/utilities"

	"github.com/yyewolf/arikawa/v3/api"
	"github.com/yyewolf/arikawa/v3/discord"
	"github.com/yyewolf/arikawa/v3/session/shard"
	"github.com/yyewolf/arikawa/v3/state"
)

type componentHandler struct {
	S *state.State
	M *shard.Manager
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
		S: s,
		M: m,
	}

	return h
}

func (h *componentHandler) HandleInteraction(e *discord.InteractionEvent) *api.InteractionResponse {
	if e.Data.InteractionType() != discord.ComponentInteractionType {
		return nil
	}
	menuid := e.Message.ID
	var menuidI discord.InteractionID
	if e.Message.Interaction != nil {
		menuidI = e.Message.Interaction.ID
	}

	menu, found := utilities.Cache.Get(menuid.String())
	if !found {
		menu, found = utilities.Cache.Get(menuidI.String())
		if !found {
			// Just ack to discord
			return &api.InteractionResponse{
				Type: api.DeferredMessageUpdate,
			}
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
