package safety

import (
	"context"
	"errors"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakemongo"
)

func TestToggleAntiScamCreatesOpenConfigWhenMissing(t *testing.T) {
	repo := fakemongo.NewAntiScamConfigRepository()
	service := NewConfigService(repo)
	config, err := service.Toggle(context.Background(), " guild-1 ")
	if err != nil {
		t.Fatalf("toggle anti-scam: %v", err)
	}
	if config.GuildID != "guild-1" || !config.Open {
		t.Fatalf("config = %#v", config)
	}
	saved, ok := repo.Last()
	if !ok || saved != config {
		t.Fatalf("saved = %#v ok=%v", saved, ok)
	}
}

func TestToggleAntiScamFlipsExistingConfig(t *testing.T) {
	repo := fakemongo.NewAntiScamConfigRepository()
	repo.Configs["guild-1"] = domain.AntiScamConfig{GuildID: "guild-1", Open: true}
	service := NewConfigService(repo)
	config, err := service.Toggle(context.Background(), "guild-1")
	if err != nil {
		t.Fatalf("toggle anti-scam: %v", err)
	}
	if config.Open {
		t.Fatalf("expected open=false, got %#v", config)
	}

	config, err = service.Toggle(context.Background(), "guild-1")
	if err != nil {
		t.Fatalf("toggle anti-scam again: %v", err)
	}
	if !config.Open {
		t.Fatalf("expected open=true, got %#v", config)
	}
}

func TestToggleAntiScamRejectsMissingRepository(t *testing.T) {
	service := NewConfigService(nil)
	_, err := service.Toggle(context.Background(), "guild-1")
	if !errors.Is(err, domain.ErrInvalidAntiScamConfig) {
		t.Fatalf("expected invalid anti-scam config, got %v", err)
	}
}

func TestToggleAntiScamRejectsMissingGuild(t *testing.T) {
	service := NewConfigService(fakemongo.NewAntiScamConfigRepository())
	_, err := service.Toggle(context.Background(), " ")
	if !errors.Is(err, domain.ErrInvalidAntiScamConfig) {
		t.Fatalf("expected invalid anti-scam config, got %v", err)
	}
}

func TestToggleAntiScamPropagatesRepositoryError(t *testing.T) {
	wantErr := errors.New("repo unavailable")
	repo := fakemongo.NewAntiScamConfigRepository()
	repo.Err = wantErr
	service := NewConfigService(repo)
	_, err := service.Toggle(context.Background(), "guild-1")
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected repo error, got %v", err)
	}
}
