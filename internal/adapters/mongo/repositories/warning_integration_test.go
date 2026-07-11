package repositories

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	mhcatmongo "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/adapters/mongo"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
	drivermongo "go.mongodb.org/mongo-driver/v2/mongo"
)

func TestWarningMongoIntegrationPreservesDuplicateIdentityAndMixedContent(t *testing.T) {
	database := warningIntegrationDatabase(t)
	repository, err := NewWarningHistoryRepositoryFromDatabase(database)
	if err != nil {
		t.Fatalf("new warning repository: %v", err)
	}
	collection := database.Collection(WarningCollectionName)
	firstID := bson.NewObjectID()
	secondID := bson.NewObjectID()
	if _, err := collection.InsertMany(context.Background(), []any{
		bson.D{
			{Key: "_id", Value: firstID},
			{Key: "guild", Value: "guild-1"},
			{Key: "user", Value: "user-1"},
			{Key: "content", Value: bson.A{bson.D{{Key: "moderator", Value: "mod-1"}, {Key: "reason", Value: int32(7)}, {Key: "time", Value: false}}, "raw"}},
		},
		bson.D{
			{Key: "_id", Value: secondID},
			{Key: "guild", Value: "guild-1"},
			{Key: "user", Value: "user-1"},
			{Key: "content", Value: bson.A{bson.D{{Key: "moderator", Value: "mod-2"}, {Key: "reason", Value: "second"}, {Key: "time", Value: "time"}}}},
		},
	}); err != nil {
		t.Fatalf("seed duplicate warnings: %v", err)
	}
	history, err := repository.GetWarningHistory(context.Background(), " guild-1 ", " user-1 ")
	if err != nil || len(history.Entries) != 2 {
		t.Fatalf("warning history = %#v err=%v", history, err)
	}

	result, err := repository.AddWarning(context.Background(), domain.WarningIssue{
		GuildID: "guild-1", UserID: "user-1", ModeratorID: "mod-3", Reason: "added", Time: "new-time",
	})
	if err != nil {
		t.Fatalf("add warning: %v", err)
	}
	if result.Created || len(result.History.Entries) != 3 || result.History.Entries[2].Reason != "added" {
		t.Fatalf("add result = %#v", result)
	}
	assertWarningContentLength(t, collection, firstID, 3)
	assertWarningContentLength(t, collection, secondID, 1)

	if err := repository.RemoveWarning(context.Background(), domain.WarningRemoval{GuildID: "guild-1", UserID: "user-1", Index: 2}); err != nil {
		t.Fatalf("remove mixed warning: %v", err)
	}
	assertWarningContentLength(t, collection, firstID, 2)
	assertWarningContentLength(t, collection, secondID, 1)
	var first bson.M
	if err := collection.FindOne(context.Background(), bson.D{{Key: "_id", Value: firstID}}).Decode(&first); err != nil {
		t.Fatalf("decode first warning: %v", err)
	}
	content, ok := first["content"].(bson.A)
	if !ok || len(content) != 2 {
		t.Fatalf("first content = %#v", first["content"])
	}
	added, ok := content[1].(bson.D)
	var addedReason any
	for _, element := range added {
		if element.Key == "reason" {
			addedReason = element.Value
			break
		}
	}
	if !ok || addedReason != "added" {
		t.Fatalf("preserved appended entry = %#v", content[1])
	}
	if err := repository.RemoveWarning(context.Background(), domain.WarningRemoval{GuildID: "guild-1", UserID: "user-1", Index: 99}); err != nil {
		t.Fatalf("large-index no-op: %v", err)
	}
	assertWarningContentLength(t, collection, firstID, 2)
	if err := repository.RemoveWarning(context.Background(), domain.WarningRemoval{GuildID: "guild-1", UserID: "user-1", Index: 0}); err != nil {
		t.Fatalf("zero-index last removal: %v", err)
	}
	assertWarningContentLength(t, collection, firstID, 1)

	if err := repository.RemoveAllWarnings(context.Background(), domain.WarningRemoval{GuildID: "guild-1", UserID: "user-1"}); err != nil {
		t.Fatalf("remove all duplicate warnings: %v", err)
	}
	count, err := collection.CountDocuments(context.Background(), bson.D{{Key: "guild", Value: "guild-1"}, {Key: "user", Value: "user-1"}})
	if err != nil || count != 0 {
		t.Fatalf("remaining warning count = %d, err=%v", count, err)
	}
}

func TestWarningMongoIntegrationNormalizesScalarContentOnAppend(t *testing.T) {
	database := warningIntegrationDatabase(t)
	repository, err := NewWarningHistoryRepositoryFromDatabase(database)
	if err != nil {
		t.Fatalf("new warning repository: %v", err)
	}
	collection := database.Collection(WarningCollectionName)
	id := bson.NewObjectID()
	if _, err := collection.InsertOne(context.Background(), bson.D{
		{Key: "_id", Value: id},
		{Key: "guild", Value: "guild-1"},
		{Key: "user", Value: "user-1"},
		{Key: "content", Value: "raw"},
	}); err != nil {
		t.Fatalf("seed scalar warning: %v", err)
	}

	result, err := repository.AddWarning(context.Background(), domain.WarningIssue{
		GuildID: "guild-1", UserID: "user-1", ModeratorID: "mod-1", Reason: "  added  ", Time: "time",
	})
	if err != nil {
		t.Fatalf("append warning: %v", err)
	}
	if result.Created || len(result.History.Entries) != 2 || result.History.Entries[1].Reason != "  added  " {
		t.Fatalf("result = %#v", result)
	}
	assertWarningContentLength(t, collection, id, 2)
}

func TestWarningSettingsMongoIntegrationPreservesThresholdScalarsAndDuplicateWrites(t *testing.T) {
	database := warningIntegrationDatabase(t)
	repository, err := NewWarningSettingsRepositoryFromDatabase(database)
	if err != nil {
		t.Fatalf("new warning settings repository: %v", err)
	}
	collection := database.Collection(WarningSettingsCollectionName)
	if _, err := collection.InsertMany(context.Background(), []any{
		bson.D{{Key: "guild", Value: "guild-1"}, {Key: "ban_count", Value: "2.5"}, {Key: "move", Value: domain.WarningSettingsActionKick}},
		bson.D{{Key: "guild", Value: "guild-1"}, {Key: "ban_count", Value: nil}, {Key: "move", Value: domain.WarningSettingsActionBan}},
	}); err != nil {
		t.Fatalf("seed warning settings: %v", err)
	}

	settings, err := repository.GetWarningSettings(context.Background(), "guild-1")
	if err != nil {
		t.Fatalf("get warning settings: %v", err)
	}
	if settings.Threshold != 2.5 || settings.Action != domain.WarningSettingsActionKick {
		t.Fatalf("settings = %#v", settings)
	}
	if err := repository.SaveWarningSettings(context.Background(), domain.WarningSettings{GuildID: "guild-1", Threshold: -2, Action: domain.WarningSettingsActionBan}); err != nil {
		t.Fatalf("save warning settings: %v", err)
	}
	cursor, err := collection.Find(context.Background(), bson.D{{Key: "guild", Value: "guild-1"}})
	if err != nil {
		t.Fatalf("find warning settings: %v", err)
	}
	defer cursor.Close(context.Background())
	var documents []bson.Raw
	if err := cursor.All(context.Background(), &documents); err != nil {
		t.Fatalf("decode warning settings: %v", err)
	}
	if len(documents) != 2 {
		t.Fatalf("settings documents = %d", len(documents))
	}
	for index, document := range documents {
		if document.Lookup("ban_count").Type != bson.TypeString || document.Lookup("ban_count").StringValue() != "-2" || document.Lookup("move").StringValue() != domain.WarningSettingsActionBan {
			t.Fatalf("settings document %d = %v", index, document)
		}
	}
}

func TestWarningSettingsMongoIntegrationReadsUnknownLegacyAction(t *testing.T) {
	database := warningIntegrationDatabase(t)
	repository, err := NewWarningSettingsRepositoryFromDatabase(database)
	if err != nil {
		t.Fatalf("new warning settings repository: %v", err)
	}
	collection := database.Collection(WarningSettingsCollectionName)
	if _, err := collection.InsertOne(context.Background(), bson.D{
		{Key: "guild", Value: "guild-1"},
		{Key: "ban_count", Value: "2"},
		{Key: "move", Value: "mute"},
	}); err != nil {
		t.Fatalf("seed warning settings: %v", err)
	}

	settings, err := repository.GetWarningSettings(context.Background(), "guild-1")
	if err != nil {
		t.Fatalf("get warning settings: %v", err)
	}
	if settings.Threshold != 2 || settings.Action != "mute" {
		t.Fatalf("settings = %#v", settings)
	}
}

func assertWarningContentLength(t *testing.T, collection *drivermongo.Collection, id bson.ObjectID, want int) {
	t.Helper()
	var document struct {
		Content bson.RawValue `bson:"content"`
	}
	if err := collection.FindOne(context.Background(), bson.D{{Key: "_id", Value: id}}).Decode(&document); err != nil {
		t.Fatalf("decode warning %s: %v", id.Hex(), err)
	}
	array, ok := document.Content.ArrayOK()
	if !ok {
		t.Fatalf("warning %s content type = %s", id.Hex(), document.Content.Type)
	}
	values, err := array.Values()
	if err != nil || len(values) != want {
		t.Fatalf("warning %s content length = %d, err=%v, want %d", id.Hex(), len(values), err, want)
	}
}

func warningIntegrationDatabase(t *testing.T) *drivermongo.Database {
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
		Database:       fmt.Sprintf("mhcat_warning_test_%d", time.Now().UnixNano()),
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
