package announcements

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
	legacySuccessColor       = 0x53FF53
	legacyErrorColor         = 0xED4245
)

func (m Module) ConfigHandler() interactions.Handler {
	return func(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
		if err := responder.Defer(ctx, responses.DeferOptions{}); err != nil {
			return err
		}
		if !interaction.Actor.HasPermission(permissionManageMessages) {
			return responder.EditOriginal(ctx, announcementErrorMessage("你需要有`undefined`才能使用此指令"))
		}
		switch interaction.Subcommand {
		case subcommandOnce:
			return m.handleOnce(ctx, interaction, responder)
		case subcommandBound:
			return m.handleBound(ctx, interaction, responder)
		case subcommandDeleteBound:
			return m.handleDeleteBound(ctx, interaction, responder)
		default:
			return responder.EditOriginal(ctx, announcementErrorMessage("很抱歉，出現了未知的錯誤，請重試!"))
		}
	}
}

func (m Module) handleOnce(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
	channelID := firstOption(interaction, optionChannel)
	created, err := m.service.SetAnnouncementChannel(ctx, domain.AnnouncementChannelConfig{
		GuildID:   interaction.Actor.GuildID,
		ChannelID: channelID,
	})
	if err != nil {
		return responder.EditOriginal(ctx, announcementUnknownError(err))
	}
	if err := responder.EditOriginal(ctx, onceSuccessMessage(channelID, created)); err != nil {
		return err
	}
	return m.track(ctx, interaction)
}

func (m Module) handleBound(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
	channelID := firstOption(interaction, optionChannel)
	color := firstRawOption(interaction, optionColor)
	if !domain.ValidLegacyBoundAnnouncementColor(color) {
		return responder.EditOriginal(ctx, announcementErrorMessage("你傳送的並不是顏色(色碼)"))
	}
	created, err := m.service.SetBoundAnnouncement(ctx, domain.BoundAnnouncementConfig{
		GuildID:   interaction.Actor.GuildID,
		ChannelID: channelID,
		Tag:       firstRawOption(interaction, optionTag),
		Color:     color,
		Title:     firstRawOption(interaction, optionTitle),
	})
	if err != nil {
		return responder.EditOriginal(ctx, announcementUnknownError(err))
	}
	if err := responder.EditOriginal(ctx, boundSuccessMessage(channelID, created)); err != nil {
		return err
	}
	return m.track(ctx, interaction)
}

func (m Module) handleDeleteBound(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
	channelID := firstOption(interaction, optionChannel)
	err := m.service.DeleteBoundAnnouncement(ctx, interaction.Actor.GuildID, channelID)
	if errors.Is(err, ports.ErrBoundAnnouncementConfigMissing) {
		return responder.EditOriginal(ctx, announcementErrorMessage("你沒有對這個頻道設定過綁定型公告!"))
	}
	if err != nil {
		return responder.EditOriginal(ctx, announcementUnknownError(err))
	}
	if err := responder.EditOriginal(ctx, boundDeleteSuccessMessage(channelID)); err != nil {
		return err
	}
	return m.track(ctx, interaction)
}

func onceSuccessMessage(channelID string, created bool) responses.Message {
	action := "更新"
	if created {
		action = "創建"
	}
	return responses.Message{
		Embeds: []responses.Embed{{
			Title:       "<:megaphone:985943890148327454> 公告系統",
			Description: "<:Channel:994524759289233438> **您的公告頻道成功__" + action + "__!!**\n**您目前的公告頻道為**:" + channelMention(channelID),
			Color:       legacySuccessColor,
		}},
		AllowedMentions: &responses.AllowedMentions{},
	}
}

func boundSuccessMessage(channelID string, created bool) responses.Message {
	action := "更新"
	if created {
		action = "創建"
	}
	return responses.Message{
		Embeds: []responses.Embed{{
			Title:       "<:megaphone:985943890148327454> 綁定型公告系統",
			Description: "<:Channel:994524759289233438> **您的綁定型公告頻道成功__" + action + "__!!**\n**新增綁定型公告頻道為**:" + channelMention(channelID),
			Color:       legacySuccessColor,
		}},
		AllowedMentions: &responses.AllowedMentions{},
	}
}

func boundDeleteSuccessMessage(channelID string) responses.Message {
	return responses.Message{
		Embeds: []responses.Embed{{
			Title:       "<:megaphone:985943890148327454> 綁定型公告系統",
			Description: "<:trashbin:995991389043163257> **您的綁定型公告頻道成功__刪除__!!**\n**刪除的綁定型公告頻道為**:" + channelMention(channelID),
			Color:       legacySuccessColor,
		}},
		AllowedMentions: &responses.AllowedMentions{},
	}
}

func announcementUnknownError(err error) responses.Message {
	_ = err
	return announcementErrorMessage("很抱歉，出現了未知的錯誤，請重試!")
}

func announcementErrorMessage(content string) responses.Message {
	return responses.Message{
		Embeds: []responses.Embed{{
			Title: "<a:Discord_AnimatedNo:1015989839809757295> | " + content,
			Color: legacyErrorColor,
		}},
		AllowedMentions: &responses.AllowedMentions{},
	}
}

func channelMention(channelID string) string {
	return "<#" + strings.TrimSpace(channelID) + ">"
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

func (m Module) track(ctx context.Context, interaction interactions.Interaction) error {
	return m.trackCommand(ctx, interaction, ConfigCommandName, "announcement-config")
}

func (m Module) trackCommand(ctx context.Context, interaction interactions.Interaction, commandName string, feature string) error {
	if m.usage == nil {
		return nil
	}
	return m.usage.TrackCommand(ctx, ports.UsageEvent{
		CommandName: commandName,
		UserID:      interaction.Actor.UserID,
		GuildID:     interaction.Actor.GuildID,
		Feature:     feature,
	})
}
