package birthday

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
	birthdaySuccessColor     = 0x57F287
	birthdayErrorColor       = 0xED4245
)

func (m Module) Handler() interactions.Handler {
	return func(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
		if err := responder.Defer(ctx, responses.DeferOptions{}); err != nil {
			return err
		}
		switch interaction.Subcommand {
		case subcommandConfig:
			return m.handleConfig(ctx, interaction, responder)
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
		Message:                    firstOption(interaction, optionMessage),
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
		return responder.EditOriginal(ctx, birthdayErrorMessage("你需要有`訊息管理`才能使用此指令", "allcommands/生日系統/birthday_date_add"))
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
		return responder.EditOriginal(ctx, birthdayErrorMessage("還沒有任何人有進行生日設置喔!", "allcommands/生日系統/birthday_date_add"))
	}
	if err := responder.EditOriginal(ctx, listMessage(profiles)); err != nil {
		return err
	}
	return m.track(ctx, interaction)
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

func listMessage(profiles []domain.BirthdayProfile) responses.Message {
	fileLines := make([]string, 0, len(profiles))
	mentionLines := make([]string, 0, len(profiles))
	for _, profile := range profiles {
		date := profileDate(profile)
		userID := strings.TrimSpace(profile.UserID)
		fileLines = append(fileLines, "找不到使用者!("+userID+")  | 生日日期(YYYY/MM/DD):"+date)
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
			Color:       birthdaySuccessColor,
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
	case errors.Is(err, domain.ErrInvalidBirthdayConfig):
		return birthdayErrorMessage("很抱歉，出現了未知的錯誤，請重試!", "")
	default:
		return birthdayErrorMessage("很抱歉，出現了未知的錯誤，請重試!", "")
	}
}

func birthdayProfileErrorMessage(err error, missingText string) responses.Message {
	switch {
	case errors.Is(err, ports.ErrBirthdayProfileMissing):
		return birthdayErrorMessage(missingText, "allcommands/生日系統/birthday_date_add")
	case errors.Is(err, domain.ErrInvalidBirthdayProfile):
		return birthdayErrorMessage("很抱歉，出現了未知的錯誤，請重試!", "")
	default:
		return birthdayErrorMessage("很抱歉，出現了未知的錯誤，請重試!", "")
	}
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
