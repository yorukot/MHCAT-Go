package repositories

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	mhcatmongo "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/adapters/mongo"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	"go.mongodb.org/mongo-driver/v2/bson"
	drivermongo "go.mongodb.org/mongo-driver/v2/mongo"
)

func TestGachaRepositoryDependenciesAreRequired(t *testing.T) {
	database := xpAccrualIntegrationDatabase(t)
	if _, err := NewGachaRepository(nil, database.Collection(GiftChangeCollectionName), database.Collection(CoinCollectionName)); err == nil {
		t.Fatal("expected gifts collection error")
	}
	if _, err := NewGachaRepository(database.Collection(GiftCollectionName), nil, database.Collection(CoinCollectionName)); err == nil {
		t.Fatal("expected gift changes collection error")
	}
	if _, err := NewGachaRepository(database.Collection(GiftCollectionName), database.Collection(GiftChangeCollectionName), nil); err == nil {
		t.Fatal("expected coins collection error")
	}
	if _, err := NewGachaRepositoryFromDatabase(nil); err == nil {
		t.Fatal("expected database error")
	}
	repository, err := NewGachaRepositoryFromDatabase(database)
	if err != nil {
		t.Fatalf("new gacha repository: %v", err)
	}
	if err := repository.SetGachaDrawTransactionRunner(nil); !errors.Is(err, ErrGachaDrawTransactionsRequired) {
		t.Fatalf("nil transaction runner error = %v", err)
	}
	_, err = repository.DrawGacha(context.Background(), domain.GachaDrawRequest{
		GuildID: "guild-1", UserID: "user-1", PaidDraws: 1, ActualDraws: 1, RandomValues: []float64{0},
	})
	if !errors.Is(err, ErrGachaDrawTransactionsRequired) {
		t.Fatalf("draw without transaction runner error = %v", err)
	}
}

func TestGachaMongoIntegrationLifecycleAndTransactionalDraw(t *testing.T) {
	database, repository := gachaIntegrationRepository(t)
	ctx := context.Background()
	seedGachaEconomy(t, database, "guild-1", "user-1", 1000, 100)

	prize := domain.GachaPrizeConfig{
		GuildID: " guild-1 ", Name: "limited", Code: "old-code", Chance: 100,
		AutoDelete: true, Count: 2, GiveCoin: 5,
	}
	if err := repository.CreateGachaPrize(ctx, prize); err != nil {
		t.Fatalf("create gacha prize: %v", err)
	}
	if err := repository.CreateGachaPrize(ctx, prize); !errors.Is(err, ports.ErrGachaPrizeExists) {
		t.Fatalf("duplicate gacha prize error = %v", err)
	}
	count, err := repository.CountGachaPrizes(ctx, "guild-1")
	if err != nil || count != 1 {
		t.Fatalf("gacha prize count = %d err=%v", count, err)
	}

	edited, err := repository.EditGachaPrize(ctx, domain.GachaPrizeEdit{
		GuildID: " guild-1 ", Name: "limited", Code: "new-code", Chance: 100,
		ChanceSet: true, AutoDelete: true, Count: 3, GiveCoin: 7,
	})
	if err != nil || edited.Code != "new-code" || edited.Chance != 100 || edited.Count != 3 || edited.GiveCoin != 7 || !edited.AutoDelete {
		t.Fatalf("edited gacha prize = %#v err=%v", edited, err)
	}
	prizes, err := repository.ListGachaPrizes(ctx, "guild-1")
	if err != nil || len(prizes) != 1 || prizes[0].Count != 3 || prizes[0].Name != "limited" {
		t.Fatalf("gacha prizes = %#v err=%v", prizes, err)
	}

	result, err := repository.DrawGacha(ctx, domain.GachaDrawRequest{
		GuildID: " guild-1 ", UserID: " user-1 ", PaidDraws: 2, ActualDraws: 3,
		RandomValues: []float64{0, 0, 0},
	})
	if err != nil {
		t.Fatalf("draw gacha: %v", err)
	}
	if result.BalanceBefore != 1000 || result.BalanceAfter != 821 || result.Cost != 200 || len(result.Prizes) != 3 {
		t.Fatalf("gacha draw result = %#v", result)
	}
	for _, drawn := range result.Prizes {
		if drawn.Name != "limited" || drawn.Code != "new-code" || drawn.GiveCoin != 7 || drawn.Air {
			t.Fatalf("drawn prize = %#v", drawn)
		}
	}
	if count, err := repository.CountGachaPrizes(ctx, "guild-1"); err != nil || count != 0 {
		t.Fatalf("depleted gacha prize count = %d err=%v", count, err)
	}
	if _, err := repository.DeleteGachaPrize(ctx, "guild-1", "limited"); !errors.Is(err, ports.ErrGachaPrizeMissing) {
		t.Fatalf("delete depleted prize error = %v", err)
	}
	assertGachaBalance(t, database, "guild-1", "user-1", 821)

	persistent := domain.GachaPrizeConfig{GuildID: "guild-1", Name: "persistent", Chance: 100, Count: 1, GiveCoin: 3}
	if err := repository.CreateGachaPrize(ctx, persistent); err != nil {
		t.Fatalf("create persistent prize: %v", err)
	}
	result, err = repository.DrawGacha(ctx, domain.GachaDrawRequest{
		GuildID: "guild-1", UserID: "user-1", PaidDraws: 1, ActualDraws: 1, RandomValues: []float64{0},
	})
	if err != nil || result.BalanceAfter != 724 {
		t.Fatalf("persistent gacha draw = %#v err=%v", result, err)
	}
	deleted, err := repository.DeleteGachaPrize(ctx, "guild-1", "persistent")
	if err != nil || deleted.Name != "persistent" {
		t.Fatalf("deleted gacha prize = %#v err=%v", deleted, err)
	}
}

func TestGachaMongoIntegrationRejectsMissingAndInsufficientBalances(t *testing.T) {
	database, repository := gachaIntegrationRepository(t)
	ctx := context.Background()
	if err := repository.CreateGachaPrize(ctx, domain.GachaPrizeConfig{GuildID: "missing", Name: "prize", Chance: 100, Count: 1}); err != nil {
		t.Fatalf("create missing-balance prize: %v", err)
	}
	_, err := repository.DrawGacha(ctx, domain.GachaDrawRequest{
		GuildID: "missing", UserID: "user-1", PaidDraws: 1, ActualDraws: 1, RandomValues: []float64{0},
	})
	if !errors.Is(err, ports.ErrCoinBalanceNotFound) {
		t.Fatalf("missing balance error = %v", err)
	}

	seedGachaEconomy(t, database, "low", "user-1", 5, 100)
	if err := repository.CreateGachaPrize(ctx, domain.GachaPrizeConfig{GuildID: "low", Name: "prize", Chance: 100, Count: 1}); err != nil {
		t.Fatalf("create low-balance prize: %v", err)
	}
	result, err := repository.DrawGacha(ctx, domain.GachaDrawRequest{
		GuildID: "low", UserID: "user-1", PaidDraws: 1, ActualDraws: 1, RandomValues: []float64{0},
	})
	if !errors.Is(err, ports.ErrGachaInsufficientCoins) || result.BalanceBefore != 5 || result.Cost != 100 {
		t.Fatalf("insufficient draw result = %#v err=%v", result, err)
	}
	assertGachaBalance(t, database, "low", "user-1", 5)
}

func TestGachaMongoIntegrationConcurrentDepletionChargesOnce(t *testing.T) {
	database, repository := gachaIntegrationRepository(t)
	ctx := context.Background()
	seedGachaEconomy(t, database, "race", "user-1", 1000, 100)
	if err := repository.CreateGachaPrize(ctx, domain.GachaPrizeConfig{
		GuildID: "race", Name: "only", Chance: 100, AutoDelete: true, Count: 1,
	}); err != nil {
		t.Fatalf("create race prize: %v", err)
	}

	request := domain.GachaDrawRequest{GuildID: "race", UserID: "user-1", PaidDraws: 1, ActualDraws: 1, RandomValues: []float64{0}}
	errorsByDraw := make([]error, 2)
	var wait sync.WaitGroup
	wait.Add(2)
	for index := range errorsByDraw {
		go func() {
			defer wait.Done()
			_, errorsByDraw[index] = repository.DrawGacha(ctx, request)
		}()
	}
	wait.Wait()

	succeeded := 0
	exhausted := 0
	for _, err := range errorsByDraw {
		switch {
		case err == nil:
			succeeded++
		case errors.Is(err, ports.ErrGachaPrizePoolEmpty):
			exhausted++
		default:
			t.Fatalf("concurrent draw error = %v", err)
		}
	}
	if succeeded != 1 || exhausted != 1 {
		t.Fatalf("concurrent draws succeeded=%d exhausted=%d errors=%v", succeeded, exhausted, errorsByDraw)
	}
	assertGachaBalance(t, database, "race", "user-1", 900)
}

func gachaIntegrationRepository(t *testing.T) (*drivermongo.Database, *GachaRepository) {
	t.Helper()
	if os.Getenv("MHCAT_RUN_MONGO_TRANSACTION_INTEGRATION_TESTS") != "true" {
		t.Skip("set MHCAT_RUN_MONGO_TRANSACTION_INTEGRATION_TESTS=true to run")
	}
	uri := os.Getenv("MHCAT_MONGODB_URI")
	if uri == "" {
		t.Fatal("MHCAT_MONGODB_URI is required")
	}
	databaseName := fmt.Sprintf("mhcat_gacha_test_%d", time.Now().UnixNano())
	client, err := mhcatmongo.NewClient(mhcatmongo.Options{
		URI: uri, Database: databaseName, ConnectTimeout: 10 * time.Second, PingTimeout: 5 * time.Second,
	})
	if err != nil {
		t.Fatalf("new Mongo client: %v", err)
	}
	if err := client.Connect(context.Background()); err != nil {
		t.Fatalf("connect Mongo: %v", err)
	}
	database, err := client.Database()
	if err != nil {
		t.Fatalf("get database: %v", err)
	}
	transactions, err := mhcatmongo.NewTransactionRunner(client)
	if err != nil {
		t.Fatalf("new transaction runner: %v", err)
	}
	repository, err := NewGachaRepositoryFromDatabase(database)
	if err != nil {
		t.Fatalf("new gacha repository: %v", err)
	}
	if err := repository.SetGachaDrawTransactionRunner(transactions); err != nil {
		t.Fatalf("set gacha transaction runner: %v", err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := database.Drop(ctx); err != nil {
			t.Errorf("drop database: %v", err)
		}
		if err := client.Disconnect(ctx); err != nil {
			t.Errorf("disconnect Mongo: %v", err)
		}
	})
	return database, repository
}

func seedGachaEconomy(t *testing.T, database *drivermongo.Database, guildID string, userID string, coins int64, cost int64) {
	t.Helper()
	ctx := context.Background()
	if _, err := database.Collection(CoinCollectionName).InsertOne(ctx, bson.D{
		{Key: "guild", Value: guildID}, {Key: "member", Value: userID}, {Key: "coin", Value: coins}, {Key: "today", Value: int64(0)},
	}); err != nil {
		t.Fatalf("seed gacha balance: %v", err)
	}
	if _, err := database.Collection(GiftChangeCollectionName).InsertOne(ctx, bson.D{
		{Key: "guild", Value: guildID}, {Key: "coin_number", Value: cost},
	}); err != nil {
		t.Fatalf("seed gacha config: %v", err)
	}
}

func assertGachaBalance(t *testing.T, database *drivermongo.Database, guildID string, userID string, want int64) {
	t.Helper()
	var stored struct {
		Coins int64 `bson:"coin"`
	}
	err := database.Collection(CoinCollectionName).FindOne(context.Background(), bson.D{
		{Key: "guild", Value: guildID}, {Key: "member", Value: userID},
	}).Decode(&stored)
	if err != nil {
		t.Fatalf("read gacha balance: %v", err)
	}
	if stored.Coins != want {
		t.Fatalf("gacha balance = %d, want %d", stored.Coins, want)
	}
}
