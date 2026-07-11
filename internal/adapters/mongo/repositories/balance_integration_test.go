package repositories

import (
	"context"
	"errors"
	"fmt"
	"math"
	"os"
	"testing"
	"time"

	mhcatmongo "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/adapters/mongo"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	"go.mongodb.org/mongo-driver/v2/bson"
	drivermongo "go.mongodb.org/mongo-driver/v2/mongo"
)

func TestBalanceMongoIntegrationPreservesMongooseValuesAndFirstMatch(t *testing.T) {
	database := balanceIntegrationDatabase(t)
	repository, err := NewBalanceRepositoryFromDatabase(database)
	if err != nil {
		t.Fatalf("new balance repository: %v", err)
	}
	collection := database.Collection(BalanceCollectionName)
	rows := []any{
		bson.D{{Key: "guild", Value: "null"}, {Key: "price", Value: nil}},
		bson.D{{Key: "guild", Value: "empty"}, {Key: "price", Value: ""}},
		bson.D{{Key: "guild", Value: "hex"}, {Key: "price", Value: "0x10"}},
		bson.D{{Key: "guild", Value: "rounded"}, {Key: "price", Value: int64(9_007_199_254_740_993)}},
		bson.D{{Key: "guild", Value: "nan"}, {Key: "price", Value: math.NaN()}},
		bson.D{{Key: "guild", Value: "binary"}, {Key: "price", Value: bson.Binary{Subtype: 0, Data: []byte("5")}}},
		bson.D{{Key: "guild", Value: "other-guild"}, {Key: "price", Value: 77}},
	}
	if _, err := collection.InsertMany(context.Background(), rows); err != nil {
		t.Fatalf("seed mixed balances: %v", err)
	}
	for _, test := range []struct {
		guild string
		want  string
	}{
		{guild: "null", want: "null"},
		{guild: "empty", want: "null"},
		{guild: "hex", want: "16"},
		{guild: "rounded", want: "9007199254740992"},
		{guild: "nan", want: "undefined"},
		{guild: "binary", want: "5"},
	} {
		balance, err := repository.GetBalance(context.Background(), test.guild)
		if err != nil {
			t.Fatalf("get %s: %v", test.guild, err)
		}
		if balance.GuildID != test.guild || balance.Amount != test.want {
			t.Fatalf("balance %s = %#v, want %q", test.guild, balance, test.want)
		}
	}

	duplicates, err := collection.InsertMany(context.Background(), []any{
		bson.D{{Key: "guild", Value: "duplicate"}, {Key: "price", Value: 1}},
		bson.D{{Key: "guild", Value: "duplicate"}, {Key: "price", Value: 2}},
	})
	if err != nil {
		t.Fatalf("seed duplicates: %v", err)
	}
	if balance, err := repository.GetBalance(context.Background(), "duplicate"); err != nil || balance.Amount != "1" {
		t.Fatalf("first duplicate balance = %#v, err=%v", balance, err)
	}
	if _, err := collection.DeleteOne(context.Background(), bson.D{{Key: "_id", Value: duplicates.InsertedIDs[0]}}); err != nil {
		t.Fatalf("delete first duplicate: %v", err)
	}
	if balance, err := repository.GetBalance(context.Background(), "duplicate"); err != nil || balance.Amount != "2" {
		t.Fatalf("second duplicate balance = %#v, err=%v", balance, err)
	}
	if _, err := repository.GetBalance(context.Background(), "missing"); !errors.Is(err, ports.ErrBalanceMissing) {
		t.Fatalf("missing balance error = %v", err)
	}
	count, err := collection.CountDocuments(context.Background(), bson.D{})
	if err != nil || count != int64(len(rows)+1) {
		t.Fatalf("read-only row count = %d, err=%v", count, err)
	}
	other, err := repository.GetBalance(context.Background(), "other-guild")
	if err != nil || other.Amount != "77" {
		t.Fatalf("other guild balance = %#v, err=%v", other, err)
	}
}

func balanceIntegrationDatabase(t *testing.T) *drivermongo.Database {
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
		Database:       fmt.Sprintf("mhcat_balance_test_%d", time.Now().UnixNano()),
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
