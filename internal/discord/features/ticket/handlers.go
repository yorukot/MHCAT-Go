package ticket

import (
	"context"
	"errors"
	"fmt"
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
			return m.repo.SaveTicketConfig(ctx, config)
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
	title := strings.TrimSpace(fields["tickettitle"])
	content := strings.TrimSpace(fields["ticketcontent"])
	if title == "" || content == "" {
		return responder.EditOriginal(ctx, ticketEditErrorMessage("請完整填寫私人頻道標題與內文。"))
	}
	if saveConfig != nil {
		if err := saveConfig(ctx); err != nil {
			return err
		}
	}
	if _, err := m.messages.SendMessage(ctx, interaction.ChannelID, ticketPanelOutboundMessage(title, content, color)); err != nil {
		return err
	}
	if err := responder.EditOriginal(ctx, ticketSetupSuccessMessage()); err != nil {
		return err
	}
	return m.track(ctx, interaction, "私人頻道設置")
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
				Color:       0xFF0000,
			}},
		}); err != nil {
			return err
		}
		return m.track(ctx, interaction, "私人頻道刪除")
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
			if interaction.MessageID != "" {
				_ = m.messages.DeleteMessage(ctx, ports.MessageRef{ChannelID: interaction.ChannelID, MessageID: interaction.MessageID})
			}
			return responder.Reply(ctx, ticketDeletedConfigMessage())
		}
		if err != nil {
			return err
		}

		created, err := m.channels.CreateChannel(ctx, ports.ChannelCreateRequest{
			GuildID:              interaction.Actor.GuildID,
			ParentID:             config.CategoryID,
			Name:                 interaction.Actor.UserID,
			Type:                 discordChannelTypeGuildText,
			PermissionOverwrites: m.ticketOpenPermissionOverwrites(config, interaction.Actor.UserID, ticketBotUserID(interaction.ApplicationID, m.botUserID)),
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
		return m.track(ctx, interaction, "私人頻道開啟")
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
		if !interaction.Actor.HasPermission(permissionManageMessages) && interaction.ChannelName != "" && interaction.ChannelName != interaction.Actor.UserID {
			return responder.Reply(ctx, ticketCloseDeniedMessage())
		}
		if !interaction.Actor.HasPermission(permissionManageMessages) && interaction.ChannelName == "" {
			return responder.Reply(ctx, ticketCloseDeniedMessage())
		}
		if err := m.channels.DeleteChannel(ctx, interaction.ChannelID); err != nil {
			return err
		}
		return m.track(ctx, interaction, "私人頻道刪除頻道")
	}
}

func (m Module) ticketOpenPermissionOverwrites(config domain.TicketConfig, userID string, botUserID string) []ports.PermissionOverwrite {
	allow := int64(permissionViewChannel | permissionSendMessages | permissionReadMessageHistory)
	denyInvite := int64(permissionCreateInstantInvite)
	overwrites := []ports.PermissionOverwrite{
		{ID: config.AdminRoleID, Type: permissionOverwriteRole, Allow: allow, Deny: denyInvite},
		{ID: config.EveryoneRoleID, Type: permissionOverwriteRole, Deny: int64(permissionViewChannel)},
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

func (m Module) track(ctx context.Context, interaction interactions.Interaction, command string) error {
	if m.usage == nil {
		return nil
	}
	return m.usage.TrackCommand(ctx, ports.UsageEvent{
		CommandName: command,
		UserID:      interaction.Actor.UserID,
		GuildID:     interaction.Actor.GuildID,
		Feature:     m.feature,
	})
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
			Color: 0xFF0000,
		}},
	}
}

func ticketEditErrorMessage(content string) responses.Message {
	return responses.Message{
		Embeds: []responses.Embed{{
			Title: content,
			Color: 0xFF0000,
		}},
	}
}

func ticketPermissionDeniedMessage(permission string) responses.Message {
	return responses.Message{
		Ephemeral: true,
		Embeds: []responses.Embed{{
			Title: "<a:Discord_AnimatedNo:1015989839809757295> | 你需要有`" + permission + "`才能使用此指令",
			Color: 0xFF0000,
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
			Color:       0xFF0000,
		}},
	}
}

func ticketSetupSuccessMessage() responses.Message {
	return responses.Message{
		Embeds: []responses.Embed{{
			Title: "<a:green_tick:994529015652163614> | 成功創建私人頻道",
			Color: 0x00DB00,
		}},
	}
}

func ticketAlreadyOpenMessage() responses.Message {
	return responses.Message{
		Ephemeral: true,
		Embeds: []responses.Embed{{
			Title:       "__**客服頻道**__",
			Description: ":warning: 你已經有一個客服頻道了!",
			Color:       0xFF0000,
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
			Color:       0x00DB00,
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
			Color:       0x00DB00,
		}},
	}
}

func ticketCloseDeniedMessage() responses.Message {
	return responses.Message{
		Embeds: []responses.Embed{{
			Title:       "__**私人頻道**__",
			Description: "你開啟了一個私人頻道，請靜候客服人員的回復!",
			Color:       0xFF0000,
		}},
	}
}

func parseLegacyColor(value string) (int, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, false
	}
	if strings.HasPrefix(value, "#") {
		parsed, err := parseHexColor(value)
		return parsed, err == nil
	}
	if parsed, ok := cssColorName(value); ok {
		return parsed, true
	}
	return 0, false
}

func parseHexColor(value string) (int, error) {
	raw := strings.TrimPrefix(value, "#")
	if len(raw) == 3 {
		raw = fmt.Sprintf("%c%c%c%c%c%c", raw[0], raw[0], raw[1], raw[1], raw[2], raw[2])
	}
	if len(raw) != 6 {
		return 0, fmt.Errorf("invalid hex color")
	}
	parsed, err := strconv.ParseInt(raw, 16, 32)
	if err != nil {
		return 0, err
	}
	return int(parsed), nil
}

func cssColorName(value string) (int, bool) {
	parsed, ok := cssNamedColors[strings.ToLower(value)]
	return parsed, ok
}

var cssNamedColors = map[string]int{
	"aliceblue":            0xF0F8FF,
	"antiquewhite":         0xFAEBD7,
	"aqua":                 0x00FFFF,
	"aquamarine":           0x7FFFD4,
	"azure":                0xF0FFFF,
	"beige":                0xF5F5DC,
	"bisque":               0xFFE4C4,
	"black":                0x000000,
	"blanchedalmond":       0xFFEBCD,
	"blue":                 0x0000FF,
	"blueviolet":           0x8A2BE2,
	"brown":                0xA52A2A,
	"burlywood":            0xDEB887,
	"cadetblue":            0x5F9EA0,
	"chartreuse":           0x7FFF00,
	"chocolate":            0xD2691E,
	"coral":                0xFF7F50,
	"cornflowerblue":       0x6495ED,
	"cornsilk":             0xFFF8DC,
	"crimson":              0xDC143C,
	"cyan":                 0x00FFFF,
	"darkblue":             0x00008B,
	"darkcyan":             0x008B8B,
	"darkgoldenrod":        0xB8860B,
	"darkgray":             0xA9A9A9,
	"darkgreen":            0x006400,
	"darkgrey":             0xA9A9A9,
	"darkkhaki":            0xBDB76B,
	"darkmagenta":          0x8B008B,
	"darkolivegreen":       0x556B2F,
	"darkorange":           0xFF8C00,
	"darkorchid":           0x9932CC,
	"darkred":              0x8B0000,
	"darksalmon":           0xE9967A,
	"darkseagreen":         0x8FBC8F,
	"darkslateblue":        0x483D8B,
	"darkslategray":        0x2F4F4F,
	"darkslategrey":        0x2F4F4F,
	"darkturquoise":        0x00CED1,
	"darkviolet":           0x9400D3,
	"deeppink":             0xFF1493,
	"deepskyblue":          0x00BFFF,
	"dimgray":              0x696969,
	"dimgrey":              0x696969,
	"dodgerblue":           0x1E90FF,
	"firebrick":            0xB22222,
	"floralwhite":          0xFFFAF0,
	"forestgreen":          0x228B22,
	"fuchsia":              0xFF00FF,
	"gainsboro":            0xDCDCDC,
	"ghostwhite":           0xF8F8FF,
	"gold":                 0xFFD700,
	"goldenrod":            0xDAA520,
	"gray":                 0x808080,
	"green":                0x008000,
	"greenyellow":          0xADFF2F,
	"grey":                 0x808080,
	"honeydew":             0xF0FFF0,
	"hotpink":              0xFF69B4,
	"indianred":            0xCD5C5C,
	"indigo":               0x4B0082,
	"ivory":                0xFFFFF0,
	"khaki":                0xF0E68C,
	"lavender":             0xE6E6FA,
	"lavenderblush":        0xFFF0F5,
	"lawngreen":            0x7CFC00,
	"lemonchiffon":         0xFFFACD,
	"lightblue":            0xADD8E6,
	"lightcoral":           0xF08080,
	"lightcyan":            0xE0FFFF,
	"lightgoldenrodyellow": 0xFAFAD2,
	"lightgray":            0xD3D3D3,
	"lightgreen":           0x90EE90,
	"lightgrey":            0xD3D3D3,
	"lightpink":            0xFFB6C1,
	"lightsalmon":          0xFFA07A,
	"lightseagreen":        0x20B2AA,
	"lightskyblue":         0x87CEFA,
	"lightslategray":       0x778899,
	"lightslategrey":       0x778899,
	"lightsteelblue":       0xB0C4DE,
	"lightyellow":          0xFFFFE0,
	"lime":                 0x00FF00,
	"limegreen":            0x32CD32,
	"linen":                0xFAF0E6,
	"magenta":              0xFF00FF,
	"maroon":               0x800000,
	"mediumaquamarine":     0x66CDAA,
	"mediumblue":           0x0000CD,
	"mediumorchid":         0xBA55D3,
	"mediumpurple":         0x9370DB,
	"mediumseagreen":       0x3CB371,
	"mediumslateblue":      0x7B68EE,
	"mediumspringgreen":    0x00FA9A,
	"mediumturquoise":      0x48D1CC,
	"mediumvioletred":      0xC71585,
	"midnightblue":         0x191970,
	"mintcream":            0xF5FFFA,
	"mistyrose":            0xFFE4E1,
	"moccasin":             0xFFE4B5,
	"navajowhite":          0xFFDEAD,
	"navy":                 0x000080,
	"oldlace":              0xFDF5E6,
	"olive":                0x808000,
	"olivedrab":            0x6B8E23,
	"orange":               0xFFA500,
	"orangered":            0xFF4500,
	"orchid":               0xDA70D6,
	"palegoldenrod":        0xEEE8AA,
	"palegreen":            0x98FB98,
	"paleturquoise":        0xAFEEEE,
	"palevioletred":        0xDB7093,
	"papayawhip":           0xFFEFD5,
	"peachpuff":            0xFFDAB9,
	"peru":                 0xCD853F,
	"pink":                 0xFFC0CB,
	"plum":                 0xDDA0DD,
	"powderblue":           0xB0E0E6,
	"purple":               0x800080,
	"rebeccapurple":        0x663399,
	"red":                  0xFF0000,
	"rosybrown":            0xBC8F8F,
	"royalblue":            0x4169E1,
	"saddlebrown":          0x8B4513,
	"salmon":               0xFA8072,
	"sandybrown":           0xF4A460,
	"seagreen":             0x2E8B57,
	"seashell":             0xFFF5EE,
	"sienna":               0xA0522D,
	"silver":               0xC0C0C0,
	"skyblue":              0x87CEEB,
	"slateblue":            0x6A5ACD,
	"slategray":            0x708090,
	"slategrey":            0x708090,
	"snow":                 0xFFFAFA,
	"springgreen":          0x00FF7F,
	"steelblue":            0x4682B4,
	"tan":                  0xD2B48C,
	"teal":                 0x008080,
	"thistle":              0xD8BFD8,
	"tomato":               0xFF6347,
	"turquoise":            0x40E0D0,
	"violet":               0xEE82EE,
	"wheat":                0xF5DEB3,
	"white":                0xFFFFFF,
	"whitesmoke":           0xF5F5F5,
	"yellow":               0xFFFF00,
	"yellowgreen":          0x9ACD32,
}
