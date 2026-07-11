package economy

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	coreeconomy "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/services/economy"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/interactions"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/responses"
)

const (
	rockPaperScissorsErrorColor   = 0xED4245
	rockPaperScissorsOptionWager  = "使用多少代幣來進行"
	rockPaperScissorsOptionChoice = "剪刀石頭或布"
)

func (m Module) RockPaperScissorsHandler() interactions.Handler {
	return func(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
		if err := responder.Defer(ctx, responses.DeferOptions{}); err != nil {
			return err
		}
		command, ok := m.rockPaperScissorsCommandFromInteraction(interaction)
		if !ok {
			return responder.EditOriginal(ctx, rockPaperScissorsErrorMessage("很抱歉，出現了未知的錯誤，請重試!"))
		}
		if command.Wager > coreeconomy.MaxLegacyCoinBalance {
			return responder.EditOriginal(ctx, rockPaperScissorsErrorMessage("最高代幣設定數只能是999999999"))
		}
		if command.Wager <= 0 {
			return responder.EditOriginal(ctx, rockPaperScissorsErrorMessage("至少要大於1!!"))
		}
		result, err := m.rps.Play(ctx, command)
		if err != nil {
			return responder.EditOriginal(ctx, rockPaperScissorsErrorFromError(err))
		}
		avatarURL := strings.TrimSpace(interaction.Actor.GuildAvatarURL)
		if avatarURL == "" {
			avatarURL = interaction.Actor.AvatarURL
		}
		if err := responder.EditOriginal(ctx, rockPaperScissorsSuccessMessage(result, avatarURL, m.color())); err != nil {
			return err
		}
		return m.trackCommand(ctx, interaction, RockPaperScissorsCommandName)
	}
}

func (m Module) rockPaperScissorsCommandFromInteraction(interaction interactions.Interaction) (domain.RockPaperScissorsCommand, bool) {
	wager, ok := integerOption(interaction, rockPaperScissorsOptionWager)
	if !ok {
		return domain.RockPaperScissorsCommand{}, false
	}
	playerChoice := domain.RockPaperScissorsChoice(strings.TrimSpace(interaction.Options[rockPaperScissorsOptionChoice]))
	if option, ok := interaction.CommandOptions[rockPaperScissorsOptionChoice]; ok && option.String != "" {
		playerChoice = domain.RockPaperScissorsChoice(strings.TrimSpace(option.String))
	}
	choice := m.rpsChoice
	if choice == nil {
		choice = legacyRandomRockPaperScissorsChoice
	}
	command := domain.RockPaperScissorsCommand{
		GuildID:        interaction.Actor.GuildID,
		UserID:         interaction.Actor.UserID,
		Wager:          wager,
		PlayerChoice:   playerChoice,
		ComputerChoice: choice(),
	}.Normalize()
	return command, command.PlayerChoice.Valid() && command.ComputerChoice.Valid()
}

func rockPaperScissorsErrorFromError(err error) responses.Message {
	switch {
	case errors.Is(err, ports.ErrCoinBalanceNotFound):
		return rockPaperScissorsErrorMessage("你沒有足夠的代幣進行此次遊玩!")
	case errors.Is(err, ports.ErrCoinNegativeBalance):
		return rockPaperScissorsErrorMessage("你沒有足夠的代幣進行此次遊玩")
	case errors.Is(err, domain.ErrInvalidRockPaperScissorsCommand):
		return rockPaperScissorsErrorMessage("很抱歉，出現了未知的錯誤，請重試!")
	default:
		return rockPaperScissorsErrorMessage("很抱歉，出現了未知的錯誤，請重試!")
	}
}

func rockPaperScissorsErrorMessage(content string) responses.Message {
	return responses.Message{
		Embeds: []responses.Embed{{
			Title: "<a:Discord_AnimatedNo:1015989839809757295> | " + content,
			Color: rockPaperScissorsErrorColor,
		}},
		AllowedMentions: &responses.AllowedMentions{},
	}
}

func rockPaperScissorsSuccessMessage(result domain.RockPaperScissorsResult, avatarURL string, color int) responses.Message {
	label := "你失去了"
	if result.Outcome == domain.RockPaperScissorsOutcomeWin {
		label = "你獲得了"
	}
	amount := result.Delta
	if amount < 0 {
		amount = -amount
	}
	return responses.Message{
		Embeds: []responses.Embed{{
			Title:       "<a:girl:983775481100914788> __**剪刀石頭布!**__",
			Description: fmt.Sprintf("**你出了:**`%s%s`\n**我出了:**`%s%s`\n**%s:**`%d`個代幣", rockPaperScissorsEmoji(result.PlayerChoice), result.PlayerChoice, rockPaperScissorsEmoji(result.ComputerChoice), result.ComputerChoice, label, amount),
			Footer: &responses.EmbedFooter{
				Text:    "剪刀石頭布! | MHCAT",
				IconURL: avatarURL,
			},
			Color: color,
		}},
		AllowedMentions: &responses.AllowedMentions{},
	}
}

func rockPaperScissorsEmoji(choice domain.RockPaperScissorsChoice) string {
	switch choice.Normalize() {
	case domain.RockPaperScissorsChoiceScissors:
		return "✂️"
	case domain.RockPaperScissorsChoiceRock:
		return "🪨"
	default:
		return "🖐"
	}
}
