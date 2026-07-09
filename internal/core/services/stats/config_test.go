package stats

import (
	"context"
	"errors"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakemongo"
)

func TestConfigServiceDeletesStatsConfig(t *testing.T) {
	repo := fakemongo.NewStatsConfigRepository()
	repo.Put(domain.StatsConfig{GuildID: "guild-1", ParentID: "parent-1"})

	config, err := (ConfigService{Repository: repo}).Delete(context.Background(), " guild-1 ")
	if err != nil {
		t.Fatalf("delete stats config: %v", err)
	}
	if config.GuildID != "guild-1" || config.ParentID != "parent-1" {
		t.Fatalf("config = %#v", config)
	}
	if _, ok := repo.Configs["guild-1"]; ok {
		t.Fatal("stats config should be deleted")
	}
}

func TestConfigServiceMissingStatsConfig(t *testing.T) {
	_, err := (ConfigService{Repository: fakemongo.NewStatsConfigRepository()}).Delete(context.Background(), "guild-1")
	if !errors.Is(err, ports.ErrStatsConfigMissing) {
		t.Fatalf("expected ErrStatsConfigMissing, got %v", err)
	}
}

func TestConfigServiceRejectsInvalidRequest(t *testing.T) {
	_, err := (ConfigService{Repository: fakemongo.NewStatsConfigRepository()}).Delete(context.Background(), "")
	if !errors.Is(err, domain.ErrInvalidStatsConfigRequest) {
		t.Fatalf("expected invalid request, got %v", err)
	}
	_, err = (ConfigService{}).Delete(context.Background(), "guild-1")
	if !errors.Is(err, domain.ErrInvalidStatsConfigRequest) {
		t.Fatalf("expected invalid request for nil repo, got %v", err)
	}
}

func TestConfigServicePropagatesRepositoryErrors(t *testing.T) {
	repo := fakemongo.NewStatsConfigRepository()
	repo.Err = ports.ErrCoinLimitExceeded

	_, err := (ConfigService{Repository: repo}).Delete(context.Background(), "guild-1")
	if !errors.Is(err, ports.ErrCoinLimitExceeded) {
		t.Fatalf("expected repo error, got %v", err)
	}
}
