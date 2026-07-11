package repositories

import (
	"context"
	"sort"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestEconomyCoinAdminMongoIntegrationPreservesScalarsAndOneDuplicate(t *testing.T) {
	database := economyQueryIntegrationDatabase(t)
	repository, err := NewEconomyRepositoryFromDatabase(database)
	if err != nil {
		t.Fatalf("new repository: %v", err)
	}
	ctx := context.Background()
	if _, err := database.Collection(CoinCollectionName).InsertMany(ctx, []any{
		bson.D{{Key: "guild", Value: "guild-1"}, {Key: "member", Value: "duplicate"}, {Key: "coin", Value: 10.5}, {Key: "today", Value: 1}},
		bson.D{{Key: "guild", Value: "guild-1"}, {Key: "member", Value: "duplicate"}, {Key: "coin", Value: 20.25}, {Key: "today", Value: 1}},
		bson.D{{Key: "guild", Value: "guild-1"}, {Key: "member", Value: "null"}, {Key: "coin", Value: nil}, {Key: "today", Value: 1}},
	}); err != nil {
		t.Fatalf("seed balances: %v", err)
	}

	result, err := repository.AdjustCoinBalance(ctx, domain.CoinAdminCommand{GuildID: "guild-1", UserID: "duplicate", Operation: domain.CoinAdminOperationAdd, Amount: 1})
	if err != nil || result.Balance.CoinsText != "11.5" {
		t.Fatalf("adjust duplicate = %#v, err=%v", result, err)
	}
	cursor, err := database.Collection(CoinCollectionName).Find(ctx, bson.D{{Key: "guild", Value: "guild-1"}, {Key: "member", Value: "duplicate"}})
	if err != nil {
		t.Fatalf("find duplicates: %v", err)
	}
	var rows []struct {
		Coin float64 `bson:"coin"`
	}
	if err := cursor.All(ctx, &rows); err != nil {
		t.Fatalf("decode duplicates: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("duplicate rows = %#v", rows)
	}
	coins := []float64{rows[0].Coin, rows[1].Coin}
	sort.Float64s(coins)
	if coins[0] != 11.5 || coins[1] != 20.25 {
		t.Fatalf("duplicate coins = %#v", coins)
	}

	nullResult, err := repository.AdjustCoinBalance(ctx, domain.CoinAdminCommand{GuildID: "guild-1", UserID: "null", Operation: domain.CoinAdminOperationAdd, Amount: 5})
	if err != nil || nullResult.Balance.CoinsText != "5" {
		t.Fatalf("adjust null = %#v, err=%v", nullResult, err)
	}
}
