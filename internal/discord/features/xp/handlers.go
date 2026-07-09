package xp

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/interactions"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/responses"
)

const (
	permissionManageMessages = int64(8192)
	textXPErrorColor         = 0xED4245
	textXPSuccessColor       = 0x57F287
	legacyLineEmoji          = "<:line:992363971803881493>"
)

func (m Module) SetHandler() interactions.Handler {
	return func(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
		if err := responder.Defer(ctx, responses.DeferOptions{}); err != nil {
			return err
		}
		if !interaction.Actor.HasPermission(permissionManageMessages) {
			return responder.EditOriginal(ctx, textXPErrorMessage("你需要有`訊息管理`才能使用此指令"))
		}
		channelID := firstOption(interaction, "頻道")
		color := firstOption(interaction, "顏色")
		message := firstRawOption(interaction, "訊息")
		if strings.TrimSpace(color) != "" && !domain.ValidLegacyColor(color) {
			return responder.EditOriginal(ctx, textXPErrorMessage("你傳送的並不是顏色(色碼)"))
		}
		config := domain.TextXPConfig{
			GuildID:   interaction.Actor.GuildID,
			ChannelID: channelID,
			Color:     color,
			Message:   message,
		}
		if err := m.service.Save(ctx, config); err != nil {
			return responder.EditOriginal(ctx, textXPUnknownError(err))
		}
		if err := responder.EditOriginal(ctx, textXPSuccessMessage(channelID)); err != nil {
			return err
		}
		if message != "" && m.messages != nil && strings.TrimSpace(interaction.ChannelID) != "" {
			_, _ = m.messages.SendMessage(ctx, interaction.ChannelID, ports.OutboundMessage{
				Content:         textXPPreviewContent(message),
				AllowedMentions: ports.AllowedMentions{},
			})
		}
		return m.track(ctx, interaction, TextXPSetCommandName)
	}
}

func (m Module) DeleteHandler() interactions.Handler {
	return func(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
		if err := responder.Defer(ctx, responses.DeferOptions{}); err != nil {
			return err
		}
		if !interaction.Actor.HasPermission(permissionManageMessages) {
			return responder.EditOriginal(ctx, textXPErrorMessage("你需要有`訊息管理`才能使用此指令"))
		}
		if err := m.service.Delete(ctx, interaction.Actor.GuildID); err != nil {
			if errors.Is(err, ports.ErrTextXPConfigMissing) {
				return responder.EditOriginal(ctx, textXPErrorMessage("你本來就沒有對聊天經驗設定喔!"))
			}
			return responder.EditOriginal(ctx, textXPUnknownError(err))
		}
		if err := responder.EditOriginal(ctx, textXPDeleteSuccessMessage()); err != nil {
			return err
		}
		return m.track(ctx, interaction, TextXPDeleteCommandName)
	}
}

func (m VoiceModule) SetHandler() interactions.Handler {
	return func(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
		if err := responder.Defer(ctx, responses.DeferOptions{}); err != nil {
			return err
		}
		if !interaction.Actor.HasPermission(permissionManageMessages) {
			return responder.EditOriginal(ctx, textXPErrorMessage("你需要有`訊息管理`才能使用此指令"))
		}
		channelID := firstOption(interaction, "頻道")
		color := firstOption(interaction, "顏色")
		message := firstRawOption(interaction, "訊息")
		if strings.TrimSpace(color) != "" && !domain.ValidLegacyColor(color) {
			return responder.EditOriginal(ctx, textXPErrorMessage("你傳送的並不是顏色(色碼)"))
		}
		config := domain.VoiceXPConfig{
			GuildID:   interaction.Actor.GuildID,
			ChannelID: channelID,
			Color:     color,
			Message:   message,
		}
		if err := m.service.Save(ctx, config); err != nil {
			return responder.EditOriginal(ctx, textXPUnknownError(err))
		}
		if err := responder.EditOriginal(ctx, voiceXPSuccessMessage(channelID)); err != nil {
			return err
		}
		if message != "" && m.messages != nil && strings.TrimSpace(interaction.ChannelID) != "" {
			_, _ = m.messages.SendMessage(ctx, interaction.ChannelID, ports.OutboundMessage{
				Content:         textXPPreviewContent(message),
				AllowedMentions: ports.AllowedMentions{},
			})
		}
		return m.track(ctx, interaction, VoiceXPSetCommandName)
	}
}

func (m VoiceModule) DeleteHandler() interactions.Handler {
	return func(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
		if err := responder.Defer(ctx, responses.DeferOptions{}); err != nil {
			return err
		}
		if !interaction.Actor.HasPermission(permissionManageMessages) {
			return responder.EditOriginal(ctx, textXPErrorMessage("你需要有`訊息管理`才能使用此指令"))
		}
		if err := m.service.Delete(ctx, interaction.Actor.GuildID); err != nil {
			if errors.Is(err, ports.ErrVoiceXPConfigMissing) {
				return responder.EditOriginal(ctx, textXPErrorMessage("你本來就沒有對語音經驗設定喔!"))
			}
			return responder.EditOriginal(ctx, textXPUnknownError(err))
		}
		if err := responder.EditOriginal(ctx, voiceXPDeleteSuccessMessage()); err != nil {
			return err
		}
		return m.track(ctx, interaction, VoiceXPDeleteCommandName)
	}
}

func textXPSuccessMessage(channelID string) responses.Message {
	return responses.Message{
		Embeds: []responses.Embed{{
			Title:       "聊天經驗系統",
			Description: fmt.Sprintf("您的聊天經驗升等頻道成功創建\n您目前的升等通知頻道為 <#%s>", strings.TrimSpace(channelID)),
			Color:       textXPSuccessColor,
		}},
		AllowedMentions: &responses.AllowedMentions{},
	}
}

func voiceXPSuccessMessage(channelID string) responses.Message {
	return responses.Message{
		Embeds: []responses.Embed{{
			Title:       "語音經驗系統",
			Description: fmt.Sprintf("您的語音經驗升等頻道成功創建\n您目前的升等通知頻道為 <#%s>", strings.TrimSpace(channelID)),
			Color:       textXPSuccessColor,
		}},
		AllowedMentions: &responses.AllowedMentions{},
	}
}

func textXPDeleteSuccessMessage() responses.Message {
	return responses.Message{
		Embeds: []responses.Embed{{
			Title:       "聊天經驗系統",
			Description: "成功刪除!",
			Color:       textXPSuccessColor,
		}},
		AllowedMentions: &responses.AllowedMentions{},
	}
}

func voiceXPDeleteSuccessMessage() responses.Message {
	return responses.Message{
		Embeds: []responses.Embed{{
			Title:       "語音經驗系統",
			Description: "成功刪除!",
			Color:       textXPSuccessColor,
		}},
		AllowedMentions: &responses.AllowedMentions{},
	}
}

func textXPPreviewContent(message string) string {
	line := legacyLineEmoji + "我" + legacyLineEmoji + "只" + legacyLineEmoji + "是" + legacyLineEmoji + "分" + legacyLineEmoji + "隔" + legacyLineEmoji + "線" + legacyLineEmoji
	return "以下為你的訊息預覽:\n" + line + "\n\n" + message
}

func textXPUnknownError(err error) responses.Message {
	_ = err
	return textXPErrorMessage("很抱歉，出現了未知的錯誤，請重試!")
}

func textXPErrorMessage(content string) responses.Message {
	return responses.Message{
		Embeds: []responses.Embed{{
			Title: "<a:Discord_AnimatedNo:1015989839809757295> | " + content,
			Color: textXPErrorColor,
		}},
		AllowedMentions: &responses.AllowedMentions{},
	}
}

func firstOption(interaction interactions.Interaction, names ...string) string {
	for _, name := range names {
		if value := strings.TrimSpace(interaction.Options[name]); value != "" {
			return value
		}
		if option, ok := interaction.CommandOptions[name]; ok {
			if value := strings.TrimSpace(option.String); value != "" {
				return value
			}
		}
	}
	return ""
}

func firstRawOption(interaction interactions.Interaction, names ...string) string {
	for _, name := range names {
		if value, ok := interaction.Options[name]; ok && value != "" {
			return value
		}
		if option, ok := interaction.CommandOptions[name]; ok && option.String != "" {
			return option.String
		}
	}
	return ""
}

func (m Module) track(ctx context.Context, interaction interactions.Interaction, commandName string) error {
	if m.usage == nil {
		return nil
	}
	return m.usage.TrackCommand(ctx, ports.UsageEvent{
		CommandName: commandName,
		UserID:      interaction.Actor.UserID,
		GuildID:     interaction.Actor.GuildID,
		Feature:     "text-xp-config",
	})
}

func (m VoiceModule) track(ctx context.Context, interaction interactions.Interaction, commandName string) error {
	if m.usage == nil {
		return nil
	}
	return m.usage.TrackCommand(ctx, ports.UsageEvent{
		CommandName: commandName,
		UserID:      interaction.Actor.UserID,
		GuildID:     interaction.Actor.GuildID,
		Feature:     "voice-xp-config",
	})
}
