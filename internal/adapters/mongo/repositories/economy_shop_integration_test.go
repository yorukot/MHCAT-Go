package repositories

import (
	"context"
	"errors"
	"math"
	"slices"
	"sort"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestEconomyShopMongoIntegrationPreservesBalanceScalars(t *testing.T) {
	database := economyQueryIntegrationDatabase(t)
	repository, err := NewEconomyRepositoryFromDatabase(database)
	if err != nil {
		t.Fatalf("new repository: %v", err)
	}
	ctx := context.Background()
	if _, err := database.Collection(ShopItemCollectionName).InsertOne(ctx, bson.D{
		{Key: "guild", Value: "guild-1"}, {Key: "commodity_id", Value: 1},
		{Key: "name", Value: "item"}, {Key: "need_coin", Value: 20.5},
		{Key: "commodity_description", Value: "desc"}, {Key: "code", Value: nil},
		{Key: "auto_delete", Value: false}, {Key: "role", Value: nil}, {Key: "commodity_count", Value: 1},
	}); err != nil {
		t.Fatalf("seed item: %v", err)
	}
	if _, err := database.Collection(CoinCollectionName).InsertMany(ctx, []any{
		bson.D{{Key: "guild", Value: "guild-1"}, {Key: "member", Value: "decimal"}, {Key: "coin", Value: 50.5}},
		bson.D{{Key: "guild", Value: "guild-1"}, {Key: "member", Value: "null"}, {Key: "coin", Value: nil}},
		bson.D{{Key: "guild", Value: "guild-1"}, {Key: "member", Value: "infinity"}, {Key: "coin", Value: math.Inf(1)}},
	}); err != nil {
		t.Fatalf("seed balances: %v", err)
	}

	for member, want := range map[string]string{"decimal": "30", "infinity": "Infinity"} {
		result, err := repository.PurchaseShopItem(ctx, domain.ShopPurchaseCommand{GuildID: "guild-1", UserID: member, CommodityID: 1, Quantity: 1})
		if err != nil || result.Balance.CoinsText != want {
			t.Fatalf("purchase %s = %#v, err=%v", member, result, err)
		}
	}
	if _, err := repository.PurchaseShopItem(ctx, domain.ShopPurchaseCommand{GuildID: "guild-1", UserID: "null", CommodityID: 1, Quantity: 1}); !errors.Is(err, ports.ErrShopInsufficientCoin) {
		t.Fatalf("null purchase error = %v", err)
	}
	nullBalance, err := repository.GetCoinBalance(ctx, "guild-1", "null")
	if err != nil || nullBalance.CoinsText != "null" {
		t.Fatalf("null balance = %#v, err=%v", nullBalance, err)
	}
	if _, err := database.Collection(ShopItemCollectionName).InsertOne(ctx, bson.D{
		{Key: "guild", Value: "guild-1"}, {Key: "commodity_id", Value: 2}, {Key: "name", Value: "free"},
		{Key: "need_coin", Value: nil}, {Key: "commodity_description", Value: "desc"},
		{Key: "code", Value: nil}, {Key: "auto_delete", Value: false}, {Key: "role", Value: nil}, {Key: "commodity_count", Value: 1},
	}); err != nil {
		t.Fatalf("seed null-price item: %v", err)
	}
	free, err := repository.PurchaseShopItem(ctx, domain.ShopPurchaseCommand{GuildID: "guild-1", UserID: "decimal", CommodityID: 2, Quantity: 1})
	if err != nil || free.TotalCost != 0 || free.Balance.CoinsText != "30" {
		t.Fatalf("free purchase = %#v, err=%v", free, err)
	}
}

func TestEconomyShopMongoIntegrationUpdatesOneBalanceDuplicate(t *testing.T) {
	database := economyQueryIntegrationDatabase(t)
	repository, err := NewEconomyRepositoryFromDatabase(database)
	if err != nil {
		t.Fatalf("new repository: %v", err)
	}
	ctx := context.Background()
	if _, err := database.Collection(ShopItemCollectionName).InsertOne(ctx, bson.D{
		{Key: "guild", Value: "guild-1"}, {Key: "commodity_id", Value: 1}, {Key: "name", Value: "item"},
		{Key: "need_coin", Value: 20}, {Key: "commodity_description", Value: "desc"},
		{Key: "code", Value: nil}, {Key: "auto_delete", Value: false}, {Key: "role", Value: nil}, {Key: "commodity_count", Value: 1},
	}); err != nil {
		t.Fatalf("seed item: %v", err)
	}
	if _, err := database.Collection(CoinCollectionName).InsertMany(ctx, []any{
		bson.D{{Key: "guild", Value: "guild-1"}, {Key: "member", Value: "user-1"}, {Key: "coin", Value: 50}},
		bson.D{{Key: "guild", Value: "guild-1"}, {Key: "member", Value: "user-1"}, {Key: "coin", Value: 80}},
	}); err != nil {
		t.Fatalf("seed duplicate balances: %v", err)
	}
	if _, err := repository.PurchaseShopItem(ctx, domain.ShopPurchaseCommand{GuildID: "guild-1", UserID: "user-1", CommodityID: 1, Quantity: 1}); err != nil {
		t.Fatalf("purchase: %v", err)
	}

	cursor, err := database.Collection(CoinCollectionName).Find(ctx, bson.D{{Key: "guild", Value: "guild-1"}, {Key: "member", Value: "user-1"}})
	if err != nil {
		t.Fatalf("find duplicate balances: %v", err)
	}
	defer cursor.Close(ctx)
	var rows []struct {
		Coin float64 `bson:"coin"`
	}
	if err := cursor.All(ctx, &rows); err != nil {
		t.Fatalf("decode duplicate balances: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("duplicate rows = %#v", rows)
	}
	got := []float64{rows[0].Coin, rows[1].Coin}
	sort.Float64s(got)
	validOutcomes := [][]float64{{30, 50}, {30, 80}, {50, 60}, {60, 80}}
	if !slices.ContainsFunc(validOutcomes, func(want []float64) bool { return slices.Equal(got, want) }) {
		t.Fatalf("balances = %v, want one arbitrary row updated from one arbitrary read", got)
	}
}
