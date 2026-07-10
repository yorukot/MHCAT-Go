package safety

import (
	"context"
	"errors"
	"testing"

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
	if len(catalog.Checked) != 1 || catalog.Checked[0] != "visit https://bad.example!" {
		t.Fatalf("checked content = %#v", catalog.Checked)
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
