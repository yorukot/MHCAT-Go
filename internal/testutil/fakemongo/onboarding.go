package fakemongo

import (
	"context"
	"strings"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
)

type JoinRoleConfigRepository struct {
	Configs map[string]domain.JoinRoleConfig
	Err     error
}

type JoinMessageConfigRepository struct {
	Configs map[string]domain.JoinMessageConfig
	Err     error
}

type LeaveMessageConfigRepository struct {
	Configs map[string]domain.LeaveMessageConfig
	Err     error
}

type VerificationConfigRepository struct {
	Configs map[string]domain.VerificationConfig
	Err     error
}

type AccountAgeConfigRepository struct {
	Configs map[string]domain.AccountAgeConfig
	Err     error
}

func NewJoinRoleConfigRepository() *JoinRoleConfigRepository {
	return &JoinRoleConfigRepository{Configs: map[string]domain.JoinRoleConfig{}}
}

func NewJoinMessageConfigRepository() *JoinMessageConfigRepository {
	return &JoinMessageConfigRepository{Configs: map[string]domain.JoinMessageConfig{}}
}

func NewLeaveMessageConfigRepository() *LeaveMessageConfigRepository {
	return &LeaveMessageConfigRepository{Configs: map[string]domain.LeaveMessageConfig{}}
}

func NewVerificationConfigRepository() *VerificationConfigRepository {
	return &VerificationConfigRepository{Configs: map[string]domain.VerificationConfig{}}
}

func NewAccountAgeConfigRepository() *AccountAgeConfigRepository {
	return &AccountAgeConfigRepository{Configs: map[string]domain.AccountAgeConfig{}}
}

func (r *JoinRoleConfigRepository) CreateJoinRoleConfig(ctx context.Context, config domain.JoinRoleConfig) error {
	if err := r.ready(ctx); err != nil {
		return err
	}
	key := joinRoleKey(config.GuildID, config.RoleID)
	if _, exists := r.Configs[key]; exists {
		return ports.ErrJoinRoleConfigExists
	}
	r.Configs[key] = config
	return nil
}

func (r *JoinRoleConfigRepository) ListJoinRoleConfigs(ctx context.Context, guildID string) ([]domain.JoinRoleConfig, error) {
	if err := r.ready(ctx); err != nil {
		return nil, err
	}
	guildID = strings.TrimSpace(guildID)
	var configs []domain.JoinRoleConfig
	for _, config := range r.Configs {
		if strings.TrimSpace(config.GuildID) == guildID {
			configs = append(configs, config)
		}
	}
	return configs, nil
}

func (r *JoinRoleConfigRepository) DeleteJoinRoleConfig(ctx context.Context, guildID string, roleID string) error {
	if err := r.ready(ctx); err != nil {
		return err
	}
	key := joinRoleKey(guildID, roleID)
	if _, ok := r.Configs[key]; !ok {
		return ports.ErrJoinRoleConfigMissing
	}
	delete(r.Configs, key)
	return nil
}

func (r *JoinRoleConfigRepository) ready(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return r.Err
}

func joinRoleKey(guildID string, roleID string) string {
	return strings.TrimSpace(guildID) + "/" + strings.TrimSpace(roleID)
}

var _ ports.JoinRoleConfigRepository = (*JoinRoleConfigRepository)(nil)

func (r *JoinMessageConfigRepository) GetJoinMessageConfig(ctx context.Context, guildID string) (domain.JoinMessageConfig, error) {
	if err := r.ready(ctx); err != nil {
		return domain.JoinMessageConfig{}, err
	}
	guildID = strings.TrimSpace(guildID)
	config, ok := r.Configs[guildID]
	if !ok {
		return domain.JoinMessageConfig{}, ports.ErrJoinMessageConfigMissing
	}
	return config, nil
}

func (r *JoinMessageConfigRepository) ready(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return r.Err
}

var _ ports.JoinMessageConfigReader = (*JoinMessageConfigRepository)(nil)

func (r *LeaveMessageConfigRepository) PrepareLeaveMessageConfig(ctx context.Context, guildID string, channelID string) (domain.LeaveMessageConfig, error) {
	if err := r.ready(ctx); err != nil {
		return domain.LeaveMessageConfig{}, err
	}
	guildID = strings.TrimSpace(guildID)
	channelID = strings.TrimSpace(channelID)
	config, ok := r.Configs[guildID]
	if !ok {
		config = domain.LeaveMessageConfig{GuildID: guildID}
	}
	config.ChannelID = channelID
	r.Configs[guildID] = config
	return config, nil
}

func (r *LeaveMessageConfigRepository) SaveLeaveMessageContent(ctx context.Context, config domain.LeaveMessageConfig) error {
	if err := r.ready(ctx); err != nil {
		return err
	}
	guildID := strings.TrimSpace(config.GuildID)
	existing, ok := r.Configs[guildID]
	if !ok {
		return ports.ErrLeaveMessageConfigMissing
	}
	existing.MessageContent = config.MessageContent
	existing.Title = config.Title
	existing.Color = strings.TrimSpace(config.Color)
	r.Configs[guildID] = existing
	return nil
}

func (r *LeaveMessageConfigRepository) GetLeaveMessageConfig(ctx context.Context, guildID string) (domain.LeaveMessageConfig, error) {
	if err := r.ready(ctx); err != nil {
		return domain.LeaveMessageConfig{}, err
	}
	guildID = strings.TrimSpace(guildID)
	config, ok := r.Configs[guildID]
	if !ok {
		return domain.LeaveMessageConfig{}, ports.ErrLeaveMessageConfigMissing
	}
	return config, nil
}

func (r *LeaveMessageConfigRepository) ready(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return r.Err
}

var _ ports.LeaveMessageConfigRepository = (*LeaveMessageConfigRepository)(nil)

func (r *VerificationConfigRepository) SaveVerificationConfig(ctx context.Context, config domain.VerificationConfig) error {
	if err := r.ready(ctx); err != nil {
		return err
	}
	config.GuildID = strings.TrimSpace(config.GuildID)
	config.RoleID = strings.TrimSpace(config.RoleID)
	if err := config.Validate(); err != nil {
		return err
	}
	r.Configs[config.GuildID] = config
	return nil
}

func (r *VerificationConfigRepository) GetVerificationConfig(ctx context.Context, guildID string) (domain.VerificationConfig, error) {
	if err := r.ready(ctx); err != nil {
		return domain.VerificationConfig{}, err
	}
	config, ok := r.Configs[strings.TrimSpace(guildID)]
	if !ok {
		return domain.VerificationConfig{}, ports.ErrVerificationConfigMissing
	}
	return config, nil
}

func (r *VerificationConfigRepository) ready(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return r.Err
}

var _ ports.VerificationConfigRepository = (*VerificationConfigRepository)(nil)

func (r *AccountAgeConfigRepository) SaveAccountAgeRequirement(ctx context.Context, guildID string, requiredSeconds int64) (domain.AccountAgeConfig, error) {
	if err := r.ready(ctx); err != nil {
		return domain.AccountAgeConfig{}, err
	}
	guildID = strings.TrimSpace(guildID)
	config := domain.AccountAgeConfig{GuildID: guildID, RequiredSeconds: float64(requiredSeconds)}
	if err := config.Validate(); err != nil {
		return domain.AccountAgeConfig{}, err
	}
	if existing, ok := r.Configs[guildID]; ok {
		config.ChannelID = existing.ChannelID
	}
	r.Configs[guildID] = config
	return config, nil
}

func (r *AccountAgeConfigRepository) SetAccountAgeLogChannel(ctx context.Context, guildID string, channelID string) (domain.AccountAgeConfig, error) {
	if err := r.ready(ctx); err != nil {
		return domain.AccountAgeConfig{}, err
	}
	guildID = strings.TrimSpace(guildID)
	channelID = strings.TrimSpace(channelID)
	if guildID == "" || channelID == "" {
		return domain.AccountAgeConfig{}, domain.ErrInvalidAccountAgeConfig
	}
	config, ok := r.Configs[guildID]
	if !ok {
		return domain.AccountAgeConfig{}, ports.ErrAccountAgeConfigMissing
	}
	config.ChannelID = channelID
	r.Configs[guildID] = config
	return config, nil
}

func (r *AccountAgeConfigRepository) DeleteAccountAgeConfig(ctx context.Context, guildID string) error {
	if err := r.ready(ctx); err != nil {
		return err
	}
	guildID = strings.TrimSpace(guildID)
	if _, ok := r.Configs[guildID]; !ok {
		return ports.ErrAccountAgeConfigMissing
	}
	delete(r.Configs, guildID)
	return nil
}

func (r *AccountAgeConfigRepository) DeleteAccountAgeLogChannel(ctx context.Context, guildID string) error {
	if err := r.ready(ctx); err != nil {
		return err
	}
	guildID = strings.TrimSpace(guildID)
	config, ok := r.Configs[guildID]
	if !ok {
		return ports.ErrAccountAgeConfigMissing
	}
	config.ChannelID = ""
	r.Configs[guildID] = config
	return nil
}

func (r *AccountAgeConfigRepository) GetAccountAgeConfig(ctx context.Context, guildID string) (domain.AccountAgeConfig, error) {
	if err := r.ready(ctx); err != nil {
		return domain.AccountAgeConfig{}, err
	}
	config, ok := r.Configs[strings.TrimSpace(guildID)]
	if !ok {
		return domain.AccountAgeConfig{}, ports.ErrAccountAgeConfigMissing
	}
	return config, nil
}

func (r *AccountAgeConfigRepository) ready(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return r.Err
}

var _ ports.AccountAgeConfigRepository = (*AccountAgeConfigRepository)(nil)
