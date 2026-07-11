package repositories

import (
	"context"
	"fmt"
	"math"
	"os"
	"testing"
	"time"

	mhcatmongo "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/adapters/mongo"
	"go.mongodb.org/mongo-driver/v2/bson"
	drivermongo "go.mongodb.org/mongo-driver/v2/mongo"
)

func TestEconomyQueryMongoIntegrationStartupDoesNotMutateDatabase(t *testing.T) {
	database := economyQueryIntegrationDatabase(t)
	if _, err := NewEconomyRepositoryFromDatabase(database); err != nil {
		t.Fatalf("new repository: %v", err)
	}
	names, err := database.ListCollectionNames(context.Background(), bson.D{})
	if err != nil {
		t.Fatalf("list collections: %v", err)
	}
	if len(names) != 0 {
		t.Fatalf("startup collections = %#v", names)
	}
}

func TestEconomyQueryMongoIntegrationPreservesLegacyScalars(t *testing.T) {
	database := economyQueryIntegrationDatabase(t)
	repository, err := NewEconomyRepositoryFromDatabase(database)
	if err != nil {
		t.Fatalf("new repository: %v", err)
	}

	tests := []struct {
		name  string
		value any
		want  string
	}{
		{name: "missing", want: "undefined"},
		{name: "null", value: nil, want: "null"},
		{name: "malformed", value: bson.D{{Key: "invalid", Value: true}}, want: "undefined"},
		{name: "decimal", value: 125.5, want: "125.5"},
		{name: "positive-infinity", value: math.Inf(1), want: "Infinity"},
		{name: "negative-infinity", value: math.Inf(-1), want: "-Infinity"},
	}
	for _, test := range tests {
		coin := bson.D{{Key: "guild", Value: "guild-1"}, {Key: "member", Value: test.name}}
		config := bson.D{{Key: "guild", Value: test.name}}
		if test.name != "missing" {
			coin = append(coin, bson.E{Key: "coin", Value: test.value})
			config = append(config, bson.E{Key: "coin_number", Value: test.value})
		}
		if _, err := database.Collection(CoinCollectionName).InsertOne(context.Background(), coin); err != nil {
			t.Fatalf("insert %s coin: %v", test.name, err)
		}
		if _, err := database.Collection(GiftChangeCollectionName).InsertOne(context.Background(), config); err != nil {
			t.Fatalf("insert %s config: %v", test.name, err)
		}

		balance, err := repository.GetCoinBalance(context.Background(), "guild-1", test.name)
		if err != nil || balance.CoinsText != test.want {
			t.Fatalf("%s balance = %#v, err=%v", test.name, balance, err)
		}
		gotConfig, err := repository.GetEconomyConfig(context.Background(), test.name)
		if err != nil || gotConfig.GachaCostText != test.want {
			t.Fatalf("%s config = %#v, err=%v", test.name, gotConfig, err)
		}
	}
}

func TestEconomyQueryMongoIntegrationAllowsLegacyDuplicates(t *testing.T) {
	database := economyQueryIntegrationDatabase(t)
	repository, err := NewEconomyRepositoryFromDatabase(database)
	if err != nil {
		t.Fatalf("new repository: %v", err)
	}
	if _, err := database.Collection(CoinCollectionName).InsertMany(context.Background(), []any{
		bson.D{{Key: "guild", Value: "guild-1"}, {Key: "member", Value: "user-1"}, {Key: "coin", Value: 10}},
		bson.D{{Key: "guild", Value: "guild-1"}, {Key: "member", Value: "user-1"}, {Key: "coin", Value: 20}},
	}); err != nil {
		t.Fatalf("insert duplicate coins: %v", err)
	}
	if _, err := database.Collection(GiftChangeCollectionName).InsertMany(context.Background(), []any{
		bson.D{{Key: "guild", Value: "guild-1"}, {Key: "coin_number", Value: 500}},
		bson.D{{Key: "guild", Value: "guild-1"}, {Key: "coin_number", Value: 700}},
	}); err != nil {
		t.Fatalf("insert duplicate configs: %v", err)
	}

	balance, err := repository.GetCoinBalance(context.Background(), "guild-1", "user-1")
	if err != nil || (balance.CoinsText != "10" && balance.CoinsText != "20") {
		t.Fatalf("duplicate balance = %#v, err=%v", balance, err)
	}
	config, err := repository.GetEconomyConfig(context.Background(), "guild-1")
	if err != nil || (config.GachaCostText != "500" && config.GachaCostText != "700") {
		t.Fatalf("duplicate config = %#v, err=%v", config, err)
	}
}

func economyQueryIntegrationDatabase(t *testing.T) *drivermongo.Database {
	t.Helper()
	if os.Getenv("MHCAT_RUN_MONGO_INTEGRATION_TESTS") != "true" {
		t.Skip("set MHCAT_RUN_MONGO_INTEGRATION_TESTS=true to run")
	}
	uri := os.Getenv("MHCAT_MONGODB_URI")
	if uri == "" {
		t.Fatal("MHCAT_MONGODB_URI is required")
	}
	client, err := mhcatmongo.NewClient(mhcatmongo.Options{
		URI:            uri,
		Database:       fmt.Sprintf("mhcat_economy_query_test_%d", time.Now().UnixNano()),
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
	return database
}
