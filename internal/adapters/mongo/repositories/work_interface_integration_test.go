package repositories

import (
	"context"
	"math"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestWorkInterfaceMongoIntegrationEnsuresUserOnce(t *testing.T) {
	database := economyQueryIntegrationDatabase(t)
	repository, err := NewWorkInterfaceRepositoryFromDatabase(database)
	if err != nil {
		t.Fatalf("new repository: %v", err)
	}
	ctx := context.Background()
	created, err := repository.EnsureWorkUser(ctx, "guild-1", "user-1", 20, "")
	if err != nil || created.State != LegacyIdleWorkState || created.Energy != 20 || created.EndTimeUnix != 0 || created.GetCoin != 0 {
		t.Fatalf("created = %#v, err=%v", created, err)
	}
	if _, err := database.Collection(WorkUserCollectionName).UpdateOne(ctx,
		bson.D{{Key: "guild", Value: "guild-1"}, {Key: "user", Value: "user-1"}},
		bson.D{{Key: "$set", Value: bson.D{{Key: "energi", Value: 7}, {Key: "state", Value: "busy"}}}},
	); err != nil {
		t.Fatalf("mutate existing user: %v", err)
	}
	existing, err := repository.EnsureWorkUser(ctx, "guild-1", "user-1", 99, "")
	if err != nil || existing.Energy != 7 || existing.State != "busy" {
		t.Fatalf("existing = %#v, err=%v", existing, err)
	}
	count, err := database.Collection(WorkUserCollectionName).CountDocuments(ctx, bson.D{{Key: "guild", Value: "guild-1"}, {Key: "user", Value: "user-1"}})
	if err != nil || count != 1 {
		t.Fatalf("row count = %d, err=%v", count, err)
	}
}

func TestWorkInterfaceMongoIntegrationInitializesScalarMaxEnergy(t *testing.T) {
	database := economyQueryIntegrationDatabase(t)
	repository, err := NewWorkInterfaceRepositoryFromDatabase(database)
	if err != nil {
		t.Fatalf("new repository: %v", err)
	}
	created, err := repository.EnsureWorkUser(context.Background(), "scalar-guild", "scalar-user", 12, "12.5")
	if err != nil || created.EnergyText != "12.5" {
		t.Fatalf("created = %#v, err=%v", created, err)
	}
}

func TestWorkInterfaceMongoIntegrationPreservesLegacyStartArithmetic(t *testing.T) {
	database := economyQueryIntegrationDatabase(t)
	repository, err := NewWorkInterfaceRepositoryFromDatabase(database)
	if err != nil {
		t.Fatalf("new repository: %v", err)
	}
	ctx := context.Background()
	tests := []struct {
		name       string
		raw        string
		wantEnergy string
		wantEnd    string
		wantReward string
	}{
		{name: "decimal", raw: "2.5", wantEnergy: "7.5", wantEnd: "102.5", wantReward: "2.5"},
		{name: "negative", raw: "-2", wantEnergy: "12", wantEnd: "98", wantReward: "-2"},
		{name: "null", raw: "null", wantEnergy: "10", wantEnd: "100", wantReward: "0"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			userID := "start-" + test.name
			if _, err := database.Collection(WorkUserCollectionName).InsertOne(ctx, bson.D{
				{Key: "guild", Value: "guild-1"}, {Key: "user", Value: userID},
				{Key: "state", Value: LegacyIdleWorkState}, {Key: "end_time", Value: 0},
				{Key: "energi", Value: 10}, {Key: "get_coin", Value: 0},
			}); err != nil {
				t.Fatalf("insert user: %v", err)
			}
			updated, err := repository.StartWork(ctx, domain.WorkStartCommand{
				GuildID: "guild-1", UserID: userID, WorkName: "work", MaxEnergy: 10, NowUnix: 100,
				DurationText: test.raw, EnergyCostText: test.raw, CoinRewardText: test.raw,
			})
			if err != nil {
				t.Fatalf("start work: %v", err)
			}
			if updated.EnergyText != test.wantEnergy || updated.EndTimeText != test.wantEnd || updated.GetCoinText != test.wantReward {
				t.Fatalf("updated = %#v", updated)
			}
		})
	}
}

func TestWorkInterfaceMongoIntegrationCoercesExistingEnergyOnStart(t *testing.T) {
	database := economyQueryIntegrationDatabase(t)
	repository, err := NewWorkInterfaceRepositoryFromDatabase(database)
	if err != nil {
		t.Fatalf("new repository: %v", err)
	}
	ctx := context.Background()
	tests := []struct {
		name       string
		energy     any
		cost       string
		wantEnergy string
	}{
		{name: "decimal", energy: 5.5, cost: "2", wantEnergy: "3.5"},
		{name: "numeric-string", energy: "5.5", cost: "2", wantEnergy: "3.5"},
		{name: "null", energy: nil, cost: "-2", wantEnergy: "2"},
		{name: "boolean", energy: true, cost: "1", wantEnergy: "0"},
		{name: "infinity", energy: math.Inf(1), cost: "2", wantEnergy: "Infinity"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			userID := "energy-" + test.name
			if _, err := database.Collection(WorkUserCollectionName).InsertOne(ctx, bson.D{
				{Key: "guild", Value: "energy-guild"}, {Key: "user", Value: userID},
				{Key: "state", Value: LegacyIdleWorkState}, {Key: "end_time", Value: 0},
				{Key: "energi", Value: test.energy}, {Key: "get_coin", Value: 0},
			}); err != nil {
				t.Fatalf("insert user: %v", err)
			}
			updated, err := repository.StartWork(ctx, domain.WorkStartCommand{
				GuildID: "energy-guild", UserID: userID, WorkName: "work", NowUnix: 100,
				DurationText: "1", EnergyCostText: test.cost, CoinRewardText: "1",
			})
			if err != nil {
				t.Fatalf("start work: %v", err)
			}
			if updated.EnergyText != test.wantEnergy {
				t.Fatalf("updated = %#v", updated)
			}
		})
	}
}

func TestWorkInterfaceMongoIntegrationUsesLegacyNaturalItemOrder(t *testing.T) {
	database := economyQueryIntegrationDatabase(t)
	repository, err := NewWorkInterfaceRepositoryFromDatabase(database)
	if err != nil {
		t.Fatalf("new repository: %v", err)
	}
	ctx := context.Background()
	collection := database.Collection(WorkSomethingCollectionName)
	if _, err := collection.InsertMany(ctx, []any{
		bson.D{{Key: "_id", Value: "z"}, {Key: "guild", Value: "order-guild"}, {Key: "name", Value: "first"}},
		bson.D{{Key: "_id", Value: "a"}, {Key: "guild", Value: "order-guild"}, {Key: "name", Value: "second"}},
	}); err != nil {
		t.Fatalf("insert items: %v", err)
	}
	items, err := repository.ListWorkItems(ctx, "order-guild")
	if err != nil {
		t.Fatalf("list items: %v", err)
	}
	if len(items) != 2 || items[0].Name != "first" || items[1].Name != "second" {
		t.Fatalf("items = %#v", items)
	}
}

func TestWorkInterfaceMongoIntegrationDeletesOneDuplicateItem(t *testing.T) {
	database := economyQueryIntegrationDatabase(t)
	repository, err := NewWorkInterfaceRepositoryFromDatabase(database)
	if err != nil {
		t.Fatalf("new repository: %v", err)
	}
	ctx := context.Background()
	filter := bson.D{{Key: "guild", Value: "delete-guild"}, {Key: "name", Value: "duplicate"}}
	if _, err := database.Collection(WorkSomethingCollectionName).InsertMany(ctx, []any{filter, filter}); err != nil {
		t.Fatalf("insert duplicate items: %v", err)
	}
	if err := repository.DeleteWorkItem(ctx, domain.WorkDeleteItemCommand{GuildID: "delete-guild", Name: "duplicate"}); err != nil {
		t.Fatalf("delete item: %v", err)
	}
	count, err := database.Collection(WorkSomethingCollectionName).CountDocuments(ctx, filter)
	if err != nil || count != 1 {
		t.Fatalf("duplicate count = %d, err=%v", count, err)
	}
}

func TestWorkInterfaceMongoIntegrationReplacesOneDuplicateConfig(t *testing.T) {
	database := economyQueryIntegrationDatabase(t)
	repository, err := NewWorkInterfaceRepositoryFromDatabase(database)
	if err != nil {
		t.Fatalf("new repository: %v", err)
	}
	ctx := context.Background()
	collection := database.Collection(WorkSetCollectionName)
	if _, err := collection.InsertMany(ctx, []any{
		bson.D{{Key: "guild", Value: "config-guild"}, {Key: "get_energy", Value: 1}, {Key: "max_energy", Value: 10}},
		bson.D{{Key: "guild", Value: "config-guild"}, {Key: "get_energy", Value: 2}, {Key: "max_energy", Value: 20}},
	}); err != nil {
		t.Fatalf("insert duplicate configs: %v", err)
	}
	if _, err := repository.SaveWorkConfig(ctx, domain.WorkConfigCommand{GuildID: "config-guild", DailyEnergy: 3, MaxEnergy: 30, Captcha: true}); err != nil {
		t.Fatalf("save config: %v", err)
	}
	cursor, err := collection.Find(ctx, bson.D{{Key: "guild", Value: "config-guild"}})
	if err != nil {
		t.Fatalf("find configs: %v", err)
	}
	var rows []struct {
		DailyEnergy int64 `bson:"get_energy"`
		MaxEnergy   int64 `bson:"max_energy"`
	}
	if err := cursor.All(ctx, &rows); err != nil {
		t.Fatalf("decode configs: %v", err)
	}
	if len(rows) != 2 || rows[0].DailyEnergy != 2 || rows[0].MaxEnergy != 20 || rows[1].DailyEnergy != 3 || rows[1].MaxEnergy != 30 {
		t.Fatalf("configs = %#v", rows)
	}
}

func TestWorkInterfaceMongoIntegrationAcceptsSignedEnergyGrants(t *testing.T) {
	database := economyQueryIntegrationDatabase(t)
	repository, err := NewWorkInterfaceRepositoryFromDatabase(database)
	if err != nil {
		t.Fatalf("new repository: %v", err)
	}
	ctx := context.Background()
	if _, err := database.Collection(WorkUserCollectionName).InsertMany(ctx, []any{
		bson.D{{Key: "guild", Value: "grant-guild"}, {Key: "user", Value: "one"}, {Key: "energi", Value: 5}},
		bson.D{{Key: "guild", Value: "grant-guild"}, {Key: "user", Value: "two"}, {Key: "energi", Value: 1}},
	}); err != nil {
		t.Fatalf("insert users: %v", err)
	}
	updated, err := repository.GrantWorkEnergy(ctx, domain.WorkEnergyGrantCommand{GuildID: "grant-guild", UserID: "one", Amount: -7, MaxEnergy: 10})
	if err != nil || updated.Energy != -2 {
		t.Fatalf("negative grant = %#v, err=%v", updated, err)
	}
	result, err := repository.GrantWorkEnergyToAll(ctx, domain.WorkEnergyGrantAllCommand{GuildID: "grant-guild", Amount: 0, MaxEnergy: 10})
	if err != nil || result.Matched != 2 || result.Modified != 0 {
		t.Fatalf("zero grant = %#v, err=%v", result, err)
	}
}
