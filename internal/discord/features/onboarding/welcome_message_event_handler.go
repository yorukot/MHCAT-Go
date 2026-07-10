package onboarding

import (
	"context"
	"strings"

	coreservice "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/services/onboarding"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/events"
)

func (m Module) WelcomeMessageDeliveryHandler() events.Handler {
	return func(ctx context.Context, event events.Event) error {
		if event.Type != events.TypeMemberAdd {
			return nil
		}
		member := event.Member
		userID := strings.TrimSpace(event.UserID)
		username := event.Username
		userTag := strings.TrimSpace(event.UserTag)
		avatarURL := strings.TrimSpace(event.AvatarURL)
		if member != nil {
			if member.UserID != "" {
				userID = member.UserID
			}
			if member.Username != "" {
				username = member.Username
			}
			if member.UserTag != "" {
				userTag = member.UserTag
			}
			if member.AvatarURL != "" {
				avatarURL = member.AvatarURL
			}
		}
		if strings.TrimSpace(event.GuildID) == "" || userID == "" {
			return nil
		}
		return m.welcomeService.SendOnJoin(ctx, coreservice.WelcomeMemberEvent{
			GuildID:      event.GuildID,
			GuildName:    event.GuildName,
			GuildIconURL: event.GuildIconURL,
			BotUserID:    event.BotUserID,
			BotAvatarURL: event.BotAvatarURL,
			UserID:       userID,
			Username:     username,
			UserTag:      userTag,
			AvatarURL:    avatarURL,
			Now:          event.CreatedAt,
		})
	}
}
