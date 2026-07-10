package birthday

import (
	"context"
	"errors"
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
	permissionManageMessages = int64(8192)
	birthdaySuccessColor     = 0x57F287
	birthdayErrorColor       = 0xED4245
	birthdayAddTimeout       = 5 * time.Minute
	birthdayDateAddDocsPath  = "allcommands/生日系統/birthday_date_add"
)

func (m Module) Handler() interactions.Handler {
	return func(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
		if err := responder.Defer(ctx, responses.DeferOptions{}); err != nil {
			return err
		}
		switch interaction.Subcommand {
		case subcommandConfig:
			return m.handleConfig(ctx, interaction, responder)
		case subcommandAdd:
			return m.handleAdd(ctx, interaction, responder)
		case subcommandDelete:
			return m.handleDelete(ctx, interaction, responder)
		case subcommandAllowAdmin:
			return m.handleAllowAdmin(ctx, interaction, responder)
		case subcommandList:
			return m.handleList(ctx, interaction, responder)
		default:
			return responder.EditOriginal(ctx, stagedUnavailableMessage())
		}
	}
}

func (m Module) handleAdd(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
	year, ok := optionalIntOption(interaction, optionBirthdayYear)
	if !ok {
		invalidYear := 1899
		year = &invalidYear
	}
	month, ok := intOption(interaction, optionBirthdayMonth)
	if !ok {
		month = 0
	}
	day, ok := intOption(interaction, optionBirthdayDay)
	if !ok {
		day = 0
	}
	targetUserID := firstOption(interaction, optionUser)
	if strings.TrimSpace(targetUserID) == "" {
		targetUserID = interaction.Actor.UserID
	}
	now := m.now()
	profile, err := m.profileService.PrepareAdd(ctx, domain.BirthdayAddRequest{
		GuildID:                interaction.Actor.GuildID,
		ActorUserID:            interaction.Actor.UserID,
		TargetUserID:           targetUserID,
		ActorCanManageMessages: interaction.Actor.HasPermission(permissionManageMessages),
		BirthdayYear:           year,
		BirthdayMonth:          month,
		BirthdayDay:            day,
		CurrentYear:            now.Year(),
	})
	if err != nil {
		return responder.EditOriginal(ctx, birthdayErrorFromError(err))
	}
	if m.pendingAdds == nil {
		return responder.EditOriginal(ctx, birthdayErrorMessage("很抱歉，出現了未知的錯誤，請重試!", ""))
	}
	expiresAt := now.Add(birthdayAddTimeout)
	stateID := m.pendingAdds.create(now, pendingBirthdayAdd{
		OwnerUserID: interaction.Actor.UserID,
		Profile:     profile,
		ExpiresAt:   expiresAt,
	})
	customID, err := birthdayAddCustomID("hour", stateID)
	if err != nil {
		m.pendingAdds.delete(stateID)
		return responder.EditOriginal(ctx, birthdayErrorFromError(err))
	}
	if err := responder.EditOriginal(ctx, m.hourSelectMessage(customID, expiresAt, interaction.Actor.AvatarURL)); err != nil {
		m.pendingAdds.delete(stateID)
		return err
	}
	return m.track(ctx, interaction)
}

func (m Module) HourSelectHandler() interactions.Handler {
	return func(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
		stateID, ok := birthdayStateIDFromCustomID(interaction.CustomID, "hour")
		if !ok || m.pendingAdds == nil {
			return responder.Reply(ctx, birthdayAddEphemeralError("請重新執行`/生日系統 增加`設定生日日期"))
		}
		now := m.now()
		entry, ok := m.pendingAdds.get(stateID, now)
		if !ok {
			return responder.Reply(ctx, birthdayAddEphemeralError("請重新執行`/生日系統 增加`設定生日日期"))
		}
		if entry.OwnerUserID != interaction.Actor.UserID {
			return responder.Reply(ctx, birthdayAddEphemeralError("你不能操作這個生日設定選單!"))
		}
		hour, ok := selectedInt(interaction)
		if !ok || hour < 0 || hour > 23 {
			return responder.Reply(ctx, birthdayAddEphemeralError("很抱歉，出現了未知的錯誤，請重試!"))
		}
		entry, ok = m.pendingAdds.setHour(stateID, now, hour)
		if !ok {
			return responder.Reply(ctx, birthdayAddEphemeralError("請重新執行`/生日系統 增加`設定生日日期"))
		}
		customID, err := birthdayAddCustomID("minute", stateID)
		if err != nil {
			return responder.Reply(ctx, birthdayAddEphemeralError("很抱歉，出現了未知的錯誤，請重試!"))
		}
		return responder.UpdateMessage(ctx, m.minuteSelectMessage(customID, entry.ExpiresAt, interaction.Actor.AvatarURL))
	}
}

func (m Module) MinuteSelectHandler() interactions.Handler {
	return func(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
		stateID, ok := birthdayStateIDFromCustomID(interaction.CustomID, "minute")
		if !ok || m.pendingAdds == nil {
			return responder.Reply(ctx, birthdayAddEphemeralError("請重新執行`/生日系統 增加`設定生日日期"))
		}
		entry, ok := m.pendingAdds.get(stateID, m.now())
		if !ok || !entry.HasHour {
			return responder.Reply(ctx, birthdayAddEphemeralError("請重新執行`/生日系統 增加`設定生日日期"))
		}
		if entry.OwnerUserID != interaction.Actor.UserID {
			return responder.Reply(ctx, birthdayAddEphemeralError("你不能操作這個生日設定選單!"))
		}
		minute, ok := selectedInt(interaction)
		if !ok || minute < 0 || minute > 55 || minute%5 != 0 {
			return responder.Reply(ctx, birthdayAddEphemeralError("很抱歉，出現了未知的錯誤，請重試!"))
		}
		if err := m.profileService.SaveDateTime(ctx, entry.Profile, entry.Hour, minute); err != nil {
			return responder.Reply(ctx, birthdayAddEphemeralErrorMessage(birthdayErrorFromError(err)))
		}
		m.pendingAdds.delete(stateID)
		return responder.UpdateMessage(ctx, m.birthdayAddSuccessMessage(entry.Profile, entry.Hour, minute))
	}
}

func (m Module) handleConfig(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
	if !interaction.Actor.HasPermission(permissionManageMessages) {
		return responder.EditOriginal(ctx, birthdayErrorMessage("你需要有`訊息管理`才能使用此指令", "allcommands/生日系統/birthday_message_set"))
	}
	canSet, ok := boolOption(interaction, optionEveryoneCanSet)
	if !ok {
		return responder.EditOriginal(ctx, birthdayErrorFromError(domain.ErrInvalidBirthdayConfig))
	}
	config := domain.BirthdayConfig{
		GuildID:                    interaction.Actor.GuildID,
		Message:                    stringOption(interaction, optionMessage),
		UTCOffset:                  firstOption(interaction, optionUTC),
		ChannelID:                  firstOption(interaction, optionChannel),
		EveryoneCanSetBirthdayDate: canSet,
		RoleID:                     firstOption(interaction, optionRole),
	}
	if err := m.configService.Save(ctx, config); err != nil {
		return responder.EditOriginal(ctx, birthdayErrorFromError(err))
	}
	if err := responder.EditOriginal(ctx, configSuccessMessage(config)); err != nil {
		return err
	}
	return m.track(ctx, interaction)
}

func (m Module) handleDelete(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
	if !interaction.Actor.HasPermission(permissionManageMessages) {
		return responder.EditOriginal(ctx, birthdayErrorMessage("你需要有`訊息管理`才能使用此指令", birthdayDateAddDocsPath))
	}
	userID := firstOption(interaction, optionUser)
	if err := m.profileService.Delete(ctx, interaction.Actor.GuildID, userID); err != nil {
		return responder.EditOriginal(ctx, birthdayProfileErrorMessage(err, "沒有這位使用者的資料!"))
	}
	if err := responder.EditOriginal(ctx, deleteSuccessMessage(userID)); err != nil {
		return err
	}
	return m.track(ctx, interaction)
}

func (m Module) handleAllowAdmin(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
	allow, ok := boolOption(interaction, optionAllowAdmin)
	if !ok {
		return responder.EditOriginal(ctx, birthdayErrorFromError(domain.ErrInvalidBirthdayProfile))
	}
	if err := m.profileService.SetAllowAdmin(ctx, interaction.Actor.GuildID, interaction.Actor.UserID, allow); err != nil {
		return responder.EditOriginal(ctx, birthdayProfileErrorMessage(err, "很抱歉，出現了未知的錯誤，請重試!"))
	}
	if err := responder.EditOriginal(ctx, allowAdminSuccessMessage(allow)); err != nil {
		return err
	}
	return m.track(ctx, interaction)
}

func (m Module) handleList(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
	profiles, err := m.profileService.List(ctx, interaction.Actor.GuildID)
	if err != nil {
		return responder.EditOriginal(ctx, birthdayProfileErrorMessage(err, "很抱歉，出現了未知的錯誤，請重試!"))
	}
	if len(profiles) == 0 {
		return responder.EditOriginal(ctx, birthdayErrorMessage("還沒有任何人有進行生日設置喔!", birthdayDateAddDocsPath))
	}
	if err := responder.EditOriginal(ctx, listMessage(profiles, m.cachedBirthdayUserTags(ctx, interaction.Actor.GuildID, profiles), m.legacyColor())); err != nil {
		return err
	}
	return m.track(ctx, interaction)
}

func (m Module) hourSelectMessage(customID string, expiresAt time.Time, avatarURL string) responses.Message {
	return responses.Message{
		Embeds: []responses.Embed{{
			Title:       "<:cake:1065654305983570041> 生日系統祝福語設定",
			Description: "**<:24hours:1022059604747747379> 請選取你的生日通知要在幾點發送**\n**<a:warn:1000814885506129990> 你必須在<t:" + legacyBirthdayExpiryTimestamp(expiresAt) + ":R>選取完畢(超過時間將會無法選取)**",
			Color:       m.legacyColor(),
			Footer:      &responses.EmbedFooter{Text: "有問題都可以前往支援伺服器詢問", IconURL: avatarURL},
		}},
		Components: []responses.ComponentRow{{Components: []responses.Component{{
			Type:        responses.ComponentTypeSelect,
			CustomID:    customID,
			Placeholder: "請選擇要在幾點發送(24hr制)",
			MinValues:   1,
			MaxValues:   1,
			Options:     legacyBirthdayHourOptions(),
		}}}},
		AllowedMentions: &responses.AllowedMentions{},
	}
}

func (m Module) minuteSelectMessage(customID string, expiresAt time.Time, avatarURL string) responses.Message {
	return responses.Message{
		Embeds: []responses.Embed{{
			Title:       "<:cake:1065654305983570041> 生日系統祝福語設定",
			Description: "<:60minutes:1022059603153924156> **請選取你的生日通知要在幾分發送**\n**<a:warn:1000814885506129990> 你必須在<t:" + legacyBirthdayExpiryTimestamp(expiresAt) + ":R>選取完畢(超過時間將會無法選取)**",
			Color:       m.legacyColor(),
			Footer:      &responses.EmbedFooter{Text: "有問題都可以前往支援伺服器詢問", IconURL: avatarURL},
		}},
		Components: []responses.ComponentRow{{Components: []responses.Component{{
			Type:        responses.ComponentTypeSelect,
			CustomID:    customID,
			Placeholder: "請選擇要在幾分發送",
			MinValues:   1,
			MaxValues:   1,
			Options:     legacyBirthdayMinuteOptions(),
		}}}},
		AllowedMentions: &responses.AllowedMentions{},
	}
}

func legacyBirthdayExpiryTimestamp(expiresAt time.Time) string {
	return strconv.FormatInt(expiresAt.Add(500*time.Millisecond).Unix(), 10)
}

func (m Module) birthdayAddSuccessMessage(profile domain.BirthdayProfile, hour int, minute int) responses.Message {
	return responses.Message{
		Embeds: []responses.Embed{{
			Title: "<:cake:1065654305983570041> 生日系統祝福語設定",
			Description: "<a:green_tick:994529015652163614> 恭喜你設定完成了!\n" +
				"**<a:arrow_pink:996242460294512690> 以下是<@" + strings.TrimSpace(profile.UserID) + ">的生日日期:**`" + profileDate(profile) + "`\n" +
				"**通知時間為:**`" + strconv.Itoa(hour) + ":" + strconv.Itoa(minute) + "`",
			Color: m.legacyColor(),
		}},
		AllowedMentions: &responses.AllowedMentions{},
	}
}

func configSuccessMessage(config domain.BirthdayConfig) responses.Message {
	role := "null"
	if strings.TrimSpace(config.RoleID) != "" {
		role = "<@&" + strings.TrimSpace(config.RoleID) + ">"
	}
	return responses.Message{
		Embeds: []responses.Embed{{
			Title: "<:cake:1065654305983570041> 生日系統祝福語設定",
			Description: "**你成功設定了祝福語!**\n" +
				"<:confetti:1065654294071738399> **祝福語為:**\n" + config.Message +
				"\n<:utc:1065654078417412168> **時區為:** `UTC" + config.UTCOffset + "`" +
				"\n**<:decisionmaking:1065935264352063559> 使用者是否可以自行設定生日日期:** `" + strconv.FormatBool(config.EveryoneCanSetBirthdayDate) + "`" +
				"\n <:Channel:994524759289233438> **通知頻道: <#" + config.ChannelID + ">**" +
				"\n <:roleplaying:985945121264635964> 身分組: " + role,
			Color: birthdaySuccessColor,
		}},
		AllowedMentions: &responses.AllowedMentions{},
	}
}

func stagedUnavailableMessage() responses.Message {
	return responses.Message{
		Embeds: []responses.Embed{{
			Title: "<a:Discord_AnimatedNo:1015989839809757295> | 此生日系統功能尚未在Go版本啟用",
			Color: birthdayErrorColor,
		}},
		AllowedMentions: &responses.AllowedMentions{},
	}
}

func deleteSuccessMessage(userID string) responses.Message {
	return responses.Message{
		Embeds: []responses.Embed{{
			Title:       "<:trashbin:995991389043163257> 刪除生日日期資料",
			Description: "<a:green_tick:994529015652163614> **你成功刪除了<@" + strings.TrimSpace(userID) + ">的資料!**",
			Color:       birthdaySuccessColor,
		}},
		AllowedMentions: &responses.AllowedMentions{},
	}
}

func allowAdminSuccessMessage(allow bool) responses.Message {
	return responses.Message{
		Embeds: []responses.Embed{{
			Title:       "<a:green_tick:994529015652163614> 成功變更資料",
			Description: "<a:green_tick:994529015652163614> **你成功將是否允許管理員設定生日資料設為**`" + strconv.FormatBool(allow) + "`!",
			Footer:      &responses.EmbedFooter{Text: "本人還是可以設定喔!"},
			Color:       birthdaySuccessColor,
		}},
		AllowedMentions: &responses.AllowedMentions{},
	}
}

func (m Module) cachedBirthdayUserTags(ctx context.Context, guildID string, profiles []domain.BirthdayProfile) map[string]string {
	tags := make(map[string]string, len(profiles))
	if m.cachedUsers == nil {
		return tags
	}
	for _, profile := range profiles {
		userID := strings.TrimSpace(profile.UserID)
		if userID == "" {
			continue
		}
		if _, exists := tags[userID]; exists {
			continue
		}
		info, ok, err := m.cachedUsers.CachedUserInfo(ctx, guildID, userID)
		if err == nil && ok {
			tags[userID] = info.Username + "#" + info.Discriminator
		}
	}
	return tags
}

func listMessage(profiles []domain.BirthdayProfile, cachedUserTags map[string]string, color int) responses.Message {
	fileLines := make([]string, 0, len(profiles))
	mentionLines := make([]string, 0, len(profiles))
	for _, profile := range profiles {
		date := profileDate(profile)
		userID := strings.TrimSpace(profile.UserID)
		userTag := "找不到使用者!"
		if cachedTag, ok := cachedUserTags[userID]; ok {
			userTag = cachedTag
		}
		fileLines = append(fileLines, userTag+"("+userID+")  | 生日日期(YYYY/MM/DD):"+date)
		mentionLines = append(mentionLines, "<@"+userID+">  | 生日日期(YYYY/MM/DD):"+date)
	}
	description := "<:list:992002476360343602>**目前共有**`" + strconv.Itoa(len(profiles)) + "`**人的生日數據**\n\n"
	if len(mentionLines) < 100 {
		description += "┃ " + strings.Join(mentionLines, "\n") + "┃"
	} else {
		description += "**由於人數過多，無法顯示所有成員名稱!\n請使用`.txt`檔案觀看**"
	}
	return responses.Message{
		Embeds: []responses.Embed{{
			Title:       "🎂 生日列表",
			Description: description,
			Color:       color,
		}},
		Files: []responses.File{{
			Name:        "discord.txt",
			ContentType: "text/plain; charset=utf-8",
			Data:        []byte(strings.Join(fileLines, "\n")),
		}},
		AllowedMentions: &responses.AllowedMentions{},
	}
}

func profileDate(profile domain.BirthdayProfile) string {
	return optionalIntString(profile.BirthdayYear) + "/" + optionalIntString(profile.BirthdayMonth) + "/" + optionalIntString(profile.BirthdayDay)
}

func optionalIntString(value *int) string {
	if value == nil {
		return "null"
	}
	return strconv.Itoa(*value)
}

func birthdayErrorFromError(err error) responses.Message {
	switch {
	case errors.Is(err, ports.ErrBirthdayConfigMissing):
		return birthdayErrorMessage("請先請管理員進行祝福語設定", birthdayDateAddDocsPath)
	case errors.Is(err, domain.ErrBirthdayManageMessagesRequired):
		return birthdayErrorMessage("你需要有`訊息管理`才能使用此指令", birthdayDateAddDocsPath)
	case errors.Is(err, domain.ErrBirthdaySelfOnly):
		return birthdayErrorMessage("你只擁有設定自己生日日期的權限!", birthdayDateAddDocsPath)
	case errors.Is(err, domain.ErrBirthdayAdminNotAllowed):
		return birthdayErrorMessage("該名使用者不允許管理員設定他的生日日期!", birthdayDateAddDocsPath)
	case errors.Is(err, domain.ErrInvalidBirthdayYear):
		return birthdayErrorMessage("請輸入有效的年份!", birthdayDateAddDocsPath)
	case errors.Is(err, domain.ErrInvalidBirthdayMonth):
		return birthdayErrorMessage("請輸入有效的月份!", birthdayDateAddDocsPath)
	case errors.Is(err, domain.ErrInvalidBirthdayDay):
		return birthdayErrorMessage("請輸入有效的日期!", birthdayDateAddDocsPath)
	case errors.Is(err, domain.ErrInvalidBirthdayTime):
		return birthdayErrorMessage("很抱歉，出現了未知的錯誤，請重試!", "")
	case errors.Is(err, domain.ErrInvalidBirthdayConfig):
		return birthdayErrorMessage("很抱歉，出現了未知的錯誤，請重試!", "")
	default:
		return birthdayErrorMessage("很抱歉，出現了未知的錯誤，請重試!", "")
	}
}

func birthdayProfileErrorMessage(err error, missingText string) responses.Message {
	switch {
	case errors.Is(err, ports.ErrBirthdayProfileMissing):
		return birthdayErrorMessage(missingText, birthdayDateAddDocsPath)
	case errors.Is(err, domain.ErrInvalidBirthdayProfile):
		return birthdayErrorMessage("很抱歉，出現了未知的錯誤，請重試!", "")
	default:
		return birthdayErrorMessage("很抱歉，出現了未知的錯誤，請重試!", "")
	}
}

func birthdayAddEphemeralError(content string) responses.Message {
	message := birthdayErrorMessage(content, birthdayDateAddDocsPath)
	message.Ephemeral = true
	return message
}

func birthdayAddEphemeralErrorMessage(message responses.Message) responses.Message {
	message.Ephemeral = true
	return message
}

func birthdayErrorMessage(content string, docsPath string) responses.Message {
	description := ""
	if strings.TrimSpace(docsPath) != "" {
		description = "<:MHCATdarkdsadsadsadsadsadas1:1079853990541529208> [立即前往官方文檔查看問題](https://docsmhcat.yorukot.me/" + docsPath + ")"
	}
	return responses.Message{
		Embeds: []responses.Embed{{
			Title:       "<a:Discord_AnimatedNo:1015989839809757295> | " + content,
			Description: description,
			Color:       birthdayErrorColor,
		}},
		AllowedMentions: &responses.AllowedMentions{},
	}
}

func legacyBirthdayHourOptions() []responses.SelectOption {
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

func legacyBirthdayMinuteOptions() []responses.SelectOption {
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

func birthdayAddCustomID(action string, stateID string) (string, error) {
	payload, err := customid.StateIDPayload(stateID)
	if err != nil {
		return "", err
	}
	return customid.Encode(customid.InteractionKindComponent, "birthday", action, payload)
}

func birthdayStateIDFromCustomID(raw string, action string) (string, bool) {
	parsed, err := customid.ParseComponent(raw)
	if err != nil || parsed.Version != customid.VersionV1 || parsed.Feature != "birthday" || parsed.Action != action {
		return "", false
	}
	if parsed.Payload.Kind != customid.PayloadState || strings.TrimSpace(parsed.Payload.StateID) == "" {
		return "", false
	}
	return parsed.Payload.StateID, true
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

func stringOption(interaction interactions.Interaction, name string) string {
	if value, ok := interaction.Options[name]; ok {
		return value
	}
	if option, ok := interaction.CommandOptions[name]; ok {
		return option.String
	}
	return ""
}

func intOption(interaction interactions.Interaction, name string) (int, bool) {
	if option, ok := interaction.CommandOptions[name]; ok && option.Type == interactions.CommandOptionInteger {
		return int(option.Int), true
	}
	value, ok := interaction.Options[name]
	if !ok {
		return 0, false
	}
	parsed, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil {
		return 0, false
	}
	return parsed, true
}

func optionalIntOption(interaction interactions.Interaction, name string) (*int, bool) {
	if option, ok := interaction.CommandOptions[name]; ok && option.Type == interactions.CommandOptionInteger {
		value := int(option.Int)
		return &value, true
	}
	value, ok := interaction.Options[name]
	if !ok || strings.TrimSpace(value) == "" {
		return nil, true
	}
	parsed, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil {
		return nil, false
	}
	return &parsed, true
}

func selectedInt(interaction interactions.Interaction) (int, bool) {
	if len(interaction.Values) == 0 {
		return 0, false
	}
	parsed, err := strconv.Atoi(strings.TrimSpace(interaction.Values[0]))
	if err != nil {
		return 0, false
	}
	return parsed, true
}

func boolOption(interaction interactions.Interaction, name string) (bool, bool) {
	if option, ok := interaction.CommandOptions[name]; ok && option.Type == interactions.CommandOptionBoolean {
		return option.Bool, true
	}
	value, ok := interaction.Options[name]
	if !ok {
		return false, false
	}
	parsed, err := strconv.ParseBool(strings.TrimSpace(value))
	if err != nil {
		return false, false
	}
	return parsed, true
}

func (m Module) track(ctx context.Context, interaction interactions.Interaction) error {
	if m.usage == nil {
		return nil
	}
	return m.usage.TrackCommand(ctx, ports.UsageEvent{
		CommandName: BirthdayCommandName,
		UserID:      interaction.Actor.UserID,
		GuildID:     interaction.Actor.GuildID,
		Feature:     "birthday-config",
	})
}
