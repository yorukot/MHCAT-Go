package economy

import (
	"context"
	"strings"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/commands"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/interactions"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakediscord"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakemongo"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakeusage"
)

func TestRockPaperScissorsDefinitionMatchesLegacyCommand(t *testing.T) {
	definition := RockPaperScissorsDefinition()
	if definition.Name != "剪刀石頭布" || definition.Description != "跟電腦剪刀時候布來獲得代幣(有賺有賠)" {
		t.Fatalf("definition = %#v", definition)
	}
	if len(definition.Options) != 2 {
		t.Fatalf("options = %#v", definition.Options)
	}
	if definition.Options[0].Name != "使用多少代幣來進行" || definition.Options[0].Type != commands.OptionTypeInteger || !definition.Options[0].Required {
		t.Fatalf("wager option = %#v", definition.Options[0])
	}
	choice := definition.Options[1]
	if choice.Name != "剪刀石頭或布" || len(choice.Choices) != 3 || !choice.Required {
		t.Fatalf("choice option = %#v", choice)
	}
}

func TestRockPaperScissorsWinUpdatesBalanceAndRendersLegacyEmbed(t *testing.T) {
	repo := fakemongo.NewEconomyRepository()
	repo.PutBalance(domain.CoinBalance{GuildID: "guild-1", UserID: "user-1", Coins: 100})
	usage := &fakeusage.Tracker{}
	module := NewRockPaperScissorsModule(repo, nil, usage)
	module.rpsChoice = func() domain.RockPaperScissorsChoice { return domain.RockPaperScissorsChoicePaper }
	module.color = func() int { return 0x123456 }
	responder := fakediscord.NewResponder()
	interaction := rockPaperScissorsInteraction(25, domain.RockPaperScissorsChoiceScissors)
	interaction.Actor.AvatarURL = "https://example.test/avatar.png"
	interaction.Actor.GuildAvatarURL = "https://example.test/guild-avatar.gif"
	if err := module.RockPaperScissorsHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	balance, err := repo.GetCoinBalance(context.Background(), "guild-1", "user-1")
	if err != nil || balance.Coins != 125 {
		t.Fatalf("balance = %#v err=%v", balance, err)
	}
	if len(responder.Defers) != 1 || responder.Defers[0].Ephemeral {
		t.Fatalf("defers = %#v", responder.Defers)
	}
	if len(responder.Edits) != 1 {
		t.Fatalf("edits = %#v", responder.Edits)
	}
	embed := responder.Edits[0].Embeds[0]
	if embed.Title != "<a:girl:983775481100914788> __**剪刀石頭布!**__" {
		t.Fatalf("title = %q", embed.Title)
	}
	wantDescription := "**你出了:**`✂️剪刀`\n**我出了:**`🖐布`\n**你獲得了:**`25`個代幣"
	if embed.Description != wantDescription {
		t.Fatalf("description = %q", embed.Description)
	}
	if embed.Footer == nil || embed.Footer.Text != "剪刀石頭布! | MHCAT" || embed.Footer.IconURL != "https://example.test/guild-avatar.gif" {
		t.Fatalf("footer = %#v", embed.Footer)
	}
	if embed.Color != 0x123456 {
		t.Fatalf("color = %#v", embed.Color)
	}
	if len(usage.Events) != 1 || usage.Events[0].CommandName != RockPaperScissorsCommandName || usage.Events[0].Feature != "economy-rps" {
		t.Fatalf("usage = %#v", usage.Events)
	}
}

func TestRockPaperScissorsTieLosesFlooredHalfWager(t *testing.T) {
	repo := fakemongo.NewEconomyRepository()
	repo.PutBalance(domain.CoinBalance{GuildID: "guild-1", UserID: "user-1", Coins: 100})
	module := NewRockPaperScissorsModule(repo, nil, nil)
	module.rpsChoice = func() domain.RockPaperScissorsChoice { return domain.RockPaperScissorsChoiceRock }
	responder := fakediscord.NewResponder()
	if err := module.RockPaperScissorsHandler()(context.Background(), rockPaperScissorsInteraction(25, domain.RockPaperScissorsChoiceRock), responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	balance, err := repo.GetCoinBalance(context.Background(), "guild-1", "user-1")
	if err != nil || balance.Coins != 88 {
		t.Fatalf("balance = %#v err=%v", balance, err)
	}
	if len(responder.Edits) != 1 || !strings.Contains(responder.Edits[0].Embeds[0].Description, "**你失去了:**`12`個代幣") {
		t.Fatalf("edits = %#v", responder.Edits)
	}
}

func TestRockPaperScissorsRejectsInvalidWagerWithLegacyErrors(t *testing.T) {
	module := NewRockPaperScissorsModule(fakemongo.NewEconomyRepository(), nil, nil)
	cases := []struct {
		name  string
		wager int64
		want  string
	}{
		{name: "too high", wager: 1000000000, want: "最高代幣設定數只能是999999999"},
		{name: "non positive", wager: 0, want: "至少要大於1!!"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			responder := fakediscord.NewResponder()
			err := module.RockPaperScissorsHandler()(context.Background(), rockPaperScissorsInteraction(tc.wager, domain.RockPaperScissorsChoiceScissors), responder)
			if err != nil {
				t.Fatalf("handler: %v", err)
			}
			if len(responder.Edits) != 1 || !strings.Contains(responder.Edits[0].Embeds[0].Title, tc.want) {
				t.Fatalf("edits = %#v", responder.Edits)
			}
		})
	}
}

func TestRockPaperScissorsRejectsMissingAndInsufficientBalanceWithLegacyErrors(t *testing.T) {
	repo := fakemongo.NewEconomyRepository()
	module := NewRockPaperScissorsModule(repo, nil, nil)
	responder := fakediscord.NewResponder()
	if err := module.RockPaperScissorsHandler()(context.Background(), rockPaperScissorsInteraction(5, domain.RockPaperScissorsChoiceScissors), responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Edits) != 1 || responder.Edits[0].Embeds[0].Title != "<a:Discord_AnimatedNo:1015989839809757295> | 你沒有足夠的代幣進行此次遊玩!" {
		t.Fatalf("missing balance edit = %#v", responder.Edits)
	}

	repo.PutBalance(domain.CoinBalance{GuildID: "guild-1", UserID: "user-1", Coins: 4})
	responder = fakediscord.NewResponder()
	if err := module.RockPaperScissorsHandler()(context.Background(), rockPaperScissorsInteraction(5, domain.RockPaperScissorsChoiceScissors), responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Edits) != 1 || responder.Edits[0].Embeds[0].Title != "<a:Discord_AnimatedNo:1015989839809757295> | 你沒有足夠的代幣進行此次遊玩" {
		t.Fatalf("insufficient edit = %#v", responder.Edits)
	}
}

func rockPaperScissorsInteraction(wager int64, choice domain.RockPaperScissorsChoice) interactions.Interaction {
	interaction := fakediscord.SlashInteractionWithOptions(RockPaperScissorsCommandName, "", map[string]string{
		rockPaperScissorsOptionWager:  "0",
		rockPaperScissorsOptionChoice: string(choice),
	})
	interaction.CommandOptions = map[string]interactions.CommandOptionValue{
		rockPaperScissorsOptionWager:  {Type: interactions.CommandOptionInteger, Int: wager},
		rockPaperScissorsOptionChoice: {Type: interactions.CommandOptionString, String: string(choice)},
	}
	return interaction
}
