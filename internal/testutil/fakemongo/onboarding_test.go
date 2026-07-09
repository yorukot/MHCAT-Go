package fakemongo

import (
	"context"
	"errors"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
)

func TestJoinRoleConfigRepositoryFake(t *testing.T) {
	repo := NewJoinRoleConfigRepository()
	config := domain.JoinRoleConfig{GuildID: "guild", RoleID: "role", GiveTo: domain.JoinRoleGiveAllUsers}
	if err := repo.CreateJoinRoleConfig(context.Background(), config); err != nil {
		t.Fatalf("create: %v", err)
	}
	if err := repo.CreateJoinRoleConfig(context.Background(), config); !errors.Is(err, ports.ErrJoinRoleConfigExists) {
		t.Fatalf("expected duplicate, got %v", err)
	}
	if err := repo.DeleteJoinRoleConfig(context.Background(), "guild", "role"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if err := repo.DeleteJoinRoleConfig(context.Background(), "guild", "role"); !errors.Is(err, ports.ErrJoinRoleConfigMissing) {
		t.Fatalf("expected missing, got %v", err)
	}
}
