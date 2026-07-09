package discordgo

import (
	"fmt"

	dgo "github.com/bwmarrin/discordgo"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/interactions"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/responses"
	discordruntime "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/runtime"
)

func (s *Session) RuntimeInteraction(event *dgo.InteractionCreate) (interactions.Interaction, responses.Responder, error) {
	if s == nil || s.session == nil {
		return interactions.Interaction{}, nil, fmt.Errorf("%w: discord session is nil", discordruntime.ErrRuntimeNotConfigured)
	}
	if event == nil || event.Interaction == nil {
		return interactions.Interaction{}, nil, fmt.Errorf("%w: interaction event is nil", discordruntime.ErrInvalidRuntimeEvent)
	}
	interaction, err := InteractionFromDiscord(event)
	if err != nil {
		return interactions.Interaction{}, nil, err
	}
	return interaction, NewInteractionResponder(s.session, event.Interaction), nil
}
