package notifications

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
	permissionManageMessages    = int64(8192)
	autoNotificationErrorColor  = 0xED4245
	autoNotificationGreenColor  = 0x57F287
	autoNotificationFallbackGld = "這個伺服器"
)

func (m Module) ListHandler() interactions.Handler {
	return func(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
		if !interaction.Actor.HasPermission(permissionManageMessages) {
			return responder.Reply(ctx, autoNotificationErrorMessage("你需要有`訊息管理`才能使用此指令"))
		}
		schedules, err := m.service.List(ctx, interaction.Actor.GuildID)
		if err != nil {
			return responder.Reply(ctx, autoNotificationErrorFromError(err))
		}
		guildName := m.guildName(ctx, interaction.Actor.GuildID)
		if err := responder.Reply(ctx, autoNotificationListMessage(guildName, schedules, m.color())); err != nil {
			return err
		}
		return m.track(ctx, interaction, AutoNotificationListCommandName)
	}
}

func (m Module) DeleteHandler() interactions.Handler {
	return func(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
		if err := responder.Defer(ctx, responses.DeferOptions{}); err != nil {
			return err
		}
		if !interaction.Actor.HasPermission(permissionManageMessages) {
			return responder.EditOriginal(ctx, autoNotificationErrorMessage("你需要有`訊息管理`才能使用此指令"))
		}
		id := firstOption(interaction, optionID)
		if err := m.service.Delete(ctx, interaction.Actor.GuildID, id); err != nil {
			if errors.Is(err, ports.ErrAutoNotificationScheduleMissing) {
				return responder.EditOriginal(ctx, autoNotificationErrorMessage("找不到這個id的自動通知，請使用`/自動通知列表`進行查詢"))
			}
			return responder.EditOriginal(ctx, autoNotificationErrorFromError(err))
		}
		if err := responder.EditOriginal(ctx, autoNotificationDeleteMessage()); err != nil {
			return err
		}
		return m.track(ctx, interaction, AutoNotificationDeleteCommandName)
	}
}

func (m Module) guildName(ctx context.Context, guildID string) string {
	if m.discord == nil || strings.TrimSpace(guildID) == "" {
		return autoNotificationFallbackGld
	}
	info, err := m.discord.GuildInfo(ctx, guildID)
	if err != nil || strings.TrimSpace(info.Name) == "" {
		return autoNotificationFallbackGld
	}
	return info.Name
}

func autoNotificationListMessage(guildName string, schedules []domain.AutoNotificationSchedule, color int) responses.Message {
	if strings.TrimSpace(guildName) == "" {
		guildName = autoNotificationFallbackGld
	}
	rows := make([]string, 0, len(schedules))
	for index, schedule := range schedules {
		rows = append(rows, fmt.Sprintf("\n**❰%d❱ id:`%s` cron設定:`%s` 頻道:**<#%s>", index+1, schedule.ID, schedule.Cron, schedule.ChannelID))
	}
	return responses.Message{
		Embeds: []responses.Embed{{
			Title:       "<:list:992002476360343602> 以下是" + guildName + "的所有自動通知id",
			Description: "輸入`/自動通知刪除 id`可進行刪除之前設定的自動通知\n" + strings.Join(rows, " "),
			Color:       color,
		}},
		AllowedMentions: &responses.AllowedMentions{},
	}
}

func autoNotificationDeleteMessage() responses.Message {
	return responses.Message{
		Embeds: []responses.Embed{{
			Title:       "<a:green_tick:994529015652163614>自動通知系統",
			Description: "<:trashbin:995991389043163257>成功刪除該自動通知",
			Color:       autoNotificationGreenColor,
		}},
		AllowedMentions: &responses.AllowedMentions{},
	}
}

func autoNotificationErrorFromError(err error) responses.Message {
	_ = err
	return autoNotificationErrorMessage("很抱歉，出現了未知的錯誤，請重試!")
}

func autoNotificationErrorMessage(content string) responses.Message {
	return responses.Message{
		Embeds: []responses.Embed{{
			Title:       "<a:Discord_AnimatedNo:1015989839809757295> | " + content,
			Description: "<a:arrow_pink:996242460294512690> [點我前往教學網址](https://youtu.be/D43zPrZU5Fw)",
			Color:       autoNotificationErrorColor,
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
		Feature:     "auto-notification-config",
	})
}
