package onboarding

import (
	"context"
	"errors"
	"strings"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
)

type JoinRoleAssignmentService struct {
	Repository     ports.JoinRoleConfigReader
	Roles          ports.DiscordRolePort
	RoleInspector  ports.DiscordRoleInspector
	Guilds         ports.DiscordInfoProvider
	DirectMessages ports.DiscordDirectMessagePort
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
		if s.RoleInspector != nil {
			assignable, err := s.RoleInspector.CanAssignRole(ctx, guildID, config.RoleID)
			if errors.Is(err, ports.ErrDiscordRoleMissing) {
				continue
			}
			if err != nil {
				errs = append(errs, err)
				continue
			}
			if !assignable {
				if err := s.notifyOwnerRoleHierarchy(ctx, guildID, config.RoleID); err != nil {
					errs = append(errs, err)
				}
				continue
			}
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

func (s JoinRoleAssignmentService) notifyOwnerRoleHierarchy(ctx context.Context, guildID string, roleID string) error {
	if s.Guilds == nil || s.DirectMessages == nil {
		return ports.ErrDiscordRoleNotAssignable
	}
	guild, err := s.Guilds.GuildInfo(ctx, guildID)
	if err != nil {
		return err
	}
	ownerID := strings.TrimSpace(guild.OwnerID)
	if ownerID == "" {
		return ports.ErrDiscordRoleNotAssignable
	}
	_, err = s.DirectMessages.SendDirectMessage(ctx, ownerID, ports.OutboundMessage{
		Content:         "很抱歉，我沒有權限給他加入的成員身分組\n麻煩請將我的身份組位階調高!\n身分組:<@" + strings.TrimSpace(roleID) + ">",
		AllowedMentions: ports.AllowedMentions{},
	})
	return err
}
