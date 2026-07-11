package repositories

import (
	"context"
	"errors"
	"fmt"
	"math"
	"os"
	"sort"
	"sync"
	"testing"
	"time"

	mhcatmongo "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/adapters/mongo"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	"go.mongodb.org/mongo-driver/v2/bson"
	drivermongo "go.mongodb.org/mongo-driver/v2/mongo"
)

func TestAutoChatPaidMongoTransactionIntegrationStartupDoesNotMutateDatabase(t *testing.T) {
	_, database := autoChatPaidIntegrationRepository(t)
	ctx := context.Background()
	names, err := database.ListCollectionNames(ctx, bson.D{})
	if err != nil {
		t.Fatalf("list collections: %v", err)
	}
	if len(names) != 0 {
		t.Fatalf("startup collections = %#v", names)
	}
}

func TestAutoChatPaidMongoTransactionIntegrationLifecycle(t *testing.T) {
	repository, database := autoChatPaidIntegrationRepository(t)
	ctx := context.Background()
	requestTime := int64(1_700_000_000_000)
	if _, err := database.Collection(BalanceCollectionName).InsertOne(ctx, bson.D{
		{Key: "_id", Value: bson.NewObjectID()},
		{Key: "guild", Value: "guild-1"},
		{Key: "price", Value: 10.0},
	}); err != nil {
		t.Fatalf("insert balance: %v", err)
	}

	first, err := repository.QueuePaidAutoChat(ctx, domain.AutoChatPaidRequest{
		GuildID:          "guild-1",
		Content:          "first prompt",
		Cost:             0.001,
		RequestedAtMilli: requestTime,
	})
	if err != nil {
		t.Fatalf("queue first request: %v", err)
	}
	if first.RequestTimeMilli != requestTime || first.ConversationReset {
		t.Fatalf("first dispatch = %#v", first)
	}
	assertAutoChatPaidBalance(t, database, "guild-1", 9.999)
	assertAutoChatPaidHandoff(t, database, "guild-1", "first prompt", requestTime, nil, nil, false)

	if _, err := database.Collection(AutoChatPaidCollectionName).UpdateOne(ctx,
		bson.D{{Key: "guild", Value: "guild-1"}},
		bson.D{{Key: "$set", Value: bson.D{
			{Key: "resid_c", Value: "conversation-1"},
			{Key: "resid_p", Value: "parent-1"},
			{Key: "reply", Value: true},
			{Key: "message", Value: "first answer"},
		}}},
	); err != nil {
		t.Fatalf("simulate worker response: %v", err)
	}
	response, err := repository.GetPaidAutoChatResponse(ctx, "guild-1", requestTime)
	if err != nil || response.Content != "first answer" || !response.Reply {
		t.Fatalf("response=%#v err=%v", response, err)
	}

	_, err = repository.QueuePaidAutoChat(ctx, domain.AutoChatPaidRequest{
		GuildID:          "guild-1",
		Content:          "too soon",
		Cost:             1,
		RequestedAtMilli: requestTime + 9_999,
	})
	if !errors.Is(err, ports.ErrAutoChatPaidBusy) {
		t.Fatalf("busy error = %v", err)
	}
	assertAutoChatPaidBalance(t, database, "guild-1", 9.999)

	secondTime := requestTime + 10_000
	second, err := repository.QueuePaidAutoChat(ctx, domain.AutoChatPaidRequest{
		GuildID:          "guild-1",
		Content:          "second prompt",
		Cost:             0.002,
		RequestedAtMilli: secondTime,
	})
	if err != nil {
		t.Fatalf("queue second request: %v", err)
	}
	if second.ConversationReset {
		t.Fatalf("second dispatch = %#v", second)
	}
	assertAutoChatPaidBalance(t, database, "guild-1", 9.997)
	assertAutoChatPaidHandoff(t, database, "guild-1", "second prompt", secondTime, "conversation-1", "parent-1", false)
	if _, err := repository.GetPaidAutoChatResponse(ctx, "guild-1", requestTime); !errors.Is(err, ports.ErrAutoChatPaidResponseMissing) {
		t.Fatalf("stale response error = %v", err)
	}

	thirdTime := secondTime + 40_001
	third, err := repository.QueuePaidAutoChat(ctx, domain.AutoChatPaidRequest{
		GuildID:          "guild-1",
		Content:          "new conversation",
		Cost:             0.003,
		RequestedAtMilli: thirdTime,
	})
	if err != nil {
		t.Fatalf("queue reset request: %v", err)
	}
	if !third.ConversationReset {
		t.Fatalf("third dispatch = %#v", third)
	}
	assertAutoChatPaidBalance(t, database, "guild-1", 9.994)
	assertAutoChatPaidHandoff(t, database, "guild-1", "new conversation", thirdTime, nil, nil, false)
}

func TestAutoChatPaidMongoTransactionIntegrationRollsBackDebit(t *testing.T) {
	repository, database := autoChatPaidIntegrationRepository(t)
	ctx := context.Background()
	if err := database.RunCommand(ctx, bson.D{
		{Key: "create", Value: AutoChatPaidCollectionName},
		{Key: "validator", Value: bson.D{{Key: "message", Value: bson.D{{Key: "$type", Value: "int"}}}}},
	}).Err(); err != nil {
		t.Fatalf("create validated handoff collection: %v", err)
	}
	if _, err := database.Collection(BalanceCollectionName).InsertOne(ctx, bson.D{
		{Key: "guild", Value: "guild-rollback"},
		{Key: "price", Value: 5.0},
	}); err != nil {
		t.Fatalf("insert balance: %v", err)
	}

	_, err := repository.QueuePaidAutoChat(ctx, domain.AutoChatPaidRequest{
		GuildID:          "guild-rollback",
		Content:          "validator rejects this string",
		Cost:             1,
		RequestedAtMilli: 1_700_000_000_000,
	})
	if err == nil {
		t.Fatal("expected handoff validation error")
	}
	assertAutoChatPaidBalance(t, database, "guild-rollback", 5)
	count, countErr := database.Collection(AutoChatPaidCollectionName).CountDocuments(ctx, bson.D{})
	if countErr != nil || count != 0 {
		t.Fatalf("handoff count=%d err=%v", count, countErr)
	}
}

func TestAutoChatPaidMongoTransactionIntegrationAllowsLegacyOverdraw(t *testing.T) {
	repository, database := autoChatPaidIntegrationRepository(t)
	ctx := context.Background()
	if _, err := database.Collection(BalanceCollectionName).InsertOne(ctx, bson.D{
		{Key: "guild", Value: "guild-overdraw"},
		{Key: "price", Value: 0.001},
	}); err != nil {
		t.Fatalf("insert balance: %v", err)
	}

	dispatch, err := repository.QueuePaidAutoChat(ctx, domain.AutoChatPaidRequest{
		GuildID:          "guild-overdraw",
		Content:          "expensive prompt",
		Cost:             1,
		RequestedAtMilli: 1_700_000_000_000,
	})
	if err != nil {
		t.Fatalf("queue overdraw request: %v", err)
	}
	if dispatch.ConversationReset {
		t.Fatalf("dispatch = %#v", dispatch)
	}
	assertAutoChatPaidBalance(t, database, "guild-overdraw", -0.999)
	assertAutoChatPaidHandoff(t, database, "guild-overdraw", "expensive prompt", 1_700_000_000_000, nil, nil, false)
}

func TestAutoChatPaidMongoTransactionIntegrationRejectsDuplicatesWithoutWrites(t *testing.T) {
	t.Run("balance rows", func(t *testing.T) {
		repository, database := autoChatPaidIntegrationRepository(t)
		ctx := context.Background()
		if _, err := database.Collection(BalanceCollectionName).InsertMany(ctx, []any{
			bson.D{{Key: "guild", Value: "guild-duplicate-balance"}, {Key: "price", Value: 5.0}},
			bson.D{{Key: "guild", Value: "guild-duplicate-balance"}, {Key: "price", Value: 7.0}},
		}); err != nil {
			t.Fatalf("insert duplicate balances: %v", err)
		}

		_, err := repository.QueuePaidAutoChat(ctx, domain.AutoChatPaidRequest{
			GuildID:          "guild-duplicate-balance",
			Content:          "prompt",
			Cost:             1,
			RequestedAtMilli: 1_700_000_000_000,
		})
		if !errors.Is(err, ports.ErrAutoChatPaidStateConflict) {
			t.Fatalf("queue error = %v", err)
		}
		assertAutoChatPaidPrices(t, database, "guild-duplicate-balance", []float64{5, 7})
		if count, countErr := database.Collection(AutoChatPaidCollectionName).CountDocuments(ctx, bson.D{}); countErr != nil || count != 0 {
			t.Fatalf("handoff count=%d err=%v", count, countErr)
		}
	})

	t.Run("handoff rows", func(t *testing.T) {
		repository, database := autoChatPaidIntegrationRepository(t)
		ctx := context.Background()
		if _, err := database.Collection(BalanceCollectionName).InsertOne(ctx, bson.D{
			{Key: "guild", Value: "guild-duplicate-handoff"},
			{Key: "price", Value: 5.0},
		}); err != nil {
			t.Fatalf("insert balance: %v", err)
		}
		if _, err := database.Collection(AutoChatPaidCollectionName).InsertMany(ctx, []any{
			bson.D{{Key: "guild", Value: "guild-duplicate-handoff"}, {Key: "message", Value: "first"}, {Key: "time", Value: int64(1)}},
			bson.D{{Key: "guild", Value: "guild-duplicate-handoff"}, {Key: "message", Value: "second"}, {Key: "time", Value: int64(2)}},
		}); err != nil {
			t.Fatalf("insert duplicate handoffs: %v", err)
		}

		_, err := repository.QueuePaidAutoChat(ctx, domain.AutoChatPaidRequest{
			GuildID:          "guild-duplicate-handoff",
			Content:          "prompt",
			Cost:             1,
			RequestedAtMilli: 1_700_000_000_000,
		})
		if !errors.Is(err, ports.ErrAutoChatPaidStateConflict) {
			t.Fatalf("queue error = %v", err)
		}
		assertAutoChatPaidBalance(t, database, "guild-duplicate-handoff", 5)
		if count, countErr := database.Collection(AutoChatPaidCollectionName).CountDocuments(ctx, bson.D{{Key: "guild", Value: "guild-duplicate-handoff"}}); countErr != nil || count != 2 {
			t.Fatalf("handoff count=%d err=%v", count, countErr)
		}
	})
}

func TestAutoChatPaidMongoTransactionIntegrationConcurrentQueue(t *testing.T) {
	repository, database := autoChatPaidIntegrationRepository(t)
	ctx := context.Background()
	if _, err := database.Collection(BalanceCollectionName).InsertOne(ctx, bson.D{
		{Key: "guild", Value: "guild-concurrent"},
		{Key: "price", Value: 5.0},
	}); err != nil {
		t.Fatalf("insert balance: %v", err)
	}

	var wg sync.WaitGroup
	errs := make([]error, 2)
	for i := range errs {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			_, errs[index] = repository.QueuePaidAutoChat(ctx, domain.AutoChatPaidRequest{
				GuildID:          "guild-concurrent",
				Content:          fmt.Sprintf("prompt-%d", index),
				Cost:             1,
				RequestedAtMilli: 1_700_000_000_000 + int64(index),
			})
		}(i)
	}
	wg.Wait()
	succeeded := 0
	busy := 0
	for _, err := range errs {
		switch {
		case err == nil:
			succeeded++
		case errors.Is(err, ports.ErrAutoChatPaidBusy):
			busy++
		default:
			t.Fatalf("queue error = %v", err)
		}
	}
	if succeeded != 1 || busy != 1 {
		t.Fatalf("succeeded=%d busy=%d errors=%v", succeeded, busy, errs)
	}
	assertAutoChatPaidBalance(t, database, "guild-concurrent", 4)
	count, err := database.Collection(AutoChatPaidCollectionName).CountDocuments(ctx, bson.D{{Key: "guild", Value: "guild-concurrent"}})
	if err != nil || count != 1 {
		t.Fatalf("handoff count=%d err=%v", count, err)
	}
}

func autoChatPaidIntegrationRepository(t *testing.T) (*AutoChatPaidRepository, *drivermongo.Database) {
	t.Helper()
	if os.Getenv("MHCAT_RUN_MONGO_TRANSACTION_INTEGRATION_TESTS") != "true" {
		t.Skip("set MHCAT_RUN_MONGO_TRANSACTION_INTEGRATION_TESTS=true to run")
	}
	uri := os.Getenv("MHCAT_MONGODB_URI")
	if uri == "" {
		t.Fatal("MHCAT_MONGODB_URI is required")
	}
	databaseName := fmt.Sprintf("mhcat_autochat_paid_test_%d", time.Now().UnixNano())
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
	repository, err := NewAutoChatPaidRepositoryFromDatabase(database, transactions)
	if err != nil {
		t.Fatalf("new paid autochat repository: %v", err)
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

func assertAutoChatPaidBalance(t *testing.T, database *drivermongo.Database, guildID string, want float64) {
	t.Helper()
	var row struct {
		Price float64 `bson:"price"`
	}
	if err := database.Collection(BalanceCollectionName).FindOne(context.Background(), bson.D{{Key: "guild", Value: guildID}}).Decode(&row); err != nil {
		t.Fatalf("decode balance: %v", err)
	}
	if math.Abs(row.Price-want) > 1e-9 {
		t.Fatalf("balance = %.12f, want %.12f", row.Price, want)
	}
}

func assertAutoChatPaidPrices(t *testing.T, database *drivermongo.Database, guildID string, want []float64) {
	t.Helper()
	cursor, err := database.Collection(BalanceCollectionName).Find(context.Background(), bson.D{{Key: "guild", Value: guildID}})
	if err != nil {
		t.Fatalf("find balances: %v", err)
	}
	defer cursor.Close(context.Background())
	var rows []struct {
		Price float64 `bson:"price"`
	}
	if err := cursor.All(context.Background(), &rows); err != nil {
		t.Fatalf("decode balances: %v", err)
	}
	if len(rows) != len(want) {
		t.Fatalf("balance rows = %#v, want %#v", rows, want)
	}
	prices := make([]float64, len(rows))
	for index := range rows {
		prices[index] = rows[index].Price
	}
	sort.Float64s(prices)
	for index := range want {
		if math.Abs(prices[index]-want[index]) > 1e-9 {
			t.Fatalf("balance rows = %#v, want %#v", rows, want)
		}
	}
}

func assertAutoChatPaidHandoff(t *testing.T, database *drivermongo.Database, guildID string, message string, requestTime int64, residC any, residP any, reply bool) {
	t.Helper()
	var row bson.M
	if err := database.Collection(AutoChatPaidCollectionName).FindOne(context.Background(), bson.D{{Key: "guild", Value: guildID}}).Decode(&row); err != nil {
		t.Fatalf("decode handoff: %v", err)
	}
	if row["message"] != message || row["time"] != requestTime || row["reply"] != reply || row["resid_c"] != residC || row["resid_p"] != residP {
		t.Fatalf("handoff = %#v", row)
	}
}
