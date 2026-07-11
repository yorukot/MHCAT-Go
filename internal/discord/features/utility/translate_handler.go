package utility

import (
	"context"
	"errors"
	"strings"
	"time"
	"unicode/utf16"

	coreutility "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/services/utility"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/interactions"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/responses"
)

const (
	TranslateProviderTimeout = 10 * time.Second
	translateLoadingColor    = 0x57F287
	translateErrorColor      = 0xEA0000
)

func (m Module) TranslateHandler() interactions.Handler {
	return func(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
		if err := responder.Defer(ctx, responses.DeferOptions{}); err != nil {
			return err
		}
		loadingMessageID, err := responder.CreateFollowUp(ctx, translateLoadingMessage())
		if err != nil {
			return err
		}
		source := interaction.Options["要的翻譯"]
		targetLanguage := interaction.Options["目標語言"]
		providerCtx := ctx
		cancel := func() {}
		if m.translateTimeout > 0 {
			providerCtx, cancel = context.WithTimeout(ctx, m.translateTimeout)
		}
		result, err := m.translate.Translate(providerCtx, source, targetLanguage)
		cancel()
		if err != nil {
			return responder.EditFollowUp(ctx, loadingMessageID, translateErrorMessage(err))
		}
		if err := responder.EditFollowUp(ctx, loadingMessageID, translateResultMessage(interaction, source, targetLanguage, result.Text, m.randomTranslateColor())); err != nil {
			return err
		}
		return m.track(ctx, interaction, "翻譯")
	}
}

func translateLoadingMessage() responses.Message {
	return responses.Message{
		Embeds: []responses.Embed{{
			Title: "<a:load:986319593444352071> | 我正在玩命幫你翻譯!",
			Color: translateLoadingColor,
		}},
		AllowedMentions: &responses.AllowedMentions{},
	}
}

func translateResultMessage(interaction interactions.Interaction, source string, targetLanguage string, translated string, color int) responses.Message {
	footerText := strings.TrimSpace(interaction.Actor.UserTag)
	if footerText == "" {
		footerText = strings.TrimSpace(interaction.Actor.UserID)
	}
	if footerText != "" {
		footerText += "的查詢"
	}
	return responses.Message{
		Embeds: []responses.Embed{{
			Title: "<:translate:986870996147507231> 翻譯系統",
			Color: color,
			Fields: []responses.EmbedField{
				{Name: "**<:edittext:986873966884962304> 原文**:", Value: codeField(source), Inline: false},
				{Name: "**<:answer:986873630178832414> 目標語言:**", Value: codeField(targetLanguage), Inline: false},
				{Name: "**<:translate1:986873633483939901> 譯文:**", Value: codeField(translated), Inline: false},
			},
			Footer: &responses.EmbedFooter{
				Text:    footerText,
				IconURL: interaction.Actor.AvatarURL,
			},
		}},
		AllowedMentions: &responses.AllowedMentions{},
	}
}

func translateErrorMessage(err error) responses.Message {
	content := "很抱歉，翻譯失敗，請稍後再試!"
	if errors.Is(err, coreutility.ErrInvalidTranslateInput) || errors.Is(err, coreutility.ErrUnsupportedTranslateLanguage) {
		content = "請輸入有效的翻譯內容與目標語言!"
	}
	return responses.Message{
		Embeds: []responses.Embed{{
			Title: "<a:Discord_AnimatedNo:1015989839809757295> | " + content,
			Color: translateErrorColor,
		}},
		AllowedMentions: &responses.AllowedMentions{},
	}
}

func codeField(value string) string {
	if value == "" {
		value = " "
	}
	value = strings.ReplaceAll(value, "`", "'")
	const (
		fieldLimit     = 1024
		wrapperLength  = 2
		ellipsisLength = 3
	)
	if len(utf16.Encode([]rune(value))) > fieldLimit-wrapperLength {
		limit := fieldLimit - wrapperLength - ellipsisLength
		var truncated strings.Builder
		used := 0
		for _, r := range value {
			width := utf16.RuneLen(r)
			if width < 1 || used+width > limit {
				break
			}
			truncated.WriteRune(r)
			used += width
		}
		value = truncated.String() + "..."
	}
	return "`" + value + "`"
}
