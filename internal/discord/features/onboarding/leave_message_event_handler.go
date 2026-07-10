package onboarding

import (
	"context"
	"strings"

	coreservice "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/services/onboarding"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/events"
)

func (m Module) LeaveMessageDeliveryHandler() events.Handler {
	return func(ctx context.Context, event events.Event) error {
		if event.Type != events.TypeMemberRemove {
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
		err := m.deliveryService.SendOnLeave(ctx, coreservice.LeaveMemberEvent{
			GuildID:   event.GuildID,
			UserID:    userID,
			Username:  username,
			UserTag:   userTag,
			AvatarURL: avatarURL,
			Now:       event.CreatedAt,
		})
		if err != nil && ctx.Err() == nil {
			return events.ContinueOnError(err)
		}
		return err
	}
}

func usernameFromTag(tag string) string {
	tag = strings.TrimSpace(tag)
	if tag == "" {
		return ""
	}
	if index := strings.Index(tag, "#"); index > 0 {
		return tag[:index]
	}
	return tag
}
