package fakemongo

import (
	"context"
	"sync"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
)

type EconomyProfileRepository struct {
	mu sync.Mutex

	Balances   map[string]domain.CoinBalance
	balanceSeq []string
	Configs    map[string]domain.EconomyConfig

	WorkConfigs map[string]domain.WorkConfig
	WorkUsers   map[string]domain.WorkUserState

	TextXP   map[string]domain.XPProfile
	textSeq  []string
	VoiceXP  map[string]domain.XPProfile
	voiceSeq []string

	Err error
}

func NewEconomyProfileRepository() *EconomyProfileRepository {
	return &EconomyProfileRepository{
		Balances:    map[string]domain.CoinBalance{},
		Configs:     map[string]domain.EconomyConfig{},
		WorkConfigs: map[string]domain.WorkConfig{},
		WorkUsers:   map[string]domain.WorkUserState{},
		TextXP:      map[string]domain.XPProfile{},
		VoiceXP:     map[string]domain.XPProfile{},
	}
}

func (r *EconomyProfileRepository) PutBalance(balance domain.CoinBalance) {
	r.mu.Lock()
	defer r.mu.Unlock()
	key := economyBalanceKey(balance.GuildID, balance.UserID)
	if _, ok := r.Balances[key]; !ok {
		r.balanceSeq = append(r.balanceSeq, key)
	}
	r.Balances[key] = balance
}

func (r *EconomyProfileRepository) PutConfig(config domain.EconomyConfig) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.Configs[config.GuildID] = config
}

func (r *EconomyProfileRepository) PutWorkConfig(config domain.WorkConfig) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.WorkConfigs[config.GuildID] = config
}

func (r *EconomyProfileRepository) PutWorkUser(user domain.WorkUserState) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.WorkUsers[workUserKey(user.GuildID, user.UserID)] = user
}

func (r *EconomyProfileRepository) PutTextXP(profile domain.XPProfile) {
	r.mu.Lock()
	defer r.mu.Unlock()
	key := economyBalanceKey(profile.GuildID, profile.UserID)
	if _, ok := r.TextXP[key]; !ok {
		r.textSeq = append(r.textSeq, key)
	}
	r.TextXP[key] = profile
}

func (r *EconomyProfileRepository) PutVoiceXP(profile domain.XPProfile) {
	r.mu.Lock()
	defer r.mu.Unlock()
	key := economyBalanceKey(profile.GuildID, profile.UserID)
	if _, ok := r.VoiceXP[key]; !ok {
		r.voiceSeq = append(r.voiceSeq, key)
	}
	r.VoiceXP[key] = profile
}

func (r *EconomyProfileRepository) GetCoinBalance(ctx context.Context, guildID string, userID string) (domain.CoinBalance, error) {
	if err := r.ready(ctx); err != nil {
		return domain.CoinBalance{}, err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	balance, ok := r.Balances[economyBalanceKey(guildID, userID)]
	if !ok {
		return domain.CoinBalance{}, ports.ErrCoinBalanceNotFound
	}
	return balance, nil
}

func (r *EconomyProfileRepository) GetEconomyConfig(ctx context.Context, guildID string) (domain.EconomyConfig, error) {
	if err := r.ready(ctx); err != nil {
		return domain.EconomyConfig{}, err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	config, ok := r.Configs[guildID]
	if !ok {
		return domain.EconomyConfig{GuildID: guildID}, ports.ErrEconomyConfigMissing
	}
	return config, nil
}

func (r *EconomyProfileRepository) ListCoinBalances(ctx context.Context, guildID string) ([]domain.CoinBalance, error) {
	if err := r.ready(ctx); err != nil {
		return nil, err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	balances := []domain.CoinBalance{}
	for _, key := range r.balanceSeq {
		balance := r.Balances[key]
		if balance.GuildID == guildID {
			balances = append(balances, balance)
		}
	}
	return balances, nil
}

func (r *EconomyProfileRepository) GetWorkConfig(ctx context.Context, guildID string) (domain.WorkConfig, error) {
	if err := r.ready(ctx); err != nil {
		return domain.WorkConfig{}, err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	config, ok := r.WorkConfigs[guildID]
	if !ok {
		return domain.WorkConfig{}, ports.ErrWorkConfigMissing
	}
	return config, nil
}

func (r *EconomyProfileRepository) GetWorkUser(ctx context.Context, guildID string, userID string) (domain.WorkUserState, error) {
	if err := r.ready(ctx); err != nil {
		return domain.WorkUserState{}, err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	user, ok := r.WorkUsers[workUserKey(guildID, userID)]
	if !ok {
		return domain.WorkUserState{}, ports.ErrWorkUserMissing
	}
	return user, nil
}

func (r *EconomyProfileRepository) GetTextXPProfile(ctx context.Context, guildID string, userID string) (domain.XPProfile, error) {
	return r.getXP(ctx, r.TextXP, guildID, userID, ports.ErrTextXPProfileMissing)
}

func (r *EconomyProfileRepository) ListTextXPProfiles(ctx context.Context, guildID string) ([]domain.XPProfile, error) {
	return r.listXP(ctx, r.TextXP, r.textSeq, guildID)
}

func (r *EconomyProfileRepository) GetVoiceXPProfile(ctx context.Context, guildID string, userID string) (domain.XPProfile, error) {
	return r.getXP(ctx, r.VoiceXP, guildID, userID, ports.ErrVoiceXPProfileMissing)
}

func (r *EconomyProfileRepository) ListVoiceXPProfiles(ctx context.Context, guildID string) ([]domain.XPProfile, error) {
	return r.listXP(ctx, r.VoiceXP, r.voiceSeq, guildID)
}

func (r *EconomyProfileRepository) getXP(ctx context.Context, profiles map[string]domain.XPProfile, guildID string, userID string, missing error) (domain.XPProfile, error) {
	if err := r.ready(ctx); err != nil {
		return domain.XPProfile{}, err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	profile, ok := profiles[economyBalanceKey(guildID, userID)]
	if !ok {
		return domain.XPProfile{}, missing
	}
	return profile, nil
}

func (r *EconomyProfileRepository) listXP(ctx context.Context, profiles map[string]domain.XPProfile, seq []string, guildID string) ([]domain.XPProfile, error) {
	if err := r.ready(ctx); err != nil {
		return nil, err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	result := []domain.XPProfile{}
	for _, key := range seq {
		profile := profiles[key]
		if profile.GuildID == guildID {
			result = append(result, profile)
		}
	}
	return result, nil
}

func (r *EconomyProfileRepository) ready(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return r.Err
}

var _ ports.EconomyProfileRepository = (*EconomyProfileRepository)(nil)
