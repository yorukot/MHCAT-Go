package moderation

import (
	"context"
	cryptorand "crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/interactions"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/responses"
)

const (
	warningManageMessagesPermission = int64(8192)
	warningErrorColor               = 0xED4245
	warningSettingsSuccessColor     = 0x57F287
	warningRemovalSuccessColor      = 0x57F287
	warningRemovalDMColor           = 0x00DB00
	warningIssueSuccessColor        = 0x57F287
	warningIssueDMColor             = 0xEA0000
	cleanupAdminPermission          = int64(8)
	cleanupMaxMessages              = 1000
	cleanupAdminThreshold           = 200
	cleanupPermissionLabel          = "訊息管理(刪除超過200則需要有權限)"
	cleanupSuccessColor             = 0x53FF53
	deleteDataFallbackColor         = 0xEA0000
	deleteDataCollectorLifetime     = time.Hour
	discordEpochMilliseconds        = uint64(1420070400000)
)

func (m Module) WarningHistoryHandler() interactions.Handler {
	return func(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
		if err := responder.Defer(ctx, responses.DeferOptions{}); err != nil {
			return err
		}
		userID := strings.TrimSpace(interaction.Options["使用者"])
		if userID == "" {
			return responder.EditOriginal(ctx, warningErrorMessage("這位使用者沒有任何警告!"))
		}
		history, err := m.warnings.History(ctx, interaction.Actor.GuildID, userID)
		if err != nil {
			return responder.EditOriginal(ctx, warningHistoryErrorMessage(err))
		}
		targetName := m.targetUsername(ctx, interaction.Actor.GuildID, userID)
		moderatorTags := m.moderatorTags(ctx, interaction.Actor.GuildID, history)
		if err := responder.EditOriginal(ctx, warningHistoryMessage(history, targetName, moderatorTags)); err != nil {
			return err
		}
		return nil
	}
}

func (m Module) WarningSettingsHandler() interactions.Handler {
	return func(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
		if err := responder.Defer(ctx, responses.DeferOptions{}); err != nil {
			return err
		}
		if !interaction.Actor.HasPermission(warningManageMessagesPermission) {
			return responder.EditOriginal(ctx, warningErrorMessage("你需要有`訊息管理`才能使用此指令"))
		}
		action := warningStringOption(interaction, warningSettingsOptionAction)
		threshold, ok := warningIntegerOption(interaction, warningSettingsOptionThreshold)
		if !ok {
			return responder.EditOriginal(ctx, warningErrorMessage("很抱歉，出現了未知的錯誤，請重試!"))
		}
		settings := domain.WarningSettings{
			GuildID:   interaction.Actor.GuildID,
			Threshold: float64(threshold),
			Action:    action,
		}
		if err := m.settings.Configure(ctx, settings); err != nil {
			return responder.EditOriginal(ctx, warningSettingsErrorMessage(err))
		}
		if err := responder.EditOriginal(ctx, warningSettingsMessage(settings)); err != nil {
			return err
		}
		return nil
	}
}

func (m Module) WarningRemoveHandler() interactions.Handler {
	return func(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
		if err := responder.Defer(ctx, responses.DeferOptions{}); err != nil {
			return err
		}
		if !interaction.Actor.HasPermission(warningManageMessagesPermission) {
			return responder.EditOriginal(ctx, warningErrorMessage("你需要有`訊息管理`才能使用此指令"))
		}
		userID := warningStringOption(interaction, warningOptionUser)
		index, ok := warningIntegerOption(interaction, warningRemoveOptionIndex)
		if !ok {
			return responder.EditOriginal(ctx, warningErrorMessage("這位使用者沒有任何警告!"))
		}
		removal := domain.WarningRemoval{GuildID: interaction.Actor.GuildID, UserID: userID, Index: index}
		if err := m.removal.RemoveOne(ctx, removal); err != nil {
			return responder.EditOriginal(ctx, warningRemoveOneErrorMessage(err))
		}
		if err := responder.EditOriginal(ctx, warningRemovalSuccessMessage()); err != nil {
			return err
		}
		m.sendWarningRemovalDM(ctx, interaction, userID, false)
		return nil
	}
}

func (m Module) WarningRemoveAllHandler() interactions.Handler {
	return func(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
		if err := responder.Defer(ctx, responses.DeferOptions{}); err != nil {
			return err
		}
		if !interaction.Actor.HasPermission(warningManageMessagesPermission) {
			return responder.EditOriginal(ctx, warningErrorMessage("你需要有`訊息管理`才能使用此指令"))
		}
		userID := warningStringOption(interaction, warningOptionUser)
		removal := domain.WarningRemoval{GuildID: interaction.Actor.GuildID, UserID: userID}
		if err := m.removal.RemoveAll(ctx, removal); err != nil {
			return responder.EditOriginal(ctx, warningRemoveAllErrorMessage(err))
		}
		if err := responder.EditOriginal(ctx, warningRemovalSuccessMessage()); err != nil {
			return err
		}
		m.sendWarningRemovalDM(ctx, interaction, userID, true)
		return nil
	}
}

func (m Module) WarningIssueHandler() interactions.Handler {
	return func(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
		if err := responder.Defer(ctx, responses.DeferOptions{}); err != nil {
			return err
		}
		if !interaction.Actor.HasPermission(warningManageMessagesPermission) {
			return responder.EditOriginal(ctx, warningErrorMessage("你需要有`訊息管理`才能使用此指令"))
		}
		userID := warningStringOption(interaction, warningOptionUser)
		reason := warningRawStringOption(interaction, warningIssueOptionReason)
		if m.hierarchy != nil {
			allowed, err := m.hierarchy.ActorCanModerate(ctx, interaction.Actor.GuildID, interaction.Actor.RoleIDs, userID)
			if err != nil {
				return responder.EditOriginal(ctx, warningErrorMessage("很抱歉，出現了未知的錯誤，請重試!"))
			}
			if !allowed {
				return responder.EditOriginal(ctx, warningErrorMessage("你沒有權限警告這位使用者(身分組位階比他低)!"))
			}
		}
		result, err := m.issue.Issue(ctx, domain.WarningIssue{
			GuildID:     interaction.Actor.GuildID,
			UserID:      userID,
			ModeratorID: interaction.Actor.UserID,
			Reason:      reason,
			Time:        warningIssueTimestamp(m.clock),
		})
		if err != nil {
			return responder.EditOriginal(ctx, warningIssueErrorMessage(err))
		}
		if err := responder.EditOriginal(ctx, warningIssueSuccessMessage()); err != nil {
			return err
		}
		m.sendWarningIssueDM(ctx, interaction, userID, reason)
		if !result.Created {
			if message, failed := m.applyWarningThreshold(ctx, interaction, userID, len(result.History.Entries)); failed {
				return responder.EditOriginal(ctx, message)
			}
		}
		return nil
	}
}

func (m Module) CleanupHandler() interactions.Handler {
	return func(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
		if err := responder.Defer(ctx, responses.DeferOptions{Ephemeral: true}); err != nil {
			return err
		}
		if !interaction.Actor.HasPermission(warningManageMessagesPermission) {
			return responder.EditOriginal(ctx, cleanupErrorMessage(fmt.Sprintf("你需要有`%s`才能使用此指令", cleanupPermissionLabel)))
		}
		count, ok := warningIntegerOption(interaction, cleanupOptionCount)
		if !ok {
			return responder.EditOriginal(ctx, cleanupErrorMessage("很抱歉，出現了未知的錯誤，請重試!"))
		}
		if count <= 0 {
			return responder.EditOriginal(ctx, cleanupErrorMessage("很抱歉，出現了未知的錯誤，請重試!"))
		}
		if count > cleanupMaxMessages {
			return responder.EditOriginal(ctx, cleanupErrorMessage("不可刪除超過1000則消息!!!!!"))
		}
		if count > cleanupAdminThreshold && !interaction.Actor.HasPermission(cleanupAdminPermission) {
			return responder.EditOriginal(ctx, cleanupErrorMessage(fmt.Sprintf("你需要有`%s`才能使用此指令", cleanupPermissionLabel)))
		}
		if m.cleaner == nil {
			return responder.EditOriginal(ctx, cleanupErrorMessage("很抱歉，出現了未知的錯誤，請重試!"))
		}
		deleted, err := m.cleaner.CleanupMessages(ctx, ports.MessageCleanupRequest{
			ChannelID: interaction.ChannelID,
			Limit:     int(count),
			UserID:    warningStringOption(interaction, cleanupOptionUser),
		})
		if err != nil {
			return responder.EditOriginal(ctx, cleanupErrorMessage("很抱歉，出現了未知的錯誤，請重試!"))
		}
		if err := responder.EditOriginal(ctx, cleanupSuccessMessage(deleted, int(count))); err != nil {
			return err
		}
		return nil
	}
}

func (m Module) DeleteDataPromptHandler() interactions.Handler {
	return func(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
		if err := responder.Defer(ctx, responses.DeferOptions{}); err != nil {
			return err
		}
		if !interaction.Actor.HasPermission(warningManageMessagesPermission) {
			return responder.EditOriginal(ctx, warningErrorMessage("你需要有`訊息管理`才能使用此指令"))
		}
		if err := responder.FollowUp(ctx, deleteDataPromptMessage()); err != nil {
			return err
		}
		return nil
	}
}

func (m Module) DeleteDataSelectHandler() interactions.Handler {
	return func(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
		if err := responder.Defer(ctx, responses.DeferOptions{}); err != nil {
			return err
		}
		ownerID := strings.TrimSpace(interaction.OriginalInteractionUserID)
		if ownerID != "" {
			if ownerID != strings.TrimSpace(interaction.Actor.UserID) {
				return responder.EditOriginal(ctx, deleteDataContentMessage("<a:Discord_AnimatedNo:1015989839809757295> **| 你沒有設定過這個選項!**"))
			}
		} else if !interaction.Actor.HasPermission(warningManageMessagesPermission) {
			return responder.EditOriginal(ctx, deleteDataContentMessage("<a:Discord_AnimatedNo:1015989839809757295> **| 你需要有`訊息管理`才能使用此指令**"))
		}
		if m.deleteDataPromptExpired(interaction.OriginalInteractionID) {
			return responder.EditOriginal(ctx, deleteDataContentMessage("<a:Discord_AnimatedNo:1015989839809757295> **| 你沒有設定過這個選項!**"))
		}
		target, ok := selectedDeleteDataTarget(interaction)
		if !ok {
			return responder.EditOriginal(ctx, deleteDataContentMessage("<a:Discord_AnimatedNo:1015989839809757295> **| 你沒有設定過這個選項!**"))
		}
		if err := m.deleteData.Delete(ctx, domain.DeleteDataRequest{GuildID: interaction.Actor.GuildID, Target: target}); err != nil {
			if errors.Is(err, ports.ErrDeleteDataTargetMissing) {
				return responder.EditOriginal(ctx, deleteDataContentMessage("<a:Discord_AnimatedNo:1015989839809757295> **| 你沒有設定過這個選項!**"))
			}
			return responder.EditOriginal(ctx, deleteDataContentMessage("<a:Discord_AnimatedNo:1015989839809757295> **| 你沒有設定過這個選項!**"))
		}
		if err := responder.EditOriginal(ctx, deleteDataContentMessage("<a:green_tick:994529015652163614> **| 成功刪除該設定!**")); err != nil {
			return err
		}
		return nil
	}
}

func (m Module) deleteDataPromptExpired(originalInteractionID string) bool {
	id, err := strconv.ParseUint(strings.TrimSpace(originalInteractionID), 10, 64)
	if err != nil {
		return false
	}
	createdAt := time.UnixMilli(int64((id >> 22) + discordEpochMilliseconds))
	now := time.Now()
	if m.clock != nil {
		now = m.clock.Now()
	}
	return !now.Before(createdAt.Add(deleteDataCollectorLifetime))
}

func (m Module) targetUsername(ctx context.Context, guildID string, userID string) string {
	if m.discord == nil {
		return userID
	}
	info, err := m.discord.UserInfo(ctx, guildID, userID)
	if err != nil || strings.TrimSpace(info.Username) == "" {
		return userID
	}
	return info.Username
}

func (m Module) moderatorTags(ctx context.Context, guildID string, history domain.WarningHistory) map[string]string {
	ids := make([]string, 0, len(history.Entries))
	seen := map[string]struct{}{}
	for _, entry := range history.Entries {
		id := strings.TrimSpace(entry.ModeratorID)
		if id == "" {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		ids = append(ids, id)
	}
	if len(ids) == 0 || m.members == nil {
		return map[string]string{}
	}
	tags, err := m.members.MemberTags(ctx, guildID, ids)
	if err != nil {
		return map[string]string{}
	}
	return tags
}

func warningHistoryMessage(history domain.WarningHistory, targetName string, moderatorTags map[string]string) responses.Message {
	if strings.TrimSpace(targetName) == "" {
		targetName = history.UserID
	}
	lines := make([]string, 0, len(history.Entries))
	for index, entry := range history.Entries {
		moderator := strings.TrimSpace(moderatorTags[strings.TrimSpace(entry.ModeratorID)])
		if moderator == "" {
			moderator = strings.TrimSpace(entry.ModeratorID)
		}
		if moderator == "" {
			moderator = "未知"
		}
		lines = append(lines, fmt.Sprintf("\n%d ```- 警告者: %s\n- 原因: %s\n- 時間: %s```",
			index+1,
			moderator,
			entry.Reason,
			entry.Time,
		))
	}
	return responses.Message{
		Embeds: []responses.Embed{{
			Title:       fmt.Sprintf("以下是%s的警告紀錄", targetName),
			Description: strings.Join(lines, " "),
			Color:       discordRandomColor(),
		}},
		AllowedMentions: &responses.AllowedMentions{},
	}
}

func warningHistoryErrorMessage(err error) responses.Message {
	switch {
	case errors.Is(err, ports.ErrWarningsNotFound):
		return warningErrorMessage("這位使用者沒有任何警告!")
	case errors.Is(err, domain.ErrInvalidWarningQuery):
		return warningErrorMessage("這位使用者沒有任何警告!")
	default:
		return warningErrorMessage("很抱歉，出現了未知的錯誤，請重試!")
	}
}

func warningSettingsErrorMessage(err error) responses.Message {
	switch {
	case errors.Is(err, domain.ErrInvalidWarningSettings):
		return warningErrorMessage("很抱歉，出現了未知的錯誤，請重試!")
	case errors.Is(err, ports.ErrWarningSettingsUnavailable):
		return warningErrorMessage("很抱歉，出現了未知的錯誤，請重試!")
	default:
		return warningErrorMessage("很抱歉，出現了未知的錯誤，請重試!")
	}
}

func warningRemoveOneErrorMessage(err error) responses.Message {
	switch {
	case errors.Is(err, ports.ErrWarningsNotFound):
		return warningErrorMessage("這位使用者沒有任何警告!")
	case errors.Is(err, domain.ErrInvalidWarningRemoval):
		return warningErrorMessage("這位使用者沒有任何警告!")
	case errors.Is(err, ports.ErrWarningRemovalUnavailable):
		return warningErrorMessage("很抱歉，出現了未知的錯誤，請重試!")
	default:
		return warningErrorMessage("很抱歉，出現了未知的錯誤，請重試!")
	}
}

func warningRemoveAllErrorMessage(err error) responses.Message {
	switch {
	case errors.Is(err, ports.ErrWarningsNotFound):
		return warningErrorMessage("這位使用者沒有任何警告")
	case errors.Is(err, domain.ErrInvalidWarningRemoval):
		return warningErrorMessage("這位使用者沒有任何警告")
	case errors.Is(err, ports.ErrWarningRemovalUnavailable):
		return warningErrorMessage("很抱歉，出現了未知的錯誤，請重試!")
	default:
		return warningErrorMessage("很抱歉，出現了未知的錯誤，請重試!")
	}
}

func warningIssueErrorMessage(err error) responses.Message {
	switch {
	case errors.Is(err, domain.ErrInvalidWarningIssue):
		return warningErrorMessage("很抱歉，出現了未知的錯誤，請重試!")
	case errors.Is(err, ports.ErrWarningIssueUnavailable):
		return warningErrorMessage("很抱歉，出現了未知的錯誤，請重試!")
	default:
		return warningErrorMessage("很抱歉，出現了未知的錯誤，請重試!")
	}
}

func warningSettingsMessage(settings domain.WarningSettings) responses.Message {
	return responses.Message{
		Embeds: []responses.Embed{{
			Title:       "警告系統",
			Description: fmt.Sprintf("警告成功設為警告%s次後\n執行%s", warningThresholdText(settings.Threshold), strings.TrimSpace(settings.Action)),
			Color:       warningSettingsSuccessColor,
		}},
		AllowedMentions: &responses.AllowedMentions{},
	}
}

func warningIssueSuccessMessage() responses.Message {
	return responses.Message{
		Embeds: []responses.Embed{{
			Title: "<a:greentick:980496858445135893> | 成功警告這位使用者!",
			Color: warningIssueSuccessColor,
		}},
		AllowedMentions: &responses.AllowedMentions{},
	}
}

func (m Module) applyWarningThreshold(ctx context.Context, interaction interactions.Interaction, userID string, warningCount int) (responses.Message, bool) {
	if m.settings.Repository == nil {
		return responses.Message{}, false
	}
	settings, err := m.settings.Settings(ctx, interaction.Actor.GuildID)
	if err != nil {
		return responses.Message{}, false
	}
	if !(float64(warningCount) >= settings.Threshold) {
		return responses.Message{}, false
	}
	switch settings.Action {
	case domain.WarningSettingsActionBan:
		if m.memberActions == nil {
			return warningErrorMessage("我沒有權限ban掉他"), true
		}
		if err := m.memberActions.BanMember(ctx, interaction.Actor.GuildID, userID, "", 0); err != nil {
			return warningErrorMessage("我沒有權限ban掉他"), true
		}
		m.sendWarningActionMessage(ctx, interaction.ChannelID, domain.WarningSettingsActionBan)
	default:
		if m.memberActions == nil {
			return warningErrorMessage("我沒有權限踢出他"), true
		}
		if err := m.memberActions.KickMember(ctx, interaction.Actor.GuildID, userID, ""); err != nil {
			return warningErrorMessage("我沒有權限踢出他"), true
		}
		m.sendWarningActionMessage(ctx, interaction.ChannelID, domain.WarningSettingsActionKick)
	}
	return responses.Message{}, false
}

func warningThresholdText(threshold float64) string {
	return strconv.FormatFloat(threshold, 'f', -1, 64)
}

func (m Module) sendWarningActionMessage(ctx context.Context, channelID string, action string) {
	if m.messages == nil || strings.TrimSpace(channelID) == "" {
		return
	}
	_, _ = m.messages.SendMessage(ctx, channelID, ports.OutboundMessage{
		Embeds: []ports.OutboundEmbed{{
			Title: fmt.Sprintf("<a:greentick:980496858445135893> | 這位使用者已到達警告須執行條件，成功對他執行`%s`!", strings.TrimSpace(action)),
			Color: warningIssueSuccessColor,
		}},
		AllowedMentions: ports.AllowedMentions{},
	})
}

func warningRemovalSuccessMessage() responses.Message {
	return responses.Message{
		Embeds: []responses.Embed{{
			Title: "<a:greentick:980496858445135893> | 這位使用者的警告成功移除!",
			Color: warningRemovalSuccessColor,
		}},
		AllowedMentions: &responses.AllowedMentions{},
	}
}

func cleanupSuccessMessage(deleted int, requested int) responses.Message {
	return responses.Message{
		Embeds: []responses.Embed{{
			Title:       "<a:green_tick:994529015652163614> | 清理完成!",
			Description: fmt.Sprintf("**成功清除:**`%d`/`%d`\n**<:deletebutton:981971559679950848> 如果沒有成功清完全\n代表可能超過14天或沒這麼多訊息給清**", deleted, requested),
			Color:       cleanupSuccessColor,
		}},
		AllowedMentions: &responses.AllowedMentions{},
		Ephemeral:       true,
	}
}

func deleteDataPromptMessage() responses.Message {
	return responses.Message{
		Embeds: []responses.Embed{{
			Title:       "<:trashbin:995991389043163257> 刪除資料",
			Description: "<a:NukeExplosion:986558305885368321>這邊刪除的都是全刪!!!\n<:warning:985590881698590730> 一但刪除將__**無法復原**__，請三思!\n<:warning:985590881698590730> 一但刪除將__**無法復原**__，請三思!",
			Color:       discordRandomColor(),
			Footer: &responses.EmbedFooter{
				Text:    "請三思!!!",
				IconURL: "https://media.discordapp.net/attachments/991337796960784424/996749656161779853/6lnjr0.gif",
			},
			Thumbnail: &responses.EmbedImage{URL: "https://media.discordapp.net/attachments/991337796960784424/996749656161779853/6lnjr0.gif"},
		}},
		Components: []responses.ComponentRow{{Components: []responses.Component{{
			Type:        responses.ComponentTypeSelect,
			CustomID:    "delete-data",
			Placeholder: "🗑 選擇你要刪除的資料!",
			MinValues:   1,
			MaxValues:   1,
			Options:     deleteDataSelectOptions(),
		}}}},
		AllowedMentions: &responses.AllowedMentions{},
	}
}

func deleteDataSelectOptions() []responses.SelectOption {
	emojis := map[domain.DeleteDataTarget]string{
		domain.DeleteDataTargetJoinMessage:  "<:joines:953970547849592884>",
		domain.DeleteDataTargetLeaveMessage: "<:leaves:956444050792280084>",
		domain.DeleteDataTargetLogging:      "<:logfile:985948561625710663>",
		domain.DeleteDataTargetStats:        "<:statistics:986108146747600928>",
		domain.DeleteDataTargetTextXP:       "<:xp:990254386792005663>",
		domain.DeleteDataTargetVoiceXP:      "<:Voice:994844272790610011>",
		domain.DeleteDataTargetAutoChat:     "<:ChatBot:956863473910947850>",
		domain.DeleteDataTargetVerification: "<:tickmark:985949769224556614>",
		domain.DeleteDataTargetTicket:       "<:ticket:985945491093205073>",
	}
	targets := domain.LegacyDeleteDataTargets()
	options := make([]responses.SelectOption, 0, len(targets))
	for _, target := range targets {
		label := string(target)
		options = append(options, responses.SelectOption{
			Label:       label,
			Description: "🗑 " + label + " 刪除!",
			Value:       label,
			Emoji:       emojis[target],
		})
	}
	return options
}

func deleteDataContentMessage(content string) responses.Message {
	return responses.Message{
		Content:         content,
		AllowedMentions: &responses.AllowedMentions{},
	}
}

func selectedDeleteDataTarget(interaction interactions.Interaction) (domain.DeleteDataTarget, bool) {
	if len(interaction.Values) == 0 {
		return "", false
	}
	return domain.ParseDeleteDataTarget(interaction.Values[0])
}

func discordRandomColor() int {
	max := big.NewInt(0x1000000)
	value, err := cryptorand.Int(cryptorand.Reader, max)
	if err != nil {
		return deleteDataFallbackColor
	}
	return int(value.Int64())
}

func (m Module) sendWarningRemovalDM(ctx context.Context, interaction interactions.Interaction, userID string, all bool) {
	if m.direct == nil || strings.TrimSpace(userID) == "" {
		return
	}
	guildName := m.guildName(ctx, interaction.Actor.GuildID)
	scope := "一個__警告__"
	if all {
		scope = "所有__警告__"
	}
	_, _ = m.direct.SendDirectMessage(ctx, userID, ports.OutboundMessage{
		Embeds: []ports.OutboundEmbed{{
			Title:       "<:warning:985590881698590730> | 警告系統",
			Description: fmt.Sprintf("<:KannaSip:997764767433379850> **你在%s的%s被刪除了!**\n<:implementation:1002170846292488232> **執行者:**%s(id:%s)", guildName, scope, interaction.Actor.Username, interaction.Actor.UserID),
			Color:       warningRemovalDMColor,
		}},
		AllowedMentions: ports.AllowedMentions{},
	})
}

func (m Module) sendWarningIssueDM(ctx context.Context, interaction interactions.Interaction, userID string, reason string) {
	if m.direct == nil || strings.TrimSpace(userID) == "" {
		return
	}
	guildName := m.guildName(ctx, interaction.Actor.GuildID)
	_, _ = m.direct.SendDirectMessage(ctx, userID, ports.OutboundMessage{
		Embeds: []ports.OutboundEmbed{{
			Title:       "<:warning:985590881698590730> | 警告系統",
			Description: fmt.Sprintf("<:KannaSip:997764767433379850> **你在%s被__警告__了!**\n<:lightbulb:1002169670574546964> **原因:**%s\n<:implementation:1002170846292488232> **執行者:**%s(id:%s)", guildName, reason, interaction.Actor.Username, interaction.Actor.UserID),
			Color:       warningIssueDMColor,
		}},
		AllowedMentions: ports.AllowedMentions{},
	})
}

func (m Module) guildName(ctx context.Context, guildID string) string {
	if m.discord == nil {
		return guildID
	}
	info, err := m.discord.GuildInfo(ctx, guildID)
	if err != nil || strings.TrimSpace(info.Name) == "" {
		return guildID
	}
	return info.Name
}

func warningErrorMessage(content string) responses.Message {
	return responses.Message{
		Embeds: []responses.Embed{{
			Title: "<a:Discord_AnimatedNo:1015989839809757295> | " + content,
			Color: warningErrorColor,
		}},
		AllowedMentions: &responses.AllowedMentions{},
	}
}

func cleanupErrorMessage(content string) responses.Message {
	message := warningErrorMessage(content)
	message.Ephemeral = true
	return message
}

func warningStringOption(interaction interactions.Interaction, name string) string {
	if value, ok := interaction.CommandOptions[name]; ok {
		return strings.TrimSpace(value.String)
	}
	return strings.TrimSpace(interaction.Options[name])
}

func warningRawStringOption(interaction interactions.Interaction, name string) string {
	if value, ok := interaction.CommandOptions[name]; ok {
		return value.String
	}
	return interaction.Options[name]
}

func warningIntegerOption(interaction interactions.Interaction, name string) (int64, bool) {
	if value, ok := interaction.CommandOptions[name]; ok {
		if value.Type == interactions.CommandOptionInteger {
			return value.Int, true
		}
		if strings.TrimSpace(value.String) != "" {
			parsed, err := strconv.ParseInt(strings.TrimSpace(value.String), 10, 64)
			return parsed, err == nil
		}
	}
	parsed, err := strconv.ParseInt(strings.TrimSpace(interaction.Options[name]), 10, 64)
	return parsed, err == nil
}

func warningIssueTimestamp(clock ports.Clock) string {
	now := time.Now()
	if clock != nil {
		now = clock.Now()
	}
	location, err := time.LoadLocation("Asia/Taipei")
	if err != nil {
		location = time.FixedZone("Asia/Taipei", 8*60*60)
	}
	return now.In(location).Format("2006年01月02日 15點04分")
}
