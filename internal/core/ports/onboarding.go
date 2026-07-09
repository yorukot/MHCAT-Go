package ports

import (
	"context"
	"errors"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
)

var ErrJoinRoleConfigExists = errors.New("join role config already exists")
var ErrJoinRoleConfigMissing = errors.New("join role config is missing")
var ErrJoinMessageConfigMissing = errors.New("join message config is missing")
var ErrDiscordRoleNotAssignable = errors.New("discord role is not assignable")
var ErrDiscordRoleMissing = errors.New("discord role is missing")
var ErrLeaveMessageConfigMissing = errors.New("leave message config is missing")
var ErrVerificationConfigMissing = errors.New("verification config is missing")
var ErrAccountAgeConfigMissing = errors.New("account age config is missing")

type JoinRoleConfigRepository interface {
	JoinRoleConfigReader
	CreateJoinRoleConfig(ctx context.Context, config domain.JoinRoleConfig) error
	DeleteJoinRoleConfig(ctx context.Context, guildID string, roleID string) error
}

type JoinRoleConfigReader interface {
	ListJoinRoleConfigs(ctx context.Context, guildID string) ([]domain.JoinRoleConfig, error)
}

type JoinMessageConfigReader interface {
	GetJoinMessageConfig(ctx context.Context, guildID string) (domain.JoinMessageConfig, error)
}

type DiscordRoleInspector interface {
	CanAssignRole(ctx context.Context, guildID string, roleID string) (bool, error)
}

type LeaveMessageConfigRepository interface {
	LeaveMessageConfigReader
	PrepareLeaveMessageConfig(ctx context.Context, guildID string, channelID string) (domain.LeaveMessageConfig, error)
	SaveLeaveMessageContent(ctx context.Context, config domain.LeaveMessageConfig) error
}

type LeaveMessageConfigReader interface {
	GetLeaveMessageConfig(ctx context.Context, guildID string) (domain.LeaveMessageConfig, error)
}

type VerificationConfigRepository interface {
	VerificationConfigReader
	SaveVerificationConfig(ctx context.Context, config domain.VerificationConfig) error
}

type VerificationConfigReader interface {
	GetVerificationConfig(ctx context.Context, guildID string) (domain.VerificationConfig, error)
}

type AccountAgeConfigRepository interface {
	AccountAgeConfigReader
	SaveAccountAgeRequirement(ctx context.Context, guildID string, requiredSeconds int64) (domain.AccountAgeConfig, error)
	SetAccountAgeLogChannel(ctx context.Context, guildID string, channelID string) (domain.AccountAgeConfig, error)
	DeleteAccountAgeConfig(ctx context.Context, guildID string) error
	DeleteAccountAgeLogChannel(ctx context.Context, guildID string) error
}

type AccountAgeConfigReader interface {
	GetAccountAgeConfig(ctx context.Context, guildID string) (domain.AccountAgeConfig, error)
}
