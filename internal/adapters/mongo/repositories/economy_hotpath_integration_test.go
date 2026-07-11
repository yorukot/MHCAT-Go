package repositories

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestEconomyXPCoinRewardMongoIntegrationUsesAtomicUpsert(t *testing.T) {
	database := economyQueryIntegrationDatabase(t)
	repository, err := NewEconomyRepositoryFromDatabase(database)
	if err != nil {
		t.Fatalf("new economy repository: %v", err)
	}
	ctx := context.Background()
	if _, err := database.Collection(GiftChangeCollectionName).InsertOne(ctx, bson.D{
		{Key: "guild", Value: "guild-1"}, {Key: "xp_multiple", Value: 2.5},
	}); err != nil {
		t.Fatalf("seed economy config: %v", err)
	}
	if _, err := database.Collection(CoinCollectionName).InsertOne(ctx, bson.D{
		{Key: "guild", Value: "guild-1"}, {Key: "member", Value: "existing"}, {Key: "coin", Value: "10"},
	}); err != nil {
		t.Fatalf("seed existing balance: %v", err)
	}

	balance, err := repository.ApplyTextXPCoinReward(ctx, " guild-1 ", " existing ", 3)
	if err != nil || balance.Coins != 17 || balance.Today != 0 {
		t.Fatalf("text XP reward balance = %#v err=%v", balance, err)
	}
	balance, err = repository.ApplyVoiceXPCoinReward(ctx, "guild-1", "existing", 2)
	if err != nil || balance.Coins != 22 {
		t.Fatalf("voice XP reward balance = %#v err=%v", balance, err)
	}
	balance, err = repository.ApplyTextXPCoinReward(ctx, "guild-1", "new", 3)
	if err != nil || balance.Coins != 7 || balance.Today != 0 {
		t.Fatalf("upserted XP reward balance = %#v err=%v", balance, err)
	}
	if _, err := repository.ApplyTextXPCoinReward(ctx, " ", "user", 1); !errors.Is(err, domain.ErrInvalidEconomyQuery) {
		t.Fatalf("invalid XP reward error = %v", err)
	}
}

func TestEconomyXPCoinRewardMongoIntegrationConcurrentUpdatesDoNotLoseRewards(t *testing.T) {
	database := economyQueryIntegrationDatabase(t)
	repository, err := NewEconomyRepositoryFromDatabase(database)
	if err != nil {
		t.Fatalf("new economy repository: %v", err)
	}
	ctx := context.Background()
	if _, err := database.Collection(GiftChangeCollectionName).InsertOne(ctx, bson.D{
		{Key: "guild", Value: "guild-1"}, {Key: "xp_multiple", Value: 1},
	}); err != nil {
		t.Fatalf("seed economy config: %v", err)
	}
	if _, err := database.Collection(CoinCollectionName).InsertOne(ctx, bson.D{
		{Key: "guild", Value: "guild-1"}, {Key: "member", Value: "user-1"}, {Key: "coin", Value: 0}, {Key: "today", Value: 9},
	}); err != nil {
		t.Fatalf("seed reward balance: %v", err)
	}

	const updates = 20
	errorsByUpdate := make([]error, updates)
	var wait sync.WaitGroup
	wait.Add(updates)
	for index := range errorsByUpdate {
		go func() {
			defer wait.Done()
			_, errorsByUpdate[index] = repository.ApplyTextXPCoinReward(ctx, "guild-1", "user-1", 1)
		}()
	}
	wait.Wait()
	for _, err := range errorsByUpdate {
		if err != nil {
			t.Fatalf("concurrent XP reward error = %v", err)
		}
	}
	balance, err := repository.GetCoinBalance(ctx, "guild-1", "user-1")
	if err != nil || balance.Coins != updates || balance.Today != 9 {
		t.Fatalf("concurrent XP reward balance = %#v err=%v", balance, err)
	}
}

func TestEconomyCoinGameMongoIntegrationChecksBalances(t *testing.T) {
	repository, database := economyGameIntegrationRepository(t)
	insertEconomyGameBalances(t, database, 50, 60)
	checked, err := repository.CheckCoinGameBalances(context.Background(), economyGameIntegrationCommand(40))
	if err != nil || checked.Challenger.Coins != 50 || checked.Opponent.Coins != 60 || checked.Wager != 40 {
		t.Fatalf("checked coin game balances = %#v err=%v", checked, err)
	}
}

func TestEconomyShopMongoIntegrationLifecycle(t *testing.T) {
	repository, _ := economyGameIntegrationRepository(t)
	ctx := context.Background()
	item := domain.ShopItem{
		GuildID: " guild-1 ", CommodityID: 1, Name: "item", NeedCoins: 10,
		Description: "description", Code: "code", AutoDelete: true, RoleID: " role-1 ", Count: 2,
	}
	created, err := repository.CreateShopItem(ctx, item)
	if err != nil || created.GuildID != "guild-1" || created.RoleID != "role-1" {
		t.Fatalf("created shop item = %#v err=%v", created, err)
	}
	if _, err := repository.CreateShopItem(ctx, item); !errors.Is(err, ports.ErrShopItemExists) {
		t.Fatalf("duplicate shop item error = %v", err)
	}
	items, err := repository.ListShopItems(ctx, " guild-1 ")
	if err != nil || len(items) != 1 || items[0].CommodityID != 1 {
		t.Fatalf("shop items = %#v err=%v", items, err)
	}
	deleted, err := repository.DeleteShopItem(ctx, " guild-1 ", 1)
	if err != nil || deleted.Name != "item" {
		t.Fatalf("deleted shop item = %#v err=%v", deleted, err)
	}
	if _, err := repository.DeleteShopItem(ctx, "guild-1", 1); !errors.Is(err, ports.ErrShopItemMissing) {
		t.Fatalf("missing shop item delete error = %v", err)
	}
}

func TestEconomyShopMongoTransactionIntegrationDeletesFullyPurchasedStock(t *testing.T) {
	repository, database := economyGameIntegrationRepository(t)
	ctx := context.Background()
	if _, err := repository.CreateShopItem(ctx, domain.ShopItem{
		GuildID: "guild-1", CommodityID: 1, Name: "item", NeedCoins: 10,
		Description: "description", AutoDelete: true, Count: 2,
	}); err != nil {
		t.Fatalf("create shop item: %v", err)
	}
	if _, err := database.Collection(CoinCollectionName).InsertOne(ctx, bson.D{
		{Key: "guild", Value: "guild-1"}, {Key: "member", Value: "user-1"}, {Key: "coin", Value: 100},
	}); err != nil {
		t.Fatalf("seed shop balance: %v", err)
	}
	result, err := repository.PurchaseShopItem(ctx, domain.ShopPurchaseCommand{
		GuildID: "guild-1", UserID: "user-1", CommodityID: 1, Quantity: 2,
	})
	if err != nil || result.Balance.Coins != 80 || result.TotalCost != 20 {
		t.Fatalf("full-stock purchase = %#v err=%v", result, err)
	}
	if _, err := repository.GetShopItem(ctx, "guild-1", 1); !errors.Is(err, ports.ErrShopItemMissing) {
		t.Fatalf("fully purchased item error = %v", err)
	}
}

func TestEconomyShopMongoTransactionIntegrationConcurrentLastItemChargesOnce(t *testing.T) {
	repository, database := economyGameIntegrationRepository(t)
	ctx := context.Background()
	if _, err := repository.CreateShopItem(ctx, domain.ShopItem{
		GuildID: "guild-1", CommodityID: 1, Name: "last", NeedCoins: 10,
		Description: "description", AutoDelete: true, Count: 1,
	}); err != nil {
		t.Fatalf("create shop item: %v", err)
	}
	if _, err := database.Collection(CoinCollectionName).InsertOne(ctx, bson.D{
		{Key: "guild", Value: "guild-1"}, {Key: "member", Value: "user-1"}, {Key: "coin", Value: 100},
	}); err != nil {
		t.Fatalf("seed shop balance: %v", err)
	}

	command := domain.ShopPurchaseCommand{GuildID: "guild-1", UserID: "user-1", CommodityID: 1, Quantity: 1}
	errorsByPurchase := make([]error, 2)
	var wait sync.WaitGroup
	wait.Add(2)
	for index := range errorsByPurchase {
		go func() {
			defer wait.Done()
			_, errorsByPurchase[index] = repository.PurchaseShopItem(ctx, command)
		}()
	}
	wait.Wait()
	succeeded := 0
	missing := 0
	for _, err := range errorsByPurchase {
		switch {
		case err == nil:
			succeeded++
		case errors.Is(err, ports.ErrShopItemMissing):
			missing++
		default:
			t.Fatalf("concurrent shop purchase error = %v", err)
		}
	}
	if succeeded != 1 || missing != 1 {
		t.Fatalf("concurrent purchases succeeded=%d missing=%d errors=%v", succeeded, missing, errorsByPurchase)
	}
	balance, err := repository.GetCoinBalance(ctx, "guild-1", "user-1")
	if err != nil || balance.Coins != 90 {
		t.Fatalf("balance after concurrent purchase = %#v err=%v", balance, err)
	}
}

func TestEconomyShopMongoTransactionIntegrationRollsBackInventoryWhenChargeFails(t *testing.T) {
	repository, database := economyGameIntegrationRepository(t)
	ctx := context.Background()
	if _, err := repository.CreateShopItem(ctx, domain.ShopItem{
		GuildID: "guild-1", CommodityID: 1, Name: "item", NeedCoins: 10,
		Description: "description", AutoDelete: true, Count: 2,
	}); err != nil {
		t.Fatalf("create shop item: %v", err)
	}
	if _, err := database.Collection(CoinCollectionName).InsertOne(ctx, bson.D{
		{Key: "guild", Value: "guild-1"}, {Key: "member", Value: "user-1"}, {Key: "coin", Value: 100},
	}); err != nil {
		t.Fatalf("seed shop balance: %v", err)
	}
	if err := database.RunCommand(ctx, bson.D{
		{Key: "collMod", Value: CoinCollectionName},
		{Key: "validator", Value: bson.D{{Key: "$expr", Value: bson.D{{Key: "$eq", Value: bson.A{"$coin", 100}}}}}},
		{Key: "validationLevel", Value: "strict"},
		{Key: "validationAction", Value: "error"},
	}).Err(); err != nil {
		t.Fatalf("configure coin validator: %v", err)
	}

	if _, err := repository.PurchaseShopItem(ctx, domain.ShopPurchaseCommand{
		GuildID: "guild-1", UserID: "user-1", CommodityID: 1, Quantity: 1,
	}); err == nil {
		t.Fatal("expected shop charge failure")
	}
	item, err := repository.GetShopItem(ctx, "guild-1", 1)
	if err != nil || item.CountText != "2" {
		t.Fatalf("shop item after rollback = %#v err=%v", item, err)
	}
	balance, err := repository.GetCoinBalance(ctx, "guild-1", "user-1")
	if err != nil || balance.Coins != 100 {
		t.Fatalf("shop balance after rollback = %#v err=%v", balance, err)
	}
}
