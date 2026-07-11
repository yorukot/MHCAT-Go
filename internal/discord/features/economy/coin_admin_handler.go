package economy

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
	coinAdminManageMessagesPermission = int64(8192)
	coinAdminErrorColor               = 0xED4245
	coinAdminOptionUser               = "使用者"
	coinAdminOptionOperation          = "增加或減少"
	coinAdminOptionAmount             = "數量"
)

func (m Module) CoinAdminHandler() interactions.Handler {
	return func(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
		if err := responder.Defer(ctx, responses.DeferOptions{}); err != nil {
			return err
		}
		if !interaction.Actor.HasPermission(coinAdminManageMessagesPermission) {
			return responder.EditOriginal(ctx, coinAdminErrorMessage("你需要有`訊息管理`才能使用此指令"))
		}
		command, ok := coinAdminCommandFromInteraction(interaction)
		if !ok {
			return responder.EditOriginal(ctx, coinAdminErrorMessage("很抱歉，出現了未知的錯誤，請重試!"))
		}
		result, err := m.coinAdmin.Adjust(ctx, command)
		if err != nil {
			return responder.EditOriginal(ctx, coinAdminErrorFromError(err))
		}
		targetName := m.lookupUsername(ctx, interaction.Actor.GuildID, command.UserID)
		avatarURL := strings.TrimSpace(interaction.Actor.GuildAvatarURL)
		if avatarURL == "" {
			avatarURL = interaction.Actor.AvatarURL
		}
		if err := responder.EditOriginal(ctx, coinAdminSuccessMessage(result, targetName, avatarURL, m.randomColor())); err != nil {
			return err
		}
		return m.trackCommand(ctx, interaction, CoinAdminCommandName)
	}
}

func coinAdminCommandFromInteraction(interaction interactions.Interaction) (domain.CoinAdminCommand, bool) {
	amount, ok := integerOption(interaction, coinAdminOptionAmount)
	if !ok {
		return domain.CoinAdminCommand{}, false
	}
	command := domain.CoinAdminCommand{
		GuildID:   interaction.Actor.GuildID,
		UserID:    strings.TrimSpace(interaction.Options[coinAdminOptionUser]),
		Operation: domain.CoinAdminOperation(strings.TrimSpace(interaction.Options[coinAdminOptionOperation])),
		Amount:    amount,
	}
	if option, ok := interaction.CommandOptions[coinAdminOptionUser]; ok && option.String != "" {
		command.UserID = strings.TrimSpace(option.String)
	}
	if option, ok := interaction.CommandOptions[coinAdminOptionOperation]; ok && option.String != "" {
		command.Operation = domain.CoinAdminOperation(strings.TrimSpace(option.String))
	}
	command = command.Normalize()
	return command, command.Validate() == nil
}

func coinAdminErrorFromError(err error) responses.Message {
	switch {
	case errors.Is(err, ports.ErrCoinNegativeBalance):
		return coinAdminErrorMessage("不可減到負數!")
	case errors.Is(err, ports.ErrCoinLimitExceeded):
		return coinAdminErrorMessage("不可以加超過`999999999`!!")
	case errors.Is(err, domain.ErrInvalidCoinAdminCommand):
		return coinAdminErrorMessage("很抱歉，出現了未知的錯誤，請重試!")
	default:
		return coinAdminErrorMessage("很抱歉，出現了未知的錯誤，請重試!")
	}
}

func coinAdminErrorMessage(content string) responses.Message {
	return responses.Message{
		Embeds: []responses.Embed{{
			Title: "<a:error:980086028113182730> | " + content,
			Color: coinAdminErrorColor,
		}},
		AllowedMentions: &responses.AllowedMentions{},
	}
}

func coinAdminSuccessMessage(result domain.CoinAdminResult, targetName string, avatarURL string, color int) responses.Message {
	action := "增加"
	if result.Delta < 0 {
		action = "減少"
	}
	amount := result.Delta
	if amount < 0 {
		amount = -amount
	}
	targetName = strings.TrimSpace(targetName)
	if targetName == "" {
		targetName = result.Balance.UserID
	}
	return responses.Message{
		Embeds: []responses.Embed{{
			Title: fmt.Sprintf("<:money:997374193026994236>已為%s`%s`:`%d`個代幣!", targetName, action, amount),
			Footer: &responses.EmbedFooter{
				Text:    fmt.Sprintf("%s%d", action, amount),
				IconURL: avatarURL,
			},
			Color: color,
		}},
		AllowedMentions: &responses.AllowedMentions{},
	}
}
