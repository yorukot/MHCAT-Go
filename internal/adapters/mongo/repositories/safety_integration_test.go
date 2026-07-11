package repositories

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestSafetyMongoIntegrationConfigAndCatalogLifecycle(t *testing.T) {
	database := xpAccrualIntegrationDatabase(t)
	configs, err := NewAntiScamConfigRepositoryFromDatabase(database)
	if err != nil {
		t.Fatalf("new anti-scam config repository: %v", err)
	}
	catalog, err := NewScamURLCatalogRepositoryFromDatabase(database)
	if err != nil {
		t.Fatalf("new scam URL catalog repository: %v", err)
	}
	ctx := context.Background()

	if err := configs.SaveAntiScamConfig(ctx, domain.AntiScamConfig{GuildID: " guild-1 ", Open: true}); err != nil {
		t.Fatalf("create anti-scam config: %v", err)
	}
	config, err := configs.FindAntiScamConfig(ctx, " guild-1 ")
	if err != nil || config.GuildID != "guild-1" || !config.Open {
		t.Fatalf("created config = %#v err=%v", config, err)
	}
	if _, err := database.Collection(GoodWebConfigCollectionName).InsertOne(ctx, bson.D{
		{Key: "guild", Value: "guild-1"}, {Key: "open", Value: "yes"}, {Key: "dashboard", Value: "preserve"},
	}); err != nil {
		t.Fatalf("seed duplicate anti-scam config: %v", err)
	}
	if err := configs.SaveAntiScamConfig(ctx, domain.AntiScamConfig{GuildID: "guild-1", Open: false}); err != nil {
		t.Fatalf("align anti-scam configs: %v", err)
	}
	matched, err := database.Collection(GoodWebConfigCollectionName).CountDocuments(ctx, bson.D{
		{Key: "guild", Value: "guild-1"}, {Key: "open", Value: false},
	})
	if err != nil || matched != 2 {
		t.Fatalf("aligned anti-scam configs = %d err=%v", matched, err)
	}
	var duplicate bson.M
	if err := database.Collection(GoodWebConfigCollectionName).FindOne(ctx, bson.D{{Key: "dashboard", Value: "preserve"}}).Decode(&duplicate); err != nil {
		t.Fatalf("read preserved anti-scam config: %v", err)
	}
	if duplicate["dashboard"] != "preserve" {
		t.Fatalf("unrelated config field changed: %#v", duplicate)
	}
	if _, err := configs.FindAntiScamConfig(ctx, "missing"); !errors.Is(err, ports.ErrAntiScamConfigMissing) {
		t.Fatalf("missing anti-scam config error = %v", err)
	}
	if err := configs.SaveAntiScamConfig(ctx, domain.AntiScamConfig{}); !errors.Is(err, domain.ErrInvalidAntiScamConfig) {
		t.Fatalf("invalid anti-scam config error = %v", err)
	}

	metacharURL := "https://bad.example/a?x=(.*)"
	if _, err := database.Collection(NotAGoodWebCollectionName).InsertMany(ctx, []any{
		bson.D{{Key: "web", Value: "https://first.bad"}},
		bson.D{{Key: "web", Value: "prefix https://reported.bad/path suffix"}},
		bson.D{{Key: "web", Value: metacharURL}},
		bson.D{{Key: "web", Value: int64(12345)}},
		bson.D{{Key: "web", Value: true}},
		bson.D{{Key: "web", Value: ""}},
		bson.D{{Key: "web", Value: nil}},
		bson.D{{Key: "web", Value: bson.A{"https://compound.bad"}}},
	}); err != nil {
		t.Fatalf("seed scam URL catalog: %v", err)
	}
	contains, err := catalog.ContainsScamURL(ctx, "https://reported.bad/path")
	if err != nil || !contains {
		t.Fatalf("known scam URL contains=%t err=%v", contains, err)
	}
	contains, err = catalog.ContainsScamURL(ctx, metacharURL)
	if err != nil || !contains {
		t.Fatalf("metachar scam URL contains=%t err=%v", contains, err)
	}
	contains, err = catalog.ContainsScamURL(ctx, "")
	if err != nil || contains {
		t.Fatalf("empty scam URL contains=%t err=%v", contains, err)
	}
	contains, err = catalog.ContainsScamURL(ctx, "https://unknown.bad")
	if err != nil || contains {
		t.Fatalf("unknown scam URL contains=%t err=%v", contains, err)
	}

	webs, err := catalog.ListScamURLs(ctx)
	if err != nil {
		t.Fatalf("list scam URLs: %v", err)
	}
	wantWebs := []string{"https://first.bad", "prefix https://reported.bad/path suffix", metacharURL, "12345", "true"}
	if !slices.Equal(webs, wantWebs) {
		t.Fatalf("scam URL list = %#v, want %#v", webs, wantWebs)
	}
	matchedURL, found, err := catalog.FindScamURLInContent(ctx, "visit https://first.bad and avoid it")
	if err != nil || !found || matchedURL != "https://first.bad" {
		t.Fatalf("content match=%q found=%t err=%v", matchedURL, found, err)
	}
	matchedURL, found, err = catalog.FindScamURLInContent(ctx, "clean message")
	if err != nil || found || matchedURL != "" {
		t.Fatalf("clean content match=%q found=%t err=%v", matchedURL, found, err)
	}
}

func TestSafetyMongoIntegrationHonorsCanceledContext(t *testing.T) {
	database := xpAccrualIntegrationDatabase(t)
	configs, err := NewAntiScamConfigRepositoryFromDatabase(database)
	if err != nil {
		t.Fatalf("new anti-scam config repository: %v", err)
	}
	catalog, err := NewScamURLCatalogRepositoryFromDatabase(database)
	if err != nil {
		t.Fatalf("new scam URL catalog repository: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := configs.FindAntiScamConfig(ctx, "guild-1"); !errors.Is(err, context.Canceled) {
		t.Fatalf("config read cancellation error = %v", err)
	}
	if err := configs.SaveAntiScamConfig(ctx, domain.AntiScamConfig{GuildID: "guild-1"}); !errors.Is(err, context.Canceled) {
		t.Fatalf("config save cancellation error = %v", err)
	}
	if _, err := catalog.ContainsScamURL(ctx, "https://bad.example"); !errors.Is(err, context.Canceled) {
		t.Fatalf("contains cancellation error = %v", err)
	}
	if _, _, err := catalog.FindScamURLInContent(ctx, "https://bad.example"); !errors.Is(err, context.Canceled) {
		t.Fatalf("content cancellation error = %v", err)
	}
	if _, err := catalog.ListScamURLs(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("list cancellation error = %v", err)
	}
}
