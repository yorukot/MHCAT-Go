package logging

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/events"
)

func NewVoiceEventModule(repo ports.LoggingConfigReader, messages ports.DiscordMessagePort) Module {
	return Module{
		configReader:       repo,
		messages:           messages,
		voiceEventsEnabled: repo != nil && messages != nil,
	}
}

func (m Module) VoiceStateHandler() events.Handler {
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
		currentChannelID := strings.TrimSpace(voice.ChannelID)
		beforeChannelID := strings.TrimSpace(voice.BeforeChannel)
		joined := beforeChannelID == "" && currentChannelID != ""
		left := beforeChannelID != "" && currentChannelID == ""
		if guildID == "" || userID == "" || (!joined && !left) {
			return nil
		}

		config, err := m.configReader.GetLoggingConfig(ctx, guildID)
		if err != nil {
			if errors.Is(err, ports.ErrLoggingConfigMissing) {
				return nil
			}
			return err
		}
		if !config.MemberVoiceUpdate || strings.TrimSpace(config.ChannelID) == "" {
			return nil
		}

		embed := ports.OutboundEmbed{
			AuthorIconURL: loggingEventAvatarURL(event),
			FooterText:    loggingFooterText,
			FooterIconURL: loggingBotAvatarURL(event.BotAvatarURL),
			Timestamp:     time.Now(),
		}
		if joined {
			embed.AuthorName = loggingEventUsername(event) + " | 使用者加入語音頻道"
			embed.Color = 0xF235FA
			embed.Description = "**<:user:986064391139115028> 使用者: <@" + userID + "> | <:voice:1086216862355951636> 頻道: <#" + currentChannelID + ">**"
			embed.Fields = []ports.OutboundEmbedField{{
				Name:  "**<:joines:1086217186256900098> 加入頻道:**",
				Value: "<#" + currentChannelID + ">(`" + voice.ChannelName + "`)",
			}}
		} else {
			embed.AuthorName = loggingEventUsername(event) + " | 使用者退出語音頻道"
			embed.Color = 0xFA359A
			embed.Description = "**<:user:986064391139115028> 使用者: <@" + userID + "> | <:voice:1086216862355951636> 頻道: <#" + beforeChannelID + ">**"
			embed.Fields = []ports.OutboundEmbedField{{
				Name:  "**<:leaves:1086219523264356513> 退出頻道:**",
				Value: "<#" + beforeChannelID + ">(`" + voice.BeforeChannelName + "`)",
			}}
		}

		_, err = m.messages.SendMessage(ctx, config.ChannelID, ports.OutboundMessage{
			Embeds:          []ports.OutboundEmbed{embed},
			AllowedMentions: ports.AllowedMentions{},
		})
		return err
	}
}
