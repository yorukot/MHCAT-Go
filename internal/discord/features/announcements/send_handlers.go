package announcements

import (
	"context"
	"errors"
	"strings"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/customid"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/interactions"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/responses"
)

const (
	announcementFeature = "announcement"
	sendModalAction     = "submit"
	confirmAction       = "confirm"
	cancelAction        = "cancel"

	sendModalLegacyID = "nal"

	fieldTag     = "anntag"
	fieldColor   = "anncolor"
	fieldTitle   = "anntitle"
	fieldContent = "anncontent"

	confirmColor = 0x00FF19
)

func (m Module) SendHandler() interactions.Handler {
	return func(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
		if !interaction.Actor.HasPermission(permissionManageMessages) {
			return responder.Reply(ctx, announcementErrorMessage("你需要有`訊息管理`才能使用此指令"))
		}
		modalID, err := customid.Encode(customid.InteractionKindModal, announcementFeature, sendModalAction, customid.EmptyPayload())
		if err != nil {
			return err
		}
		return responder.ShowModal(ctx, announcementSendModal(modalID))
	}
}

func (m Module) SendModalHandler() interactions.Handler {
	return func(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
		if err := responder.Defer(ctx, responses.DeferOptions{}); err != nil {
			return err
		}
		if _, ok := domain.ParseLegacyColorValue(firstFieldValue(interaction.ModalFields, fieldColor)); !ok {
			return responder.EditOriginal(ctx, modalErrorMessage("你傳送的並不是顏色(色碼)"))
		}
		draft, err := draftFromInteraction(interaction)
		if err != nil {
			return responder.EditOriginal(ctx, announcementErrorMessage("請完整填寫公告內容"))
		}
		stateID, err := m.draftStore().Put(draft)
		if err != nil {
			return err
		}
		if err := responder.EditOriginal(ctx, previewMessage(draft)); err != nil {
			m.draftStore().Delete(stateID)
			return err
		}
		return responder.FollowUp(ctx, confirmationMessage(stateID))
	}
}

func (m Module) ConfirmHandler() interactions.Handler {
	return func(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
		if err := responder.Defer(ctx, responses.DeferOptions{}); err != nil {
			return err
		}
		stateID, err := stateIDFromInteraction(interaction)
		if err != nil {
			return responder.EditOriginal(ctx, announcementErrorMessage("已取消"))
		}
		draft, err := m.draftStore().Take(stateID)
		if errors.Is(err, ErrAnnouncementDraftNotFound) {
			return responder.EditOriginal(ctx, announcementErrorMessage("已取消"))
		}
		if err != nil {
			return err
		}
		if draft.UserID != "" && draft.UserID != interaction.Actor.UserID {
			return responder.EditOriginal(ctx, announcementErrorMessage("這個公告確認按鈕不是給你使用的"))
		}
		if m.reader == nil || m.messages == nil {
			return responder.EditOriginal(ctx, announcementErrorMessage("很抱歉，出現了未知的錯誤，請重試!"))
		}
		config, err := m.reader.GetAnnouncementChannel(ctx, interaction.Actor.GuildID)
		if errors.Is(err, ports.ErrAnnouncementChannelMissing) {
			return responder.EditOriginal(ctx, missingAnnouncementChannelMessage())
		}
		if err != nil {
			return responder.EditOriginal(ctx, announcementUnknownError(err))
		}
		if _, err := m.messages.SendMessage(ctx, config.ChannelID, outboundAnnouncementMessage(draft)); err != nil {
			return responder.EditOriginal(ctx, announcementUnknownError(err))
		}
		if err := responder.EditOriginal(ctx, sendSuccessMessage()); err != nil {
			return err
		}
		return m.trackCommand(ctx, interaction, SendCommandName, "announcement-send")
	}
}

func (m Module) CancelHandler() interactions.Handler {
	return func(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
		if stateID, err := stateIDFromInteraction(interaction); err == nil {
			m.draftStore().Delete(stateID)
		}
		return responder.Reply(ctx, responses.Message{
			Content:         "已取消",
			AllowedMentions: &responses.AllowedMentions{},
		})
	}
}

func announcementSendModal(customID string) responses.Modal {
	return responses.Modal{
		CustomID: customID,
		Title:    "公告系統",
		Rows: []responses.ModalRow{
			{Inputs: []responses.TextInput{{CustomID: fieldTag, Label: "請輸入你要tag誰", Style: responses.TextInputStyleShort, Required: true}}},
			{Inputs: []responses.TextInput{{CustomID: fieldColor, Label: "請輸入你的公告要甚麼顏色", Style: responses.TextInputStyleShort, Required: true}}},
			{Inputs: []responses.TextInput{{CustomID: fieldTitle, Label: "請輸入你的公告標題", Style: responses.TextInputStyleShort, Required: true}}},
			{Inputs: []responses.TextInput{{CustomID: fieldContent, Label: "請輸入公告內文", Style: responses.TextInputStyleParagraph, Required: true}}},
		},
	}
}

func draftFromInteraction(interaction interactions.Interaction) (AnnouncementDraft, error) {
	tag := firstFieldValue(interaction.ModalFields, fieldTag)
	colorValue := firstFieldValue(interaction.ModalFields, fieldColor)
	title := firstFieldValue(interaction.ModalFields, fieldTitle)
	content := firstFieldValue(interaction.ModalFields, fieldContent)
	color, ok := domain.ParseLegacyColorValue(colorValue)
	if !ok || strings.TrimSpace(tag) == "" || strings.TrimSpace(title) == "" || strings.TrimSpace(content) == "" {
		return AnnouncementDraft{}, domain.ErrInvalidAnnouncementConfig
	}
	return AnnouncementDraft{
		GuildID:   interaction.Actor.GuildID,
		UserID:    interaction.Actor.UserID,
		UserTag:   interaction.Actor.UserTag,
		AvatarURL: interaction.Actor.AvatarURL,
		Tag:       tag,
		Color:     color,
		Title:     title,
		Content:   content,
	}, nil
}

func previewMessage(draft AnnouncementDraft) responses.Message {
	return responses.Message{
		Content:         draft.Tag,
		Embeds:          []responses.Embed{announcementEmbed(draft)},
		AllowedMentions: &responses.AllowedMentions{},
	}
}

func confirmationMessage(stateID string) responses.Message {
	confirmID := announcementStateComponentID(confirmAction, stateID)
	cancelID := announcementStateComponentID(cancelAction, stateID)
	return responses.Message{
		Embeds: []responses.Embed{{
			Title: "是否將此訊息送往公告?(請於六秒內點擊:P)",
			Color: confirmColor,
		}},
		Components: []responses.ComponentRow{{Components: []responses.Component{
			{
				Type:     responses.ComponentTypeButton,
				CustomID: confirmID,
				Emoji:    "✅",
				Label:    "是",
				Style:    responses.ButtonStylePrimary,
			},
			{
				Type:     responses.ComponentTypeButton,
				CustomID: cancelID,
				Emoji:    "❎",
				Label:    "否",
				Style:    responses.ButtonStyleDanger,
			},
		}}},
		AllowedMentions: &responses.AllowedMentions{},
	}
}

func announcementEmbed(draft AnnouncementDraft) responses.Embed {
	footer := "來自" + draft.UserTag + "的公告"
	return responses.Embed{
		Title:       draft.Title,
		Description: draft.Content,
		Color:       draft.Color,
		Footer:      &responses.EmbedFooter{Text: footer, IconURL: draft.AvatarURL},
	}
}

func outboundAnnouncementMessage(draft AnnouncementDraft) ports.OutboundMessage {
	footer := "來自" + draft.UserTag + "的公告"
	return ports.OutboundMessage{
		Content: draft.Tag,
		Embeds: []ports.OutboundEmbed{{
			Title:         draft.Title,
			Description:   draft.Content,
			Color:         draft.Color,
			FooterText:    footer,
			FooterIconURL: draft.AvatarURL,
		}},
		AllowedMentions: ports.AllowedMentions{},
	}
}

func sendSuccessMessage() responses.Message {
	return responses.Message{
		Content:         "<a:green_tick:994529015652163614> | 成功發送!",
		AllowedMentions: &responses.AllowedMentions{},
	}
}

func missingAnnouncementChannelMessage() responses.Message {
	return responses.Message{
		Content:         "很抱歉!\n你還沒有對您的公告頻道進行選擇!\n命令:`<> 公告頻道設置 [公告頻道id]`\n有問題歡迎打`<>幫助`",
		AllowedMentions: &responses.AllowedMentions{},
	}
}

func modalErrorMessage(content string) responses.Message {
	return responses.Message{
		Embeds: []responses.Embed{{
			Title: content,
			Color: legacyErrorColor,
		}},
		AllowedMentions: &responses.AllowedMentions{},
	}
}

func announcementStateComponentID(action string, stateID string) string {
	payload, err := customid.StateIDPayload(stateID)
	if err != nil {
		return "mhcat:v1:announcement:" + action + ":"
	}
	id, err := customid.Encode(customid.InteractionKindComponent, announcementFeature, action, payload)
	if err != nil {
		return "mhcat:v1:announcement:" + action + ":"
	}
	return id
}

func stateIDFromInteraction(interaction interactions.Interaction) (string, error) {
	id, err := customid.ParseComponent(interaction.CustomID)
	if err != nil {
		return "", err
	}
	if id.Payload.StateID == "" {
		return "", customid.ErrInvalidPayload
	}
	return id.Payload.StateID, nil
}

func firstFieldValue(fields []customid.ModalField, customID string) string {
	for _, field := range fields {
		if field.CustomID == customID {
			return field.Value
		}
	}
	return ""
}

func (m Module) draftStore() *DraftStore {
	if m.drafts != nil {
		return m.drafts
	}
	return NewDraftStore()
}
