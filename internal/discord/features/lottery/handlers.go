package lottery

import (
	"context"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/interactions"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/responses"
)

const (
	legacyUnavailableTitle = "<a:green_tick:994529015652163614> | 這個指令暫時無法使用造成困擾非常抱歉!"
	legacySuccessColor     = 0x53FF53
)

func (m Module) CreateHandler() interactions.Handler {
	return func(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
		if err := responder.Defer(ctx, responses.DeferOptions{Ephemeral: true}); err != nil {
			return err
		}
		if err := responder.EditOriginal(ctx, unavailableMessage()); err != nil {
			return err
		}
		return m.track(ctx, interaction)
	}
}

func unavailableMessage() responses.Message {
	return responses.Message{
		Embeds: []responses.Embed{{
			Title: legacyUnavailableTitle,
			Color: legacySuccessColor,
		}},
		AllowedMentions: &responses.AllowedMentions{},
		Ephemeral:       true,
	}
}

func (m Module) track(ctx context.Context, interaction interactions.Interaction) error {
	if m.usage == nil {
		return nil
	}
	return m.usage.TrackCommand(ctx, ports.UsageEvent{
		CommandName: LotteryCreateCommandName,
		UserID:      interaction.Actor.UserID,
		GuildID:     interaction.Actor.GuildID,
		Feature:     "lottery-disabled-command",
	})
}
