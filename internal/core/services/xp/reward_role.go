package xp

import (
	"context"
	"errors"
	"strings"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
)

const legacyRewardRoleLimit = 119

type TextRewardRoleService struct {
	Repository    ports.TextXPRewardRoleRepository
	RoleInspector ports.DiscordRoleInspector
}

type VoiceRewardRoleService struct {
	Repository    ports.VoiceXPRewardRoleRepository
	RoleInspector ports.DiscordRoleInspector
}

func (s TextRewardRoleService) Add(ctx context.Context, config domain.XPRewardRoleConfig) error {
	if s.Repository == nil {
		return domain.ErrInvalidXPRewardRoleConfig
	}
	config = config.Normalize()
	if err := config.Validate(); err != nil {
		return err
	}
	if err := ensureAssignableRewardRole(ctx, s.RoleInspector, config); err != nil {
		return err
	}
	existing, err := s.Repository.ListTextXPRewardRoles(ctx, config.GuildID)
	if err != nil {
		return err
	}
	if len(existing) > legacyRewardRoleLimit {
		return ports.ErrXPRewardRoleLimitExceeded
	}
	return s.Repository.SaveTextXPRewardRole(ctx, config)
}

func (s TextRewardRoleService) Delete(ctx context.Context, guildID string, level int64, roleID string) error {
	if s.Repository == nil {
		return domain.ErrInvalidXPRewardRoleConfig
	}
	guildID = strings.TrimSpace(guildID)
	roleID = strings.TrimSpace(roleID)
	if guildID == "" || roleID == "" {
		return domain.ErrInvalidXPRewardRoleConfig
	}
	return s.Repository.DeleteTextXPRewardRole(ctx, guildID, level, roleID)
}

func (s TextRewardRoleService) List(ctx context.Context, guildID string) ([]domain.XPRewardRoleConfig, error) {
	if s.Repository == nil {
		return nil, domain.ErrInvalidXPRewardRoleConfig
	}
	guildID = strings.TrimSpace(guildID)
	if guildID == "" {
		return nil, domain.ErrInvalidXPRewardRoleConfig
	}
	return s.Repository.ListTextXPRewardRoles(ctx, guildID)
}

func (s VoiceRewardRoleService) Add(ctx context.Context, config domain.XPRewardRoleConfig) error {
	if s.Repository == nil {
		return domain.ErrInvalidXPRewardRoleConfig
	}
	config = config.Normalize()
	if err := config.Validate(); err != nil {
		return err
	}
	if err := ensureAssignableRewardRole(ctx, s.RoleInspector, config); err != nil {
		return err
	}
	existing, err := s.Repository.ListVoiceXPRewardRoles(ctx, config.GuildID)
	if err != nil {
		return err
	}
	if len(existing) > legacyRewardRoleLimit {
		return ports.ErrXPRewardRoleLimitExceeded
	}
	return s.Repository.SaveVoiceXPRewardRole(ctx, config)
}

func (s VoiceRewardRoleService) Delete(ctx context.Context, guildID string, level int64, roleID string) error {
	if s.Repository == nil {
		return domain.ErrInvalidXPRewardRoleConfig
	}
	guildID = strings.TrimSpace(guildID)
	roleID = strings.TrimSpace(roleID)
	if guildID == "" || roleID == "" {
		return domain.ErrInvalidXPRewardRoleConfig
	}
	return s.Repository.DeleteVoiceXPRewardRole(ctx, guildID, level, roleID)
}

func (s VoiceRewardRoleService) List(ctx context.Context, guildID string) ([]domain.XPRewardRoleConfig, error) {
	if s.Repository == nil {
		return nil, domain.ErrInvalidXPRewardRoleConfig
	}
	guildID = strings.TrimSpace(guildID)
	if guildID == "" {
		return nil, domain.ErrInvalidXPRewardRoleConfig
	}
	return s.Repository.ListVoiceXPRewardRoles(ctx, guildID)
}

func ensureAssignableRewardRole(ctx context.Context, inspector ports.DiscordRoleInspector, config domain.XPRewardRoleConfig) error {
	if inspector == nil {
		return nil
	}
	ok, err := inspector.CanAssignRole(ctx, config.GuildID, config.RoleID)
	if err != nil {
		if errors.Is(err, ports.ErrDiscordRoleMissing) {
			return ports.ErrDiscordRoleNotAssignable
		}
		return err
	}
	if !ok {
		return ports.ErrDiscordRoleNotAssignable
	}
	return nil
}
