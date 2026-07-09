package onboarding

import (
	"context"
	"errors"
	"strings"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
)

type JoinRoleService struct {
	Repository    ports.JoinRoleConfigRepository
	RoleInspector ports.DiscordRoleInspector
}

func (s JoinRoleService) Create(ctx context.Context, config domain.JoinRoleConfig) error {
	if s.Repository == nil {
		return domain.ErrInvalidJoinRoleConfig
	}
	config = normalizeJoinRoleConfig(config)
	if err := config.Validate(); err != nil {
		return err
	}
	if s.RoleInspector != nil {
		ok, err := s.RoleInspector.CanAssignRole(ctx, config.GuildID, config.RoleID)
		if err != nil {
			if errors.Is(err, ports.ErrDiscordRoleMissing) {
				return ports.ErrDiscordRoleNotAssignable
			}
			return err
		}
		if !ok {
			return ports.ErrDiscordRoleNotAssignable
		}
	}
	return s.Repository.CreateJoinRoleConfig(ctx, config)
}

func (s JoinRoleService) Delete(ctx context.Context, guildID string, roleID string) error {
	if s.Repository == nil {
		return domain.ErrInvalidJoinRoleConfig
	}
	guildID = strings.TrimSpace(guildID)
	roleID = strings.TrimSpace(roleID)
	if guildID == "" || roleID == "" {
		return domain.ErrInvalidJoinRoleConfig
	}
	return s.Repository.DeleteJoinRoleConfig(ctx, guildID, roleID)
}

func normalizeJoinRoleConfig(config domain.JoinRoleConfig) domain.JoinRoleConfig {
	config.GuildID = strings.TrimSpace(config.GuildID)
	config.RoleID = strings.TrimSpace(config.RoleID)
	config.GiveTo = domain.NormalizeJoinRoleGiveTo(config.GiveTo)
	return config
}
