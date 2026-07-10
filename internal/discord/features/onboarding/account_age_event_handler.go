package onboarding

import (
	"context"
	"strings"
	"time"

	coreservice "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/services/onboarding"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/events"
)

func (m Module) AccountAgeGateHandler() events.Handler {
	return func(ctx context.Context, event events.Event) error {
		if event.Type != events.TypeMemberAdd {
			return nil
		}
		member := event.Member
		userID := strings.TrimSpace(event.UserID)
		userTag := strings.TrimSpace(event.UserTag)
		avatarURL := strings.TrimSpace(event.AvatarURL)
		var accountCreatedAt time.Time
		if member != nil {
			if member.UserID != "" {
				userID = member.UserID
			}
			if member.UserTag != "" {
				userTag = member.UserTag
			}
			if member.AvatarURL != "" {
				avatarURL = member.AvatarURL
			}
			if !member.AccountCreatedAt.IsZero() {
				accountCreatedAt = member.AccountCreatedAt
			}
		}
		if strings.TrimSpace(event.GuildID) == "" || userID == "" {
			return nil
		}
		result, err := m.accountAgePolicy.GateMemberAdd(ctx, coreservice.AccountAgeMemberEvent{
			GuildID:          event.GuildID,
			GuildName:        event.GuildName,
			UserID:           userID,
			UserTag:          userTag,
			AvatarURL:        avatarURL,
			AccountCreatedAt: accountCreatedAt,
		})
		if err != nil {
			if !result.Matched && ctx.Err() == nil {
				return events.ContinueOnError(err)
			}
			return err
		}
		if result.Matched {
			return events.ErrStopPropagation
		}
		return nil
	}
}
