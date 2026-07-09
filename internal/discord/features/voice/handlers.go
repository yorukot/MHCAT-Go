package voice

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/interactions"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/responses"
)

const (
	permissionManageMessages = int64(8192)
	voiceErrorColor          = 0xED4245
	voiceSuccessColor        = 0x57F287
	legacyVoiceEmoji         = "<:Voice:994844272790610011>"
	legacyDoneEmoji          = "<a:green_tick:994529015652163614>"
	legacyDeleteEmoji        = "<:trashbin:995991389043163257>"
	discordChannelTypeVoice  = 2
)

func (m Module) SetHandler() interactions.Handler {
	return func(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
		if err := responder.Defer(ctx, responses.DeferOptions{}); err != nil {
			return err
		}
		if !interaction.Actor.HasPermission(permissionManageMessages) {
			return responder.EditOriginal(ctx, voiceErrorMessage("你需要有`訊息管理`才能使用此指令"))
		}
		trigger := firstCommandOption(interaction, optionTriggerChannel)
		triggerID := strings.TrimSpace(trigger.String)
		limit, ok := limitOption(interaction, optionUserLimit)
		if ok && (limit < 1 || limit > 99) {
			return responder.EditOriginal(ctx, voiceErrorMessage("必須為1-99的整數!"))
		}
		config := domain.VoiceRoomConfig{
			GuildID:          interaction.Actor.GuildID,
			TriggerChannelID: triggerID,
			ParentID:         trigger.ChannelParentID,
			Name:             firstOption(interaction, optionRoomName),
			Limit:            limit,
			Lock:             boolOption(interaction, optionOwnerLock),
		}
		if err := m.service.Save(ctx, config); err != nil {
			return responder.EditOriginal(ctx, voiceUnknownError(err))
		}
		if err := responder.EditOriginal(ctx, voiceSetSuccessMessage()); err != nil {
			return err
		}
		return m.track(ctx, interaction, VoiceRoomSetCommandName)
	}
}

func (m Module) DeleteHandler() interactions.Handler {
	return func(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
		if err := responder.Defer(ctx, responses.DeferOptions{}); err != nil {
			return err
		}
		if !interaction.Actor.HasPermission(permissionManageMessages) {
			return responder.EditOriginal(ctx, voiceErrorMessage("你需要有`訊息管理`才能使用此指令"))
		}
		selected := firstCommandOption(interaction, optionChannelOrGroup)
		channelID := strings.TrimSpace(selected.String)
		var err error
		if selected.ChannelType == discordChannelTypeVoice {
			err = m.service.DeleteByTrigger(ctx, interaction.Actor.GuildID, channelID)
			if err != nil {
				if errors.Is(err, ports.ErrVoiceRoomConfigMissing) {
					return responder.EditOriginal(ctx, voiceErrorMessage("你沒有對這個頻道做出設定過喔!"))
				}
				return responder.EditOriginal(ctx, voiceUnknownError(err))
			}
			if err := responder.EditOriginal(ctx, voiceDeleteTriggerSuccessMessage()); err != nil {
				return err
			}
		} else {
			err = m.service.DeleteByParent(ctx, interaction.Actor.GuildID, channelID)
			if err != nil {
				if errors.Is(err, ports.ErrVoiceRoomConfigMissing) {
					return responder.EditOriginal(ctx, voiceErrorMessage("你沒有對這個類別沒有設定喔!"))
				}
				return responder.EditOriginal(ctx, voiceUnknownError(err))
			}
			if err := responder.EditOriginal(ctx, voiceDeleteParentSuccessMessage()); err != nil {
				return err
			}
		}
		return m.track(ctx, interaction, VoiceRoomDeleteCommandName)
	}
}

func voiceSetSuccessMessage() responses.Message {
	return responses.Message{
		Embeds: []responses.Embed{{
			Title:       legacyDoneEmoji + " | 成功進行設定",
			Description: legacyVoiceEmoji + " 你成功對語音包廂進行`設定`",
			Color:       voiceSuccessColor,
		}},
		AllowedMentions: &responses.AllowedMentions{},
	}
}

func voiceDeleteTriggerSuccessMessage() responses.Message {
	return responses.Message{
		Embeds: []responses.Embed{{
			Title:       legacyDoneEmoji + "成功進行刪除",
			Description: legacyDeleteEmoji + "你成功對這個設定刪除",
			Color:       voiceSuccessColor,
		}},
		AllowedMentions: &responses.AllowedMentions{},
	}
}

func voiceDeleteParentSuccessMessage() responses.Message {
	return responses.Message{
		Embeds: []responses.Embed{{
			Title:       "成功進行刪除",
			Description: "你成功對這個設定刪除",
			Color:       voiceSuccessColor,
		}},
		AllowedMentions: &responses.AllowedMentions{},
	}
}

func voiceUnknownError(err error) responses.Message {
	_ = err
	return voiceErrorMessage("很抱歉，出現了未知的錯誤，請重試!")
}

func voiceErrorMessage(content string) responses.Message {
	return responses.Message{
		Embeds: []responses.Embed{{
			Title: "<a:Discord_AnimatedNo:1015989839809757295> | " + content,
			Color: voiceErrorColor,
		}},
		AllowedMentions: &responses.AllowedMentions{},
	}
}

func firstCommandOption(interaction interactions.Interaction, name string) interactions.CommandOptionValue {
	if value, ok := interaction.CommandOptions[name]; ok {
		if strings.TrimSpace(value.String) != "" {
			return value
		}
	}
	return interactions.CommandOptionValue{
		Type:   interactions.CommandOptionChannel,
		String: strings.TrimSpace(interaction.Options[name]),
	}
}

func firstOption(interaction interactions.Interaction, name string) string {
	if value := strings.TrimSpace(interaction.Options[name]); value != "" {
		return value
	}
	if option, ok := interaction.CommandOptions[name]; ok {
		return strings.TrimSpace(option.String)
	}
	return ""
}

func boolOption(interaction interactions.Interaction, name string) bool {
	if option, ok := interaction.CommandOptions[name]; ok {
		return option.Bool
	}
	return strings.EqualFold(strings.TrimSpace(interaction.Options[name]), "true")
}

func limitOption(interaction interactions.Interaction, name string) (int, bool) {
	if option, ok := interaction.CommandOptions[name]; ok {
		return int(option.Int), true
	}
	raw := strings.TrimSpace(interaction.Options[name])
	if raw == "" {
		return 0, false
	}
	parsed, err := strconv.Atoi(raw)
	if err != nil {
		return 0, true
	}
	return parsed, true
}

func (m Module) track(ctx context.Context, interaction interactions.Interaction, commandName string) error {
	if m.usage == nil {
		return nil
	}
	return m.usage.TrackCommand(ctx, ports.UsageEvent{
		CommandName: commandName,
		UserID:      interaction.Actor.UserID,
		GuildID:     interaction.Actor.GuildID,
		Feature:     "voice-room-config",
	})
}
