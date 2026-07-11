package repositories

import (
	"context"
	"reflect"
	"testing"

	coreeconomy "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/services/economy"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestEconomyCoinRankMongoIntegrationPreservesOrderDuplicatesAndScalars(t *testing.T) {
	database := economyQueryIntegrationDatabase(t)
	repository, err := NewEconomyRepositoryFromDatabase(database)
	if err != nil {
		t.Fatalf("new repository: %v", err)
	}
	if _, err := database.Collection(CoinCollectionName).InsertMany(context.Background(), []any{
		bson.D{{Key: "guild", Value: "guild-1"}, {Key: "member", Value: "viewer"}, {Key: "coin", Value: int64(10)}},
		bson.D{{Key: "guild", Value: "guild-1"}, {Key: "member", Value: "decimal"}, {Key: "coin", Value: 20.5}},
		bson.D{{Key: "guild", Value: "guild-1"}, {Key: "member", Value: "malformed"}, {Key: "coin", Value: bson.D{{Key: "bad", Value: true}}}},
		bson.D{{Key: "guild", Value: "guild-1"}, {Key: "member", Value: "viewer"}, {Key: "coin", Value: int64(30)}},
		bson.D{{Key: "guild", Value: "other"}, {Key: "member", Value: "excluded"}, {Key: "coin", Value: int64(999)}},
	}); err != nil {
		t.Fatalf("seed coin ranks: %v", err)
	}

	balances, err := repository.ListCoinBalances(context.Background(), "guild-1")
	if err != nil {
		t.Fatalf("list balances: %v", err)
	}
	got := make([]string, 0, len(balances))
	for _, balance := range balances {
		got = append(got, balance.UserID+":"+balance.CoinsText)
	}
	if want := []string{"viewer:10", "decimal:20.5", "malformed:undefined", "viewer:30"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("balances = %#v want %#v", got, want)
	}

	page, err := (coreeconomy.CoinRankService{Repository: repository}).Query(context.Background(), coreeconomy.CoinRankQuery{
		GuildID: "guild-1", ViewerID: "viewer", Page: 0,
	})
	if err != nil {
		t.Fatalf("query rank: %v", err)
	}
	got = got[:0]
	for _, entry := range page.Entries {
		got = append(got, entry.Balance.UserID+":"+entry.Balance.CoinsText)
	}
	if want := []string{"viewer:30", "malformed:undefined", "decimal:20.5", "viewer:10"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("rank entries = %#v want %#v", got, want)
	}
	if page.ViewerRank != 4 || !page.ViewerHasBalance {
		t.Fatalf("viewer rank = %d has=%t", page.ViewerRank, page.ViewerHasBalance)
	}
}
