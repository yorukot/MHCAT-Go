package notifications

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/customid"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/interactions"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/responses"
)

const (
	permissionManageMessages    = int64(8192)
	autoNotificationErrorColor  = 0xED4245
	autoNotificationGreenColor  = 0x57F287
	autoNotificationFallbackGld = "這個伺服器"
	fieldCron                   = "cron_setcron"
	fieldMessage                = "cron_setmsg"
	fieldColor                  = "cron_setcolor"
	fieldTitle                  = "cron_settitle"
	fieldContent                = "cron_setcontent"
)

func (m Module) SetupHandler() interactions.Handler {
	return func(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
		if !interaction.Actor.HasPermission(permissionManageMessages) {
			return responder.Reply(ctx, autoNotificationErrorMessage("你需要有`訊息管理`才能使用此指令"))
		}
		id := setupID(m.now())
		if err := m.service.StartSetup(ctx, domain.AutoNotificationSetupDraft{
			GuildID:   interaction.Actor.GuildID,
			ID:        id,
			ChannelID: firstOption(interaction, optionChannel, "頻道"),
		}); err != nil {
			if errors.Is(err, ports.ErrAutoNotificationScheduleLimit) {
				return responder.Reply(ctx, autoNotificationErrorMessage("一個伺服器最多只能設置10個自動通知，請使用`/備份列表`進行刪除"))
			}
			return responder.Reply(ctx, autoNotificationErrorFromError(err))
		}
		if err := responder.ShowModal(ctx, autoNotificationSetupModal(id)); err != nil {
			return err
		}
		return m.track(ctx, interaction, AutoNotificationSetupCommandName)
	}
}

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

func (m Module) SetupModalHandler() interactions.Handler {
	return func(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
		if err := responder.Defer(ctx, responses.DeferOptions{}); err != nil {
			return err
		}
		fields := autoNotificationModalFields(interaction.ModalFields)
		message := domain.AutoNotificationMessage{
			Content:          fields[fieldMessage],
			EmbedTitle:       fields[fieldTitle],
			EmbedDescription: fields[fieldContent],
			EmbedColor:       fields[fieldColor],
		}.Normalized()
		if message.EmbedColor != "" {
			if !validLegacyAutoNotificationColor(message.EmbedColor) {
				return responder.EditOriginal(ctx, autoNotificationErrorMessage("你傳送的並不是顏色(色碼)"))
			}
		}
		if message.Empty() {
			return responder.EditOriginal(ctx, autoNotificationErrorMessage("你都沒輸入你要發送甚麼，我要怎麼發送啦!"))
		}
		if message.HasEmbed() && message.EmbedColor == "Random" {
			message = resolveAutoNotificationMessageColor(message, m.color())
		}
		cron := fields[fieldCron]
		switch validateDirectCron(cron, m.now()) {
		case directCronTooFrequent:
			return responder.EditOriginal(ctx, autoNotificationErrorMessage("傳送訊息的間隔必須大於15分鐘!"))
		case directCronInvalid:
			return m.startSimplifiedWizard(ctx, interaction, responder, message)
		}
		id := strings.TrimSpace(interaction.CustomID)
		if err := m.service.CompleteSetup(ctx, domain.AutoNotificationSetup{
			GuildID: interaction.Actor.GuildID,
			ID:      id,
			Cron:    cron,
			Message: message,
		}); err != nil {
			return responder.EditOriginal(ctx, autoNotificationErrorFromError(err))
		}
		if err := responder.EditOriginal(ctx, autoNotificationSetupCompleteMessage(id)); err != nil {
			return err
		}
		if m.messages != nil && strings.TrimSpace(interaction.ChannelID) != "" {
			_, _ = m.messages.SendMessage(ctx, interaction.ChannelID, autoNotificationPreviewOutbound(message, m.color()))
		}
		return nil
	}
}

func validLegacyAutoNotificationColor(value string) bool {
	if value == "Random" {
		return true
	}
	if len(value) == 7 && value[0] == '#' {
		_, ok := domain.ParseLegacyColorValue(value)
		return ok
	}
	switch value {
	case "White", "Aqua", "Green", "Blue", "Yellow", "Purple", "Fuchsia", "Gold", "Orange", "Red", "Navy", "DarkGreen", "DarkBlue", "DarkOrange", "DarkRed":
		return true
	default:
		return false
	}
}

func resolveAutoNotificationMessageColor(message domain.AutoNotificationMessage, randomColor int) domain.AutoNotificationMessage {
	message = message.Normalized()
	if message.HasEmbed() && message.EmbedColor == "Random" {
		message.EmbedColor = fmt.Sprintf("#%06X", randomColor&0xFFFFFF)
	}
	return message
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

func autoNotificationSetupModal(id string) responses.Modal {
	return responses.Modal{
		CustomID: id,
		Title:    "自動發送通知系統!",
		Rows: []responses.ModalRow{
			{Inputs: []responses.TextInput{{
				CustomID: fieldCron,
				Label:    "請輸入corn表達式(如想用簡化版，請直接輸入取消或cancel就可以簡易設置corn)",
				Style:    responses.TextInputStyleShort,
				Required: true,
			}}},
			{Inputs: []responses.TextInput{{
				CustomID: fieldMessage,
				Label:    "請輸入文字(如不輸入這項請務必輸入下面三項)",
				Style:    responses.TextInputStyleParagraph,
			}}},
			{Inputs: []responses.TextInput{{
				CustomID: fieldColor,
				Label:    "請輸入你的嵌入訊息顏色(如不輸入嵌入訊息相關，請務必輸入文字)",
				Style:    responses.TextInputStyleShort,
			}}},
			{Inputs: []responses.TextInput{{
				CustomID: fieldTitle,
				Label:    "請輸入你的嵌入標題(如不輸入嵌入訊息相關，請務必輸入文字)",
				Style:    responses.TextInputStyleShort,
			}}},
			{Inputs: []responses.TextInput{{
				CustomID: fieldContent,
				Label:    "請輸入嵌入內文(如不輸入嵌入訊息相關，請務必輸入文字)",
				Style:    responses.TextInputStyleParagraph,
			}}},
		},
	}
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

func autoNotificationSetupCompleteMessage(id string) responses.Message {
	return responses.Message{
		Content:         fmt.Sprintf(":white_check_mark:**以下是該自動通知id:**`%s`\n使用`/自動通知刪除 id:%s`進行刪除\n~~我只是個分隔線，下面是你的訊息預覽~~", id, id),
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

func autoNotificationPreviewOutbound(message domain.AutoNotificationMessage, color int) ports.OutboundMessage {
	message = message.Normalized()
	result := ports.OutboundMessage{
		Content:         message.Content,
		AllowedMentions: ports.AllowedMentions{},
	}
	if !message.HasEmbed() {
		return result
	}
	embedColor := 0
	if message.EmbedColor == "Random" {
		embedColor = color
	} else if message.EmbedColor != "" {
		if parsed, ok := domain.ParseLegacyColorValue(message.EmbedColor); ok {
			embedColor = parsed
		}
	}
	result.Embeds = []ports.OutboundEmbed{{
		Title:       message.EmbedTitle,
		Description: message.EmbedDescription,
		Color:       embedColor,
	}}
	return result
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

func setupID(createdAt time.Time) string {
	if createdAt.IsZero() {
		createdAt = time.Now()
	}
	return strconv.FormatInt(createdAt.UnixMilli(), 10)
}

func autoNotificationModalFields(fields []customid.ModalField) map[string]string {
	values := make(map[string]string, len(fields))
	for _, field := range fields {
		values[field.CustomID] = field.Value
	}
	return values
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
