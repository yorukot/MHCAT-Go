package notifications

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/customid"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/interactions"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/responses"
)

func (m Module) startSimplifiedWizard(ctx context.Context, interaction interactions.Interaction, responder responses.Responder, message domain.AutoNotificationMessage) error {
	if m.pendingWizards == nil {
		return responder.EditOriginal(ctx, autoNotificationErrorFromError(domain.ErrInvalidAutoNotificationSchedule))
	}
	ownerUserID := strings.TrimSpace(interaction.Actor.UserID)
	guildID := strings.TrimSpace(interaction.Actor.GuildID)
	scheduleID := strings.TrimSpace(interaction.CustomID)
	if ownerUserID == "" || guildID == "" || scheduleID == "" {
		return responder.EditOriginal(ctx, autoNotificationErrorFromError(domain.ErrInvalidAutoNotificationSchedule))
	}
	now := m.now()
	expiresAt := now.Add(autoNotificationWizardTTL)
	stateID, err := m.pendingWizards.create(now, pendingAutoNotificationWizard{
		OwnerUserID:      ownerUserID,
		GuildID:          guildID,
		ScheduleID:       scheduleID,
		PreviewChannelID: strings.TrimSpace(interaction.ChannelID),
		Message:          message.Normalized(),
		ExpiresAt:        expiresAt,
	})
	if err != nil {
		return responder.EditOriginal(ctx, autoNotificationErrorFromError(err))
	}
	customID, err := autoNotificationWizardCustomID("week", stateID)
	if err != nil {
		m.pendingWizards.delete(stateID)
		return responder.EditOriginal(ctx, autoNotificationErrorFromError(err))
	}
	if err := responder.EditOriginal(ctx, m.autoNotificationWeekMessage(customID, expiresAt, interaction.Actor.AvatarURL)); err != nil {
		m.pendingWizards.delete(stateID)
		return err
	}
	return nil
}

func (m Module) WeekSelectHandler() interactions.Handler {
	return func(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
		stateID, ok := autoNotificationWizardStateID(interaction.CustomID, "week")
		week, valid := autoNotificationWeekExpression(interaction.Values)
		if !ok || !valid || m.pendingWizards == nil {
			return responder.Reply(ctx, autoNotificationWizardErrorMessage())
		}
		entry, ok := m.pendingWizards.setWeek(stateID, interaction.Actor.UserID, m.now(), week)
		if !ok {
			return responder.Reply(ctx, autoNotificationWizardErrorMessage())
		}
		customID, err := autoNotificationWizardCustomID("hour", stateID)
		if err != nil {
			return responder.Reply(ctx, autoNotificationWizardErrorMessage())
		}
		return responder.UpdateMessage(ctx, m.autoNotificationHourMessage(customID, entry.ExpiresAt, interaction.Actor.AvatarURL))
	}
}

func (m Module) HourSelectHandler() interactions.Handler {
	return func(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
		stateID, ok := autoNotificationWizardStateID(interaction.CustomID, "hour")
		hours, valid := autoNotificationHourExpression(interaction.Values)
		if !ok || !valid || m.pendingWizards == nil {
			return responder.Reply(ctx, autoNotificationWizardErrorMessage())
		}
		entry, ok := m.pendingWizards.setHours(stateID, interaction.Actor.UserID, m.now(), hours)
		if !ok {
			return responder.Reply(ctx, autoNotificationWizardErrorMessage())
		}
		customID, err := autoNotificationWizardCustomID("minute", stateID)
		if err != nil {
			return responder.Reply(ctx, autoNotificationWizardErrorMessage())
		}
		return responder.UpdateMessage(ctx, m.autoNotificationMinuteMessage(customID, entry.ExpiresAt, interaction.Actor.AvatarURL))
	}
}

func (m Module) MinuteSelectHandler() interactions.Handler {
	return func(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
		stateID, ok := autoNotificationWizardStateID(interaction.CustomID, "minute")
		minutes, valid := autoNotificationMinuteExpression(interaction.Values)
		if !ok || !valid || m.pendingWizards == nil {
			return responder.Reply(ctx, autoNotificationWizardErrorMessage())
		}
		entry, ok := m.pendingWizards.ready(stateID, interaction.Actor.UserID, m.now())
		if !ok {
			return responder.Reply(ctx, autoNotificationWizardErrorMessage())
		}
		cron := minutes + " " + entry.Hours + " * * " + entry.Week
		if err := m.service.CompleteSetup(ctx, domain.AutoNotificationSetup{
			GuildID: entry.GuildID,
			ID:      entry.ScheduleID,
			Cron:    cron,
			Message: entry.Message,
		}); err != nil {
			message := autoNotificationErrorFromError(err)
			message.Ephemeral = true
			return responder.Reply(ctx, message)
		}
		m.pendingWizards.delete(stateID)
		if err := responder.UpdateMessage(ctx, m.autoNotificationWizardCompleteMessage(entry.ScheduleID)); err != nil {
			return err
		}
		previewChannelID := entry.PreviewChannelID
		if previewChannelID == "" {
			previewChannelID = strings.TrimSpace(interaction.ChannelID)
		}
		if m.messages != nil && previewChannelID != "" {
			_, _ = m.messages.SendMessage(ctx, previewChannelID, autoNotificationPreviewOutbound(entry.Message, m.color()))
		}
		return nil
	}
}

func (m Module) autoNotificationWeekMessage(customID string, expiresAt time.Time, avatarURL string) responses.Message {
	return autoNotificationWizardSelectMessage(
		customID,
		"**<:7days:1022059380725784626> 請選取你的定時要在星期幾執行__(可複選)__**",
		"請選擇要在星期幾發送(可複選)",
		1,
		7,
		legacyAutoNotificationWeekOptions(),
		expiresAt,
		avatarURL,
		m.color(),
	)
}

func (m Module) autoNotificationHourMessage(customID string, expiresAt time.Time, avatarURL string) responses.Message {
	return autoNotificationWizardSelectMessage(
		customID,
		"**<:24hours:1022059604747747379> 請選取你的定時要在幾點執行__(可複選)__**",
		"請選擇要在幾點發送(可複選)24hr制",
		1,
		24,
		legacyAutoNotificationHourOptions(),
		expiresAt,
		avatarURL,
		m.color(),
	)
}

func (m Module) autoNotificationMinuteMessage(customID string, expiresAt time.Time, avatarURL string) responses.Message {
	return autoNotificationWizardSelectMessage(
		customID,
		"<:60minutes:1022059603153924156> **請選取你的定時要在幾分執行__(可複選)__**",
		"請選擇要在幾分發送(可複選)24hr制",
		1,
		6,
		legacyAutoNotificationMinuteOptions(),
		expiresAt,
		avatarURL,
		m.color(),
	)
}

func autoNotificationWizardSelectMessage(customID string, prompt string, placeholder string, minValues int, maxValues int, options []responses.SelectOption, expiresAt time.Time, avatarURL string, color int) responses.Message {
	return responses.Message{
		Embeds: []responses.Embed{{
			Title:       "<:dailytasks:1022041880394989669> 設定corn",
			Description: prompt + "\n**<a:warn:1000814885506129990> 你必須在<t:" + legacyAutoNotificationExpiryTimestamp(expiresAt) + ":R>選取完畢(超過時間將會無法選取)**",
			Color:       color,
			Footer: &responses.EmbedFooter{
				Text:    "有問題都可以前往支援伺服器詢問",
				IconURL: avatarURL,
			},
		}},
		Components: []responses.ComponentRow{{Components: []responses.Component{{
			Type:        responses.ComponentTypeSelect,
			CustomID:    customID,
			Placeholder: placeholder,
			MinValues:   minValues,
			MaxValues:   maxValues,
			Options:     options,
		}}}},
		AllowedMentions: &responses.AllowedMentions{},
	}
}

func legacyAutoNotificationExpiryTimestamp(expiresAt time.Time) string {
	return strconv.FormatInt(expiresAt.Add(500*time.Millisecond).Unix(), 10)
}

func (m Module) autoNotificationWizardCompleteMessage(id string) responses.Message {
	return responses.Message{
		Embeds: []responses.Embed{{
			Title: "<:dailytasks:1022041880394989669> 設定corn",
			Description: "<a:green_tick:994529015652163614> 恭喜你設定完成了!\n" +
				"**以下是該自動通知id:**`" + id + "`\n" +
				"使用`/自動通知刪除 id:" + id + "`進行刪除\n" +
				"~~我只是個分隔線，下面是你的訊息預覽~~",
			Color: m.color(),
		}},
		AllowedMentions: &responses.AllowedMentions{},
	}
}

func autoNotificationWizardErrorMessage() responses.Message {
	message := autoNotificationErrorMessage("這個簡易設定已過期或不屬於你，請重新執行指令")
	message.Ephemeral = true
	return message
}

func autoNotificationWizardCustomID(action string, stateID string) (string, error) {
	payload, err := customid.StateIDPayload(stateID)
	if err != nil {
		return "", err
	}
	return customid.Encode(customid.InteractionKindComponent, "cron", action, payload)
}

func autoNotificationWizardStateID(raw string, action string) (string, bool) {
	parsed, err := customid.ParseComponent(raw)
	if err != nil || parsed.Version != customid.VersionV1 || parsed.Feature != "cron" || parsed.Action != action {
		return "", false
	}
	if parsed.Payload.Kind != customid.PayloadState || strings.TrimSpace(parsed.Payload.StateID) == "" {
		return "", false
	}
	return parsed.Payload.StateID, true
}

func autoNotificationWeekExpression(values []string) (string, bool) {
	selected, ok := orderedAutoNotificationSelection(values, []string{"1", "2", "3", "4", "5", "6", "0"}, 7)
	if !ok {
		return "", false
	}
	if len(selected) == 7 {
		return "*", true
	}
	return strings.Join(selected, ","), true
}

func autoNotificationHourExpression(values []string) (string, bool) {
	order := make([]string, 0, 24)
	for hour := 1; hour <= 23; hour++ {
		order = append(order, strconv.Itoa(hour))
	}
	order = append(order, "0")
	selected, ok := orderedAutoNotificationSelection(values, order, 24)
	return strings.Join(selected, ","), ok
}

func autoNotificationMinuteExpression(values []string) (string, bool) {
	order := make([]string, 0, 12)
	for minute := 0; minute < 60; minute += 5 {
		order = append(order, strconv.Itoa(minute))
	}
	selected, ok := orderedAutoNotificationSelection(values, order, 6)
	return strings.Join(selected, ","), ok
}

func orderedAutoNotificationSelection(values []string, order []string, maxValues int) ([]string, bool) {
	if len(values) == 0 || len(values) > maxValues {
		return nil, false
	}
	selected := make(map[string]struct{}, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			return nil, false
		}
		if _, exists := selected[value]; exists {
			return nil, false
		}
		selected[value] = struct{}{}
	}
	result := make([]string, 0, len(selected))
	for _, value := range order {
		if _, exists := selected[value]; exists {
			result = append(result, value)
			delete(selected, value)
		}
	}
	if len(selected) != 0 || len(result) == 0 {
		return nil, false
	}
	return result, true
}

func legacyAutoNotificationWeekOptions() []responses.SelectOption {
	return []responses.SelectOption{
		{Label: "禮拜一", Description: "禮拜一執行", Value: "1", Emoji: "<:monday:1022040759614050314>"},
		{Label: "禮拜二", Description: "禮拜二執行", Value: "2", Emoji: "<:tuesday:1022040763044986931>"},
		{Label: "禮拜三", Description: "禮拜三執行", Value: "3", Emoji: "<:wednesday:1022040757764378686>"},
		{Label: "禮拜四", Description: "禮拜四執行", Value: "4", Emoji: "<:thursday:1022040755834990695>"},
		{Label: "禮拜五", Description: "禮拜五執行", Value: "5", Emoji: "<:friday:1022040752722825237>"},
		{Label: "禮拜六", Description: "禮拜六執行", Value: "6", Emoji: "<:saturday:1022040761165955134>"},
		{Label: "禮拜日", Description: "禮拜日執行", Value: "0", Emoji: "<:sunday:1022040754643812352>"},
	}
}

func legacyAutoNotificationHourOptions() []responses.SelectOption {
	return []responses.SelectOption{
		{Label: "1點", Description: "凌晨1點", Value: "1", Emoji: "<:moon:1022055227194605599>"},
		{Label: "2點", Description: "凌晨2點", Value: "2", Emoji: "<:moon:1022055227194605599>"},
		{Label: "3點", Description: "凌晨3點", Value: "3", Emoji: "<:moon:1022055227194605599>"},
		{Label: "4點", Description: "凌晨4點", Value: "4", Emoji: "<:moon:1022055227194605599>"},
		{Label: "5點", Description: "早上5點", Value: "5", Emoji: "<:morning:1022055616203726888>"},
		{Label: "6點", Description: "早上6點", Value: "6", Emoji: "<:morning:1022055616203726888>"},
		{Label: "7點", Description: "早上7點", Value: "7", Emoji: "<:morning:1022055616203726888>"},
		{Label: "8點", Description: "早上8點", Value: "8", Emoji: "<:morning:1022055616203726888>"},
		{Label: "9點", Description: "早上9點", Value: "9", Emoji: "<:morning:1022055616203726888>"},
		{Label: "10點", Description: "早上10點", Value: "10", Emoji: "<:morning:1022055616203726888>"},
		{Label: "11點", Description: "中午11點", Value: "11", Emoji: "<:sun:1022055614458904596>"},
		{Label: "12點", Description: "中午12點", Value: "12", Emoji: "<:sun:1022055614458904596>"},
		{Label: "13點", Description: "中午1點", Value: "13", Emoji: "<:sun:1022055614458904596>"},
		{Label: "14點", Description: "下午2點", Value: "14", Emoji: "<:sun1:1022055612294647839>"},
		{Label: "15點", Description: "下午3點", Value: "15", Emoji: "<:sun1:1022055612294647839>"},
		{Label: "16點", Description: "下午4點", Value: "16", Emoji: "<:sun1:1022055612294647839>"},
		{Label: "17點", Description: "下午5點", Value: "17", Emoji: "<:sun1:1022055612294647839>"},
		{Label: "18點", Description: "晚上6點", Value: "18", Emoji: "<:forest:1022055611044732998>"},
		{Label: "19點", Description: "晚上7點", Value: "19", Emoji: "<:forest:1022055611044732998>"},
		{Label: "20點", Description: "晚上8點", Value: "20", Emoji: "<:forest:1022055611044732998>"},
		{Label: "21點", Description: "晚上9點", Value: "21", Emoji: "<:forest:1022055611044732998>"},
		{Label: "22點", Description: "晚上10點", Value: "22", Emoji: "<:forest:1022055611044732998>"},
		{Label: "23點", Description: "晚上11點", Value: "23", Emoji: "<:forest:1022055611044732998>"},
		{Label: "24點(0點)", Description: "凌晨12點(0點)", Value: "0", Emoji: "<:moon:1022055227194605599>"},
	}
}

func legacyAutoNotificationMinuteOptions() []responses.SelectOption {
	return []responses.SelectOption{
		{Label: "0分", Description: "每個你選取的小時的0分", Value: "0", Emoji: "<:time:1022057997515640852>"},
		{Label: "5分", Description: "每個你選取的小時的5分", Value: "5", Emoji: "<:time:1022057997515640852>"},
		{Label: "10分", Description: "每個你選取的小時的10分", Value: "10", Emoji: "<:time:1022057997515640852>"},
		{Label: "15分", Description: "每個你選取的小時的15分", Value: "15", Emoji: "<:15minutes:1022058003752570933>"},
		{Label: "20分", Description: "每個你選取的小時的20分", Value: "20", Emoji: "<:15minutes:1022058003752570933>"},
		{Label: "25分", Description: "每個你選取的小時的25分", Value: "25", Emoji: "<:15minutes:1022058003752570933>"},
		{Label: "30分", Description: "每個你選取的小時的30分", Value: "30", Emoji: "<:30minutes:1022058001722527744>"},
		{Label: "35分", Description: "每個你選取的小時的35分", Value: "35", Emoji: "<:30minutes:1022058001722527744>"},
		{Label: "40分", Description: "每個你選取的小時的40分", Value: "40", Emoji: "<:30minutes:1022058001722527744>"},
		{Label: "45分", Description: "每個你選取的小時的45分", Value: "45", Emoji: "<:45minutes:1022057999881228288>"},
		{Label: "50分", Description: "每個你選取的小時的50分", Value: "50", Emoji: "<:45minutes:1022057999881228288>"},
		{Label: "55分", Description: "每個你選取的小時的55分", Value: "55", Emoji: "<:45minutes:1022057999881228288>"},
	}
}
