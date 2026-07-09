package xp

import (
	"context"
	cryptorand "crypto/rand"
	"errors"
	"fmt"
	"math"
	"math/big"
	"regexp"
	"strconv"
	"strings"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/interactions"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/responses"
)

const (
	rewardRolePageSize       = 12
	rewardRoleSuccessColor   = 0x53FF53
	rewardRoleRandomFallback = 0x5865F2
	channelEmoji             = "<:Channel:994524759289233438>"
	deleteEmoji              = "<:trashbin:995991389043163257>"
	doneEmoji                = "<a:green_tick:994529015652163614>"
	levelEmoji               = "<:levelup:990254382845157406>"
	roleEmoji                = "<:roleplaying:985945121264635964>"
	previousEmoji            = "<:previous:986067803910066256>"
	nextEmoji                = "<:next:986067802056167495>"
)

var legacyRewardRolePageRe = regexp.MustCompile(`^([0-9]+)(text_leave_role|voice_leave_role)$`)

func (m RewardRoleModule) TextHandler() interactions.Handler {
	return m.commandHandler(TextXPRewardRoleCommandName, "聊天", true)
}

func (m RewardRoleModule) VoiceHandler() interactions.Handler {
	return m.commandHandler(VoiceXPRewardRoleCommandName, "語音", false)
}

func (m RewardRoleModule) TextPageHandler() interactions.Handler {
	return m.pageHandler("聊天", true)
}

func (m RewardRoleModule) VoicePageHandler() interactions.Handler {
	return m.pageHandler("語音", false)
}

func (m RewardRoleModule) commandHandler(commandName string, label string, text bool) interactions.Handler {
	return func(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
		if err := responder.Defer(ctx, responses.DeferOptions{}); err != nil {
			return err
		}
		if !interaction.Actor.HasPermission(permissionManageMessages) {
			return responder.EditOriginal(ctx, textXPErrorMessage("你需要有`訊息管理`才能使用此指令"))
		}
		switch interaction.Subcommand {
		case "增加":
			return m.addRewardRole(ctx, interaction, responder, commandName, label, text)
		case "刪除":
			return m.deleteRewardRole(ctx, interaction, responder, commandName, label, text)
		case "設定查詢":
			return m.queryRewardRoles(ctx, interaction, responder, commandName, label, text, 0)
		default:
			return responder.EditOriginal(ctx, textXPErrorMessage("很抱歉，出現了未知的錯誤，請重試!"))
		}
	}
}

func (m RewardRoleModule) pageHandler(label string, text bool) interactions.Handler {
	return func(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
		page := rewardRolePageFromCustomID(interaction.CustomID)
		configs, err := m.listRewardRoles(ctx, interaction.Actor.GuildID, text)
		if err != nil {
			return responder.UpdateMessage(ctx, textXPErrorMessage("很抱歉，出現了未知的錯誤，請重試!"))
		}
		return responder.UpdateMessage(ctx, rewardRoleListMessage(label, configs, page, text, m.nextColor()))
	}
}

func (m RewardRoleModule) addRewardRole(ctx context.Context, interaction interactions.Interaction, responder responses.Responder, commandName string, label string, text bool) error {
	config := domain.XPRewardRoleConfig{
		GuildID:       interaction.Actor.GuildID,
		Level:         rewardRoleLevel(interaction),
		RoleID:        firstOption(interaction, "身分組"),
		DeleteWhenNot: rewardRoleDeleteWhenNot(interaction),
	}
	var err error
	if text {
		err = m.textService.Add(ctx, config)
	} else {
		err = m.voiceService.Add(ctx, config)
	}
	if err != nil {
		return responder.EditOriginal(ctx, rewardRoleErrorMessage(err))
	}
	if err := responder.EditOriginal(ctx, rewardRoleSaveMessage(label)); err != nil {
		return err
	}
	return m.track(ctx, interaction, commandName, rewardRoleFeature(text))
}

func (m RewardRoleModule) deleteRewardRole(ctx context.Context, interaction interactions.Interaction, responder responses.Responder, commandName string, label string, text bool) error {
	level := rewardRoleLevel(interaction)
	roleID := firstOption(interaction, "身分組")
	var err error
	if text {
		err = m.textService.Delete(ctx, interaction.Actor.GuildID, level, roleID)
	} else {
		err = m.voiceService.Delete(ctx, interaction.Actor.GuildID, level, roleID)
	}
	if err != nil {
		return responder.EditOriginal(ctx, rewardRoleErrorMessage(err))
	}
	if err := responder.EditOriginal(ctx, rewardRoleDeleteMessage(label)); err != nil {
		return err
	}
	return m.track(ctx, interaction, commandName, rewardRoleFeature(text))
}

func (m RewardRoleModule) queryRewardRoles(ctx context.Context, interaction interactions.Interaction, responder responses.Responder, commandName string, label string, text bool, page int) error {
	configs, err := m.listRewardRoles(ctx, interaction.Actor.GuildID, text)
	if err != nil {
		return responder.EditOriginal(ctx, textXPErrorMessage("很抱歉，出現了未知的錯誤，請重試!"))
	}
	if err := responder.EditOriginal(ctx, rewardRoleListMessage(label, configs, page, text, m.nextColor())); err != nil {
		return err
	}
	return m.track(ctx, interaction, commandName, rewardRoleFeature(text))
}

func (m RewardRoleModule) listRewardRoles(ctx context.Context, guildID string, text bool) ([]domain.XPRewardRoleConfig, error) {
	if text {
		return m.textService.List(ctx, guildID)
	}
	return m.voiceService.List(ctx, guildID)
}

func rewardRoleSaveMessage(label string) responses.Message {
	return responses.Message{
		Embeds: []responses.Embed{{
			Title:       channelEmoji + label + "經驗系統",
			Description: doneEmoji + "成功`增加`/`修改`該設定",
			Color:       rewardRoleSuccessColor,
		}},
		AllowedMentions: &responses.AllowedMentions{},
	}
}

func rewardRoleDeleteMessage(label string) responses.Message {
	return responses.Message{
		Embeds: []responses.Embed{{
			Title:       deleteEmoji + label + "經驗系統",
			Description: doneEmoji + "成功刪除該設定",
			Color:       rewardRoleSuccessColor,
		}},
		AllowedMentions: &responses.AllowedMentions{},
	}
}

func rewardRoleListMessage(label string, configs []domain.XPRewardRoleConfig, page int, text bool, color int) responses.Message {
	if page < 0 {
		page = 0
	}
	totalPages := int(math.Ceil(float64(len(configs)) / float64(rewardRolePageSize)))
	if totalPages > 0 && page >= totalPages {
		page = totalPages - 1
	}
	start := page * rewardRolePageSize
	end := start + rewardRolePageSize
	if end > len(configs) {
		end = len(configs)
	}
	fields := make([]responses.EmbedField, 0, rewardRolePageSize)
	for index, config := range configs[start:end] {
		number := start + index + 1
		fields = append(fields, responses.EmbedField{
			Name:   fmt.Sprintf("第%d個:", number),
			Value:  fmt.Sprintf("%s **等級:**`%d`\n%s **身分組:**<@&%s>\n%s **是否自動刪除身分組:**%t", levelEmoji, config.Level, roleEmoji, config.RoleID, deleteEmoji, config.DeleteWhenNot),
			Inline: true,
		})
	}
	kind := "text"
	if !text {
		kind = "voice"
	}
	return responses.Message{
		Embeds: []responses.Embed{{
			Title:  channelEmoji + " 以下是" + label + "經驗身分組的所有設定!!",
			Color:  color,
			Fields: fields,
			Footer: &responses.EmbedFooter{Text: fmt.Sprintf("總共: %d 筆資料\n第 %d / %d 頁(按按鈕會自動更新喔!", len(configs), page+1, totalPages)},
		}},
		Components: []responses.ComponentRow{{
			Components: []responses.Component{
				{Type: responses.ComponentTypeButton, CustomID: fmt.Sprintf("%d%s_leave_role", page-1, kind), Emoji: previousEmoji, Label: "上一頁", Style: responses.ButtonStyleSuccess, Disabled: page-1 == -1},
				{Type: responses.ComponentTypeButton, CustomID: fmt.Sprintf("%d%s_leave_role", page+1, kind), Emoji: nextEmoji, Label: "下一頁", Style: responses.ButtonStyleSuccess, Disabled: totalPages == 0 || page+1 >= totalPages},
			},
		}},
		AllowedMentions: &responses.AllowedMentions{},
	}
}

func rewardRoleErrorMessage(err error) responses.Message {
	switch {
	case errors.Is(err, ports.ErrDiscordRoleNotAssignable):
		return textXPErrorMessage("我沒有權限給大家這個身分組(請把我的身分組調高)!")
	case errors.Is(err, ports.ErrXPRewardRoleLimitExceeded):
		return textXPErrorMessage("你的設定已經過多，請先刪除一些!")
	case errors.Is(err, ports.ErrTextXPRewardRoleMissing), errors.Is(err, ports.ErrVoiceXPRewardRoleMissing):
		return textXPErrorMessage("你沒有設定過這個選項!")
	default:
		return textXPErrorMessage("很抱歉，出現了未知的錯誤，請重試!")
	}
}

func rewardRoleLevel(interaction interactions.Interaction) int64 {
	if option, ok := interaction.CommandOptions["等級"]; ok {
		if option.Type == interactions.CommandOptionInteger {
			return option.Int
		}
		if value := strings.TrimSpace(option.String); value != "" {
			parsed, _ := strconv.ParseInt(value, 10, 64)
			return parsed
		}
	}
	parsed, _ := strconv.ParseInt(strings.TrimSpace(interaction.Options["等級"]), 10, 64)
	return parsed
}

func rewardRoleDeleteWhenNot(interaction interactions.Interaction) bool {
	if option, ok := interaction.CommandOptions["是否自動刪除"]; ok {
		if option.Type == interactions.CommandOptionBoolean {
			return option.Bool
		}
		value := strings.TrimSpace(strings.ToLower(option.String))
		return value == "true" || value == "1"
	}
	value := strings.TrimSpace(strings.ToLower(interaction.Options["是否自動刪除"]))
	return value == "true" || value == "1"
}

func rewardRolePageFromCustomID(customID string) int {
	matches := legacyRewardRolePageRe.FindStringSubmatch(customID)
	if matches == nil {
		return 0
	}
	page, err := strconv.Atoi(matches[1])
	if err != nil || page < 0 {
		return 0
	}
	return page
}

func rewardRoleFeature(text bool) string {
	if text {
		return "text-xp-role-config"
	}
	return "voice-xp-role-config"
}

func randomXPColor() int {
	max := big.NewInt(0xFFFFFF + 1)
	value, err := cryptorand.Int(cryptorand.Reader, max)
	if err != nil {
		return rewardRoleRandomFallback
	}
	return int(value.Int64())
}

func (m RewardRoleModule) nextColor() int {
	if m.color == nil {
		return rewardRoleRandomFallback
	}
	return m.color()
}

func (m RewardRoleModule) track(ctx context.Context, interaction interactions.Interaction, commandName string, feature string) error {
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
