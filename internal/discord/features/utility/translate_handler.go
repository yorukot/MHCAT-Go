package utility

import (
	"context"
	"errors"
	"strings"

	coreutility "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/services/utility"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/interactions"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/responses"
)

const (
	translateLoadingColor = 0x57F287
	translateErrorColor   = 0xEA0000
)

func (m Module) TranslateHandler() interactions.Handler {
	return func(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
		if err := responder.Defer(ctx, responses.DeferOptions{}); err != nil {
			return err
		}
		if err := responder.EditOriginal(ctx, translateLoadingMessage()); err != nil {
			return err
		}
		source := strings.TrimSpace(interaction.Options["要的翻譯"])
		targetLanguage := strings.TrimSpace(interaction.Options["目標語言"])
		result, err := m.translate.Translate(ctx, source, targetLanguage)
		if err != nil {
			return responder.EditOriginal(ctx, translateErrorMessage(err))
		}
		if err := responder.EditOriginal(ctx, translateResultMessage(interaction, source, targetLanguage, result.Text, m.randomTranslateColor())); err != nil {
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
	value = strings.TrimSpace(value)
	if value == "" {
		value = " "
	}
	if len([]rune(value)) > 1000 {
		runes := []rune(value)
		value = string(runes[:1000]) + "..."
	}
	value = strings.ReplaceAll(value, "`", "'")
	return "`" + value + "`"
}
