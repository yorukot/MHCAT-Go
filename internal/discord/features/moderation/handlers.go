package moderation

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/interactions"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/responses"
)

const (
	warningManageMessagesPermission = int64(8192)
	warningErrorColor               = 0xEA0000
	warningHistoryColor             = 0x5865F2
	warningSettingsSuccessColor     = 0x57F287
	warningRemovalSuccessColor      = 0x57F287
	warningRemovalDMColor           = 0x00DB00
)

func (m Module) WarningHistoryHandler() interactions.Handler {
	return func(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
		if err := responder.Defer(ctx, responses.DeferOptions{}); err != nil {
			return err
		}
		if !interaction.Actor.HasPermission(warningManageMessagesPermission) {
			return responder.EditOriginal(ctx, warningErrorMessage("你需要有`訊息管理`才能使用此指令"))
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
		return m.track(ctx, interaction, WarningHistoryCommandName, "warnings")
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
			Threshold: threshold,
			Action:    action,
		}
		if err := m.settings.Configure(ctx, settings); err != nil {
			return responder.EditOriginal(ctx, warningSettingsErrorMessage(err))
		}
		if err := responder.EditOriginal(ctx, warningSettingsMessage(settings)); err != nil {
			return err
		}
		return m.track(ctx, interaction, WarningSettingsCommandName, "warning-settings")
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
		return m.track(ctx, interaction, WarningRemoveCommandName, "warning-removal")
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
		return m.track(ctx, interaction, WarningRemoveAllCommandName, "warning-removal")
	}
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
			Color:       warningHistoryColor,
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

func warningSettingsMessage(settings domain.WarningSettings) responses.Message {
	return responses.Message{
		Embeds: []responses.Embed{{
			Title:       "警告系統",
			Description: fmt.Sprintf("警告成功設為警告%d次後\n執行%s", settings.Threshold, strings.TrimSpace(settings.Action)),
			Color:       warningSettingsSuccessColor,
		}},
		AllowedMentions: &responses.AllowedMentions{},
	}
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

func warningStringOption(interaction interactions.Interaction, name string) string {
	if value, ok := interaction.CommandOptions[name]; ok {
		return strings.TrimSpace(value.String)
	}
	return strings.TrimSpace(interaction.Options[name])
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

func (m Module) track(ctx context.Context, interaction interactions.Interaction, commandName string, feature string) error {
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
