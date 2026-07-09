package ports

import (
	"context"
	"errors"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
)

var ErrRoleReactionConfigMissing = errors.New("role reaction config is missing")
var ErrRoleButtonConfigMissing = errors.New("role button config is missing")

type RoleReactionRepository interface {
	SaveRoleReactionConfig(ctx context.Context, config domain.RoleReactionConfig) error
	DeleteRoleReactionConfig(ctx context.Context, guildID string, messageID string, react string) error
	GetRoleReactionConfig(ctx context.Context, guildID string, messageID string, react string) (domain.RoleReactionConfig, error)
}

type RoleButtonRepository interface {
	SaveRoleButtonConfigs(ctx context.Context, configs ...domain.RoleButtonConfig) error
	GetRoleButtonConfig(ctx context.Context, guildID string, number string) (domain.RoleButtonConfig, error)
}

type RoleSelectionRepository interface {
	RoleReactionRepository
	RoleButtonRepository
}
