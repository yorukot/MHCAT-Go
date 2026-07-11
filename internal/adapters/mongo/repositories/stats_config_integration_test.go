package repositories

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	mhcatmongo "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/adapters/mongo"
	"go.mongodb.org/mongo-driver/v2/bson"
	drivermongo "go.mongodb.org/mongo-driver/v2/mongo"
)

func TestStatsConfigMongoIntegrationHydratesScalarsAndIsolatesMalformedRows(t *testing.T) {
	database := statsIntegrationDatabase(t)
	repository, err := NewStatsConfigRepositoryFromDatabase(database)
	if err != nil {
		t.Fatalf("new repository: %v", err)
	}
	numbers := database.Collection(StatsConfigCollectionName)
	roleNumbers := database.Collection(StatsRoleConfigCollectionName)
	roleID := bson.NewObjectID()
	if _, err := numbers.InsertMany(context.Background(), []any{
		bson.D{
			{Key: "guild", Value: "guild-1"},
			{Key: "parent", Value: " parent-channel "},
			{Key: "memberNumber", Value: bson.Binary{Data: []byte("member-channel")}},
			{Key: "memberNumber_name", Value: true},
			{Key: "userNumber", Value: bson.D{{Key: "invalid", Value: true}}},
		},
		bson.D{
			{Key: "guild", Value: bson.A{"invalid"}},
			{Key: "parent", Value: "malformed-row"},
		},
	}); err != nil {
		t.Fatalf("seed stats configs: %v", err)
	}
	if _, err := roleNumbers.InsertMany(context.Background(), []any{
		bson.D{
			{Key: "guild", Value: int64(9)},
			{Key: "channel", Value: bson.Binary{Data: []byte("role-channel")}},
			{Key: "channel_name", Value: 4.5},
			{Key: "role", Value: roleID},
		},
		bson.D{
			{Key: "guild", Value: "guild-2"},
			{Key: "channel", Value: bson.D{{Key: "invalid", Value: true}}},
			{Key: "channel_name", Value: nil},
			{Key: "role", Value: bson.A{"invalid"}},
		},
	}); err != nil {
		t.Fatalf("seed role stats configs: %v", err)
	}

	config, err := repository.GetStatsConfig(context.Background(), "guild-1")
	if err != nil {
		t.Fatalf("get scalar stats config: %v", err)
	}
	if config.ParentID != " parent-channel " || config.MemberNumberID != "member-channel" || config.MemberNumberName != "true" || config.UserNumberID != "" {
		t.Fatalf("scalar stats config = %#v", config)
	}

	configs, err := repository.ListStatsConfigs(context.Background())
	if err != nil {
		t.Fatalf("list stats configs: %v", err)
	}
	if len(configs) != 2 {
		t.Fatalf("stats configs = %#v", configs)
	}
	byGuild := make(map[string]string, len(configs))
	for _, item := range configs {
		byGuild[item.GuildID] = item.ParentID
	}
	if byGuild["guild-1"] != " parent-channel " || byGuild[""] != "malformed-row" {
		t.Fatalf("stats configs by guild = %#v", byGuild)
	}

	roles, err := repository.ListStatsRoleConfigs(context.Background())
	if err != nil {
		t.Fatalf("list role stats configs: %v", err)
	}
	if len(roles) != 2 {
		t.Fatalf("role stats configs = %#v", roles)
	}
	byRoleGuild := make(map[string]struct {
		channel string
		count   string
		role    string
	}, len(roles))
	for _, item := range roles {
		byRoleGuild[item.GuildID] = struct {
			channel string
			count   string
			role    string
		}{channel: item.ChannelID, count: item.ChannelName, role: item.RoleID}
	}
	if got := byRoleGuild["9"]; got.channel != "role-channel" || got.count != "4.5" || got.role != roleID.Hex() {
		t.Fatalf("scalar role stats config = %#v", got)
	}
	if got := byRoleGuild["guild-2"]; got.channel != "" || got.count != "" || got.role != "" {
		t.Fatalf("malformed role stats config = %#v", got)
	}
}

func statsIntegrationDatabase(t *testing.T) *drivermongo.Database {
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
		Database:       fmt.Sprintf("mhcat_stats_test_%d", time.Now().UnixNano()),
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
			t.Errorf("disconnect Mongo: %v", err)
		}
	})
	return database
}
