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
	corenotifications "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/services/notifications"
	"go.mongodb.org/mongo-driver/v2/bson"
	drivermongo "go.mongodb.org/mongo-driver/v2/mongo"
)

func TestAutoNotificationMongoIntegrationKeepsMalformedRowsIsolated(t *testing.T) {
	database := autoNotificationIntegrationDatabase(t)
	repository, err := NewAutoNotificationScheduleRepositoryFromDatabase(database)
	if err != nil {
		t.Fatalf("new repository: %v", err)
	}
	collection := database.Collection(AutoNotificationScheduleCollectionName)
	if _, err := collection.InsertMany(context.Background(), []any{
		bson.D{{Key: "guild", Value: "guild-1"}, {Key: "id", Value: "valid"}, {Key: "cron", Value: "*/30 * * * *"}, {Key: "channel", Value: "channel-1"}, {Key: "message", Value: bson.D{{Key: "content", Value: "valid"}}}},
		bson.D{{Key: "guild", Value: "guild-1"}, {Key: "id", Value: "scalar"}, {Key: "cron", Value: bson.Binary{Data: []byte("15 * * * *")}}, {Key: "channel", Value: bson.Binary{Data: []byte("channel-2")}}, {Key: "message", Value: bson.D{{Key: "content", Value: "scalar"}}}},
		bson.D{{Key: "guild", Value: "guild-1"}, {Key: "id", Value: bson.A{"invalid"}}, {Key: "cron", Value: bson.D{{Key: "invalid", Value: true}}}, {Key: "channel", Value: bson.D{{Key: "invalid", Value: true}}}, {Key: "message", Value: bson.D{{Key: "content", Value: "malformed"}}}},
		bson.D{{Key: "guild", Value: "guild-1"}, {Key: "id", Value: "bad-embeds"}, {Key: "cron", Value: "*/30 * * * *"}, {Key: "channel", Value: "channel-1"}, {Key: "message", Value: bson.D{{Key: "content", Value: "plain survives"}, {Key: "embeds", Value: bson.D{{Key: "not", Value: "array"}}}}}},
		bson.D{{Key: "guild", Value: "guild-1"}, {Key: "id", Value: "numeric-content"}, {Key: "cron", Value: "*/30 * * * *"}, {Key: "channel", Value: "channel-1"}, {Key: "message", Value: bson.D{{Key: "content", Value: int32(123)}}}},
		bson.D{{Key: "guild", Value: int64(1)}, {Key: "id", Value: "numeric-guild"}, {Key: "cron", Value: "*/30 * * * *"}, {Key: "channel", Value: "channel-1"}, {Key: "message", Value: bson.D{{Key: "content", Value: "excluded"}}}},
		bson.D{{Key: "guild", Value: "guild-1"}, {Key: "id", Value: "pending"}, {Key: "cron", Value: nil}, {Key: "channel", Value: "channel-1"}, {Key: "message", Value: bson.D{{Key: "content", Value: "excluded"}}}},
		bson.D{{Key: "guild", Value: "guild-1"}, {Key: "id", Value: "scalar-message"}, {Key: "cron", Value: "*/30 * * * *"}, {Key: "channel", Value: "channel-1"}, {Key: "message", Value: "excluded"}},
	}); err != nil {
		t.Fatalf("seed schedules: %v", err)
	}

	schedules, err := repository.ListAutoNotificationDeliveries(context.Background())
	if err != nil {
		t.Fatalf("list deliveries: %v", err)
	}
	if len(schedules) != 5 {
		t.Fatalf("deliveries = %#v", schedules)
	}
	byID := make(map[string]string, len(schedules))
	messagesByID := make(map[string]string, len(schedules))
	for _, schedule := range schedules {
		byID[schedule.ID] = schedule.Cron + "|" + schedule.ChannelID
		messagesByID[schedule.ID] = schedule.Message.Content
	}
	if byID["valid"] != "*/30 * * * *|channel-1" || byID["scalar"] != "15 * * * *|channel-2" || byID[""] != "|" {
		t.Fatalf("deliveries by id = %#v", byID)
	}
	if messagesByID["bad-embeds"] != "plain survives" || messagesByID["numeric-content"] != "" {
		t.Fatalf("delivery messages = %#v", messagesByID)
	}

	scalar, err := repository.GetAutoNotificationDelivery(context.Background(), "guild-1", "scalar")
	if err != nil || scalar.Cron != "15 * * * *" || scalar.ChannelID != "channel-2" {
		t.Fatalf("scalar delivery = %#v, err=%v", scalar, err)
	}
}

func TestAutoNotificationMongoIntegrationStartupDoesNotMutateDatabase(t *testing.T) {
	database := autoNotificationIntegrationDatabase(t)
	if _, err := NewAutoNotificationScheduleRepositoryFromDatabase(database); err != nil {
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

func TestAutoNotificationMongoIntegrationConfigLifecycle(t *testing.T) {
	database := autoNotificationIntegrationDatabase(t)
	repository, err := NewAutoNotificationScheduleRepositoryFromDatabase(database)
	if err != nil {
		t.Fatalf("new repository: %v", err)
	}
	service := corenotifications.NewScheduleService(repository)
	collection := database.Collection(AutoNotificationScheduleCollectionName)
	if _, err := collection.InsertMany(context.Background(), []any{
		bson.D{{Key: "guild", Value: "guild-1"}, {Key: "id", Value: "active"}, {Key: "cron", Value: "*/30 * * * *"}, {Key: "channel", Value: "channel-1"}, {Key: "message", Value: bson.D{{Key: "content", Value: "active"}}}},
		bson.D{{Key: "guild", Value: "guild-1"}, {Key: "id", Value: "null-draft"}, {Key: "cron", Value: nil}, {Key: "channel", Value: "channel-1"}, {Key: "message", Value: nil}},
		bson.D{{Key: "guild", Value: "guild-1"}, {Key: "id", Value: "missing-draft"}, {Key: "channel", Value: "channel-1"}, {Key: "message", Value: nil}},
		bson.D{{Key: "guild", Value: "guild-2"}, {Key: "id", Value: "other"}, {Key: "cron", Value: nil}, {Key: "channel", Value: "channel-2"}, {Key: "message", Value: nil}},
	}); err != nil {
		t.Fatalf("seed schedules: %v", err)
	}

	schedules, err := service.List(context.Background(), "guild-1")
	if err != nil {
		t.Fatalf("list schedules: %v", err)
	}
	if len(schedules) != 1 || schedules[0].ID != "active" {
		t.Fatalf("active schedules = %#v", schedules)
	}
	if count, err := collection.CountDocuments(context.Background(), bson.D{{Key: "guild", Value: "guild-1"}}); err != nil || count != 1 {
		t.Fatalf("guild-1 rows after cleanup=%d err=%v", count, err)
	}
	if count, err := collection.CountDocuments(context.Background(), bson.D{{Key: "guild", Value: "guild-2"}}); err != nil || count != 1 {
		t.Fatalf("other guild rows=%d err=%v", count, err)
	}

	draft := domain.AutoNotificationSetupDraft{GuildID: "guild-1", ID: "new", ChannelID: "channel-new"}
	if err := service.StartSetup(context.Background(), draft); err != nil {
		t.Fatalf("start setup: %v", err)
	}
	if _, err := collection.UpdateOne(context.Background(), bson.D{{Key: "guild", Value: "guild-1"}, {Key: "id", Value: "new"}}, bson.D{{Key: "$set", Value: bson.D{{Key: "marker", Value: "preserved"}}}}); err != nil {
		t.Fatalf("add marker: %v", err)
	}
	if err := service.CompleteSetup(context.Background(), domain.AutoNotificationSetup{
		GuildID: "guild-1",
		ID:      "new",
		Cron:    "0 9 * * 1",
		Message: domain.AutoNotificationMessage{Content: "hello", EmbedTitle: "Title", EmbedColor: "#123456"},
	}); err != nil {
		t.Fatalf("complete setup: %v", err)
	}
	var completed bson.M
	if err := collection.FindOne(context.Background(), bson.D{{Key: "guild", Value: "guild-1"}, {Key: "id", Value: "new"}}).Decode(&completed); err != nil {
		t.Fatalf("find completed schedule: %v", err)
	}
	if completed["cron"] != "0 9 * * 1" || completed["channel"] != "channel-new" || completed["marker"] != "preserved" {
		t.Fatalf("completed schedule = %#v", completed)
	}
	message, ok := completed["message"].(bson.D)
	if !ok || autoNotificationFilterValue(t, message, "content") != "hello" {
		t.Fatalf("completed message = %#v", completed["message"])
	}

	if err := service.Delete(context.Background(), "guild-1", "new"); err != nil {
		t.Fatalf("delete schedule: %v", err)
	}
	if err := service.Delete(context.Background(), "guild-1", "new"); !errors.Is(err, ports.ErrAutoNotificationScheduleMissing) {
		t.Fatalf("missing delete error = %v", err)
	}
	indexes, err := collection.Indexes().List(context.Background())
	if err != nil {
		t.Fatalf("list indexes: %v", err)
	}
	defer indexes.Close(context.Background())
	var indexRows []bson.M
	if err := indexes.All(context.Background(), &indexRows); err != nil {
		t.Fatalf("decode indexes: %v", err)
	}
	if len(indexRows) != 1 || indexRows[0]["name"] != "_id_" {
		t.Fatalf("indexes = %#v", indexRows)
	}
}

func TestAutoNotificationMongoIntegrationSetupUsesExactStringID(t *testing.T) {
	database := autoNotificationIntegrationDatabase(t)
	repository, err := NewAutoNotificationScheduleRepositoryFromDatabase(database)
	if err != nil {
		t.Fatalf("new repository: %v", err)
	}
	collection := database.Collection(AutoNotificationScheduleCollectionName)
	if _, err := collection.InsertOne(context.Background(), bson.D{
		{Key: "guild", Value: "guild-1"},
		{Key: "id", Value: int64(123)},
		{Key: "cron", Value: "*/30 * * * *"},
		{Key: "channel", Value: "channel-old"},
		{Key: "message", Value: bson.D{{Key: "content", Value: "old"}}},
	}); err != nil {
		t.Fatalf("seed numeric id: %v", err)
	}
	service := corenotifications.NewScheduleService(repository)
	draft := domain.AutoNotificationSetupDraft{GuildID: "guild-1", ID: "123", ChannelID: "channel-new"}
	if err := service.StartSetup(context.Background(), draft); err != nil {
		t.Fatalf("start setup: %v", err)
	}
	if count, err := collection.CountDocuments(context.Background(), bson.D{{Key: "guild", Value: "guild-1"}}); err != nil || count != 2 {
		t.Fatalf("guild rows=%d err=%v", count, err)
	}
	if count, err := collection.CountDocuments(context.Background(), bson.D{{Key: "guild", Value: "guild-1"}, {Key: "id", Value: "123"}}); err != nil || count != 1 {
		t.Fatalf("string-id rows=%d err=%v", count, err)
	}
	if err := service.StartSetup(context.Background(), draft); !errors.Is(err, ports.ErrAutoNotificationScheduleExists) {
		t.Fatalf("second setup error = %v", err)
	}
}

func autoNotificationIntegrationDatabase(t *testing.T) *drivermongo.Database {
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
		Database:       fmt.Sprintf("mhcat_auto_notification_test_%d", time.Now().UnixNano()),
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
