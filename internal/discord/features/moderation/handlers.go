package moderation

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
	warningManageMessagesPermission = int64(8192)
	warningErrorColor               = 0xEA0000
	warningHistoryColor             = 0x5865F2
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
		return m.track(ctx, interaction)
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

func warningErrorMessage(content string) responses.Message {
	return responses.Message{
		Embeds: []responses.Embed{{
			Title: "<a:Discord_AnimatedNo:1015989839809757295> | " + content,
			Color: warningErrorColor,
		}},
		AllowedMentions: &responses.AllowedMentions{},
	}
}

func (m Module) track(ctx context.Context, interaction interactions.Interaction) error {
	if m.usage == nil {
		return nil
	}
	return m.usage.TrackCommand(ctx, ports.UsageEvent{
		CommandName: WarningHistoryCommandName,
		UserID:      interaction.Actor.UserID,
		GuildID:     interaction.Actor.GuildID,
		Feature:     "warnings",
	})
}
