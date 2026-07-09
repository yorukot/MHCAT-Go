package fakeinteractions

import (
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/customid"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/interactions"
)

func Modal(customID string, fields ...customid.ModalField) interactions.Interaction {
	return interactions.Interaction{
		Type:        interactions.TypeModal,
		CustomID:    customID,
		ModalFields: fields,
		Actor:       interactions.Actor{UserID: "user-1", GuildID: "guild-1"},
	}
}
