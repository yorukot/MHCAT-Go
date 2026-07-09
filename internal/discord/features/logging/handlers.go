package logging

import (
	"context"
	"errors"
	"strings"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/customid"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/interactions"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/responses"
)

const (
	permissionManageMessages = int64(8192)
	loggingEmbedColor        = 0xFFDC35
	loggingErrorColor        = 0xEA0000
	loggingFooterText        = "MHCAT帶給你最棒的Discord體驗!"
)

var logValueToField = map[string]func(*domain.LoggingConfig){
	"訊息更新":   func(c *domain.LoggingConfig) { c.MessageUpdate = true },
	"訊息刪除":   func(c *domain.LoggingConfig) { c.MessageDelete = true },
	"頻道更新":   func(c *domain.LoggingConfig) { c.ChannelUpdate = true },
	"用戶語音更新": func(c *domain.LoggingConfig) { c.MemberVoiceUpdate = true },
}

func (m Module) ConfigPromptHandler() interactions.Handler {
	return func(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
		if err := responder.Defer(ctx, responses.DeferOptions{}); err != nil {
			return err
		}
		if !interaction.Actor.HasPermission(permissionManageMessages) {
			return responder.EditOriginal(ctx, loggingErrorMessage("你需要有`訊息管理`才能使用此指令"))
		}
		channelID := firstOption(interaction, "channel", "頻道")
		if channelID == "" {
			return responder.EditOriginal(ctx, loggingErrorMessage("很抱歉，出現了未知的錯誤，請重試!"))
		}
		return responder.EditOriginal(ctx, loggingPromptMessage(channelID, nil))
	}
}

func (m Module) ConfigSelectHandler() interactions.Handler {
	return func(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
		channelID := channelIDFromCustomID(interaction.CustomID)
		if channelID == "" {
			return responder.UpdateMessage(ctx, loggingErrorMessage("很抱歉，出現了未知的錯誤，請重試!"))
		}
		if err := responder.DeferUpdate(ctx); err != nil {
			return err
		}
		config := loggingConfigFromValues(interaction.Actor.GuildID, channelID, interaction.Values)
		if err := m.service.Save(ctx, config); err != nil {
			return responder.EditOriginal(ctx, loggingErrorFromError(err))
		}
		if err := responder.EditOriginal(ctx, loggingPromptMessage(channelID, interaction.Values)); err != nil {
			return err
		}
		return m.track(ctx, interaction)
	}
}

func (m Module) LegacyConfigSelectHandler() interactions.Handler {
	return func(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
		return responder.UpdateMessage(ctx, loggingErrorMessage("請重新執行`/set-log-channel`設定日誌頻道"))
	}
}

func loggingConfigFromValues(guildID string, channelID string, values []string) domain.LoggingConfig {
	config := domain.LoggingConfig{GuildID: strings.TrimSpace(guildID), ChannelID: strings.TrimSpace(channelID)}
	for _, value := range values {
		if apply, ok := logValueToField[value]; ok {
			apply(&config)
		}
	}
	return config
}

func loggingPromptMessage(channelID string, selected []string) responses.Message {
	selectedText := ""
	if len(selected) > 0 {
		selectedText = "`" + strings.Join(selected, "`,`") + "`"
	}
	return responses.Message{
		Embeds: []responses.Embed{{
			Title:       "<:logfile:985948561625710663> 日誌系統",
			Description: "**請選擇您需要的日誌(未來會更新更多喔)** \n目前的選擇:" + selectedText,
			Color:       loggingEmbedColor,
			Footer: &responses.EmbedFooter{
				Text: loggingFooterText,
			},
		}},
		Components: []responses.ComponentRow{{
			Components: []responses.Component{{
				Type:        responses.ComponentTypeSelect,
				CustomID:    loggingConfigCustomID(channelID),
				Placeholder: "請選擇您需要的日誌",
				MinValues:   1,
				MaxValues:   4,
				Options:     loggingSelectOptions(selected),
			}},
		}},
		AllowedMentions: &responses.AllowedMentions{},
	}
}

func loggingSelectOptions(selected []string) []responses.SelectOption {
	selectedSet := map[string]struct{}{}
	for _, value := range selected {
		selectedSet[value] = struct{}{}
	}
	options := []responses.SelectOption{
		{Label: "訊息更新", Description: "當訊息編輯時發送日誌", Value: "訊息更新"},
		{Label: "訊息刪除", Description: "當訊息刪除時發送日誌", Value: "訊息刪除"},
		{Label: "頻道更新", Description: "當頻道更新時發送日誌", Value: "頻道更新"},
		{Label: "用戶語音狀態更新", Description: "當用戶離開或加入或是靜音之類的時發送這個通知", Value: "用戶語音更新"},
	}
	for index := range options {
		_, options[index].Default = selectedSet[options[index].Value]
	}
	return options
}

func loggingConfigCustomID(channelID string) string {
	payload, err := customid.KeyValuePayload(map[string]string{"c": strings.TrimSpace(channelID)})
	if err != nil {
		return "mhcat:v1:logging:configure:"
	}
	id, err := customid.Encode(customid.InteractionKindComponent, "logging", "configure", payload)
	if err != nil {
		return "mhcat:v1:logging:configure:"
	}
	return id
}

func channelIDFromCustomID(raw string) string {
	parsed, err := customid.ParseComponent(raw)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(parsed.Payload.Values["c"])
}

func loggingErrorFromError(err error) responses.Message {
	switch {
	case errors.Is(err, domain.ErrInvalidLoggingConfig):
		return loggingErrorMessage("很抱歉，出現了未知的錯誤，請重試!")
	default:
		return loggingErrorMessage("很抱歉，出現了未知的錯誤，請重試!")
	}
}

func loggingErrorMessage(content string) responses.Message {
	return responses.Message{
		Embeds: []responses.Embed{{
			Title: "<a:Discord_AnimatedNo:1015989839809757295> | " + content,
			Color: loggingErrorColor,
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

func (m Module) track(ctx context.Context, interaction interactions.Interaction) error {
	if m.usage == nil {
		return nil
	}
	return m.usage.TrackCommand(ctx, ports.UsageEvent{
		CommandName: LoggingConfigCommandName,
		UserID:      interaction.Actor.UserID,
		GuildID:     interaction.Actor.GuildID,
		Feature:     "logging",
	})
}
