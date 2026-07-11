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

func TestRedeemMongoIntegrationPreservesFetchedCodeAndOneBalanceReplacement(t *testing.T) {
	database := redeemIntegrationDatabase(t)
	repository, err := NewRedeemRepositoryFromDatabase(database)
	if err != nil {
		t.Fatalf("new redeem repository: %v", err)
	}
	codes := database.Collection(RedeemCodeCollectionName)
	balances := database.Collection(BalanceCollectionName)
	firstCodeID := bson.NewObjectID()
	secondCodeID := bson.NewObjectID()
	firstBalanceID := bson.NewObjectID()
	secondBalanceID := bson.NewObjectID()
	if _, err := codes.InsertMany(context.Background(), []any{
		bson.D{{Key: "_id", Value: firstCodeID}, {Key: "code", Value: " raw "}, {Key: "price", Value: -2}, {Key: "time", Value: 1_700_000_000_000}},
		bson.D{{Key: "_id", Value: secondCodeID}, {Key: "code", Value: " raw "}, {Key: "price", Value: 99}, {Key: "time", Value: 1_700_000_000_000}},
	}); err != nil {
		t.Fatalf("seed codes: %v", err)
	}
	if _, err := balances.InsertMany(context.Background(), []any{
		bson.D{{Key: "_id", Value: firstBalanceID}, {Key: "guild", Value: "guild-1"}, {Key: "price", Value: 10}, {Key: "marker", Value: "first"}},
		bson.D{{Key: "_id", Value: secondBalanceID}, {Key: "guild", Value: "guild-1"}, {Key: "price", Value: 20}, {Key: "marker", Value: "second"}},
		bson.D{{Key: "guild", Value: "guild-2"}, {Key: "price", Value: 30}},
	}); err != nil {
		t.Fatalf("seed balances: %v", err)
	}

	fetched, err := repository.GetRedeemCode(context.Background(), " raw ")
	if err != nil {
		t.Fatalf("get code: %v", err)
	}
	if fetched.Identity != firstCodeID || fetched.Code != " raw " || fetched.Price != -2 {
		t.Fatalf("fetched code = %#v", fetched)
	}
	command := domain.RedeemCommand{GuildID: "guild-1", Code: " raw ", NowMS: 1_700_000_000_000}
	if err := repository.ConsumeRedeemCode(context.Background(), command, fetched); err != nil {
		t.Fatalf("consume code: %v", err)
	}
	if count, err := codes.CountDocuments(context.Background(), bson.D{{Key: "_id", Value: firstCodeID}}); err != nil || count != 0 {
		t.Fatalf("first code count = %d, err=%v", count, err)
	}
	if count, err := codes.CountDocuments(context.Background(), bson.D{{Key: "_id", Value: secondCodeID}}); err != nil || count != 1 {
		t.Fatalf("second code count = %d, err=%v", count, err)
	}
	if count, err := balances.CountDocuments(context.Background(), bson.D{{Key: "_id", Value: firstBalanceID}}); err != nil || count != 0 {
		t.Fatalf("first balance count = %d, err=%v", count, err)
	}
	var second bson.M
	if err := balances.FindOne(context.Background(), bson.D{{Key: "_id", Value: secondBalanceID}}).Decode(&second); err != nil || second["price"] != int32(20) || second["marker"] != "second" {
		t.Fatalf("second balance = %#v, err=%v", second, err)
	}
	var replacement bson.M
	if err := balances.FindOne(context.Background(), bson.D{{Key: "guild", Value: "guild-1"}, {Key: "_id", Value: bson.D{{Key: "$nin", Value: bson.A{secondBalanceID}}}}}).Decode(&replacement); err != nil {
		t.Fatalf("find replacement: %v", err)
	}
	if replacement["price"] != float64(8) || replacement["marker"] != nil || replacement["_id"] == firstBalanceID {
		t.Fatalf("replacement = %#v", replacement)
	}
	var other bson.M
	if err := balances.FindOne(context.Background(), bson.D{{Key: "guild", Value: "guild-2"}}).Decode(&other); err != nil || other["price"] != int32(30) {
		t.Fatalf("other guild = %#v, err=%v", other, err)
	}

	badCodeID := bson.NewObjectID()
	badBalanceID := bson.NewObjectID()
	if _, err := codes.InsertOne(context.Background(), bson.D{{Key: "_id", Value: badCodeID}, {Key: "code", Value: "bad"}, {Key: "price", Value: 5}, {Key: "time", Value: 1_700_000_000_000}}); err != nil {
		t.Fatalf("seed bad code: %v", err)
	}
	if _, err := balances.InsertOne(context.Background(), bson.D{{Key: "_id", Value: badBalanceID}, {Key: "guild", Value: "guild-bad"}, {Key: "price", Value: "not-a-number"}}); err != nil {
		t.Fatalf("seed bad balance: %v", err)
	}
	badCode, err := repository.GetRedeemCode(context.Background(), "bad")
	if err != nil {
		t.Fatalf("get bad code: %v", err)
	}
	err = repository.ConsumeRedeemCode(context.Background(), domain.RedeemCommand{GuildID: "guild-bad", Code: "bad", NowMS: 1_700_000_000_000}, badCode)
	if !errors.Is(err, domain.ErrInvalidRedeemCode) {
		t.Fatalf("bad balance error = %v", err)
	}
	if count, _ := codes.CountDocuments(context.Background(), bson.D{{Key: "_id", Value: badCodeID}}); count != 0 {
		t.Fatalf("bad code count = %d, want consumed before credit failure", count)
	}
	if count, _ := balances.CountDocuments(context.Background(), bson.D{{Key: "_id", Value: badBalanceID}}); count != 1 {
		t.Fatalf("bad balance count = %d, want preserved", count)
	}
}

func TestRedeemMongoIntegrationConsumesConcurrentCodeOnce(t *testing.T) {
	database := redeemIntegrationDatabase(t)
	repository, err := NewRedeemRepositoryFromDatabase(database)
	if err != nil {
		t.Fatalf("new redeem repository: %v", err)
	}
	codeID := bson.NewObjectID()
	if _, err := database.Collection(RedeemCodeCollectionName).InsertOne(context.Background(), bson.D{
		{Key: "_id", Value: codeID},
		{Key: "code", Value: "once"},
		{Key: "price", Value: 5},
		{Key: "time", Value: 1_700_000_000_000},
	}); err != nil {
		t.Fatalf("seed code: %v", err)
	}
	fetched, err := repository.GetRedeemCode(context.Background(), "once")
	if err != nil {
		t.Fatalf("get code: %v", err)
	}
	command := domain.RedeemCommand{GuildID: "guild-1", Code: "once", NowMS: 1_700_000_000_000}
	errorsByCall := make([]error, 2)
	var wait sync.WaitGroup
	for index := range errorsByCall {
		wait.Add(1)
		go func() {
			defer wait.Done()
			errorsByCall[index] = repository.ConsumeRedeemCode(context.Background(), command, fetched)
		}()
	}
	wait.Wait()
	successes := 0
	notFound := 0
	for _, err := range errorsByCall {
		switch {
		case err == nil:
			successes++
		case errors.Is(err, ports.ErrRedeemCodeNotFound):
			notFound++
		default:
			t.Fatalf("unexpected consume error: %v", err)
		}
	}
	if successes != 1 || notFound != 1 {
		t.Fatalf("successes=%d not-found=%d errors=%#v", successes, notFound, errorsByCall)
	}
	var balance struct {
		Price float64 `bson:"price"`
	}
	if err := database.Collection(BalanceCollectionName).FindOne(context.Background(), bson.D{{Key: "guild", Value: "guild-1"}}).Decode(&balance); err != nil || balance.Price != 5 {
		t.Fatalf("balance = %#v, err=%v", balance, err)
	}
}

func redeemIntegrationDatabase(t *testing.T) *drivermongo.Database {
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
		Database:       fmt.Sprintf("mhcat_redeem_test_%d", time.Now().UnixNano()),
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
