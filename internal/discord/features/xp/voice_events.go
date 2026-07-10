package xp

import (
	"context"
	"strings"

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
			return m.service.Leave(ctx, guildID, userID)
		}
		return m.service.Join(ctx, guildID, userID)
	}
}
