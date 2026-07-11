package repositories

import (
	"context"
	"reflect"
	"testing"
	"time"

	coreeconomy "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/services/economy"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestEconomyProfileMongoIntegrationStartupDoesNotMutateDatabase(t *testing.T) {
	database := economyQueryIntegrationDatabase(t)
	if _, err := NewEconomyProfileRepositoryFromDatabase(database); err != nil {
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

func TestEconomyProfileMongoIntegrationPreservesLegacyRowsAndScalars(t *testing.T) {
	database := economyQueryIntegrationDatabase(t)
	repository, err := NewEconomyProfileRepositoryFromDatabase(database)
	if err != nil {
		t.Fatalf("new repository: %v", err)
	}
	ctx := context.Background()
	fixtures := map[string][]any{
		CoinCollectionName: {
			bson.D{{Key: "guild", Value: "guild-1"}, {Key: "member", Value: "viewer"}, {Key: "coin", Value: 10.5}, {Key: "today", Value: 100.5}},
			bson.D{{Key: "guild", Value: "guild-1"}, {Key: "member", Value: "malformed"}, {Key: "coin", Value: bson.D{{Key: "bad", Value: true}}}},
			bson.D{{Key: "guild", Value: "guild-1"}, {Key: "member", Value: "viewer"}, {Key: "coin", Value: 30}},
		},
		GiftChangeCollectionName: {
			bson.D{{Key: "guild", Value: "guild-1"}, {Key: "coin_number", Value: 500.5}, {Key: "sign_coin", Value: nil}, {Key: "xp_multiple", Value: 1.5}, {Key: "time", Value: -1}},
		},
		WorkSetCollectionName: {
			bson.D{{Key: "guild", Value: "guild-1"}, {Key: "get_energy", Value: 10.5}, {Key: "max_energy", Value: nil}},
		},
		WorkUserCollectionName: {
			bson.D{{Key: "guild", Value: "guild-1"}, {Key: "user", Value: "viewer"}, {Key: "energi", Value: 20.5}, {Key: "end_time", Value: 200.5}},
		},
		TextXPCollectionName: {
			bson.D{{Key: "guild", Value: "guild-1"}, {Key: "member", Value: "viewer"}, {Key: "xp", Value: 12.5}, {Key: "leavel", Value: 2.5}},
			bson.D{{Key: "guild", Value: "guild-1"}, {Key: "member", Value: "other"}, {Key: "xp", Value: bson.D{{Key: "bad", Value: true}}}, {Key: "leavel", Value: 2}},
		},
		VoiceXPCollectionName: {
			bson.D{{Key: "guild", Value: "guild-1"}, {Key: "member", Value: "viewer"}, {Key: "xp", Value: nil}, {Key: "leavel", Value: 3.5}},
		},
	}
	for collection, rows := range fixtures {
		if _, err := database.Collection(collection).InsertMany(ctx, rows); err != nil {
			t.Fatalf("seed %s: %v", collection, err)
		}
	}

	result, err := (coreeconomy.ProfileService{Repository: repository}).Query(ctx, coreeconomy.ProfileQuery{
		GuildID: "guild-1", UserID: "viewer", Now: time.Unix(100, 500_000_000),
	})
	if err != nil {
		t.Fatalf("query profile: %v", err)
	}
	if result.CoinBalance.CoinsText != "10.5" || result.CoinBalance.TodayText != "100.5" || result.CoinRank != 3 || result.SignStatus != "未簽到" {
		t.Fatalf("coin result = %#v", result)
	}
	if result.Config.GachaCostText != "500.5" || result.Config.SignCoinsText != "null" || result.Config.XPMultipleText != "1.5" || result.Config.ResetMarkerText != "-1" {
		t.Fatalf("config = %#v", result.Config)
	}
	if result.WorkConfig.DailyEnergyText != "10.5" || result.WorkConfig.MaxEnergyText != "null" || result.WorkUser.EnergyText != "20.5" || result.WorkUser.EndTimeText != "200.5" {
		t.Fatalf("work config=%#v user=%#v", result.WorkConfig, result.WorkUser)
	}
	if result.TextXP.XPText != "12.5" || result.TextXP.LevelText != "2.5" || result.TextRank != 2 || result.VoiceXP.XPText != "null" || result.VoiceXP.LevelText != "3.5" {
		t.Fatalf("xp result = %#v", result)
	}
	balances, err := repository.ListCoinBalances(ctx, "guild-1")
	if err != nil {
		t.Fatalf("list balances: %v", err)
	}
	got := make([]string, 0, len(balances))
	for _, balance := range balances {
		got = append(got, balance.UserID+":"+balance.CoinsText)
	}
	if want := []string{"viewer:10.5", "malformed:undefined", "viewer:30"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("balances = %#v want %#v", got, want)
	}
}
