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
		announcementSent, err := m.sendLevelUpAnnouncement(ctx, event, userID, result.Profile.Level)
		if err != nil {
			firstErr = err
		}
		if err := m.applyRewardRoles(ctx, guildID, userID, result.Profile.Level, event.Member); err != nil && firstErr == nil {
			firstErr = err
		}
		if announcementSent {
			err = m.applyCoinReward(ctx, guildID, userID, result.Profile.Level)
		}
		if err != nil && firstErr == nil {
			firstErr = err
		}
		if firstErr != nil {
			return firstErr
		}
		return ctx.Err()
	}
}

func (m TextEventModule) sendLevelUpAnnouncement(ctx context.Context, event events.Event, userID string, level int64) (bool, error) {
	if m.configs == nil || m.messages == nil {
		return false, ctx.Err()
	}
	config, err := m.configs.GetTextXPConfig(ctx, event.GuildID)
	if errors.Is(err, ports.ErrTextXPConfigMissing) {
		return false, ctx.Err()
	}
	if err != nil {
		return false, err
	}
	channelID := strings.TrimSpace(config.ChannelID)
	if channelID == coreservice.LegacyTextXPAnnouncementCurrentChannel {
		channelID = strings.TrimSpace(event.ChannelID)
	}
	if channelID == "" {
		return false, ctx.Err()
	}
	channelName := channelID
	if m.channels != nil {
		channel, err := m.channels.FindChannelByID(ctx, event.GuildID, channelID)
		if errors.Is(err, ports.ErrChannelNotFound) {
			return false, m.sendMissingLevelChannelFallback(ctx, event)
		}
		if err != nil {
			return false, err
		}
		if strings.TrimSpace(channel.Name) != "" {
			channelName = strings.TrimSpace(channel.Name)
		}
	}
	content := coreservice.LegacyTextXPLevelUpAnnouncement(config.Message, level, userID)
	_, err = m.messages.SendMessage(ctx, channelID, ports.OutboundMessage{
		Content:         content,
		AllowedMentions: ports.AllowedMentions{UserIDs: []string{strings.TrimSpace(userID)}},
	})
	if err != nil {
		if m.direct != nil {
			_, dmErr := m.direct.SendDirectMessage(ctx, userID, ports.OutboundMessage{
				Content:         "你升級了，但是我沒有權限在" + channelName + "發送消息!",
				AllowedMentions: ports.AllowedMentions{},
			})
			return false, dmErr
		}
		return false, err
	}
	return true, ctx.Err()
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

func (m TextEventModule) applyCoinReward(ctx context.Context, guildID string, userID string, level int64) error {
	if m.coinRewards.Repository == nil {
		return ctx.Err()
	}
	_, err := m.coinRewards.ApplyLevelUp(ctx, guildID, userID, level)
	return err
}

func (m TextEventModule) sendMissingLevelChannelFallback(ctx context.Context, event events.Event) error {
	if strings.TrimSpace(event.ChannelID) == "" {
		return ctx.Err()
	}
	_, err := m.messages.SendMessage(ctx, event.ChannelID, ports.OutboundMessage{
		Embeds: []ports.OutboundEmbed{{
			Title: "<a:error:980086028113182730> | 群組的升等頻道被刪除了，請重新設定升等消息!",
			Color: textXPErrorColor,
		}},
		AllowedMentions: ports.AllowedMentions{},
	})
	return err
}
