package xp

import (
	"context"
	"errors"
	"strings"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	coreservice "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/services/xp"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/events"
)

func (m VoiceEventModule) VoiceStateHandler() events.Handler {
	return func(ctx context.Context, event events.Event) error {
		if event.Type != events.TypeVoiceState || event.VoiceState == nil {
			return nil
		}
		voice := event.VoiceState
		guildID := strings.TrimSpace(voice.GuildID)
		if guildID == "" {
			guildID = strings.TrimSpace(event.GuildID)
		}
		userID := strings.TrimSpace(voice.UserID)
		if userID == "" {
			userID = strings.TrimSpace(event.UserID)
		}
		isBot := event.IsBot
		if event.Member != nil {
			isBot = event.Member.IsBot
			if event.Member.UserID != "" {
				userID = strings.TrimSpace(event.Member.UserID)
			}
		}
		channelID := strings.TrimSpace(voice.ChannelID)
		beforeChannelID := strings.TrimSpace(voice.BeforeChannel)
		if guildID == "" || userID == "" || isBot || channelID == beforeChannelID {
			return nil
		}
		if channelID == "" {
			if err := m.service.Leave(ctx, guildID, userID); err != nil {
				return err
			}
			if m.worker != nil {
				m.worker.Stop(guildID, userID)
			}
			return ctx.Err()
		}
		if err := m.service.Join(ctx, guildID, userID); err != nil {
			return err
		}
		if m.worker != nil {
			m.worker.Start(guildID, userID, voiceEventRoleIDs(event.Member))
		}
		return ctx.Err()
	}
}

func voiceEventRoleIDs(member *events.Member) []string {
	if member == nil {
		return nil
	}
	return member.RoleIDs
}

func (m VoiceEventModule) TickVoiceXP(ctx context.Context, guildID string, userID string, currentRoleIDs []string) (coreservice.VoiceAccrualResult, error) {
	result, err := m.accrual.Tick(ctx, guildID, userID)
	if err != nil {
		return result, err
	}
	if !result.Active || !result.Leveled {
		return result, ctx.Err()
	}
	var firstErr error
	announcementSent, err := m.sendLevelUpAnnouncement(ctx, result.Profile.GuildID, result.Profile.UserID, result.Profile.Level)
	if err != nil {
		firstErr = err
	}
	if err := m.applyRewardRoles(ctx, result.Profile.GuildID, result.Profile.UserID, result.Profile.Level, currentRoleIDs); err != nil && firstErr == nil {
		firstErr = err
	}
	if announcementSent {
		err = m.applyCoinReward(ctx, result.Profile.GuildID, result.Profile.UserID, result.Profile.Level)
	}
	if err != nil && firstErr == nil {
		firstErr = err
	}
	if firstErr != nil {
		return result, firstErr
	}
	return result, ctx.Err()
}

func (m VoiceEventModule) sendLevelUpAnnouncement(ctx context.Context, guildID string, userID string, level int64) (bool, error) {
	if m.configs == nil || m.messages == nil {
		return false, ctx.Err()
	}
	config, err := m.configs.GetVoiceXPConfig(ctx, guildID)
	if errors.Is(err, ports.ErrVoiceXPConfigMissing) {
		return false, ctx.Err()
	}
	if err != nil {
		return false, err
	}
	channelID := strings.TrimSpace(config.ChannelID)
	if channelID == "" {
		return false, ctx.Err()
	}
	channelName := channelID
	if m.channels != nil {
		channel, err := m.channels.FindChannelByID(ctx, guildID, channelID)
		if errors.Is(err, ports.ErrChannelNotFound) {
			return false, m.sendVoiceOwnerFallback(ctx, guildID, ":x: 有人的語音頻道等級升級了，但升等頻道已經被刪除了!")
		}
		if err != nil {
			return false, err
		}
		if strings.TrimSpace(channel.Name) != "" {
			channelName = strings.TrimSpace(channel.Name)
		}
	}
	content := coreservice.LegacyVoiceXPLevelUpAnnouncement(config.Message, level, userID)
	_, err = m.messages.SendMessage(ctx, channelID, ports.OutboundMessage{
		Content:         content,
		AllowedMentions: ports.AllowedMentions{UserIDs: []string{strings.TrimSpace(userID)}},
	})
	if err != nil {
		if m.direct != nil && m.guilds != nil {
			return false, m.sendVoiceOwnerFallback(ctx, guildID, ":x: 有人的語音頻道等級升級了，但是我沒有權限在"+channelName+"發送消息!\n因為你是該伺服器擁有者，所以我找你報告: P")
		}
		return false, err
	}
	return true, ctx.Err()
}

func (m VoiceEventModule) applyRewardRoles(ctx context.Context, guildID string, userID string, level int64, currentRoleIDs []string) error {
	if m.rewardRoles.Repository == nil || m.rewardRoles.RolePort == nil {
		return ctx.Err()
	}
	return m.rewardRoles.ApplyLevelUp(ctx, guildID, userID, level, currentRoleIDs)
}

func (m VoiceEventModule) applyCoinReward(ctx context.Context, guildID string, userID string, level int64) error {
	if m.coinRewards.Repository == nil {
		return ctx.Err()
	}
	_, err := m.coinRewards.ApplyLevelUp(ctx, guildID, userID, level)
	return err
}

func (m VoiceEventModule) sendVoiceOwnerFallback(ctx context.Context, guildID string, content string) error {
	if m.direct == nil || m.guilds == nil {
		return ctx.Err()
	}
	guild, err := m.guilds.GuildInfo(ctx, strings.TrimSpace(guildID))
	if err != nil {
		return err
	}
	ownerID := strings.TrimSpace(guild.OwnerID)
	if ownerID == "" {
		return ctx.Err()
	}
	_, err = m.direct.SendDirectMessage(ctx, ownerID, ports.OutboundMessage{
		Content:         content,
		AllowedMentions: ports.AllowedMentions{},
	})
	return err
}
