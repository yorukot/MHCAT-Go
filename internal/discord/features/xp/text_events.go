package xp

import (
	"context"
	"strings"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/events"
)

func (m TextEventModule) MessageCreateHandler() events.Handler {
	return func(ctx context.Context, event events.Event) error {
		if event.Type != events.TypeMessageCreate || event.IsBot {
			return nil
		}
		guildID := strings.TrimSpace(event.GuildID)
		userID := strings.TrimSpace(event.UserID)
		if event.Member != nil {
			if event.Member.IsBot {
				return nil
			}
			if event.Member.UserID != "" {
				userID = strings.TrimSpace(event.Member.UserID)
			}
		}
		if guildID == "" || userID == "" {
			return nil
		}
		_, err := m.service.AccrueMessage(ctx, guildID, userID, event.Content)
		return err
	}
}
