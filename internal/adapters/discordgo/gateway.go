package discordgo

import (
	"context"
	"log/slog"

	dgo "github.com/bwmarrin/discordgo"
	discordruntime "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/runtime"
)

func (s *Session) RegisterInteractionHandler(handler discordruntime.Handler) func() {
	if s == nil || s.session == nil || handler == nil {
		return func() {}
	}
	return s.session.AddHandler(func(_ *dgo.Session, event *dgo.InteractionCreate) {
		ctx := context.Background()
		interaction, responder, err := s.RuntimeInteraction(event)
		if err != nil {
			slog.Default().WarnContext(ctx, "drop invalid discord interaction", "error", err.Error())
			return
		}
		_ = handler(ctx, interaction, responder)
	})
}
