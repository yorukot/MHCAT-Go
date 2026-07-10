package fakemongo

import (
	"context"
	"sync"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
)

type StatsConfigRepository struct {
	mu          sync.Mutex
	Configs     map[string]domain.StatsConfig
	RoleConfigs map[string]domain.StatsRoleConfig
	Err         error
}

func NewStatsConfigRepository() *StatsConfigRepository {
	return &StatsConfigRepository{Configs: map[string]domain.StatsConfig{}, RoleConfigs: map[string]domain.StatsRoleConfig{}}
}

func (r *StatsConfigRepository) Put(config domain.StatsConfig) {
	r.mu.Lock()
	defer r.mu.Unlock()
	config = config.Normalize()
	r.Configs[config.GuildID] = config
}

func (r *StatsConfigRepository) GetStatsConfig(ctx context.Context, guildID string) (domain.StatsConfig, error) {
	if err := ctx.Err(); err != nil {
		return domain.StatsConfig{}, err
	}
	if r.Err != nil {
		return domain.StatsConfig{}, r.Err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	config, ok := r.Configs[guildID]
	if !ok {
		return domain.StatsConfig{}, ports.ErrStatsConfigMissing
	}
	return config.Normalize(), nil
}

func (r *StatsConfigRepository) SaveStatsConfig(ctx context.Context, config domain.StatsConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if r.Err != nil {
		return r.Err
	}
	r.Put(config)
	return nil
}

func (r *StatsConfigRepository) AddStatsConfigChannel(ctx context.Context, guildID string, option string, channelID string, currentValue int) (domain.StatsConfig, error) {
	if err := ctx.Err(); err != nil {
		return domain.StatsConfig{}, err
	}
	if r.Err != nil {
		return domain.StatsConfig{}, r.Err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	config, ok := r.Configs[guildID]
	if !ok {
		return domain.StatsConfig{}, ports.ErrStatsConfigMissing
	}
	next, err := config.WithOptionalChannel(option, channelID, currentValue)
	if err != nil {
		return domain.StatsConfig{}, err
	}
	r.Configs[guildID] = next
	return next, nil
}

func (r *StatsConfigRepository) DeleteStatsConfig(ctx context.Context, guildID string) (domain.StatsConfig, error) {
	if err := ctx.Err(); err != nil {
		return domain.StatsConfig{}, err
	}
	if r.Err != nil {
		return domain.StatsConfig{}, r.Err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	config, ok := r.Configs[guildID]
	if !ok {
		return domain.StatsConfig{}, ports.ErrStatsConfigMissing
	}
	delete(r.Configs, guildID)
	return config, nil
}

func (r *StatsConfigRepository) ListStatsConfigs(ctx context.Context) ([]domain.StatsConfig, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if r.Err != nil {
		return nil, r.Err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	configs := make([]domain.StatsConfig, 0, len(r.Configs))
	for _, config := range r.Configs {
		configs = append(configs, config.Normalize())
	}
	return configs, nil
}

func (r *StatsConfigRepository) UpdateStatsConfigCounters(ctx context.Context, guildID string, update domain.StatsConfigCounterUpdate) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if r.Err != nil {
		return r.Err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	config, ok := r.Configs[guildID]
	if !ok {
		return ports.ErrStatsConfigMissing
	}
	if update.MemberNumberName != nil {
		config.MemberNumberName = *update.MemberNumberName
	}
	if update.UserNumberName != nil {
		config.UserNumberName = *update.UserNumberName
	}
	if update.BotNumberName != nil {
		config.BotNumberName = *update.BotNumberName
	}
	if update.ChannelNumberName != nil {
		config.ChannelNumberName = *update.ChannelNumberName
	}
	if update.TextNumberName != nil {
		config.TextNumberName = *update.TextNumberName
	}
	if update.VoiceNumberName != nil {
		config.VoiceNumberName = *update.VoiceNumberName
	}
	r.Configs[guildID] = config.Normalize()
	return nil
}

func (r *StatsConfigRepository) SaveStatsRoleConfig(ctx context.Context, config domain.StatsRoleConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if r.Err != nil {
		return r.Err
	}
	config = config.Normalize()
	if config.GuildID == "" || config.ChannelID == "" || config.ChannelName == "" || config.RoleID == "" {
		return domain.ErrInvalidStatsConfigRequest
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.RoleConfigs[statsRoleKey(config.GuildID, config.RoleID)] = config
	return nil
}

func (r *StatsConfigRepository) ListStatsRoleConfigs(ctx context.Context) ([]domain.StatsRoleConfig, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if r.Err != nil {
		return nil, r.Err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	configs := make([]domain.StatsRoleConfig, 0, len(r.RoleConfigs))
	for _, config := range r.RoleConfigs {
		configs = append(configs, config.Normalize())
	}
	return configs, nil
}

func (r *StatsConfigRepository) UpdateStatsRoleConfigCounter(ctx context.Context, guildID string, roleID string, currentValue string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if r.Err != nil {
		return r.Err
	}
	key := statsRoleKey(guildID, roleID)
	r.mu.Lock()
	defer r.mu.Unlock()
	config, ok := r.RoleConfigs[key]
	if !ok {
		return ports.ErrStatsConfigMissing
	}
	config.ChannelName = currentValue
	r.RoleConfigs[key] = config.Normalize()
	return nil
}

func statsRoleKey(guildID string, roleID string) string {
	return guildID + "/" + roleID
}

var _ ports.StatsConfigRepository = (*StatsConfigRepository)(nil)
var _ ports.StatsRoleConfigRepository = (*StatsConfigRepository)(nil)
var _ ports.StatsRenameRepository = (*StatsConfigRepository)(nil)
