package ports

import (
	"context"
	"errors"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
)

var ErrStatsConfigMissing = errors.New("stats config is missing")

type StatsConfigRepository interface {
	GetStatsConfig(ctx context.Context, guildID string) (domain.StatsConfig, error)
	SaveStatsConfig(ctx context.Context, config domain.StatsConfig) error
	AddStatsConfigChannel(ctx context.Context, guildID string, option string, channelID string, currentValue int) (domain.StatsConfig, error)
	DeleteStatsConfig(ctx context.Context, guildID string) (domain.StatsConfig, error)
}

type StatsRoleConfigRepository interface {
	SaveStatsRoleConfig(ctx context.Context, config domain.StatsRoleConfig) error
}

type StatsRenameRepository interface {
	ListStatsConfigs(ctx context.Context) ([]domain.StatsConfig, error)
	UpdateStatsConfigCounters(ctx context.Context, guildID string, update domain.StatsConfigCounterUpdate) error
	ListStatsRoleConfigs(ctx context.Context) ([]domain.StatsRoleConfig, error)
	UpdateStatsRoleConfigCounter(ctx context.Context, guildID string, roleID string, currentValue string) error
}

type DiscordGuildStatsReader interface {
	GuildStats(ctx context.Context, guildID string) (domain.StatsSnapshot, error)
}

type DiscordRoleStatsReader interface {
	RoleStats(ctx context.Context, guildID string, roleID string) (domain.StatsRoleSnapshot, error)
}
