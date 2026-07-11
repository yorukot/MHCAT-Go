package repositories

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	mhcatmongo "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/adapters/mongo"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	"go.mongodb.org/mongo-driver/v2/bson"
	drivermongo "go.mongodb.org/mongo-driver/v2/mongo"
)

func TestAutoChatConfigMongoIntegrationReplacesAndDeletesOneFetchedRow(t *testing.T) {
	database := autoChatConfigIntegrationDatabase(t)
	repository, err := NewAutoChatConfigRepositoryFromDatabase(database)
	if err != nil {
		t.Fatalf("new autochat config repository: %v", err)
	}
	collection := database.Collection(AutoChatConfigCollectionName)
	firstID := bson.NewObjectID()
	secondID := bson.NewObjectID()
	otherID := bson.NewObjectID()
	if _, err := collection.InsertMany(context.Background(), []any{
		bson.D{{Key: "_id", Value: firstID}, {Key: "guild", Value: "guild-1"}, {Key: "channel", Value: bson.D{{Key: "legacy", Value: true}}}, {Key: "marker", Value: "first"}},
		bson.D{{Key: "_id", Value: secondID}, {Key: "guild", Value: "guild-1"}, {Key: "channel", Value: "channel-2"}, {Key: "marker", Value: "second"}},
		bson.D{{Key: "_id", Value: otherID}, {Key: "guild", Value: "guild-2"}, {Key: "channel", Value: "other"}},
	}); err != nil {
		t.Fatalf("seed autochat configs: %v", err)
	}

	if err := repository.SaveAutoChatConfig(context.Background(), domain.AutoChatConfig{GuildID: "guild-1", ChannelID: "channel-new"}); err != nil {
		t.Fatalf("save autochat config: %v", err)
	}
	if count, err := collection.CountDocuments(context.Background(), bson.D{{Key: "guild", Value: "guild-1"}}); err != nil || count != 2 {
		t.Fatalf("guild-1 count=%d err=%v", count, err)
	}
	if count, err := collection.CountDocuments(context.Background(), bson.D{{Key: "_id", Value: firstID}}); err != nil || count != 0 {
		t.Fatalf("first row count=%d err=%v", count, err)
	}
	var untouched bson.M
	if err := collection.FindOne(context.Background(), bson.D{{Key: "_id", Value: secondID}}).Decode(&untouched); err != nil || untouched["channel"] != "channel-2" || untouched["marker"] != "second" {
		t.Fatalf("untouched duplicate=%#v err=%v", untouched, err)
	}
	var replacement bson.M
	if err := collection.FindOne(context.Background(), bson.D{{Key: "guild", Value: "guild-1"}, {Key: "channel", Value: "channel-new"}}).Decode(&replacement); err != nil {
		t.Fatalf("find replacement: %v", err)
	}
	if replacement["_id"] == firstID || replacement["_id"] == secondID || replacement["marker"] != nil || len(replacement) != 3 {
		t.Fatalf("replacement=%#v", replacement)
	}
	var other bson.M
	if err := collection.FindOne(context.Background(), bson.D{{Key: "_id", Value: otherID}}).Decode(&other); err != nil || other["channel"] != "other" {
		t.Fatalf("other guild=%#v err=%v", other, err)
	}

	if err := repository.DeleteAutoChatConfig(context.Background(), "guild-1"); err != nil {
		t.Fatalf("delete first duplicate: %v", err)
	}
	if count, err := collection.CountDocuments(context.Background(), bson.D{{Key: "guild", Value: "guild-1"}}); err != nil || count != 1 {
		t.Fatalf("count after first delete=%d err=%v", count, err)
	}
	if err := repository.DeleteAutoChatConfig(context.Background(), "guild-1"); err != nil {
		t.Fatalf("delete second duplicate: %v", err)
	}
	if err := repository.DeleteAutoChatConfig(context.Background(), "guild-1"); !errors.Is(err, ports.ErrAutoChatConfigMissing) {
		t.Fatalf("missing delete error=%v", err)
	}
}

func autoChatConfigIntegrationDatabase(t *testing.T) *drivermongo.Database {
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
		Database:       fmt.Sprintf("mhcat_autochat_config_test_%d", time.Now().UnixNano()),
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
