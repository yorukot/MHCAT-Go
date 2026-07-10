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
	creation, err := port.CreateTicketConfig(ctx, config)
	if err != nil {
		t.Fatalf("create config: %v", err)
	}
	if creation.GuildID != config.GuildID || creation.ID == "" {
		t.Fatalf("creation = %#v", creation)
	}
	got, err := port.GetTicketConfig(ctx, config.GuildID)
	if err != nil {
		t.Fatalf("get config: %v", err)
	}
	if got != config {
		t.Fatalf("config = %#v, want %#v", got, config)
	}
	conflict := config
	conflict.CategoryID = "category-2"
	if _, err := port.CreateTicketConfig(ctx, conflict); !errors.Is(err, ports.ErrTicketConfigExists) {
		t.Fatalf("create duplicate config error = %v", err)
	}
	got, err = port.GetTicketConfig(ctx, config.GuildID)
	if err != nil {
		t.Fatalf("get original config: %v", err)
	}
	if got != config {
		t.Fatalf("duplicate create changed config = %#v, want %#v", got, config)
	}
	if err := port.DeleteTicketConfig(ctx, config.GuildID); err != nil {
		t.Fatalf("delete config: %v", err)
	}
	replacement, err := port.CreateTicketConfig(ctx, conflict)
	if err != nil {
		t.Fatalf("create replacement config: %v", err)
	}
	if err := port.RollbackTicketConfigCreation(ctx, creation); err != nil {
		t.Fatalf("rollback stale creation: %v", err)
	}
	got, err = port.GetTicketConfig(ctx, config.GuildID)
	if err != nil {
		t.Fatalf("get replacement after stale rollback: %v", err)
	}
	if got != conflict {
		t.Fatalf("stale rollback changed replacement = %#v, want %#v", got, conflict)
	}
	if err := port.RollbackTicketConfigCreation(ctx, replacement); err != nil {
		t.Fatalf("rollback replacement creation: %v", err)
	}
	if _, err := port.GetTicketConfig(ctx, config.GuildID); !errors.Is(err, ports.ErrTicketConfigNotFound) {
		t.Fatalf("get rolled back config error = %v", err)
	}
}

func TestTicketConfigRepositoryContractValidation(t *testing.T) {
	repo := fakemongo.NewTicketConfigRepository()
	_, err := repo.CreateTicketConfig(context.Background(), domain.TicketConfig{GuildID: "guild-1"})
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
	if _, err := repo.CreateTicketConfig(ctx, domain.TicketConfig{}); !errors.Is(err, context.Canceled) {
		t.Fatalf("create canceled error = %v", err)
	}
	if err := repo.RollbackTicketConfigCreation(ctx, ports.TicketConfigCreation{}); !errors.Is(err, context.Canceled) {
		t.Fatalf("rollback canceled error = %v", err)
	}
	if err := repo.DeleteTicketConfig(ctx, "guild-1"); !errors.Is(err, context.Canceled) {
		t.Fatalf("delete canceled error = %v", err)
	}
}
