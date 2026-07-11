package repositories

import (
	"context"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestEconomySettingsMongoIntegrationReplacesOneDuplicateRow(t *testing.T) {
	database := economyQueryIntegrationDatabase(t)
	repository, err := NewEconomyRepositoryFromDatabase(database)
	if err != nil {
		t.Fatalf("new repository: %v", err)
	}
	ctx := context.Background()
	if _, err := database.Collection(GiftChangeCollectionName).InsertMany(ctx, []any{
		bson.D{{Key: "guild", Value: "guild-1"}, {Key: "coin_number", Value: 100}, {Key: "sign_coin", Value: 10}, {Key: "channel", Value: "old-1"}, {Key: "xp_multiple", Value: 1}, {Key: "time", Value: 0}},
		bson.D{{Key: "guild", Value: "guild-1"}, {Key: "coin_number", Value: 200}, {Key: "sign_coin", Value: 20}, {Key: "channel", Value: "old-2"}, {Key: "xp_multiple", Value: 2}, {Key: "time", Value: 3600}},
	}); err != nil {
		t.Fatalf("seed configs: %v", err)
	}

	if _, err := repository.SaveEconomyConfig(ctx, domain.EconomyConfig{
		GuildID: "guild-1", GachaCost: -5, SignCoins: -10, ChannelID: "new", XPMultiple: -0.5, ResetMarker: 7200, ResetMarkerText: "7200",
	}); err != nil {
		t.Fatalf("save config: %v", err)
	}
	var rows []bson.M
	cursor, err := database.Collection(GiftChangeCollectionName).Find(ctx, bson.D{{Key: "guild", Value: "guild-1"}})
	if err != nil {
		t.Fatalf("find configs: %v", err)
	}
	if err := cursor.All(ctx, &rows); err != nil {
		t.Fatalf("decode configs: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("configs = %#v", rows)
	}
	newRows := 0
	oldRows := 0
	for _, row := range rows {
		if row["channel"] == "new" && row["coin_number"] == int64(-5) && row["sign_coin"] == int64(-10) && row["xp_multiple"] == -0.5 && row["time"] == float64(7200) {
			newRows++
		} else if row["channel"] == "old-1" || row["channel"] == "old-2" {
			oldRows++
		}
	}
	if newRows != 1 || oldRows != 1 {
		t.Fatalf("replacement rows = %#v", rows)
	}
}
