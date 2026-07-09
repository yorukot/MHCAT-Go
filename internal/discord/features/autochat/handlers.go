package autochat

import (
	"context"
	"errors"
	"strings"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/interactions"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/responses"
)

const (
	permissionManageMessages = int64(8192)
	autoChatSuccessColor     = 0x57F287
	autoChatErrorColor       = 0xED4245
)

func (m Module) SetHandler() interactions.Handler {
	return func(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
		if err := responder.Defer(ctx, responses.DeferOptions{}); err != nil {
			return err
		}
		if !interaction.Actor.HasPermission(permissionManageMessages) {
			return responder.EditOriginal(ctx, autoChatErrorMessage("你需要有`訊息管理`才能使用此指令"))
		}
		channelID := firstOption(interaction, optionChannel)
		config := domain.AutoChatConfig{GuildID: interaction.Actor.GuildID, ChannelID: channelID}
		if err := m.service.Save(ctx, config); err != nil {
			return responder.EditOriginal(ctx, autoChatUnknownError(err))
		}
		if err := responder.EditOriginal(ctx, autoChatSetSuccessMessage(channelID)); err != nil {
			return err
		}
		return m.track(ctx, interaction, AutoChatSetCommandName)
	}
}

func (m Module) DeleteHandler() interactions.Handler {
	return func(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
		if err := responder.Defer(ctx, responses.DeferOptions{}); err != nil {
			return err
		}
		if !interaction.Actor.HasPermission(permissionManageMessages) {
			return responder.EditOriginal(ctx, autoChatErrorMessage("你需要有`訊息管理`才能使用此指令"))
		}
		if err := m.service.Delete(ctx, interaction.Actor.GuildID); err != nil {
			if errors.Is(err, ports.ErrAutoChatConfigMissing) {
				return responder.EditOriginal(ctx, autoChatErrorMessage("你沒有設定過，我不知道要刪除甚麼!"))
			}
			return responder.EditOriginal(ctx, autoChatUnknownError(err))
		}
		if err := responder.EditOriginal(ctx, autoChatDeleteSuccessMessage()); err != nil {
			return err
		}
		return m.track(ctx, interaction, AutoChatDeleteCommandName)
	}
}

func autoChatSetSuccessMessage(channelID string) responses.Message {
	return responses.Message{
		Embeds: []responses.Embed{{
			Title:       "自動聊天系統",
			Description: "您的自動聊天頻道成功創建\n您目前的自動聊天頻道為 <#" + strings.TrimSpace(channelID) + ">",
			Color:       autoChatSuccessColor,
		}},
		AllowedMentions: &responses.AllowedMentions{},
	}
}

func autoChatDeleteSuccessMessage() responses.Message {
	return responses.Message{
		Embeds: []responses.Embed{{
			Title:       "自動聊天系統",
			Description: "您的自動聊天頻道成功刪除",
			Color:       autoChatSuccessColor,
		}},
		AllowedMentions: &responses.AllowedMentions{},
	}
}

func autoChatUnknownError(err error) responses.Message {
	_ = err
	return autoChatErrorMessage("很抱歉，出現了未知的錯誤，請重試!")
}

func autoChatErrorMessage(content string) responses.Message {
	return responses.Message{
		Embeds: []responses.Embed{{
			Title: "<a:Discord_AnimatedNo:1015989839809757295> | " + content,
			Color: autoChatErrorColor,
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

func (m Module) track(ctx context.Context, interaction interactions.Interaction, commandName string) error {
	if m.usage == nil {
		return nil
	}
	return m.usage.TrackCommand(ctx, ports.UsageEvent{
		CommandName: commandName,
		UserID:      interaction.Actor.UserID,
		GuildID:     interaction.Actor.GuildID,
		Feature:     "autochat-config",
	})
}
