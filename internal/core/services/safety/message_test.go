package safety

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakemongo"
)

func TestMessageServiceDetectsScamURLWhenEnabled(t *testing.T) {
	configs := fakemongo.NewAntiScamConfigRepository()
	configs.Configs["guild-1"] = domain.AntiScamConfig{GuildID: "guild-1", Open: true}
	catalog := fakemongo.NewScamURLCatalogRepository()
	catalog.Known = []string{"https://bad.example"}
	service := NewMessageService(configs, catalog)

	result, err := service.Scan(context.Background(), " guild-1 ", "please visit https://bad.example now")
	if err != nil {
		t.Fatalf("scan message: %v", err)
	}
	if !result.Delete || result.MatchedURL != "https://bad.example" {
		t.Fatalf("result = %#v", result)
	}
}

func TestMessageServiceIgnoresClosedMissingAndCleanMessages(t *testing.T) {
	for name, setup := range map[string]func(*fakemongo.AntiScamConfigRepository, *fakemongo.ScamURLCatalogRepository){
		"closed": func(configs *fakemongo.AntiScamConfigRepository, catalog *fakemongo.ScamURLCatalogRepository) {
			configs.Configs["guild-1"] = domain.AntiScamConfig{GuildID: "guild-1", Open: false}
			catalog.Known = []string{"https://bad.example"}
		},
		"missing": func(configs *fakemongo.AntiScamConfigRepository, catalog *fakemongo.ScamURLCatalogRepository) {
			catalog.Known = []string{"https://bad.example"}
		},
		"clean": func(configs *fakemongo.AntiScamConfigRepository, catalog *fakemongo.ScamURLCatalogRepository) {
			configs.Configs["guild-1"] = domain.AntiScamConfig{GuildID: "guild-1", Open: true}
			catalog.Known = []string{"https://bad.example"}
		},
	} {
		t.Run(name, func(t *testing.T) {
			configs := fakemongo.NewAntiScamConfigRepository()
			catalog := fakemongo.NewScamURLCatalogRepository()
			setup(configs, catalog)
			content := "hello"
			if name != "clean" {
				content = "https://bad.example"
			}
			result, err := NewMessageService(configs, catalog).Scan(context.Background(), "guild-1", content)
			if err != nil {
				t.Fatalf("scan message: %v", err)
			}
			if result.Delete {
				t.Fatalf("unexpected delete result = %#v", result)
			}
		})
	}
}

func TestMessageServicePreservesRawCatalogAndMessageWhitespace(t *testing.T) {
	configs := fakemongo.NewAntiScamConfigRepository()
	configs.Configs["guild-1"] = domain.AntiScamConfig{GuildID: "guild-1", Open: true}
	catalog := fakemongo.NewScamURLCatalogRepository()
	catalog.Known = []string{" https://bad.example "}
	service := NewMessageService(configs, catalog)

	result, err := service.Scan(context.Background(), "guild-1", "visit https://bad.example!")
	if err != nil {
		t.Fatalf("scan message: %v", err)
	}
	if result.Delete {
		t.Fatalf("stored whitespace was normalized: %#v", result)
	}
	if catalog.ListCalls != 1 || len(catalog.Checked) != 0 {
		t.Fatalf("list calls=%d checked=%#v", catalog.ListCalls, catalog.Checked)
	}
}

func TestMessageServiceCachesHotPathMongoReads(t *testing.T) {
	configs := fakemongo.NewAntiScamConfigRepository()
	configs.Configs["guild-1"] = domain.AntiScamConfig{GuildID: "guild-1", Open: true}
	catalog := fakemongo.NewScamURLCatalogRepository()
	catalog.Known = []string{"https://bad.example"}
	service := NewMessageService(configs, catalog)

	for _, content := range []string{"first clean message", "second clean message"} {
		if _, err := service.Scan(context.Background(), "guild-1", content); err != nil {
			t.Fatalf("scan %q: %v", content, err)
		}
	}
	if configs.FindCalls != 1 || catalog.ListCalls != 1 {
		t.Fatalf("config reads=%d catalog reads=%d", configs.FindCalls, catalog.ListCalls)
	}
}

func TestMessageServiceRefreshesExpiredCaches(t *testing.T) {
	configs := fakemongo.NewAntiScamConfigRepository()
	configs.Configs["guild-1"] = domain.AntiScamConfig{GuildID: "guild-1", Open: true}
	catalog := fakemongo.NewScamURLCatalogRepository()
	catalog.Known = []string{"https://bad.example"}
	now := time.Date(2026, 7, 11, 0, 0, 0, 0, time.UTC)
	service := newMessageService(configs, catalog, func() time.Time { return now })

	if _, err := service.Scan(context.Background(), "guild-1", "clean"); err != nil {
		t.Fatalf("initial scan: %v", err)
	}
	now = now.Add(scamURLCatalogCacheTTL + time.Second)
	if _, err := service.Scan(context.Background(), "guild-1", "still clean"); err != nil {
		t.Fatalf("refresh scan: %v", err)
	}
	if configs.FindCalls != 2 || catalog.ListCalls != 2 {
		t.Fatalf("config reads=%d catalog reads=%d", configs.FindCalls, catalog.ListCalls)
	}
}

func TestMessageServiceBacksOffFailedCatalogRefresh(t *testing.T) {
	configs := fakemongo.NewAntiScamConfigRepository()
	configs.Configs["guild-1"] = domain.AntiScamConfig{GuildID: "guild-1", Open: true}
	catalog := fakemongo.NewScamURLCatalogRepository()
	catalog.Known = []string{"https://bad.example"}
	now := time.Date(2026, 7, 11, 0, 0, 0, 0, time.UTC)
	service := newMessageService(configs, catalog, func() time.Time { return now })

	if _, err := service.Scan(context.Background(), "guild-1", "clean"); err != nil {
		t.Fatalf("initial scan: %v", err)
	}
	now = now.Add(scamURLCatalogCacheTTL + time.Second)
	catalog.Err = errors.New("catalog unavailable")
	for range 2 {
		result, err := service.Scan(context.Background(), "guild-1", "https://bad.example")
		if err != nil || !result.Delete {
			t.Fatalf("stale scan result=%#v err=%v", result, err)
		}
	}
	if configs.FindCalls != 2 || catalog.ListCalls != 2 {
		t.Fatalf("config reads=%d catalog reads=%d", configs.FindCalls, catalog.ListCalls)
	}
}

func TestMessageServiceRejectsMissingPortsAndGuild(t *testing.T) {
	_, err := (MessageService{}).Scan(context.Background(), "guild-1", "https://bad.example")
	if !errors.Is(err, domain.ErrInvalidAntiScamConfig) {
		t.Fatalf("expected invalid config, got %v", err)
	}
	_, err = NewMessageService(fakemongo.NewAntiScamConfigRepository(), fakemongo.NewScamURLCatalogRepository()).Scan(context.Background(), "", "https://bad.example")
	if !errors.Is(err, domain.ErrInvalidAntiScamConfig) {
		t.Fatalf("expected invalid guild, got %v", err)
	}
}
