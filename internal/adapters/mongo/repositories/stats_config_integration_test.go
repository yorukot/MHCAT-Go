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

func TestStatsConfigMongoIntegrationLifecycleAndIndexes(t *testing.T) {
	database := statsIntegrationDatabase(t)
	repository, err := NewStatsConfigRepositoryFromDatabase(database)
	if err != nil {
		t.Fatalf("new repository: %v", err)
	}
	if names, err := database.ListCollectionNames(context.Background(), bson.D{}); err != nil || len(names) != 0 {
		t.Fatalf("startup collections=%#v err=%v", names, err)
	}

	base := domain.StatsConfig{
		GuildID: "guild-1", ParentID: "parent-1",
		MemberNumberID: "member-1", MemberNumberName: "10",
		UserNumberID: "user-1", UserNumberName: "8",
		BotNumberID: "bot-1", BotNumberName: "2",
	}
	if err := repository.SaveStatsConfig(context.Background(), base); err != nil {
		t.Fatalf("save base config: %v", err)
	}
	numbers := database.Collection(StatsConfigCollectionName)
	roleNumbers := database.Collection(StatsRoleConfigCollectionName)
	if _, err := numbers.UpdateOne(context.Background(), bson.D{{Key: "guild", Value: "guild-1"}}, bson.D{{Key: "$set", Value: bson.D{{Key: "marker", Value: "preserved"}}}}); err != nil {
		t.Fatalf("add marker: %v", err)
	}
	updated, err := repository.AddStatsConfigChannel(context.Background(), "guild-1", domain.StatsOptionChannelCount, "channel-count", 12)
	if err != nil {
		t.Fatalf("add optional channel: %v", err)
	}
	if updated.ChannelNumberID != "channel-count" || updated.ChannelNumberName != "12" || updated.ParentID != "parent-1" {
		t.Fatalf("updated config = %#v", updated)
	}
	memberCounter := " 11 "
	channelCounter := " 13 "
	if err := repository.UpdateStatsConfigCounters(context.Background(), " guild-1 ", domain.StatsConfigCounterUpdate{
		MemberNumberName: &memberCounter, ChannelNumberName: &channelCounter,
	}); err != nil {
		t.Fatalf("update stats counters: %v", err)
	}
	if err := repository.UpdateStatsConfigCounters(context.Background(), "guild-1", domain.StatsConfigCounterUpdate{}); err != nil {
		t.Fatalf("skip empty stats counter update: %v", err)
	}
	var stored bson.M
	if err := numbers.FindOne(context.Background(), bson.D{{Key: "guild", Value: "guild-1"}}).Decode(&stored); err != nil {
		t.Fatalf("read stored config: %v", err)
	}
	if stored["marker"] != "preserved" || stored["channelnumber"] != "channel-count" ||
		stored["memberNumber_name"] != "11" || stored["channelnumber_name"] != "13" {
		t.Fatalf("stored config = %#v", stored)
	}
	for _, field := range []string{"categoriesnumber", "rolesnumber", "statusnumber_name"} {
		value, ok := stored[field]
		if !ok || value != nil {
			t.Fatalf("stored %s = %#v present=%v", field, value, ok)
		}
	}

	if _, err := numbers.InsertMany(context.Background(), []any{
		bson.D{{Key: "guild", Value: "guild-1"}, {Key: "parent", Value: "duplicate-parent"}},
		bson.D{{Key: "guild", Value: "guild-2"}, {Key: "parent", Value: "other-parent"}},
	}); err != nil {
		t.Fatalf("seed duplicate and other config: %v", err)
	}
	deleted, err := repository.DeleteStatsConfig(context.Background(), "guild-1")
	if err != nil {
		t.Fatalf("delete config: %v", err)
	}
	if deleted.ParentID != "parent-1" {
		t.Fatalf("deleted first config = %#v", deleted)
	}
	if count, err := numbers.CountDocuments(context.Background(), bson.D{{Key: "guild", Value: "guild-1"}}); err != nil || count != 0 {
		t.Fatalf("remaining guild-1 configs=%d err=%v", count, err)
	}
	if count, err := numbers.CountDocuments(context.Background(), bson.D{{Key: "guild", Value: "guild-2"}}); err != nil || count != 1 {
		t.Fatalf("other guild configs=%d err=%v", count, err)
	}
	if _, err := repository.DeleteStatsConfig(context.Background(), "guild-1"); !errors.Is(err, ports.ErrStatsConfigMissing) {
		t.Fatalf("missing delete error = %v", err)
	}

	if _, err := roleNumbers.InsertMany(context.Background(), []any{
		bson.D{{Key: "guild", Value: "guild-1"}, {Key: "channel", Value: "old-1"}, {Key: "channel_name", Value: "1"}, {Key: "role", Value: "role-1"}, {Key: "marker", Value: true}},
		bson.D{{Key: "guild", Value: "guild-1"}, {Key: "channel", Value: "old-2"}, {Key: "channel_name", Value: "2"}, {Key: "role", Value: "role-1"}},
		bson.D{{Key: "guild", Value: "guild-2"}, {Key: "channel", Value: "other"}, {Key: "channel_name", Value: "3"}, {Key: "role", Value: "role-1"}},
	}); err != nil {
		t.Fatalf("seed role configs: %v", err)
	}
	if err := repository.SaveStatsRoleConfig(context.Background(), domain.StatsRoleConfig{
		GuildID: "guild-1", ChannelID: "new-channel", ChannelName: "4", RoleID: "role-1",
	}); err != nil {
		t.Fatalf("replace role config: %v", err)
	}
	if err := repository.UpdateStatsRoleConfigCounter(context.Background(), " guild-1 ", " role-1 ", " 5 "); err != nil {
		t.Fatalf("update role stats counter: %v", err)
	}
	var roleRows []bson.M
	cursor, err := roleNumbers.Find(context.Background(), bson.D{{Key: "guild", Value: "guild-1"}, {Key: "role", Value: "role-1"}})
	if err != nil {
		t.Fatalf("find role configs: %v", err)
	}
	if err := cursor.All(context.Background(), &roleRows); err != nil {
		t.Fatalf("decode role configs: %v", err)
	}
	if len(roleRows) != 1 || roleRows[0]["channel"] != "new-channel" || roleRows[0]["channel_name"] != "5" || roleRows[0]["marker"] != nil {
		t.Fatalf("replacement role configs = %#v", roleRows)
	}
	if count, err := roleNumbers.CountDocuments(context.Background(), bson.D{{Key: "guild", Value: "guild-2"}}); err != nil || count != 1 {
		t.Fatalf("other guild role configs=%d err=%v", count, err)
	}

	for _, collection := range []*drivermongo.Collection{numbers, roleNumbers} {
		indexes, err := collection.Indexes().List(context.Background())
		if err != nil {
			t.Fatalf("list %s indexes: %v", collection.Name(), err)
		}
		var rows []bson.M
		if err := indexes.All(context.Background(), &rows); err != nil {
			t.Fatalf("decode %s indexes: %v", collection.Name(), err)
		}
		if len(rows) != 1 || rows[0]["name"] != "_id_" {
			t.Fatalf("%s indexes = %#v", collection.Name(), rows)
		}
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
