package repositories

import (
	"context"
	"math"
	"sort"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestEconomyRPSMongoIntegrationPreservesScalarsAndOneDuplicate(t *testing.T) {
	database := economyQueryIntegrationDatabase(t)
	repository, err := NewEconomyRepositoryFromDatabase(database)
	if err != nil {
		t.Fatalf("new repository: %v", err)
	}
	ctx := context.Background()
	if _, err := database.Collection(CoinCollectionName).InsertMany(ctx, []any{
		bson.D{{Key: "guild", Value: "guild-1"}, {Key: "member", Value: "duplicate"}, {Key: "coin", Value: 10.5}},
		bson.D{{Key: "guild", Value: "guild-1"}, {Key: "member", Value: "duplicate"}, {Key: "coin", Value: 20.25}},
		bson.D{{Key: "guild", Value: "guild-1"}, {Key: "member", Value: "infinity"}, {Key: "coin", Value: math.Inf(1)}},
	}); err != nil {
		t.Fatalf("seed balances: %v", err)
	}

	win := domain.RockPaperScissorsCommand{GuildID: "guild-1", UserID: "duplicate", Wager: 5, PlayerChoice: domain.RockPaperScissorsChoiceScissors, ComputerChoice: domain.RockPaperScissorsChoicePaper}
	result, err := repository.ApplyRockPaperScissors(ctx, win)
	if err != nil || result.PreviousBalance != 10.5 || result.Balance.CoinsText != "15.5" {
		t.Fatalf("play duplicate = %#v, err=%v", result, err)
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
	if coins[0] != 15.5 || coins[1] != 20.25 {
		t.Fatalf("duplicate coins = %#v", coins)
	}

	loss := domain.RockPaperScissorsCommand{GuildID: "guild-1", UserID: "infinity", Wager: 5, PlayerChoice: domain.RockPaperScissorsChoiceScissors, ComputerChoice: domain.RockPaperScissorsChoiceRock}
	infinite, err := repository.ApplyRockPaperScissors(ctx, loss)
	if err != nil || infinite.Balance.CoinsText != "Infinity" {
		t.Fatalf("play infinity = %#v, err=%v", infinite, err)
	}
}
