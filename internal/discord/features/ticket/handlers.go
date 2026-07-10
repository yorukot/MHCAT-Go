package ticket

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/customid"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/interactions"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/responses"
)

var ErrTicketRepositoryNotConfigured = errors.New("ticket repository is not configured")
var ErrTicketSideEffectsNotConfigured = errors.New("ticket side-effect ports are not configured")

const (
	discordChannelTypeGuildText = 0

	permissionOverwriteRole   = 0
	permissionOverwriteMember = 1

	permissionCreateInstantInvite = 1
	permissionViewChannel         = 1024
	permissionSendMessages        = 2048
	permissionManageMessages      = 8192
	permissionReadMessageHistory  = 65536

	legacyDiscordNamedRed   = 0xED4245
	legacyDiscordNamedGreen = 0x57F287
	legacyTicketOpenGreen   = 0x00DB00
)

func (m Module) SetupHandler() interactions.Handler {
	return func(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
		if m.repo == nil {
			return ErrTicketRepositoryNotConfigured
		}
		if !interaction.Actor.HasPermission(permissionManageMessages) {
			return responder.Reply(ctx, ticketPermissionDeniedMessage("訊息管理"))
		}
		categoryID := firstOption(interaction, "類別", "category")
		adminRoleID := firstOption(interaction, "管理員身分組", "admin_role")
		if categoryID == "" || adminRoleID == "" {
			return responder.Reply(ctx, ticketErrorMessage("缺少必要的類別或管理員身分組。"))
		}
		if _, err := m.repo.GetTicketConfig(ctx, interaction.Actor.GuildID); err == nil {
			return responder.Reply(ctx, ticketDuplicateConfigMessage())
		} else if err != nil && !errors.Is(err, ports.ErrTicketConfigNotFound) {
			return err
		}
		payload, err := customid.KeyValuePayload(map[string]string{"c": categoryID, "r": adminRoleID})
		if err != nil {
			return err
		}
		modalID, err := customid.Encode(customid.InteractionKindModal, "ticket", "setup", payload)
		if err != nil {
			return err
		}
		return responder.ShowModal(ctx, ticketSetupModal(modalID))
	}
}

func ticketSetupModal(customID string) responses.Modal {
	return responses.Modal{
		CustomID: customID,
		Title:    "私人頻道系統!",
		Rows: []responses.ModalRow{
			{Inputs: []responses.TextInput{{
				CustomID: "ticketcolor",
				Label:    "請輸入嵌入顏色",
				Style:    responses.TextInputStyleShort,
				Required: true,
			}}},
			{Inputs: []responses.TextInput{{
				CustomID: "tickettitle",
				Label:    "請輸入標題",
				Style:    responses.TextInputStyleShort,
				Required: true,
			}}},
			{Inputs: []responses.TextInput{{
				CustomID: "ticketcontent",
				Label:    "請輸入內文",
				Style:    responses.TextInputStyleParagraph,
				Required: true,
			}}},
		},
	}
}

func (m Module) SetupModalHandler() interactions.Handler {
	return func(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
		if m.repo == nil {
			return ErrTicketRepositoryNotConfigured
		}
		parsed, err := customid.ParseModal(interaction.CustomID, interaction.ModalFields)
		if err != nil {
			return err
		}
		categoryID := parsed.Payload.Values["c"]
		adminRoleID := parsed.Payload.Values["r"]
		return m.submitTicketPanel(ctx, interaction, responder, func(ctx context.Context) error {
			config := domain.TicketConfig{
				GuildID:        interaction.Actor.GuildID,
				CategoryID:     categoryID,
				AdminRoleID:    adminRoleID,
				EveryoneRoleID: interaction.Actor.GuildID,
			}
			return m.repo.CreateTicketConfig(ctx, config)
		})
	}
}

func (m Module) LegacyPanelSubmitHandler() interactions.Handler {
	return func(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
		return m.submitTicketPanel(ctx, interaction, responder, nil)
	}
}

func (m Module) submitTicketPanel(ctx context.Context, interaction interactions.Interaction, responder responses.Responder, saveConfig func(context.Context) error) error {
	if m.messages == nil {
		return ErrTicketSideEffectsNotConfigured
	}
	if interaction.ChannelID == "" {
		return responder.Reply(ctx, ticketErrorMessage("找不到要傳送私人頻道面板的頻道。"))
	}
	if err := responder.Defer(ctx, responses.DeferOptions{}); err != nil {
		return err
	}
	fields := modalFieldMap(interaction.ModalFields)
	color, ok := parseLegacyColor(fields["ticketcolor"])
	if !ok {
		return responder.EditOriginal(ctx, ticketEditErrorMessage("你傳送的並不是顏色(色碼)"))
	}
	title := fields["tickettitle"]
	content := fields["ticketcontent"]
	if title == "" || content == "" {
		return responder.EditOriginal(ctx, ticketEditErrorMessage("請完整填寫私人頻道標題與內文。"))
	}
	if saveConfig != nil {
		if err := saveConfig(ctx); err != nil {
			if errors.Is(err, ports.ErrTicketConfigExists) {
				return responder.EditOriginal(ctx, ticketDuplicateConfigEditMessage())
			}
			return err
		}
	}
	if _, err := m.messages.SendMessage(ctx, interaction.ChannelID, ticketPanelOutboundMessage(title, content, color)); err != nil {
		return err
	}
	if err := responder.EditOriginal(ctx, ticketSetupSuccessMessage()); err != nil {
		return err
	}
	return nil
}

func (m Module) DeleteHandler() interactions.Handler {
	return func(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
		if m.repo == nil {
			return ErrTicketRepositoryNotConfigured
		}
		if err := responder.Defer(ctx, responses.DeferOptions{}); err != nil {
			return err
		}
		if !interaction.Actor.HasPermission(permissionManageMessages) {
			return responder.EditOriginal(ctx, ticketPermissionDeniedEditMessage("訊息管理"))
		}
		err := m.repo.DeleteTicketConfig(ctx, interaction.Actor.GuildID)
		description := "成功刪除私人頻道的設置\n現在你可以重新創建了!"
		if errors.Is(err, ports.ErrTicketConfigNotFound) {
			description = "你還沒有創建私人頻道的設定\n是要怎麼刪除啦!"
		} else if err != nil {
			return err
		}
		if err := responder.EditOriginal(ctx, responses.Message{
			Embeds: []responses.Embed{{
				Title:       "刪除私人頻道設定",
				Description: description,
				Color:       legacyDiscordNamedRed,
			}},
		}); err != nil {
			return err
		}
		return nil
	}
}

func (m Module) OpenHandler() interactions.Handler {
	return func(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
		if m.repo == nil {
			return ErrTicketRepositoryNotConfigured
		}
		if m.channels == nil || m.messages == nil {
			return ErrTicketSideEffectsNotConfigured
		}
		if existing, err := m.channels.FindChannelByName(ctx, interaction.Actor.GuildID, interaction.Actor.UserID, -1); err == nil && existing.ChannelID != "" {
			return responder.Reply(ctx, ticketAlreadyOpenMessage())
		} else if err != nil && !errors.Is(err, ports.ErrChannelNotFound) {
			return err
		}

		config, err := m.repo.GetTicketConfig(ctx, interaction.Actor.GuildID)
		if errors.Is(err, ports.ErrTicketConfigNotFound) {
			if err := responder.Reply(ctx, ticketDeletedConfigMessage()); err != nil {
				return err
			}
			if interaction.MessageID != "" {
				_ = m.messages.DeleteMessage(ctx, ports.MessageRef{ChannelID: interaction.ChannelID, MessageID: interaction.MessageID})
			}
			return nil
		}
		if err != nil {
			return err
		}

		created, err := m.channels.CreateChannel(ctx, ports.ChannelCreateRequest{
			GuildID:              interaction.Actor.GuildID,
			ParentID:             config.CategoryID,
			Name:                 interaction.Actor.UserID,
			Type:                 discordChannelTypeGuildText,
			PermissionOverwrites: m.ticketOpenPermissionOverwrites(config, interaction.Actor.GuildID, interaction.Actor.UserID, ticketBotUserID(interaction.ApplicationID, m.botUserID)),
		})
		if err != nil {
			return err
		}
		if _, err := m.messages.SendMessage(ctx, created.ChannelID, ticketWelcomeMessage()); err != nil {
			return err
		}
		if err := responder.Reply(ctx, ticketOpenSuccessMessage()); err != nil {
			return err
		}
		return nil
	}
}

func (m Module) CloseHandler() interactions.Handler {
	return func(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
		if m.channels == nil {
			return ErrTicketSideEffectsNotConfigured
		}
		if interaction.ChannelID == "" {
			return responder.Reply(ctx, ticketCloseDeniedMessage())
		}
		if !interaction.Actor.HasPermission(permissionManageMessages) {
			channelName := interaction.ChannelName
			if channelName == "" {
				channel, err := m.channels.FindChannelByID(ctx, interaction.Actor.GuildID, interaction.ChannelID)
				if errors.Is(err, ports.ErrChannelNotFound) {
					return responder.Reply(ctx, ticketCloseDeniedMessage())
				}
				if err != nil {
					return err
				}
				channelName = channel.Name
			}
			if channelName != interaction.Actor.UserID {
				return responder.Reply(ctx, ticketCloseDeniedMessage())
			}
		}
		if err := m.channels.DeleteChannel(ctx, interaction.ChannelID); err != nil {
			return err
		}
		return nil
	}
}

func (m Module) ticketOpenPermissionOverwrites(config domain.TicketConfig, guildID string, userID string, botUserID string) []ports.PermissionOverwrite {
	allow := int64(permissionViewChannel | permissionSendMessages | permissionReadMessageHistory)
	denyInvite := int64(permissionCreateInstantInvite)
	overwrites := []ports.PermissionOverwrite{
		{ID: config.AdminRoleID, Type: permissionOverwriteRole, Allow: allow, Deny: denyInvite},
		{ID: guildID, Type: permissionOverwriteRole, Deny: int64(permissionViewChannel)},
		{ID: userID, Type: permissionOverwriteMember, Allow: allow, Deny: denyInvite},
	}
	if botUserID != "" {
		overwrites = append(overwrites, ports.PermissionOverwrite{ID: botUserID, Type: permissionOverwriteMember, Allow: allow, Deny: denyInvite})
	}
	return overwrites
}

func ticketBotUserID(applicationID string, fallbackID string) string {
	if applicationID = strings.TrimSpace(applicationID); applicationID != "" {
		return applicationID
	}
	return strings.TrimSpace(fallbackID)
}

func firstOption(interaction interactions.Interaction, names ...string) string {
	for _, name := range names {
		if value := strings.TrimSpace(interaction.Options[name]); value != "" {
			return value
		}
	}
	return ""
}

func modalFieldMap(fields []customid.ModalField) map[string]string {
	result := map[string]string{}
	for _, field := range fields {
		if field.CustomID != "" {
			result[field.CustomID] = field.Value
		}
	}
	return result
}

func ticketPanelMessage(title string, content string, color int) responses.Message {
	return responses.Message{
		Embeds: []responses.Embed{{
			Title:       title,
			Description: content,
			Color:       color,
		}},
		Components: []responses.ComponentRow{{Components: []responses.Component{{
			Type:     responses.ComponentTypeButton,
			CustomID: "tic",
			Label:    "🎫 點我創建客服頻道!",
			Style:    responses.ButtonStylePrimary,
		}}}},
	}
}

func ticketPanelOutboundMessage(title string, content string, color int) ports.OutboundMessage {
	return ports.OutboundMessage{
		Embeds: []ports.OutboundEmbed{{
			Title:       title,
			Description: content,
			Color:       color,
		}},
		Components: []ports.OutboundComponentRow{{Components: []ports.OutboundComponent{{
			Type:     "button",
			CustomID: "tic",
			Label:    "🎫 點我創建客服頻道!",
			Style:    "primary",
		}}}},
	}
}

func ticketErrorMessage(content string) responses.Message {
	return responses.Message{
		Ephemeral: true,
		Embeds: []responses.Embed{{
			Title: content,
			Color: legacyDiscordNamedRed,
		}},
	}
}

func ticketEditErrorMessage(content string) responses.Message {
	return responses.Message{
		Embeds: []responses.Embed{{
			Title: content,
			Color: legacyDiscordNamedRed,
		}},
	}
}

func ticketPermissionDeniedMessage(permission string) responses.Message {
	return responses.Message{
		Ephemeral: true,
		Embeds: []responses.Embed{{
			Title: "<a:Discord_AnimatedNo:1015989839809757295> | 你需要有`" + permission + "`才能使用此指令",
			Color: legacyDiscordNamedRed,
		}},
	}
}

func ticketPermissionDeniedEditMessage(permission string) responses.Message {
	message := ticketPermissionDeniedMessage(permission)
	message.Ephemeral = false
	return message
}

func ticketDuplicateConfigMessage() responses.Message {
	return responses.Message{
		Ephemeral: true,
		Embeds: []responses.Embed{{
			Title:       "__**錯誤**__",
			Description: "您已經註冊一個私人頻道了，如果需要更改，請打\n`<>h 刪除私人頻道`\n將會告訴您如何刪除",
			Color:       legacyDiscordNamedRed,
		}},
	}
}

func ticketDuplicateConfigEditMessage() responses.Message {
	message := ticketDuplicateConfigMessage()
	message.Ephemeral = false
	return message
}

func ticketSetupSuccessMessage() responses.Message {
	return responses.Message{
		Embeds: []responses.Embed{{
			Title: "<a:green_tick:994529015652163614> | 成功創建私人頻道",
			Color: legacyDiscordNamedGreen,
		}},
	}
}

func ticketAlreadyOpenMessage() responses.Message {
	return responses.Message{
		Ephemeral: true,
		Embeds: []responses.Embed{{
			Title:       "__**客服頻道**__",
			Description: ":warning: 你已經有一個客服頻道了!",
			Color:       legacyDiscordNamedRed,
		}},
	}
}

func ticketDeletedConfigMessage() responses.Message {
	return responses.Message{
		Content: ":x: 這個創建私人頻道的設置已經被刪除了喔，請麻煩管理員重新創建!",
	}
}

func ticketWelcomeMessage() ports.OutboundMessage {
	return ports.OutboundMessage{
		Content: "||@everyone||",
		AllowedMentions: ports.AllowedMentions{
			ParseEveryone: false,
		},
		Embeds: []ports.OutboundEmbed{{
			Title:       "__**私人頻道**__",
			Description: "你開啟了一個私人頻道，請等待客服人員的回復!",
			Color:       legacyDiscordNamedGreen,
		}},
		Components: []ports.OutboundComponentRow{{Components: []ports.OutboundComponent{{
			Type:     "button",
			CustomID: "del",
			Label:    "🗑️ 刪除!",
			Style:    "danger",
		}}}},
	}
}

func ticketOpenSuccessMessage() responses.Message {
	return responses.Message{
		Ephemeral: true,
		Embeds: []responses.Embed{{
			Title:       "__**頻道**__",
			Description: ":white_check_mark: 你成功開啟了頻道!",
			Color:       legacyTicketOpenGreen,
		}},
	}
}

func ticketCloseDeniedMessage() responses.Message {
	return responses.Message{
		Embeds: []responses.Embed{{
			Title:       "__**私人頻道**__",
			Description: "你開啟了一個私人頻道，請靜候客服人員的回復!",
			Color:       legacyDiscordNamedRed,
		}},
	}
}

func parseLegacyColor(value string) (int, bool) {
	if len(value) == 7 && strings.HasPrefix(value, "#") {
		parsed, err := parseHexColor(value)
		return parsed, err == nil
	}
	if parsed, ok := legacyTicketNamedColors[value]; ok {
		return parsed, true
	}
	return 0, false
}

func parseHexColor(value string) (int, error) {
	if len(value) != 7 || value[0] != '#' {
		return 0, errors.New("invalid hex color")
	}
	parsed, err := strconv.ParseInt(value[1:], 16, 32)
	if err != nil {
		return 0, err
	}
	return int(parsed), nil
}

var legacyTicketNamedColors = map[string]int{
	"White":      0xFFFFFF,
	"Aqua":       0x1ABC9C,
	"Green":      0x57F287,
	"Blue":       0x3498DB,
	"Yellow":     0xFEE75C,
	"Purple":     0x9B59B6,
	"Fuchsia":    0xEB459E,
	"Gold":       0xF1C40F,
	"Orange":     0xE67E22,
	"Red":        0xED4245,
	"Grey":       0x95A5A6,
	"Navy":       0x34495E,
	"DarkGreen":  0x1F8B4C,
	"DarkBlue":   0x206694,
	"DarkOrange": 0xA84300,
	"DarkRed":    0x992D22,
	"DarkGrey":   0x979C9F,
	"LightGrey":  0xBCC0C0,
}
