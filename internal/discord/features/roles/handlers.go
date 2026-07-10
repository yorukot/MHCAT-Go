package roles

import (
	"context"
	"errors"
	"strings"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	coreservice "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/services/roles"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/customid"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/events"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/interactions"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/responses"
)

const (
	permissionManageMessages  = int64(8192)
	roleSelectionErrorPrefix  = "<a:Discord_AnimatedNo:1015989839809757295> | "
	roleSelectionActionPrefix = "<a:error:980086028113182730> | "
	roleSelectionErrorColor   = 0xED4245
	roleSelectionSuccessColor = 0x57F287
	roleSelectionDoneEmoji    = "<a:green_tick:994529015652163614>"
	roleSelectionModalID      = "nal"
	roleSelectionFieldPrefix  = "roleaddcontent"
)

func (m Module) ReactionSetHandler() interactions.Handler {
	return func(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
		if err := responder.Defer(ctx, responses.DeferOptions{}); err != nil {
			return err
		}
		if !interaction.Actor.HasPermission(permissionManageMessages) {
			return responder.EditOriginal(ctx, roleSelectionErrorMessage("你需要有`訊息管理`才能使用此指令"))
		}
		_, err := m.service.ConfigureReaction(ctx, coreservice.ReactionSetCommand{
			GuildID:    interaction.Actor.GuildID,
			MessageURL: firstOption(interaction, "訊息url"),
			RoleID:     firstOption(interaction, "身分組"),
			Emoji:      firstOption(interaction, "表情符號"),
		})
		if err != nil {
			return responder.EditOriginal(ctx, roleSelectionErrorFromError(err, "很抱歉，出現了未知的錯誤，請重試!"))
		}
		if err := responder.EditOriginal(ctx, roleSelectionSuccessTitle(roleSelectionDoneEmoji+" | 表情符號選取身分組成功設定")); err != nil {
			return err
		}
		return nil
	}
}

func (m Module) ReactionDeleteHandler() interactions.Handler {
	return func(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
		if err := responder.Defer(ctx, responses.DeferOptions{}); err != nil {
			return err
		}
		if !interaction.Actor.HasPermission(permissionManageMessages) {
			return responder.EditOriginal(ctx, roleSelectionErrorMessage("你需要有`訊息管理`才能使用此指令"))
		}
		err := m.service.DeleteReaction(ctx, coreservice.ReactionDeleteCommand{
			GuildID:    interaction.Actor.GuildID,
			MessageURL: firstOption(interaction, "訊息url"),
			Emoji:      firstOption(interaction, "表情符號"),
		})
		if err != nil {
			return responder.EditOriginal(ctx, roleSelectionErrorFromError(err, "很抱歉，出現了未知的錯誤，請重試!"))
		}
		if err := responder.EditOriginal(ctx, roleSelectionSuccessTitle("表情符號選取身分組成功刪除")); err != nil {
			return err
		}
		return nil
	}
}

func (m Module) ButtonSetupHandler() interactions.Handler {
	return func(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
		if !interaction.Actor.HasPermission(permissionManageMessages) {
			message := roleSelectionErrorMessage("你需要有`訊息管理`才能使用此指令")
			message.Ephemeral = true
			return responder.Reply(ctx, message)
		}
		baseID := m.idGenerator()
		_, err := m.service.PrepareButton(ctx, coreservice.ButtonPrepareCommand{
			GuildID: interaction.Actor.GuildID,
			RoleID:  firstOption(interaction, "身分組"),
			BaseID:  baseID,
		})
		if err != nil {
			message := roleSelectionErrorFromError(err, "很抱歉，出現了未知的錯誤，請重試!")
			message.Ephemeral = true
			return responder.Reply(ctx, message)
		}
		if err := responder.ShowModal(ctx, roleSelectionModal(baseID)); err != nil {
			return err
		}
		return nil
	}
}

func (m Module) ButtonModalHandler() interactions.Handler {
	return func(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
		if err := responder.Defer(ctx, responses.DeferOptions{}); err != nil {
			return err
		}
		fieldID, content := roleSelectionModalField(interaction.ModalFields)
		if fieldID == "" {
			return responder.EditOriginal(ctx, roleSelectionErrorMessage("很抱歉，出現了未知的錯誤，請重試!"))
		}
		baseID := strings.TrimPrefix(fieldID, roleSelectionFieldPrefix)
		if m.messages == nil {
			return responder.EditOriginal(ctx, roleSelectionErrorMessage("很抱歉，出現了未知的錯誤，請重試!"))
		}
		if _, err := m.messages.SendMessage(ctx, interaction.ChannelID, roleSelectionButtonPanelOutbound(baseID, content, interaction.BotDisplayColor)); err != nil {
			return responder.EditOriginal(ctx, roleSelectionErrorMessage("很抱歉，出現了未知的錯誤，請重試!"))
		}
		return responder.EditOriginal(ctx, roleSelectionSuccessTitle(roleSelectionDoneEmoji+" | 成功創建領取身分組"))
	}
}

func (m Module) ButtonApplyHandler(remove bool) interactions.Handler {
	return func(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
		if err := responder.Defer(ctx, responses.DeferOptions{Ephemeral: true}); err != nil {
			return err
		}
		err := m.service.ApplyButton(ctx, coreservice.ButtonApplyCommand{
			GuildID:      interaction.Actor.GuildID,
			UserID:       interaction.Actor.UserID,
			Number:       interaction.CustomID,
			Remove:       remove,
			ActorRoleIDs: interaction.Actor.RoleIDs,
		})
		if err != nil {
			return responder.EditOriginal(ctx, roleSelectionButtonError(err, remove))
		}
		if remove {
			return responder.EditOriginal(ctx, roleSelectionSuccessTitle(roleSelectionDoneEmoji+" | 成功刪除身分組!"))
		}
		return responder.EditOriginal(ctx, roleSelectionSuccessTitle(roleSelectionDoneEmoji+" | 成功增加身分組!"))
	}
}

func (m Module) ReactionEventHandler(remove bool) events.Handler {
	return func(ctx context.Context, event events.Event) error {
		if event.IsBot || event.Reaction == nil {
			return nil
		}
		react := event.Reaction.EmojiName
		if strings.TrimSpace(event.Reaction.EmojiID) != "" {
			react = event.Reaction.EmojiID
		}
		err := m.service.ApplyReaction(ctx, coreservice.ReactionApplyCommand{
			GuildID:   event.GuildID,
			MessageID: event.MessageID,
			React:     react,
			UserID:    event.UserID,
			Remove:    remove,
		})
		if errors.Is(err, ports.ErrRoleReactionConfigMissing) {
			return nil
		}
		if errors.Is(err, ports.ErrDiscordRoleMissing) || errors.Is(err, ports.ErrDiscordRoleNotAssignable) {
			if m.direct != nil {
				_, _ = m.direct.SendDirectMessage(ctx, event.UserID, roleSelectionRoleErrorOutbound(remove))
			}
			return nil
		}
		return events.ContinueOnError(err)
	}
}

func roleSelectionModal(baseID string) responses.Modal {
	return responses.Modal{
		CustomID: roleSelectionModalID,
		Title:    "領取身分系統!",
		Rows: []responses.ModalRow{{
			Inputs: []responses.TextInput{{
				CustomID: roleSelectionFieldPrefix + baseID,
				Label:    "請輸入身分訊息內文",
				Style:    responses.TextInputStyleParagraph,
			}},
		}},
	}
}

func roleSelectionButtonPanelOutbound(baseID string, content string, color int) ports.OutboundMessage {
	return ports.OutboundMessage{
		Embeds: []ports.OutboundEmbed{{
			Title:       "選取身分組",
			Description: content,
			Color:       color,
		}},
		Components: []ports.OutboundComponentRow{{
			Components: []ports.OutboundComponent{
				{Type: "button", CustomID: baseID + "add", Label: "✅點我增加身分組!", Style: "primary"},
				{Type: "button", CustomID: baseID + "delete", Label: "❎點我刪除身分組!", Style: "danger"},
			},
		}},
		AllowedMentions: ports.AllowedMentions{},
	}
}

func roleSelectionRoleErrorOutbound(remove bool) ports.OutboundMessage {
	prefix := roleSelectionActionPrefix
	if remove {
		prefix = roleSelectionErrorPrefix
	}
	return ports.OutboundMessage{
		Embeds: []ports.OutboundEmbed{{
			Title: prefix + "我沒有權限給大家這個身分組或是身分組被刪除了(請把我的身分組調高)!",
			Color: roleSelectionErrorColor,
		}},
		AllowedMentions: ports.AllowedMentions{},
	}
}

func roleSelectionModalField(fields []customid.ModalField) (string, string) {
	for _, field := range fields {
		if strings.HasPrefix(field.CustomID, roleSelectionFieldPrefix) {
			return field.CustomID, field.Value
		}
	}
	return "", ""
}

func roleSelectionButtonError(err error, remove bool) responses.Message {
	switch {
	case errors.Is(err, ports.ErrRoleButtonConfigMissing):
		return roleSelectionButtonContentError("很抱歉，出現了錯誤!")
	case errors.Is(err, coreservice.ErrRoleAlreadyAssigned):
		return roleSelectionErrorMessage("你已經擁有身分組了!")
	case errors.Is(err, coreservice.ErrRoleNotAssigned):
		return roleSelectionErrorMessage(" 你沒有這個身分組!")
	case errors.Is(err, ports.ErrDiscordRoleMissing):
		if remove {
			return roleSelectionButtonActionError("找不到這個身分組!")
		}
		return roleSelectionButtonActionError("請通知群主管裡員找不到這個身分組!")
	case errors.Is(err, ports.ErrDiscordRoleNotAssignable):
		return roleSelectionButtonActionError("請通知群主管裡員我沒有權限給你這個身分組(請把我的身分組調高)!")
	default:
		return roleSelectionButtonContentError("opps,出現了錯誤!\n有可能是你設定沒設定好\n或是我沒有權限喔(請確認我的權限比你要加的權限高，還需要管理身分組的權限)")
	}
}

func roleSelectionButtonActionError(content string) responses.Message {
	return responses.Message{
		Embeds: []responses.Embed{{
			Title: roleSelectionActionPrefix + content,
			Color: roleSelectionErrorColor,
		}},
		AllowedMentions: &responses.AllowedMentions{},
	}
}

func roleSelectionButtonContentError(content string) responses.Message {
	return responses.Message{
		Content:         content,
		AllowedMentions: &responses.AllowedMentions{},
	}
}

func roleSelectionErrorFromError(err error, fallback string) responses.Message {
	switch {
	case errors.Is(err, domain.ErrInvalidRoleSelectionEmoji):
		return roleSelectionErrorMessage("你必須輸入正確的表情符號!(表情符號所在伺服器我必須在裡面!)")
	case errors.Is(err, domain.ErrInvalidRoleSelectionConfig):
		return roleSelectionErrorMessage("你輸入的不是一個訊息連結")
	case errors.Is(err, ports.ErrChannelNotFound), errors.Is(err, ports.ErrDiscordMessageNotFound):
		return roleSelectionErrorMessage("很抱歉，找不到這個訊息")
	case errors.Is(err, ports.ErrDiscordRoleNotAssignable), errors.Is(err, ports.ErrDiscordRoleMissing):
		return roleSelectionErrorMessage("我沒有權限給大家這個身分組(請把我的身分組調高)!")
	case errors.Is(err, ports.ErrRoleReactionConfigMissing):
		return roleSelectionErrorMessage("這個表情符號沒有在這個訊息上設定")
	default:
		return roleSelectionErrorMessage(fallback)
	}
}

func roleSelectionErrorMessage(content string) responses.Message {
	return responses.Message{
		Embeds: []responses.Embed{{
			Title: roleSelectionErrorPrefix + content,
			Color: roleSelectionErrorColor,
		}},
		AllowedMentions: &responses.AllowedMentions{},
	}
}

func roleSelectionSuccessTitle(title string) responses.Message {
	return responses.Message{
		Embeds: []responses.Embed{{
			Title: title,
			Color: roleSelectionSuccessColor,
		}},
		AllowedMentions: &responses.AllowedMentions{},
	}
}

func firstOption(interaction interactions.Interaction, names ...string) string {
	for _, name := range names {
		if value, ok := interaction.Options[name]; ok {
			return strings.TrimSpace(value)
		}
		if value, ok := interaction.CommandOptions[name]; ok {
			return strings.TrimSpace(value.String)
		}
	}
	return ""
}
