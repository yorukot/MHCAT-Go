package repositories

import (
	"context"
	"math"
	"sort"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestEconomyCoinResetMongoIntegrationPreservesScalarsAndDuplicateIteration(t *testing.T) {
	database := economyQueryIntegrationDatabase(t)
	repository, err := NewEconomyRepositoryFromDatabase(database)
	if err != nil {
		t.Fatalf("new repository: %v", err)
	}
	ctx := context.Background()
	if _, err := database.Collection(CoinCollectionName).InsertMany(ctx, []any{
		bson.D{{Key: "guild", Value: "guild-1"}, {Key: "member", Value: "duplicate"}, {Key: "coin", Value: 10.5}},
		bson.D{{Key: "guild", Value: "guild-1"}, {Key: "member", Value: "duplicate"}, {Key: "coin", Value: 20.5}},
		bson.D{{Key: "guild", Value: "guild-1"}, {Key: "member", Value: "decimal"}, {Key: "coin", Value: 5.5}},
		bson.D{{Key: "guild", Value: "guild-1"}, {Key: "member", Value: "null"}, {Key: "coin", Value: nil}},
		bson.D{{Key: "guild", Value: "guild-1"}, {Key: "member", Value: "infinity"}, {Key: "coin", Value: math.Inf(1)}},
	}); err != nil {
		t.Fatalf("seed balances: %v", err)
	}

	result, err := repository.ResetCoinBalances(ctx, domain.CoinResetCommand{GuildID: "guild-1", Divisor: 2})
	if err != nil || result.AffectedCount != 5 || result.Deleted {
		t.Fatalf("reset = %#v, err=%v", result, err)
	}
	duplicateCursor, err := database.Collection(CoinCollectionName).Find(ctx, bson.D{{Key: "guild", Value: "guild-1"}, {Key: "member", Value: "duplicate"}})
	if err != nil {
		t.Fatalf("find duplicates: %v", err)
	}
	var duplicates []struct {
		Coin float64 `bson:"coin"`
	}
	if err := duplicateCursor.All(ctx, &duplicates); err != nil {
		t.Fatalf("decode duplicates: %v", err)
	}
	if len(duplicates) != 2 {
		t.Fatalf("duplicates = %#v", duplicates)
	}
	values := []float64{duplicates[0].Coin, duplicates[1].Coin}
	sort.Float64s(values)
	if values[0] != 10 || values[1] != 20.5 {
		t.Fatalf("duplicate values = %#v", values)
	}
	for member, want := range map[string]string{"decimal": "3", "null": "0", "infinity": "Infinity"} {
		balance, err := repository.GetCoinBalance(ctx, "guild-1", member)
		if err != nil || balance.CoinsText != want {
			t.Fatalf("%s balance = %#v, err=%v", member, balance, err)
		}
	}
}
