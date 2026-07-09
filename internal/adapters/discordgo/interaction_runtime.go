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
	s.populateActorVoiceState(&interaction)
	return interaction, NewInteractionResponder(s.session, event.Interaction), nil
}

func (s *Session) populateActorVoiceState(interaction *interactions.Interaction) {
	if interaction == nil || s == nil || s.session == nil || s.session.State == nil {
		return
	}
	if interaction.Actor.GuildID == "" || interaction.Actor.UserID == "" {
		return
	}
	voice, err := s.session.State.VoiceState(interaction.Actor.GuildID, interaction.Actor.UserID)
	if err != nil || voice == nil {
		return
	}
	interaction.Actor.VoiceChannelID = voice.ChannelID
}
