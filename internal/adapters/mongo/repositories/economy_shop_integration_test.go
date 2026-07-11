package repositories

import (
	"context"
	"errors"
	"math"
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
		{Key: "name", Value: "item"}, {Key: "need_coin", Value: 20},
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

	for member, want := range map[string]string{"decimal": "30.5", "infinity": "Infinity"} {
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
}
