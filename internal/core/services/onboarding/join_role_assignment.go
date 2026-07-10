package onboarding

import (
	"context"
	"errors"
	"strings"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
)

type JoinRoleAssignmentService struct {
	Repository ports.JoinRoleConfigReader
	Roles      ports.DiscordRolePort
}

func (s JoinRoleAssignmentService) AssignOnJoin(ctx context.Context, guildID string, userID string, isBot bool) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if s.Repository == nil || s.Roles == nil {
		return domain.ErrInvalidJoinRoleConfig
	}
	guildID = strings.TrimSpace(guildID)
	userID = strings.TrimSpace(userID)
	if guildID == "" || userID == "" {
		return domain.ErrInvalidJoinRoleConfig
	}
	configs, err := s.Repository.ListJoinRoleConfigs(ctx, guildID)
	if err != nil {
		return err
	}
	var errs []error
	for _, config := range configs {
		config.GuildID = strings.TrimSpace(config.GuildID)
		config.RoleID = strings.TrimSpace(config.RoleID)
		config.GiveTo = strings.TrimSpace(config.GiveTo)
		if config.GiveTo == "" {
			config.GiveTo = domain.JoinRoleGiveAllUsers
		}
		if err := config.Validate(); err != nil {
			continue
		}
		if !joinRoleAppliesToMember(config.GiveTo, isBot) {
			continue
		}
		if err := s.Roles.AddRole(ctx, guildID, userID, config.RoleID); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

func joinRoleAppliesToMember(giveTo string, isBot bool) bool {
	switch domain.NormalizeJoinRoleGiveTo(giveTo) {
	case domain.JoinRoleGiveBots:
		return isBot
	case domain.JoinRoleGiveMembers:
		return !isBot
	default:
		return true
	}
}
