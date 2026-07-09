package ports_test

import (
	"context"
	"errors"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakemongo"
)

func TestTicketConfigRepositoryContractWithFake(t *testing.T) {
	repo := fakemongo.NewTicketConfigRepository()
	var port ports.TicketConfigRepository = repo
	ctx := context.Background()

	config := domain.TicketConfig{
		GuildID:        "guild-1",
		CategoryID:     "category-1",
		AdminRoleID:    "role-1",
		EveryoneRoleID: "everyone-1",
	}
	if err := port.SaveTicketConfig(ctx, config); err != nil {
		t.Fatalf("save config: %v", err)
	}
	got, err := port.GetTicketConfig(ctx, config.GuildID)
	if err != nil {
		t.Fatalf("get config: %v", err)
	}
	if got != config {
		t.Fatalf("config = %#v, want %#v", got, config)
	}
	updated := config
	updated.CategoryID = "category-2"
	if err := port.SaveTicketConfig(ctx, updated); err != nil {
		t.Fatalf("save updated config: %v", err)
	}
	got, err = port.GetTicketConfig(ctx, config.GuildID)
	if err != nil {
		t.Fatalf("get updated config: %v", err)
	}
	if got != updated {
		t.Fatalf("updated config = %#v, want %#v", got, updated)
	}
	if err := port.DeleteTicketConfig(ctx, config.GuildID); err != nil {
		t.Fatalf("delete config: %v", err)
	}
	if _, err := port.GetTicketConfig(ctx, config.GuildID); !errors.Is(err, ports.ErrTicketConfigNotFound) {
		t.Fatalf("get deleted config error = %v", err)
	}
}

func TestTicketConfigRepositoryContractValidation(t *testing.T) {
	repo := fakemongo.NewTicketConfigRepository()
	err := repo.SaveTicketConfig(context.Background(), domain.TicketConfig{GuildID: "guild-1"})
	if !errors.Is(err, domain.ErrInvalidTicketConfig) {
		t.Fatalf("expected invalid config, got %v", err)
	}
}

func TestTicketConfigRepositoryContractContextCancellation(t *testing.T) {
	repo := fakemongo.NewTicketConfigRepository()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := repo.GetTicketConfig(ctx, "guild-1"); !errors.Is(err, context.Canceled) {
		t.Fatalf("get canceled error = %v", err)
	}
	if err := repo.SaveTicketConfig(ctx, domain.TicketConfig{}); !errors.Is(err, context.Canceled) {
		t.Fatalf("save canceled error = %v", err)
	}
	if err := repo.DeleteTicketConfig(ctx, "guild-1"); !errors.Is(err, context.Canceled) {
		t.Fatalf("delete canceled error = %v", err)
	}
}
