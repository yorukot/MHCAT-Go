package ports

import (
	"context"
	"errors"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
)

var ErrTextXPConfigMissing = errors.New("text xp config is missing")
var ErrVoiceXPConfigMissing = errors.New("voice xp config is missing")
var ErrTextXPProfileMissing = errors.New("text xp profile is missing")
var ErrVoiceXPProfileMissing = errors.New("voice xp profile is missing")
var ErrTextXPRewardRoleMissing = errors.New("text xp reward role config is missing")
var ErrVoiceXPRewardRoleMissing = errors.New("voice xp reward role config is missing")
var ErrXPRewardRoleLimitExceeded = errors.New("xp reward role config limit exceeded")

type TextXPConfigRepository interface {
	SaveTextXPConfig(ctx context.Context, config domain.TextXPConfig) error
	DeleteTextXPConfig(ctx context.Context, guildID string) error
}

type TextXPConfigReader interface {
	GetTextXPConfig(ctx context.Context, guildID string) (domain.TextXPConfig, error)
}

type VoiceXPConfigRepository interface {
	SaveVoiceXPConfig(ctx context.Context, config domain.VoiceXPConfig) error
	DeleteVoiceXPConfig(ctx context.Context, guildID string) error
}

type TextXPRewardRoleRepository interface {
	ListTextXPRewardRoles(ctx context.Context, guildID string) ([]domain.XPRewardRoleConfig, error)
	SaveTextXPRewardRole(ctx context.Context, config domain.XPRewardRoleConfig) error
	DeleteTextXPRewardRole(ctx context.Context, guildID string, level int64, roleID string) error
}

type TextXPCoinRewardRepository interface {
	ApplyTextXPCoinReward(ctx context.Context, guildID string, userID string, level int64) (domain.CoinBalance, error)
}

type VoiceXPRewardRoleRepository interface {
	ListVoiceXPRewardRoles(ctx context.Context, guildID string) ([]domain.XPRewardRoleConfig, error)
	SaveVoiceXPRewardRole(ctx context.Context, config domain.XPRewardRoleConfig) error
	DeleteVoiceXPRewardRole(ctx context.Context, guildID string, level int64, roleID string) error
}

type XPAdminRepository interface {
	GetTextXPProfile(ctx context.Context, guildID string, userID string) (domain.XPProfile, error)
	SaveTextXPProfile(ctx context.Context, profile domain.XPProfile) error
	GetVoiceXPProfile(ctx context.Context, guildID string, userID string) (domain.XPProfile, error)
	SaveVoiceXPProfile(ctx context.Context, profile domain.XPProfile) error
}

type TextXPAccrualRepository interface {
	GetTextXPProfile(ctx context.Context, guildID string, userID string) (domain.XPProfile, error)
	SaveTextXPProfile(ctx context.Context, profile domain.XPProfile) error
}

type VoiceXPSessionRepository interface {
	MarkVoiceXPJoined(ctx context.Context, guildID string, userID string) error
	MarkVoiceXPLeft(ctx context.Context, guildID string, userID string) error
}

type VoiceXPAccrualRepository interface {
	GetVoiceXPProfile(ctx context.Context, guildID string, userID string) (domain.XPProfile, error)
	SaveVoiceXPProfile(ctx context.Context, profile domain.XPProfile) error
}

type XPResetRepository interface {
	DeleteTextXPProfile(ctx context.Context, guildID string, userID string) error
	DeleteVoiceXPProfile(ctx context.Context, guildID string, userID string) error
	DeleteTextXPGuild(ctx context.Context, guildID string) error
	DeleteVoiceXPGuild(ctx context.Context, guildID string) error
}

type XPRankRepository interface {
	ListTextXPProfiles(ctx context.Context, guildID string) ([]domain.XPProfile, error)
	ListVoiceXPProfiles(ctx context.Context, guildID string) ([]domain.XPProfile, error)
}
