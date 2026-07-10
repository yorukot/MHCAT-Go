package xp

import (
	"context"
	"errors"
	"strings"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	coreservice "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/services/xp"
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
		result, err := m.service.AccrueMessage(ctx, guildID, userID, event.Content)
		if err != nil {
			return err
		}
		if !result.Leveled {
			return ctx.Err()
		}
		var firstErr error
		if err := m.sendLevelUpAnnouncement(ctx, event, userID, result.Profile.Level); err != nil {
			firstErr = err
		}
		if err := m.applyRewardRoles(ctx, guildID, userID, result.Profile.Level, event.Member); err != nil && firstErr == nil {
			firstErr = err
		}
		if firstErr != nil {
			return firstErr
		}
		return ctx.Err()
	}
}

func (m TextEventModule) sendLevelUpAnnouncement(ctx context.Context, event events.Event, userID string, level int64) error {
	if m.configs == nil || m.messages == nil {
		return ctx.Err()
	}
	config, err := m.configs.GetTextXPConfig(ctx, event.GuildID)
	if errors.Is(err, ports.ErrTextXPConfigMissing) {
		return ctx.Err()
	}
	if err != nil {
		return err
	}
	channelID := strings.TrimSpace(config.ChannelID)
	if channelID == coreservice.LegacyTextXPAnnouncementCurrentChannel {
		channelID = strings.TrimSpace(event.ChannelID)
	}
	if channelID == "" {
		return ctx.Err()
	}
	content := coreservice.LegacyTextXPLevelUpAnnouncement(config.Message, level, userID)
	_, err = m.messages.SendMessage(ctx, channelID, ports.OutboundMessage{
		Content:         content,
		AllowedMentions: ports.AllowedMentions{UserIDs: []string{strings.TrimSpace(userID)}},
	})
	if err != nil {
		return err
	}
	return ctx.Err()
}

func (m TextEventModule) applyRewardRoles(ctx context.Context, guildID string, userID string, level int64, member *events.Member) error {
	if m.rewardRoles.Repository == nil || m.rewardRoles.RolePort == nil {
		return ctx.Err()
	}
	var currentRoleIDs []string
	if member != nil {
		currentRoleIDs = member.RoleIDs
	}
	return m.rewardRoles.ApplyLevelUp(ctx, guildID, userID, level, currentRoleIDs)
}
