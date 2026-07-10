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

func TestEconomyCoinGameMongoTransactionIntegrationLifecycle(t *testing.T) {
	repository, database := economyGameIntegrationRepository(t)
	insertEconomyGameBalances(t, database, 50, 50)

	reserved, err := repository.ReserveCoinGameWager(context.Background(), economyGameIntegrationCommand(10))
	if err != nil {
		t.Fatalf("reserve wager: %v", err)
	}
	if reserved.Challenger.Coins != 40 || reserved.Opponent.Coins != 40 {
		t.Fatalf("reserved = %#v", reserved)
	}

	settled, err := repository.SettleCoinGameWager(context.Background(), domain.CoinGameSettlementCommand{
		GuildID:          "guild-1",
		ChallengerID:     "user-1",
		OpponentID:       "user-2",
		ChallengerReturn: 20,
		OpponentReturn:   0,
	})
	if err != nil {
		t.Fatalf("settle wager: %v", err)
	}
	if settled.Challenger.Coins != 60 || settled.Opponent.Coins != 40 {
		t.Fatalf("settled = %#v", settled)
	}
	assertEconomyGameBalances(t, repository, 60, 40)
}

func TestEconomyCoinGameMongoTransactionIntegrationRollsBackReserve(t *testing.T) {
	repository, database := economyGameIntegrationRepository(t)
	insertEconomyGameBalances(t, database, 50, 50)
	rejectEconomyGameOpponentBalanceChanges(t, database, 50)

	if _, err := repository.ReserveCoinGameWager(context.Background(), economyGameIntegrationCommand(10)); err == nil {
		t.Fatal("expected reserve failure")
	}
	assertEconomyGameBalances(t, repository, 50, 50)
}

func TestEconomyCoinGameMongoTransactionIntegrationRollsBackSettlement(t *testing.T) {
	repository, database := economyGameIntegrationRepository(t)
	insertEconomyGameBalances(t, database, 40, 40)
	rejectEconomyGameOpponentBalanceChanges(t, database, 40)

	_, err := repository.SettleCoinGameWager(context.Background(), domain.CoinGameSettlementCommand{
		GuildID:          "guild-1",
		ChallengerID:     "user-1",
		OpponentID:       "user-2",
		ChallengerReturn: 10,
		OpponentReturn:   10,
	})
	if err == nil {
		t.Fatal("expected settlement failure")
	}
	assertEconomyGameBalances(t, repository, 40, 40)
}

func TestEconomyCoinGameMongoTransactionIntegrationConcurrentReserve(t *testing.T) {
	repository, database := economyGameIntegrationRepository(t)
	insertEconomyGameBalances(t, database, 100, 100)

	start := make(chan struct{})
	errs := make([]error, 2)
	var wg sync.WaitGroup
	for index := range errs {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			<-start
			_, errs[index] = repository.ReserveCoinGameWager(context.Background(), economyGameIntegrationCommand(80))
		}(index)
	}
	close(start)
	wg.Wait()

	successes := 0
	insufficient := 0
	for _, err := range errs {
		switch {
		case err == nil:
			successes++
		case errors.Is(err, ports.ErrCoinGameOpponent), errors.Is(err, ports.ErrCoinGameChallenger):
			insufficient++
		default:
			t.Fatalf("unexpected reserve error: %v", err)
		}
	}
	if successes != 1 || insufficient != 1 {
		t.Fatalf("successes=%d insufficient=%d errors=%v", successes, insufficient, errs)
	}
	assertEconomyGameBalances(t, repository, 20, 20)
}

func economyGameIntegrationRepository(t *testing.T) (*EconomyRepository, *drivermongo.Database) {
	t.Helper()
	if os.Getenv("MHCAT_RUN_MONGO_TRANSACTION_INTEGRATION_TESTS") != "true" {
		t.Skip("set MHCAT_RUN_MONGO_TRANSACTION_INTEGRATION_TESTS=true to run")
	}
	uri := os.Getenv("MHCAT_MONGODB_URI")
	if uri == "" {
		t.Fatal("MHCAT_MONGODB_URI is required")
	}
	databaseName := fmt.Sprintf("mhcat_economy_game_test_%d", time.Now().UnixNano())
	client, err := mhcatmongo.NewClient(mhcatmongo.Options{
		URI:            uri,
		Database:       databaseName,
		ConnectTimeout: 10 * time.Second,
		PingTimeout:    5 * time.Second,
	})
	if err != nil {
		t.Fatalf("new Mongo client: %v", err)
	}
	if err := client.Connect(context.Background()); err != nil {
		t.Fatalf("connect Mongo: %v", err)
	}
	database, err := client.Database()
	if err != nil {
		t.Fatalf("get Mongo database: %v", err)
	}
	transactions, err := mhcatmongo.NewTransactionRunner(client)
	if err != nil {
		t.Fatalf("new transaction runner: %v", err)
	}
	repository, err := NewEconomyRepositoryFromDatabase(database)
	if err != nil {
		t.Fatalf("new economy repository: %v", err)
	}
	if err := repository.SetCoinGameTransactionRunner(transactions); err != nil {
		t.Fatalf("set coin game transaction runner: %v", err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := database.Drop(ctx); err != nil {
			t.Errorf("drop integration database: %v", err)
		}
		if err := client.Disconnect(ctx); err != nil {
			t.Errorf("disconnect integration Mongo: %v", err)
		}
	})
	return repository, database
}

func economyGameIntegrationCommand(wager int64) domain.CoinGameCommand {
	return domain.CoinGameCommand{
		GuildID:      "guild-1",
		ChallengerID: "user-1",
		OpponentID:   "user-2",
		Wager:        wager,
		Kind:         domain.CoinGameKindKnowledge,
	}
}

func insertEconomyGameBalances(t *testing.T, database *drivermongo.Database, challenger int64, opponent int64) {
	t.Helper()
	_, err := database.Collection(CoinCollectionName).InsertMany(context.Background(), []any{
		bson.D{{Key: "guild", Value: "guild-1"}, {Key: "member", Value: "user-1"}, {Key: "coin", Value: challenger}},
		bson.D{{Key: "guild", Value: "guild-1"}, {Key: "member", Value: "user-2"}, {Key: "coin", Value: opponent}},
	})
	if err != nil {
		t.Fatalf("insert coin game balances: %v", err)
	}
}

func rejectEconomyGameOpponentBalanceChanges(t *testing.T, database *drivermongo.Database, allowed int64) {
	t.Helper()
	validator := bson.D{{Key: "$expr", Value: bson.D{{Key: "$or", Value: bson.A{
		bson.D{{Key: "$ne", Value: bson.A{"$member", "user-2"}}},
		bson.D{{Key: "$eq", Value: bson.A{"$coin", allowed}}},
	}}}}}
	if err := database.RunCommand(context.Background(), bson.D{
		{Key: "collMod", Value: CoinCollectionName},
		{Key: "validator", Value: validator},
		{Key: "validationLevel", Value: "strict"},
		{Key: "validationAction", Value: "error"},
	}).Err(); err != nil {
		t.Fatalf("configure coin validator: %v", err)
	}
}

func assertEconomyGameBalances(t *testing.T, repository *EconomyRepository, challenger int64, opponent int64) {
	t.Helper()
	challengerBalance, err := repository.GetCoinBalance(context.Background(), "guild-1", "user-1")
	if err != nil {
		t.Fatalf("get challenger balance: %v", err)
	}
	opponentBalance, err := repository.GetCoinBalance(context.Background(), "guild-1", "user-2")
	if err != nil {
		t.Fatalf("get opponent balance: %v", err)
	}
	if challengerBalance.Coins != challenger || opponentBalance.Coins != opponent {
		t.Fatalf("balances challenger=%d opponent=%d, want %d/%d", challengerBalance.Coins, opponentBalance.Coins, challenger, opponent)
	}
}
