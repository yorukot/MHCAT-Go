package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestEconomySignInMongoIntegrationStartupDoesNotMutateDatabase(t *testing.T) {
	database := economyQueryIntegrationDatabase(t)
	if _, err := NewEconomyRepositoryFromDatabase(database); err != nil {
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

func TestEconomySignInMongoIntegrationPreservesRewardAndMarkerScalars(t *testing.T) {
	database := economyQueryIntegrationDatabase(t)
	repository, err := NewEconomyRepositoryFromDatabase(database)
	if err != nil {
		t.Fatalf("new repository: %v", err)
	}
	ctx := context.Background()
	if _, err := database.Collection(GiftChangeCollectionName).InsertOne(ctx, bson.D{
		{Key: "guild", Value: "guild-1"}, {Key: "sign_coin", Value: 25.5}, {Key: "time", Value: 0},
	}); err != nil {
		t.Fatalf("seed config: %v", err)
	}
	if _, err := database.Collection(CoinCollectionName).InsertOne(ctx, bson.D{
		{Key: "guild", Value: "guild-1"}, {Key: "member", Value: "existing"}, {Key: "coin", Value: nil}, {Key: "today", Value: nil},
	}); err != nil {
		t.Fatalf("seed balance: %v", err)
	}

	existing, err := repository.SignIn(ctx, signIntegrationCommand("existing"))
	if err != nil {
		t.Fatalf("sign existing: %v", err)
	}
	if existing.Balance.CoinsText != "25.5" || existing.Balance.TodayText != "1" || existing.Reward != 25.5 || !existing.Calendar.HasDay("2026", "07", "4") {
		t.Fatalf("existing result = %#v", existing)
	}
	created, err := repository.SignIn(ctx, signIntegrationCommand("new-user"))
	if err != nil {
		t.Fatalf("sign new user: %v", err)
	}
	if created.Balance.CoinsText != "25.5" || created.Balance.Today != 101 || created.Reward != 25.5 || !created.Calendar.HasDay("2026", "07", "4") {
		t.Fatalf("created result = %#v", created)
	}
}

func TestEconomySignInMongoIntegrationLeavesDuplicateRowsObservable(t *testing.T) {
	database := economyQueryIntegrationDatabase(t)
	repository, err := NewEconomyRepositoryFromDatabase(database)
	if err != nil {
		t.Fatalf("new repository: %v", err)
	}
	ctx := context.Background()
	if _, err := database.Collection(CoinCollectionName).InsertMany(ctx, []any{
		bson.D{{Key: "guild", Value: "guild-1"}, {Key: "member", Value: "duplicate"}, {Key: "coin", Value: 10}, {Key: "today", Value: 0}},
		bson.D{{Key: "guild", Value: "guild-1"}, {Key: "member", Value: "duplicate"}, {Key: "coin", Value: 20}, {Key: "today", Value: 0}},
	}); err != nil {
		t.Fatalf("seed duplicate balances: %v", err)
	}
	if _, err := database.Collection(SignListCollectionName).InsertMany(ctx, []any{
		bson.D{{Key: "guild", Value: "guild-1"}, {Key: "member", Value: "duplicate"}, {Key: "date", Value: bson.D{}}},
		bson.D{{Key: "guild", Value: "guild-1"}, {Key: "member", Value: "duplicate"}, {Key: "date", Value: bson.D{}}},
	}); err != nil {
		t.Fatalf("seed duplicate calendars: %v", err)
	}

	if _, err := repository.SignIn(ctx, signIntegrationCommand("duplicate")); err != nil {
		t.Fatalf("sign duplicate rows: %v", err)
	}
	coinCount, err := database.Collection(CoinCollectionName).CountDocuments(ctx, bson.D{{Key: "guild", Value: "guild-1"}, {Key: "member", Value: "duplicate"}})
	if err != nil || coinCount != 2 {
		t.Fatalf("coin count = %d, err=%v", coinCount, err)
	}
	calendarCount, err := database.Collection(SignListCollectionName).CountDocuments(ctx, bson.D{{Key: "guild", Value: "guild-1"}, {Key: "member", Value: "duplicate"}})
	if err != nil || calendarCount != 2 {
		t.Fatalf("calendar count = %d, err=%v", calendarCount, err)
	}
	updatedCalendars, err := database.Collection(SignListCollectionName).CountDocuments(ctx, bson.D{{Key: "date.2026.07", Value: "4"}})
	if err != nil || updatedCalendars != 1 {
		t.Fatalf("updated calendars = %d, err=%v", updatedCalendars, err)
	}
}

func signIntegrationCommand(userID string) domain.SignInCommand {
	return domain.SignInCommand{
		GuildID: "guild-1", UserID: userID,
		Now: time.Unix(100, 600_000_000), Year: "2026", Month: "07", Day: "4",
	}
}
