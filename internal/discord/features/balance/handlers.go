package balance

import (
	"context"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/interactions"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/responses"
)

const (
	successColor   = 0x57F287
	successIconURL = "https://media.discordapp.net/attachments/991337796960784424/1078883215462383697/success.gif"
)

func (m Module) Handler() interactions.Handler {
	return func(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
		if err := responder.Defer(ctx, responses.DeferOptions{Ephemeral: true}); err != nil {
			return err
		}
		balance, err := m.service.Get(ctx, interaction.Actor.GuildID)
		if err != nil {
			return err
		}
		if err := responder.EditOriginal(ctx, message(balance.Amount)); err != nil {
			return err
		}
		return m.track(ctx, interaction)
	}
}

func message(amount string) responses.Message {
	return responses.Message{
		Embeds: []responses.Embed{{
			Author: &responses.EmbedAuthor{
				Name:    "伺服器目前剩於餘額: " + amount,
				IconURL: successIconURL,
			},
			Color: successColor,
		}},
		Ephemeral:       true,
		AllowedMentions: &responses.AllowedMentions{},
	}
}
