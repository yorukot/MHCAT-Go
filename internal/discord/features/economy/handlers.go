package economy

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	coreeconomy "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/services/economy"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/interactions"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/responses"
)

const (
	coinQueryCommandName       = "代幣查詢"
	coinQuerySuccessColor      = 0x5865F2
	coinQueryErrorColor        = 0xED4245
	legacyCoinNoBalanceTitle   = "<a:Discord_AnimatedNo:1015989839809757295> | 你還沒有任何代幣欸使用`/簽到`或是多講話，都可以獲得代幣喔!"
	legacyCoinQuestionLine     = "<:question:997374195229003776>我該如何獲取代幣?"
	legacyCoinMoneyEmoji       = "<:money:997374193026994236>"
	legacyCoinCatJumpEmoji     = "<a:catjump:984807173529931837>"
	legacyCoinFallbackUsername = "你"
)

func (m Module) CoinQueryHandler() interactions.Handler {
	return func(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
		if err := responder.Defer(ctx, responses.DeferOptions{Ephemeral: true}); err != nil {
			return err
		}
		targetUserID := strings.TrimSpace(interaction.Options["使用者"])
		subjectName := legacyCoinFallbackUsername
		if targetUserID == "" {
			targetUserID = interaction.Actor.UserID
		} else {
			subjectName = interaction.CommandOptions["使用者"].UserName
			if subjectName == "" {
				subjectName = m.lookupUsername(ctx, interaction.Actor.GuildID, targetUserID)
			}
		}
		result, err := m.query.Query(ctx, interaction.Actor.GuildID, targetUserID)
		if err != nil {
			if errors.Is(err, ports.ErrCoinBalanceNotFound) {
				if editErr := responder.EditOriginal(ctx, legacyCoinNoBalanceMessage()); editErr != nil {
					return editErr
				}
				return m.track(ctx, interaction)
			}
			return err
		}
		if err := responder.EditOriginal(ctx, legacyCoinQueryMessage(result, subjectName, interaction.Actor.AvatarURL, m.randomColor())); err != nil {
			return err
		}
		return m.track(ctx, interaction)
	}
}

func (m Module) lookupUsername(ctx context.Context, guildID string, userID string) string {
	if m.discord == nil || strings.TrimSpace(guildID) == "" || strings.TrimSpace(userID) == "" {
		return userID
	}
	info, err := m.discord.UserInfo(ctx, guildID, userID)
	if err != nil || strings.TrimSpace(info.Username) == "" {
		return userID
	}
	return info.Username
}

func legacyCoinNoBalanceMessage() responses.Message {
	return responses.Message{
		Ephemeral: true,
		Embeds: []responses.Embed{{
			Title: legacyCoinNoBalanceTitle,
			Color: coinQueryErrorColor,
		}},
	}
}

func legacyCoinQueryMessage(result coreeconomy.CoinQueryResult, subjectName string, actorAvatarURL string, color int) responses.Message {
	if strings.TrimSpace(subjectName) == "" {
		subjectName = legacyCoinFallbackUsername
	}
	return responses.Message{
		Ephemeral: true,
		Embeds: []responses.Embed{{
			Title:       fmt.Sprintf("%s%s目前有:`%d`個代幣!", legacyCoinMoneyEmoji, subjectName, result.Balance.Coins),
			Description: legacyCoinDescription(result.GachaCostText),
			Color:       color,
			Footer: &responses.EmbedFooter{
				Text:    legacyCoinFooterText(result, subjectName),
				IconURL: actorAvatarURL,
			},
		}},
		AllowedMentions: &responses.AllowedMentions{},
	}
}

func legacyCoinDescription(gachaCost string) string {
	return fmt.Sprintf("%s\n使用`/簽到`或是多多聊天都可以拿到代幣喔\n%s對了對了，代幣數到了%s可以進行扭蛋喔!\n如果代幣足夠也可以到代幣商城逛逛!", legacyCoinQuestionLine, legacyCoinCatJumpEmoji, gachaCost)
}

func legacyCoinFooterText(result coreeconomy.CoinQueryResult, subjectName string) string {
	if !result.ConfigFound {
		return fmt.Sprintf("%s還差:%d", subjectName, coreeconomy.DefaultGachaCost)
	}
	if !result.CanGacha {
		return fmt.Sprintf("%s還差:你還差%s就可以扭蛋了，加油!!", subjectName, result.MissingCoinsText)
	}
	return fmt.Sprintf("%s還差:你可以扭蛋了!!使用`/扭蛋`進行扭蛋", subjectName)
}

func (m Module) track(ctx context.Context, interaction interactions.Interaction) error {
	return m.trackCommand(ctx, interaction, coinQueryCommandName)
}

func (m Module) trackCommand(ctx context.Context, interaction interactions.Interaction, commandName string) error {
	if m.usage == nil {
		return nil
	}
	return m.usage.TrackCommand(ctx, ports.UsageEvent{
		CommandName: commandName,
		UserID:      interaction.Actor.UserID,
		GuildID:     interaction.Actor.GuildID,
		Feature:     m.feature,
	})
}
