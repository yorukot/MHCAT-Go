package repositories

import (
	"context"
	"errors"
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"
	drivermongo "go.mongodb.org/mongo-driver/v2/mongo"
)

func TestDailyResetMongoIntegrationPreviewAndRun(t *testing.T) {
	database := xpAccrualIntegrationDatabase(t)
	repository, err := NewDailyResetRepositoryFromDatabase(database)
	if err != nil {
		t.Fatalf("new daily reset repository: %v", err)
	}
	ctx := context.Background()

	if _, err := database.Collection(GiftChangeCollectionName).InsertMany(ctx, []any{
		bson.D{{Key: "guild", Value: "rolling"}, {Key: "time", Value: "7200"}},
		bson.D{{Key: "guild", Value: " rolling "}, {Key: "time", Value: int64(3600)}},
		bson.D{{Key: "guild", Value: "daily"}, {Key: "time", Value: "0"}},
		bson.D{{Key: "guild", Value: "missing-time"}},
		bson.D{{Key: "guild", Value: " "}, {Key: "time", Value: int32(1)}},
	}); err != nil {
		t.Fatalf("seed economy configs: %v", err)
	}
	if _, err := database.Collection(CoinCollectionName).InsertMany(ctx, []any{
		bson.D{{Key: "guild", Value: "rolling"}, {Key: "member", Value: "u1"}, {Key: "today", Value: int64(5)}},
		bson.D{{Key: "guild", Value: "daily"}, {Key: "member", Value: "u2"}, {Key: "today", Value: int32(4)}},
		bson.D{{Key: "guild", Value: "missing-time"}, {Key: "member", Value: "u3"}, {Key: "today", Value: "3"}},
		bson.D{{Key: "guild", Value: "daily"}, {Key: "member", Value: "u4"}, {Key: "today", Value: int64(0)}},
	}); err != nil {
		t.Fatalf("seed coins: %v", err)
	}
	if _, err := database.Collection(WorkSetCollectionName).InsertMany(ctx, []any{
		bson.D{{Key: "guild", Value: "work"}, {Key: "get_energy", Value: "3"}, {Key: "max_energy", Value: int32(10)}},
		bson.D{{Key: "guild", Value: "duplicate"}, {Key: "get_energy", Value: int64(2)}, {Key: "max_energy", Value: "10"}},
		bson.D{{Key: "guild", Value: "duplicate"}, {Key: "get_energy", Value: int64(2)}, {Key: "max_energy", Value: "10"}},
		bson.D{{Key: "guild", Value: " "}, {Key: "get_energy", Value: int64(99)}, {Key: "max_energy", Value: int64(100)}},
	}); err != nil {
		t.Fatalf("seed work configs: %v", err)
	}
	if _, err := database.Collection(WorkUserCollectionName).InsertMany(ctx, []any{
		bson.D{{Key: "guild", Value: "work"}, {Key: "user", Value: "increment"}, {Key: "energi", Value: int64(5)}},
		bson.D{{Key: "guild", Value: "work"}, {Key: "user", Value: "overshoot"}, {Key: "energi", Value: int64(9)}},
		bson.D{{Key: "guild", Value: "work"}, {Key: "user", Value: "clamp"}, {Key: "energi", Value: int64(12)}},
		bson.D{{Key: "guild", Value: "work"}, {Key: "user", Value: "full"}, {Key: "energi", Value: int64(10)}},
		bson.D{{Key: "guild", Value: "duplicate"}, {Key: "user", Value: "repeat"}, {Key: "energi", Value: int64(5)}},
	}); err != nil {
		t.Fatalf("seed work users: %v", err)
	}

	preview, err := repository.PreviewDailyReset(ctx)
	if err != nil {
		t.Fatalf("preview daily reset: %v", err)
	}
	if preview.ExcludedGuilds != 1 || preview.CoinsMatched != 2 || preview.WorkGuilds != 3 ||
		preview.WorkEnergyIncrements != 4 || preview.WorkEnergyClamps != 1 {
		t.Fatalf("preview result = %#v", preview)
	}

	result, err := repository.RunDailyReset(ctx)
	if err != nil {
		t.Fatalf("run daily reset: %v", err)
	}
	if result.ExcludedGuilds != 1 || result.CoinsMatched != 2 || result.CoinsModified != 2 ||
		result.WorkGuilds != 3 || result.WorkEnergyIncrements != 4 || result.WorkEnergyClamps != 2 {
		t.Fatalf("run result = %#v", result)
	}
	assertDailyResetValue(t, database, CoinCollectionName, bson.D{{Key: "member", Value: "u1"}}, "today", int64(5))
	assertDailyResetValue(t, database, CoinCollectionName, bson.D{{Key: "member", Value: "u2"}}, "today", int64(0))
	assertDailyResetValue(t, database, CoinCollectionName, bson.D{{Key: "member", Value: "u3"}}, "today", int64(0))
	assertDailyResetValue(t, database, WorkUserCollectionName, bson.D{{Key: "user", Value: "increment"}}, "energi", int64(8))
	assertDailyResetValue(t, database, WorkUserCollectionName, bson.D{{Key: "user", Value: "overshoot"}}, "energi", int64(10))
	assertDailyResetValue(t, database, WorkUserCollectionName, bson.D{{Key: "user", Value: "clamp"}}, "energi", int64(10))
	assertDailyResetValue(t, database, WorkUserCollectionName, bson.D{{Key: "user", Value: "repeat"}}, "energi", int64(9))

	second, err := repository.RunDailyReset(ctx)
	if err != nil {
		t.Fatalf("rerun daily reset: %v", err)
	}
	if second.CoinsMatched != 0 || second.CoinsModified != 0 || second.WorkEnergyIncrements != 2 || second.WorkEnergyClamps != 2 {
		t.Fatalf("rerun result = %#v", second)
	}
	assertDailyResetValue(t, database, WorkUserCollectionName, bson.D{{Key: "user", Value: "increment"}}, "energi", int64(10))
	assertDailyResetValue(t, database, WorkUserCollectionName, bson.D{{Key: "user", Value: "repeat"}}, "energi", int64(10))
}

func TestDailyResetMongoIntegrationHonorsCanceledContext(t *testing.T) {
	database := xpAccrualIntegrationDatabase(t)
	repository, err := NewDailyResetRepositoryFromDatabase(database)
	if err != nil {
		t.Fatalf("new daily reset repository: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := repository.PreviewDailyReset(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("preview cancellation error = %v", err)
	}
	if _, err := repository.RunDailyReset(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("run cancellation error = %v", err)
	}
}

func assertDailyResetValue(t *testing.T, database *drivermongo.Database, collectionName string, filter bson.D, field string, want int64) {
	t.Helper()
	var document bson.Raw
	if err := database.Collection(collectionName).FindOne(context.Background(), filter).Decode(&document); err != nil {
		t.Fatalf("read %s value: %v", collectionName, err)
	}
	if got := rawInt64(document.Lookup(field)); got != want {
		t.Fatalf("%s.%s = %d, want %d", collectionName, field, got, want)
	}
}
