package onboarding

import (
	"context"
	cryptorand "crypto/rand"
	"errors"
	"math/big"
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
	joinRoleSuccessColor     = 0x57F287
	joinRoleErrorColor       = 0xED4245
	legacyDashboardColor     = 0xDF1F2F
)

func (m Module) SetHandler() interactions.Handler {
	return func(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
		if err := responder.Defer(ctx, responses.DeferOptions{}); err != nil {
			return err
		}
		if !interaction.Actor.HasPermission(permissionManageMessages) {
			return responder.EditOriginal(ctx, joinRoleErrorMessage("你需要有`訊息管理`才能使用此指令"))
		}
		roleID := firstOption(interaction, "身分組")
		config := domain.JoinRoleConfig{
			GuildID: interaction.Actor.GuildID,
			RoleID:  roleID,
			GiveTo:  firstOption(interaction, "給人還是給機器人"),
		}
		if err := m.service.Create(ctx, config); err != nil {
			return responder.EditOriginal(ctx, joinRoleErrorFromError(err))
		}
		if err := responder.EditOriginal(ctx, joinRoleSetSuccessMessage(roleID)); err != nil {
			return err
		}
		return m.track(ctx, interaction, JoinRoleSetCommandName)
	}
}

func (m Module) DeleteHandler() interactions.Handler {
	return func(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
		if err := responder.Defer(ctx, responses.DeferOptions{}); err != nil {
			return err
		}
		if !interaction.Actor.HasPermission(permissionManageMessages) {
			return responder.EditOriginal(ctx, joinRoleErrorMessage("你需要有`訊息管理`才能使用此指令"))
		}
		roleID := firstOption(interaction, "身分組")
		if err := m.service.Delete(ctx, interaction.Actor.GuildID, roleID); err != nil {
			return responder.EditOriginal(ctx, joinRoleErrorFromError(err))
		}
		if err := responder.EditOriginal(ctx, joinRoleDeleteSuccessMessage(roleID)); err != nil {
			return err
		}
		return m.track(ctx, interaction, JoinRoleDeleteCommandName)
	}
}

func (m Module) JoinMessageDashboardHandler() interactions.Handler {
	return func(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
		if err := responder.Reply(ctx, joinMessageDashboardMessage(interaction.Actor.GuildID)); err != nil {
			return err
		}
		return m.trackFeature(ctx, interaction, JoinMessageSetCommandName, "welcome-message-config")
	}
}

func (m Module) LeaveMessagePromptHandler() interactions.Handler {
	return func(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
		if !interaction.Actor.HasPermission(permissionManageMessages) {
			return responder.Reply(ctx, leaveMessagePermissionError("你需要有`訊息管理`才能使用此指令"))
		}
		channelID := firstOption(interaction, "頻道")
		config, err := m.leaveMessageService.Prepare(ctx, interaction.Actor.GuildID, channelID)
		if err != nil {
			return responder.Reply(ctx, leaveMessageModalError("很抱歉，出現了未知的錯誤!"))
		}
		if err := responder.ShowModal(ctx, leaveMessageModal(config)); err != nil {
			return err
		}
		return m.trackFeature(ctx, interaction, LeaveMessageSetCommandName, "welcome-message-config")
	}
}

func (m Module) LeaveMessageModalHandler() interactions.Handler {
	return func(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
		if err := responder.Defer(ctx, responses.DeferOptions{}); err != nil {
			return err
		}
		fields := modalFieldMap(interaction.ModalFields)
		config := domain.LeaveMessageConfig{
			GuildID:        interaction.Actor.GuildID,
			MessageContent: fields["leave_msgcontent"],
			Color:          fields["leave_msgcolor"],
			Title:          fields["leave_msgtitle"],
		}
		if err := config.ValidateContent(); err != nil {
			return responder.EditOriginal(ctx, leaveMessageModalError(leaveMessageContentError(err)))
		}
		if err := m.leaveMessageService.Save(ctx, config); err != nil {
			return responder.EditOriginal(ctx, leaveMessageModalError(leaveMessageContentError(err)))
		}
		if err := responder.EditOriginal(ctx, leaveMessagePreviewMessage(config, interaction.Actor.AvatarURL, time.Now())); err != nil {
			return err
		}
		return m.trackFeature(ctx, interaction, "leave_msg", "welcome-message-config")
	}
}

func joinRoleSetSuccessMessage(roleID string) responses.Message {
	return responses.Message{
		Embeds: []responses.Embed{{
			Title:       "🪂 加入身分組系統",
			Description: "<a:green_tick:994529015652163614> **成功創建加入給身分組!**\n**身分組:** <@" + strings.TrimSpace(roleID) + ">!",
			Color:       joinRoleSuccessColor,
		}},
		AllowedMentions: &responses.AllowedMentions{},
	}
}

func joinRoleDeleteSuccessMessage(roleID string) responses.Message {
	return responses.Message{
		Embeds: []responses.Embed{{
			Title:       "🪂 加入身分組系統",
			Description: "<:trashbin:986308183674990592>**成功刪除:**\n身分組: <@" + strings.TrimSpace(roleID) + ">!",
			Color:       joinRoleSuccessColor,
		}},
		AllowedMentions: &responses.AllowedMentions{},
	}
}

func joinRoleErrorFromError(err error) responses.Message {
	switch {
	case errors.Is(err, ports.ErrDiscordRoleNotAssignable):
		return joinRoleErrorMessage("我沒有權限為大家增加這個身分組，請將我的身分組位階調高")
	case errors.Is(err, ports.ErrJoinRoleConfigExists):
		return joinRoleErrorMessage("很抱歉，這個身分組已經被註冊了，請重試!")
	case errors.Is(err, ports.ErrJoinRoleConfigMissing):
		return joinRoleErrorMessage("找不到這個身份組!")
	case errors.Is(err, domain.ErrInvalidJoinRoleConfig):
		return joinRoleErrorMessage("很抱歉，出現了未知的錯誤，請重試!")
	default:
		return joinRoleErrorMessage("很抱歉，出現了未知的錯誤，請重試!")
	}
}

func joinRoleErrorMessage(content string) responses.Message {
	return responses.Message{
		Embeds: []responses.Embed{{
			Title: "<a:Discord_AnimatedNo:1015989839809757295> | " + content,
			Color: joinRoleErrorColor,
		}},
		AllowedMentions: &responses.AllowedMentions{},
	}
}

func joinMessageDashboardMessage(guildID string) responses.Message {
	return responses.Message{
		Embeds: []responses.Embed{{
			Title: "<a:announcement:1005035747197337650> | 該指令已經移往控制面板，請前往控制面板進行設定",
			Color: legacyDashboardColor,
		}},
		Components: []responses.ComponentRow{{
			Components: []responses.Component{{
				Type:  responses.ComponentTypeButton,
				Style: responses.ButtonStyleLink,
				URL:   "https://mhcat.yorukot.meguilds/" + strings.TrimSpace(guildID) + "/welcome",
				Label: "點我前往儀錶板設定!",
				Emoji: "<a:arrow:986268851786375218>",
			}},
		}},
		AllowedMentions: &responses.AllowedMentions{},
	}
}

func leaveMessagePermissionError(content string) responses.Message {
	message := joinRoleErrorMessage(content)
	message.Ephemeral = true
	return message
}

func leaveMessageModalError(content string) responses.Message {
	return responses.Message{
		Embeds: []responses.Embed{{
			Title: content,
			Color: joinRoleErrorColor,
		}},
		AllowedMentions: &responses.AllowedMentions{},
	}
}

func leaveMessageContentError(err error) string {
	if errors.Is(err, domain.ErrInvalidLeaveMessageConfig) {
		return "你傳送的並不是顏色(色碼)"
	}
	return "很抱歉，出現了未知的錯誤!"
}

func leaveMessageModal(config domain.LeaveMessageConfig) responses.Modal {
	return responses.Modal{
		CustomID: "nal",
		Title:    "退出訊息設置!",
		Rows: []responses.ModalRow{
			{Inputs: []responses.TextInput{{
				CustomID: "leave_msgcolor",
				Label:    "請輸入你的加入訊息要甚麼顏色(要隨機顏色可輸入:Random)",
				Style:    responses.TextInputStyleShort,
				Required: true,
				Value:    config.Color,
			}}},
			{Inputs: []responses.TextInput{{
				CustomID: "leave_msgtitle",
				Label:    "請輸入訊息標題",
				Style:    responses.TextInputStyleShort,
				Required: true,
				Value:    config.Title,
			}}},
			{Inputs: []responses.TextInput{{
				CustomID: "leave_msgcontent",
				Label:    "請輸入訊息內文(如要顯示用戶名可輸入: {MEMBERNAME} )",
				Style:    responses.TextInputStyleParagraph,
				Required: true,
				Value:    config.MessageContent,
			}}},
		},
	}
}

func leaveMessagePreviewMessage(config domain.LeaveMessageConfig, avatarURL string, now time.Time) responses.Message {
	return responses.Message{
		Content: "下面為預覽，想修改嗎?再次輸入指令即可修改((MEMBERNAME)在到時候會變正常喔)",
		Embeds: []responses.Embed{{
			Title:       config.Title,
			Description: config.MessageContent,
			Color:       leaveMessagePreviewColor(config.Color),
			Timestamp:   now,
			Thumbnail:   &responses.EmbedImage{URL: avatarURL},
		}},
		AllowedMentions: &responses.AllowedMentions{},
	}
}

func leaveMessagePreviewColor(value string) int {
	if strings.TrimSpace(value) == "Random" {
		return randomEmbedColor()
	}
	if parsed, ok := domain.ParseLegacyColorValue(value); ok {
		return parsed
	}
	return joinRoleErrorColor
}

func randomEmbedColor() int {
	n, err := cryptorand.Int(cryptorand.Reader, big.NewInt(0x1000000))
	if err != nil {
		return 0x5865F2
	}
	return int(n.Int64())
}

func modalFieldMap(fields []customid.ModalField) map[string]string {
	result := make(map[string]string, len(fields))
	for _, field := range fields {
		if field.CustomID != "" {
			result[field.CustomID] = field.Value
		}
	}
	return result
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

func (m Module) track(ctx context.Context, interaction interactions.Interaction, commandName string) error {
	return m.trackFeature(ctx, interaction, commandName, "join-role-config")
}

func (m Module) trackFeature(ctx context.Context, interaction interactions.Interaction, commandName string, feature string) error {
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
