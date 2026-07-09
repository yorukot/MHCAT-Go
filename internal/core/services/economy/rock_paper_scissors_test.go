package economy

import (
	"context"
	"errors"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakemongo"
)

func TestRockPaperScissorsWinAddsWagerWithoutUpperCap(t *testing.T) {
	repo := fakemongo.NewEconomyRepository()
	repo.PutBalance(domain.CoinBalance{GuildID: "guild", UserID: "user", Coins: MaxLegacyCoinBalance})
	result, err := (RockPaperScissorsService{Repository: repo}).Play(context.Background(), domain.RockPaperScissorsCommand{
		GuildID:        " guild ",
		UserID:         " user ",
		Wager:          10,
		PlayerChoice:   domain.RockPaperScissorsChoiceScissors,
		ComputerChoice: domain.RockPaperScissorsChoicePaper,
	})
	if err != nil {
		t.Fatalf("play: %v", err)
	}
	if result.Outcome != domain.RockPaperScissorsOutcomeWin || result.Delta != 10 || result.Balance.Coins != MaxLegacyCoinBalance+10 {
		t.Fatalf("unexpected result: %#v", result)
	}
}

func TestRockPaperScissorsTieSubtractsHalfWager(t *testing.T) {
	repo := fakemongo.NewEconomyRepository()
	repo.PutBalance(domain.CoinBalance{GuildID: "guild", UserID: "user", Coins: 10})
	result, err := (RockPaperScissorsService{Repository: repo}).Play(context.Background(), domain.RockPaperScissorsCommand{
		GuildID:        "guild",
		UserID:         "user",
		Wager:          5,
		PlayerChoice:   domain.RockPaperScissorsChoiceRock,
		ComputerChoice: domain.RockPaperScissorsChoiceRock,
	})
	if err != nil {
		t.Fatalf("play: %v", err)
	}
	if result.Outcome != domain.RockPaperScissorsOutcomeTie || result.Delta != -2 || result.Balance.Coins != 8 {
		t.Fatalf("unexpected tie result: %#v", result)
	}
}

func TestRockPaperScissorsRejectsMissingAndInsufficientBalance(t *testing.T) {
	repo := fakemongo.NewEconomyRepository()
	command := domain.RockPaperScissorsCommand{
		GuildID:        "guild",
		UserID:         "user",
		Wager:          5,
		PlayerChoice:   domain.RockPaperScissorsChoicePaper,
		ComputerChoice: domain.RockPaperScissorsChoiceScissors,
	}
	_, err := (RockPaperScissorsService{Repository: repo}).Play(context.Background(), command)
	if !errors.Is(err, ports.ErrCoinBalanceNotFound) {
		t.Fatalf("expected missing balance, got %v", err)
	}
	repo.PutBalance(domain.CoinBalance{GuildID: "guild", UserID: "user", Coins: 4})
	_, err = (RockPaperScissorsService{Repository: repo}).Play(context.Background(), command)
	if !errors.Is(err, ports.ErrCoinNegativeBalance) {
		t.Fatalf("expected insufficient balance, got %v", err)
	}
}

func TestRockPaperScissorsRejectsInvalidCommand(t *testing.T) {
	_, err := (RockPaperScissorsService{Repository: fakemongo.NewEconomyRepository()}).Play(context.Background(), domain.RockPaperScissorsCommand{
		GuildID:        "guild",
		UserID:         "user",
		Wager:          MaxLegacyCoinBalance + 1,
		PlayerChoice:   domain.RockPaperScissorsChoiceScissors,
		ComputerChoice: domain.RockPaperScissorsChoicePaper,
	})
	if !errors.Is(err, domain.ErrInvalidRockPaperScissorsCommand) {
		t.Fatalf("expected invalid command, got %v", err)
	}
}
