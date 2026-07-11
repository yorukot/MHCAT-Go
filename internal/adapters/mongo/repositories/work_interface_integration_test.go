package repositories

import (
	"context"
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestWorkInterfaceMongoIntegrationEnsuresUserOnce(t *testing.T) {
	database := economyQueryIntegrationDatabase(t)
	repository, err := NewWorkInterfaceRepositoryFromDatabase(database)
	if err != nil {
		t.Fatalf("new repository: %v", err)
	}
	ctx := context.Background()
	created, err := repository.EnsureWorkUser(ctx, "guild-1", "user-1", 20)
	if err != nil || created.State != LegacyIdleWorkState || created.Energy != 20 || created.EndTimeUnix != 0 || created.GetCoin != 0 {
		t.Fatalf("created = %#v, err=%v", created, err)
	}
	if _, err := database.Collection(WorkUserCollectionName).UpdateOne(ctx,
		bson.D{{Key: "guild", Value: "guild-1"}, {Key: "user", Value: "user-1"}},
		bson.D{{Key: "$set", Value: bson.D{{Key: "energi", Value: 7}, {Key: "state", Value: "busy"}}}},
	); err != nil {
		t.Fatalf("mutate existing user: %v", err)
	}
	existing, err := repository.EnsureWorkUser(ctx, "guild-1", "user-1", 99)
	if err != nil || existing.Energy != 7 || existing.State != "busy" {
		t.Fatalf("existing = %#v, err=%v", existing, err)
	}
	count, err := database.Collection(WorkUserCollectionName).CountDocuments(ctx, bson.D{{Key: "guild", Value: "guild-1"}, {Key: "user", Value: "user-1"}})
	if err != nil || count != 1 {
		t.Fatalf("row count = %d, err=%v", count, err)
	}
}
