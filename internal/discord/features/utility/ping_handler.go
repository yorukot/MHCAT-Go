package utility

import (
	"context"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/interactions"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/responses"
)

func (m Module) PingHandler() interactions.Handler {
	return func(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := responder.Reply(ctx, responses.Message{Content: m.ping.Response(interaction.CreatedAt)}); err != nil {
			return err
		}
		return m.track(ctx, interaction, "ping")
	}
}
