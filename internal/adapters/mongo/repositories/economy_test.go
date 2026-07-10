package repositories

import (
	"context"
	"errors"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakemongo"
)

func TestEconomyCollectionNames(t *testing.T) {
	if CoinCollectionName != "coins" {
		t.Fatalf("coin collection = %s", CoinCollectionName)
	}
	if GiftChangeCollectionName != "gift_changes" {
		t.Fatalf("gift_change collection = %s", GiftChangeCollectionName)
	}
	if SignListCollectionName != "sign_lists" {
		t.Fatalf("sign_list collection = %s", SignListCollectionName)
	}
}

func TestNewEconomyRepositoryRequiresCollections(t *testing.T) {
	if _, err := NewEconomyRepository(nil, nil, nil); err == nil {
		t.Fatalf("expected nil collection error")
	}
}

func TestNewEconomyRepositoryFromDatabaseRequiresDatabase(t *testing.T) {
	if _, err := NewEconomyRepositoryFromDatabase(nil); err == nil {
		t.Fatalf("expected nil database error")
	}
}

func TestEconomyCoinGameWritesRequireTransactionRunner(t *testing.T) {
	repository := &EconomyRepository{}
	_, reserveErr := repository.ReserveCoinGameWager(context.Background(), domain.CoinGameCommand{
		GuildID:      "guild-1",
		ChallengerID: "user-1",
		OpponentID:   "user-2",
		Wager:        10,
		Kind:         domain.CoinGameKindKnowledge,
	})
	if !errors.Is(reserveErr, ErrCoinGameTransactionsRequired) {
		t.Fatalf("reserve error = %v", reserveErr)
	}
	_, settleErr := repository.SettleCoinGameWager(context.Background(), domain.CoinGameSettlementCommand{
		GuildID:          "guild-1",
		ChallengerID:     "user-1",
		OpponentID:       "user-2",
		ChallengerReturn: 20,
	})
	if !errors.Is(settleErr, ErrCoinGameTransactionsRequired) {
		t.Fatalf("settle error = %v", settleErr)
	}
}

func TestEconomyCoinGameTransactionRunnerConfiguration(t *testing.T) {
	repository := &EconomyRepository{}
	if err := repository.SetCoinGameTransactionRunner(nil); !errors.Is(err, ErrCoinGameTransactionsRequired) {
		t.Fatalf("nil runner error = %v", err)
	}
	if err := repository.SetCoinGameTransactionRunner(&fakemongo.TransactionRunner{}); err != nil {
		t.Fatalf("set runner: %v", err)
	}
}
