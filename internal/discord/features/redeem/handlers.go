package redeem

import (
	"context"
	"errors"
	"strings"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/interactions"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/responses"
)

const (
	redeemSuccessColor = 0x57F287
	redeemErrorColor   = 0xED4245
	successIconURL     = "https://media.discordapp.net/attachments/991337796960784424/1078883215462383697/success.gif"
)

func (m Module) Handler() interactions.Handler {
	return func(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
		if err := responder.Defer(ctx, responses.DeferOptions{Ephemeral: true}); err != nil {
			return err
		}
		code := strings.TrimSpace(interaction.Options[optionCode])
		if option, ok := interaction.CommandOptions[optionCode]; ok {
			code = strings.TrimSpace(option.String)
		}
		if err := m.service.Redeem(ctx, interaction.Actor.GuildID, code); err != nil {
			return responder.EditOriginal(ctx, redeemErrorMessage(redeemErrorText(err)))
		}
		if err := responder.EditOriginal(ctx, redeemSuccessMessage()); err != nil {
			return err
		}
		return m.track(ctx, interaction)
	}
}

func redeemErrorText(err error) string {
	switch {
	case errors.Is(err, ports.ErrRedeemCodeNotFound):
		return "找不到這個代碼!"
	case errors.Is(err, ports.ErrRedeemCodeExpired):
		return "這個代碼為防止遭人惡意使用，已過期，如果你是代碼擁有者，請前往支援伺服器開啟客服頻道!"
	default:
		return "很抱歉，出現了未知的錯誤，請重試!"
	}
}

func redeemErrorMessage(content string) responses.Message {
	return responses.Message{
		Embeds: []responses.Embed{{
			Title: "<a:Discord_AnimatedNo:1015989839809757295> | " + content,
			Color: redeemErrorColor,
		}},
		Ephemeral:       true,
		AllowedMentions: &responses.AllowedMentions{},
	}
}

func redeemSuccessMessage() responses.Message {
	return responses.Message{
		Embeds: []responses.Embed{{
			Author: &responses.EmbedAuthor{
				Name:    "成功兌換代碼!",
				IconURL: successIconURL,
			},
			Footer: &responses.EmbedFooter{Text: "你可以使用/查看餘額進行查詢剩餘餘額"},
			Color:  redeemSuccessColor,
		}},
		Ephemeral:       true,
		AllowedMentions: &responses.AllowedMentions{},
	}
}
