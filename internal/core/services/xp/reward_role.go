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
	RolePort      ports.DiscordRolePort
}

type VoiceRewardRoleService struct {
	Repository    ports.VoiceXPRewardRoleRepository
	RoleInspector ports.DiscordRoleInspector
	RolePort      ports.DiscordRolePort
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

func (s TextRewardRoleService) ApplyLevelUp(ctx context.Context, guildID string, userID string, level int64, currentRoleIDs []string) error {
	if s.Repository == nil || s.RolePort == nil {
		return domain.ErrInvalidXPRewardRoleConfig
	}
	guildID = strings.TrimSpace(guildID)
	userID = strings.TrimSpace(userID)
	if guildID == "" || userID == "" {
		return domain.ErrInvalidXPRewardRoleConfig
	}
	configs, err := s.Repository.ListTextXPRewardRoles(ctx, guildID)
	if err != nil {
		return err
	}
	return applyXPRewardRoles(ctx, s.RolePort, configs, guildID, userID, level, currentRoleIDs)
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

func (s VoiceRewardRoleService) ApplyLevelUp(ctx context.Context, guildID string, userID string, level int64, currentRoleIDs []string) error {
	if s.Repository == nil || s.RolePort == nil {
		return domain.ErrInvalidXPRewardRoleConfig
	}
	guildID = strings.TrimSpace(guildID)
	userID = strings.TrimSpace(userID)
	if guildID == "" || userID == "" {
		return domain.ErrInvalidXPRewardRoleConfig
	}
	configs, err := s.Repository.ListVoiceXPRewardRoles(ctx, guildID)
	if err != nil {
		return err
	}
	return applyXPRewardRoles(ctx, s.RolePort, configs, guildID, userID, level, currentRoleIDs)
}

func applyXPRewardRoles(ctx context.Context, roles ports.DiscordRolePort, configs []domain.XPRewardRoleConfig, guildID string, userID string, level int64, currentRoleIDs []string) error {
	currentRoles := map[string]struct{}{}
	for _, roleID := range currentRoleIDs {
		roleID = strings.TrimSpace(roleID)
		if roleID != "" {
			currentRoles[roleID] = struct{}{}
		}
	}
	for _, config := range configs {
		config = config.Normalize()
		if !config.DeleteWhenNot || config.RoleID == "" {
			continue
		}
		if _, ok := currentRoles[config.RoleID]; !ok {
			continue
		}
		if err := roles.RemoveRole(ctx, guildID, userID, config.RoleID); err != nil {
			return err
		}
	}
	for _, config := range configs {
		config = config.Normalize()
		if config.RoleID == "" || config.Level != level {
			continue
		}
		if err := roles.AddRole(ctx, guildID, userID, config.RoleID); err != nil {
			return err
		}
	}
	return ctx.Err()
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
