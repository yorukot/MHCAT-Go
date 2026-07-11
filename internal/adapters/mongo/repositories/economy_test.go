package repositories

import (
	"context"
	"errors"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakemongo"
	"go.mongodb.org/mongo-driver/v2/bson"
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

func TestLegacyFirstSignTodayUsesConfigPresence(t *testing.T) {
	if got := legacyFirstSignToday(false, 12345); got != 1 {
		t.Fatalf("missing config marker = %d", got)
	}
	if got := legacyFirstSignToday(true, 12345); got != 12345 {
		t.Fatalf("existing config marker = %d", got)
	}
}

func TestSignInRollingFilterPreservesFractionalThreshold(t *testing.T) {
	filter := signInRollingFilter("guild", "user", 25, 90.5)
	andClauses := filter[2].Value.(bson.A)
	orClause := andClauses[0].(bson.D)[0].Value.(bson.A)
	todayLimit := orClause[0].(bson.D)[0].Value.(bson.D)[0].Value
	if got, ok := todayLimit.(float64); !ok || got != 90.5 {
		t.Fatalf("rolling threshold = %#v", todayLimit)
	}
	nullType := orClause[2].(bson.D)[0].Value.(bson.D)[0]
	if nullType.Key != "$type" || nullType.Value != "null" {
		t.Fatalf("null today clause = %#v", nullType)
	}
}

func TestSignInCoinLimitFilterPreservesFractionalReward(t *testing.T) {
	filter := coinLimitFilter(25.5)
	orClause := filter[0].Value.(bson.A)
	coinLimit := orClause[0].(bson.D)[0].Value.(bson.D)[0].Value
	if got, ok := coinLimit.(float64); !ok || got != 999999973.5 {
		t.Fatalf("coin limit = %#v", coinLimit)
	}
}

func TestSignInCoinLimitAllowsExplicitNullButNotMissingCoin(t *testing.T) {
	filter := coinLimitFilter(25)
	orClause := filter[0].Value.(bson.A)
	nullType := orClause[1].(bson.D)[0].Value.(bson.D)[0]
	if nullType.Key != "$type" || nullType.Value != "null" {
		t.Fatalf("null coin clause = %#v", nullType)
	}
}

func TestSignInUpdateAddsRewardToNullAsZero(t *testing.T) {
	pipeline := signInUpdate(25.5, 123)
	set := pipeline[0][0].Value.(bson.D)
	add := set[0].Value.(bson.D)[0]
	values := add.Value.(bson.A)
	if add.Key != "$add" || values[1] != 25.5 {
		t.Fatalf("coin update = %#v", set[0])
	}
	ifNull := values[0].(bson.D)[0]
	if ifNull.Key != "$ifNull" {
		t.Fatalf("coin fallback = %#v", ifNull)
	}
	if set[1].Key != "today" || set[1].Value != int64(123) {
		t.Fatalf("today update = %#v", set[1])
	}
}
