package xp

import (
	"context"
	"errors"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakediscord"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakemongo"
)

func TestTextRewardRoleServiceAddsAndChecksAssignableRole(t *testing.T) {
	repo := fakemongo.NewTextXPRewardRoleRepository()
	roles := fakediscord.NewSideEffects()
	roles.AssignableRoles["guild-1/role-1"] = true
	service := TextRewardRoleService{Repository: repo, RoleInspector: roles}

	err := service.Add(context.Background(), domain.XPRewardRoleConfig{
		GuildID:       " guild-1 ",
		Level:         12,
		RoleID:        " role-1 ",
		DeleteWhenNot: true,
	})
	if err != nil {
		t.Fatalf("add: %v", err)
	}
	if len(repo.Configs) != 1 || repo.Configs[0].GuildID != "guild-1" || repo.Configs[0].RoleID != "role-1" || !repo.Configs[0].DeleteWhenNot {
		t.Fatalf("saved config = %#v", repo.Configs)
	}
}

func TestTextRewardRoleServiceRejectsUnassignableAndTooMany(t *testing.T) {
	roles := fakediscord.NewSideEffects()
	service := TextRewardRoleService{Repository: fakemongo.NewTextXPRewardRoleRepository(), RoleInspector: roles}
	err := service.Add(context.Background(), domain.XPRewardRoleConfig{GuildID: "guild-1", Level: 1, RoleID: "role-1"})
	if !errors.Is(err, ports.ErrDiscordRoleNotAssignable) {
		t.Fatalf("expected role not assignable, got %v", err)
	}

	repo := fakemongo.NewTextXPRewardRoleRepository()
	for i := 0; i < 120; i++ {
		repo.Configs = append(repo.Configs, domain.XPRewardRoleConfig{GuildID: "guild-1", Level: int64(i), RoleID: "role"})
	}
	service = TextRewardRoleService{Repository: repo}
	err = service.Add(context.Background(), domain.XPRewardRoleConfig{GuildID: "guild-1", Level: 121, RoleID: "role-121"})
	if !errors.Is(err, ports.ErrXPRewardRoleLimitExceeded) {
		t.Fatalf("expected limit error, got %v", err)
	}
}

func TestRewardRoleServicesCheckLegacyLimitBeforeRoleHierarchy(t *testing.T) {
	roles := fakediscord.NewSideEffects()
	textRepo := fakemongo.NewTextXPRewardRoleRepository()
	voiceRepo := fakemongo.NewVoiceXPRewardRoleRepository()
	for i := 0; i < 120; i++ {
		config := domain.XPRewardRoleConfig{GuildID: "guild-1", Level: int64(i), RoleID: "role"}
		textRepo.Configs = append(textRepo.Configs, config)
		voiceRepo.Configs = append(voiceRepo.Configs, config)
	}

	for _, tc := range []struct {
		name string
		add  func() error
	}{
		{
			name: "text",
			add: func() error {
				return (TextRewardRoleService{Repository: textRepo, RoleInspector: roles}).Add(context.Background(), domain.XPRewardRoleConfig{GuildID: "guild-1", Level: 120, RoleID: "unassignable"})
			},
		},
		{
			name: "voice",
			add: func() error {
				return (VoiceRewardRoleService{Repository: voiceRepo, RoleInspector: roles}).Add(context.Background(), domain.XPRewardRoleConfig{GuildID: "guild-1", Level: 120, RoleID: "unassignable"})
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.add(); !errors.Is(err, ports.ErrXPRewardRoleLimitExceeded) {
				t.Fatalf("expected legacy limit error before role hierarchy, got %v", err)
			}
		})
	}
}

func TestRewardRoleServicesAllowLegacy120thConfig(t *testing.T) {
	t.Run("text", func(t *testing.T) {
		repo := fakemongo.NewTextXPRewardRoleRepository()
		for i := 0; i < 119; i++ {
			repo.Configs = append(repo.Configs, domain.XPRewardRoleConfig{GuildID: "guild-1", Level: int64(i), RoleID: "existing-role"})
		}
		roles := fakediscord.NewSideEffects()
		roles.AssignableRoles["guild-1/role-120"] = true
		service := TextRewardRoleService{Repository: repo, RoleInspector: roles}
		if err := service.Add(context.Background(), domain.XPRewardRoleConfig{GuildID: "guild-1", Level: 120, RoleID: "role-120"}); err != nil {
			t.Fatalf("add 120th config: %v", err)
		}
		if len(repo.Configs) != 120 {
			t.Fatalf("config count = %d", len(repo.Configs))
		}
		if err := service.Add(context.Background(), domain.XPRewardRoleConfig{GuildID: "guild-1", Level: 121, RoleID: "role-121"}); !errors.Is(err, ports.ErrXPRewardRoleLimitExceeded) {
			t.Fatalf("add 121st config error = %v", err)
		}
	})

	t.Run("voice", func(t *testing.T) {
		repo := fakemongo.NewVoiceXPRewardRoleRepository()
		for i := 0; i < 119; i++ {
			repo.Configs = append(repo.Configs, domain.XPRewardRoleConfig{GuildID: "guild-1", Level: int64(i), RoleID: "existing-role"})
		}
		roles := fakediscord.NewSideEffects()
		roles.AssignableRoles["guild-1/role-120"] = true
		service := VoiceRewardRoleService{Repository: repo, RoleInspector: roles}
		if err := service.Add(context.Background(), domain.XPRewardRoleConfig{GuildID: "guild-1", Level: 120, RoleID: "role-120"}); err != nil {
			t.Fatalf("add 120th config: %v", err)
		}
		if len(repo.Configs) != 120 {
			t.Fatalf("config count = %d", len(repo.Configs))
		}
		if err := service.Add(context.Background(), domain.XPRewardRoleConfig{GuildID: "guild-1", Level: 121, RoleID: "role-121"}); !errors.Is(err, ports.ErrXPRewardRoleLimitExceeded) {
			t.Fatalf("add 121st config error = %v", err)
		}
	})
}

func TestRewardRoleServicesMapMissingRolesToHierarchyError(t *testing.T) {
	roles := fakediscord.NewSideEffects()
	roles.MissingRoles["guild-1/missing-role"] = true
	config := domain.XPRewardRoleConfig{GuildID: "guild-1", Level: 1, RoleID: "missing-role"}

	for _, tc := range []struct {
		name string
		add  func() error
	}{
		{
			name: "text",
			add: func() error {
				return (TextRewardRoleService{Repository: fakemongo.NewTextXPRewardRoleRepository(), RoleInspector: roles}).Add(context.Background(), config)
			},
		},
		{
			name: "voice",
			add: func() error {
				return (VoiceRewardRoleService{Repository: fakemongo.NewVoiceXPRewardRoleRepository(), RoleInspector: roles}).Add(context.Background(), config)
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.add(); !errors.Is(err, ports.ErrDiscordRoleNotAssignable) {
				t.Fatalf("missing role error = %v", err)
			}
		})
	}
}

func TestTextRewardRoleServiceApplyLevelUpChangesConfiguredRoles(t *testing.T) {
	repo := fakemongo.NewTextXPRewardRoleRepository()
	repo.Configs = []domain.XPRewardRoleConfig{
		{GuildID: "guild-1", Level: 0, RoleID: "old-role", DeleteWhenNot: true},
		{GuildID: "guild-1", Level: 1, RoleID: "new-role"},
		{GuildID: "guild-1", Level: 1, RoleID: "kept-role", DeleteWhenNot: false},
		{GuildID: "other", Level: 1, RoleID: "other-role", DeleteWhenNot: true},
	}
	roles := fakediscord.NewSideEffects()
	service := TextRewardRoleService{Repository: repo, RolePort: roles}

	if err := service.ApplyLevelUp(context.Background(), " guild-1 ", " user-1 ", 1, []string{"old-role", "kept-role"}); err != nil {
		t.Fatalf("apply level up: %v", err)
	}
	if len(roles.RemovedRoles) != 1 || roles.RemovedRoles[0].RoleID != "old-role" {
		t.Fatalf("removed roles = %#v", roles.RemovedRoles)
	}
	if len(roles.AddedRoles) != 2 || roles.AddedRoles[0].RoleID != "new-role" || roles.AddedRoles[1].RoleID != "kept-role" {
		t.Fatalf("added roles = %#v", roles.AddedRoles)
	}
}

func TestVoiceRewardRoleServiceListAndDelete(t *testing.T) {
	repo := fakemongo.NewVoiceXPRewardRoleRepository()
	repo.Configs = []domain.XPRewardRoleConfig{
		{GuildID: "guild-1", Level: 2, RoleID: "role-2"},
		{GuildID: "other", Level: 3, RoleID: "role-3"},
	}
	service := VoiceRewardRoleService{Repository: repo}

	configs, err := service.List(context.Background(), "guild-1")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(configs) != 1 || configs[0].RoleID != "role-2" {
		t.Fatalf("configs = %#v", configs)
	}
	if err := service.Delete(context.Background(), "guild-1", 2, "role-2"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if err := service.Delete(context.Background(), "guild-1", 2, "role-2"); !errors.Is(err, ports.ErrVoiceXPRewardRoleMissing) {
		t.Fatalf("expected missing, got %v", err)
	}
}

func TestVoiceRewardRoleServiceApplyLevelUpChangesConfiguredRoles(t *testing.T) {
	repo := fakemongo.NewVoiceXPRewardRoleRepository()
	repo.Configs = []domain.XPRewardRoleConfig{
		{GuildID: "guild-1", Level: 0, RoleID: "old-role", DeleteWhenNot: true},
		{GuildID: "guild-1", Level: 1, RoleID: "new-role"},
		{GuildID: "guild-1", Level: 1, RoleID: "kept-role", DeleteWhenNot: false},
		{GuildID: "other", Level: 1, RoleID: "other-role", DeleteWhenNot: true},
	}
	roles := fakediscord.NewSideEffects()
	service := VoiceRewardRoleService{Repository: repo, RolePort: roles}

	if err := service.ApplyLevelUp(context.Background(), " guild-1 ", " user-1 ", 1, []string{"old-role", "kept-role"}); err != nil {
		t.Fatalf("apply level up: %v", err)
	}
	if len(roles.RemovedRoles) != 1 || roles.RemovedRoles[0].RoleID != "old-role" {
		t.Fatalf("removed roles = %#v", roles.RemovedRoles)
	}
	if len(roles.AddedRoles) != 2 || roles.AddedRoles[0].RoleID != "new-role" || roles.AddedRoles[1].RoleID != "kept-role" {
		t.Fatalf("added roles = %#v", roles.AddedRoles)
	}
}
