package onboarding

import (
	"context"
	"errors"
	"strings"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
)

type VerificationConfigService struct {
	Repository    ports.VerificationConfigRepository
	RoleInspector ports.DiscordRoleInspector
}

func (s VerificationConfigService) Save(ctx context.Context, config domain.VerificationConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if s.Repository == nil || s.RoleInspector == nil {
		return domain.ErrInvalidVerificationConfig
	}
	config.GuildID = strings.TrimSpace(config.GuildID)
	config.RoleID = strings.TrimSpace(config.RoleID)
	config.RenameTemplate = strings.TrimSpace(config.RenameTemplate)
	if err := config.Validate(); err != nil {
		return err
	}
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
	return s.Repository.SaveVerificationConfig(ctx, config)
}
