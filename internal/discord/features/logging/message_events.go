package logging

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/events"
)

const (
	loggingMessageDeleteAuditAction = 72
	loggingDefaultAvatarURL         = "https://i.imgur.com/B91C90T.png"
)

func NewMessageEventModule(repo ports.LoggingConfigReader, messages ports.DiscordMessagePort, auditLogs ports.DiscordAuditLogPort) Module {
	return Module{
		configReader:         repo,
		messages:             messages,
		auditLogs:            auditLogs,
		messageEventsEnabled: repo != nil && messages != nil,
	}
}

func (m Module) RegisterEventRoutes(dispatcher *events.Dispatcher) {
	if dispatcher == nil {
		return
	}
	if m.messageEventsEnabled {
		dispatcher.Register(events.TypeMessageUpdate, m.MessageUpdateHandler())
		dispatcher.Register(events.TypeMessageDelete, m.MessageDeleteHandler())
	}
	if m.channelEventsEnabled {
		dispatcher.Register(events.TypeChannelUpdate, m.ChannelUpdateHandler())
	}
	if m.voiceEventsEnabled {
		dispatcher.Register(events.TypeVoiceState, m.VoiceStateHandler())
	}
}

func (m Module) MessageUpdateHandler() events.Handler {
	return func(ctx context.Context, event events.Event) error {
		if event.Type != events.TypeMessageUpdate || event.IsBot || !event.HasOldContent {
			return nil
		}
		if strings.TrimSpace(event.GuildID) == "" || strings.TrimSpace(event.ChannelID) == "" || strings.TrimSpace(event.UserID) == "" {
			return nil
		}
		if event.OldContent == event.Content {
			return nil
		}
		config, err := m.configReader.GetLoggingConfig(ctx, event.GuildID)
		if err != nil {
			if errors.Is(err, ports.ErrLoggingConfigMissing) {
				return nil
			}
			return err
		}
		if !config.MessageUpdate || strings.TrimSpace(config.ChannelID) == "" {
			return nil
		}
		_, err = m.messages.SendMessage(ctx, config.ChannelID, ports.OutboundMessage{
			Embeds: []ports.OutboundEmbed{{
				AuthorName:    loggingEventUsername(event) + " | 訊息編輯",
				AuthorIconURL: loggingEventAvatarURL(event),
				Color:         0x46A3FF,
				Description:   "**<:edit:1084846013476511765> 訊息編輯者: <@" + event.UserID + "> | <:Channel:994524759289233438> 訊息編輯位置: <#" + event.ChannelID + ">**",
				Fields: []ports.OutboundEmbedField{
					{Name: "**<:book:1084846007545778217> 舊訊息:**", Value: loggingCodeBlock(event.OldContent)},
					{Name: "**<:new:1084846011366785135> 新訊息:**", Value: loggingCodeBlock(event.Content)},
					{Name: "**<:attachment:1084846756799455242> 附件:**", Value: loggingAttachmentText(event.Attachments)},
				},
				FooterText:    loggingFooterText,
				FooterIconURL: event.BotAvatarURL,
				Timestamp:     time.Now(),
			}},
			AllowedMentions: ports.AllowedMentions{},
		})
		return err
	}
}

func (m Module) MessageDeleteHandler() events.Handler {
	return func(ctx context.Context, event events.Event) error {
		if event.Type != events.TypeMessageDelete || event.IsBot {
			return nil
		}
		if strings.TrimSpace(event.GuildID) == "" || strings.TrimSpace(event.ChannelID) == "" || strings.TrimSpace(event.UserID) == "" {
			return nil
		}
		config, err := m.configReader.GetLoggingConfig(ctx, event.GuildID)
		if err != nil {
			if errors.Is(err, ports.ErrLoggingConfigMissing) {
				return nil
			}
			return err
		}
		if !config.MessageDelete || strings.TrimSpace(config.ChannelID) == "" {
			return nil
		}
		deleterID := m.messageDeleteActor(ctx, event)
		_, err = m.messages.SendMessage(ctx, config.ChannelID, ports.OutboundMessage{
			Embeds: []ports.OutboundEmbed{{
				AuthorName:    loggingEventUsername(event) + " | 訊息刪除",
				AuthorIconURL: loggingEventAvatarURL(event),
				Color:         0x84C1FF,
				Description:   "**<:trash:1084846016798396526> 訊息刪除者: <@" + deleterID + "> | <:user:986064391139115028> 訊息發送者:<@" + event.UserID + "> | <:Channel:994524759289233438> 訊息刪除位置: <#" + event.ChannelID + ">**",
				Fields: []ports.OutboundEmbedField{
					{Name: "**<:comments:985944111725019246> 訊息:**", Value: loggingCodeBlock(event.Content)},
					{Name: "**<:attachment:1084846756799455242> 附件:**", Value: loggingAttachmentText(event.Attachments)},
				},
				FooterText:    loggingFooterText,
				FooterIconURL: event.BotAvatarURL,
				Timestamp:     time.Now(),
			}},
			AllowedMentions: ports.AllowedMentions{},
		})
		return err
	}
}

func (m Module) messageDeleteActor(ctx context.Context, event events.Event) string {
	if m.auditLogs == nil {
		return event.UserID
	}
	entries, err := m.auditLogs.AuditLog(ctx, ports.AuditLogQuery{
		GuildID: event.GuildID,
		Action:  loggingMessageDeleteAuditAction,
		Limit:   1,
	})
	if err != nil {
		return event.UserID
	}
	for _, entry := range entries {
		if entry.TargetID != event.UserID {
			continue
		}
		if entry.ChannelID != event.ChannelID {
			continue
		}
		if entry.UserID != "" {
			return entry.UserID
		}
	}
	return event.UserID
}

func loggingEventUsername(event events.Event) string {
	if strings.TrimSpace(event.Username) != "" {
		return strings.TrimSpace(event.Username)
	}
	userTag := strings.TrimSpace(event.UserTag)
	if userTag == "" {
		return event.UserID
	}
	if before, _, ok := strings.Cut(userTag, "#"); ok && before != "" {
		return before
	}
	return userTag
}

func loggingEventAvatarURL(event events.Event) string {
	if strings.TrimSpace(event.AvatarURL) != "" {
		return event.AvatarURL
	}
	return loggingDefaultAvatarURL
}

func loggingCodeBlock(content string) string {
	return "```" + content + " ```"
}

func loggingAttachmentText(attachments []events.Attachment) string {
	urls := []string{}
	for _, attachment := range attachments {
		if strings.TrimSpace(attachment.URL) != "" {
			urls = append(urls, strings.TrimSpace(attachment.URL))
		}
	}
	if len(urls) == 0 {
		return "**沒有附件**"
	}
	return strings.Join(urls, ",")
}
