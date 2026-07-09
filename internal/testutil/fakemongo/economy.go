package fakemongo

import (
	"context"
	"sync"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
)

type EconomyRepository struct {
	mu             sync.Mutex
	Balances       map[string]domain.CoinBalance
	Configs        map[string]domain.EconomyConfig
	Calendars      map[string]domain.SignCalendar
	SignInResult   *domain.SignInResult
	SignInErr      error
	SignInCommands []domain.SignInCommand
	SavedConfigs   []domain.EconomyConfig
	Err            error
}

func NewEconomyRepository() *EconomyRepository {
	return &EconomyRepository{
		Balances:  map[string]domain.CoinBalance{},
		Configs:   map[string]domain.EconomyConfig{},
		Calendars: map[string]domain.SignCalendar{},
	}
}

func (r *EconomyRepository) GetCoinBalance(ctx context.Context, guildID string, userID string) (domain.CoinBalance, error) {
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

func (r *EconomyRepository) GetEconomyConfig(ctx context.Context, guildID string) (domain.EconomyConfig, error) {
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

func (r *EconomyRepository) PutBalance(balance domain.CoinBalance) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.Balances[economyBalanceKey(balance.GuildID, balance.UserID)] = balance
}

func (r *EconomyRepository) PutConfig(config domain.EconomyConfig) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.Configs[config.GuildID] = config
}

func (r *EconomyRepository) SaveEconomyConfig(ctx context.Context, config domain.EconomyConfig) (domain.EconomyConfig, error) {
	if err := r.ready(ctx); err != nil {
		return domain.EconomyConfig{}, err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.Configs[config.GuildID] = config
	r.SavedConfigs = append(r.SavedConfigs, config)
	return config, nil
}

func (r *EconomyRepository) SignIn(ctx context.Context, command domain.SignInCommand) (domain.SignInResult, error) {
	if err := r.ready(ctx); err != nil {
		return domain.SignInResult{}, err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.SignInCommands = append(r.SignInCommands, command)
	if r.SignInErr != nil {
		return domain.SignInResult{}, r.SignInErr
	}
	if r.SignInResult != nil {
		return *r.SignInResult, nil
	}
	balance, ok := r.Balances[economyBalanceKey(command.GuildID, command.UserID)]
	if !ok {
		balance = domain.CoinBalance{GuildID: command.GuildID, UserID: command.UserID, Coins: 25, Today: command.Now.Unix()}
	}
	config, ok := r.Configs[command.GuildID]
	result := domain.SignInResult{
		Balance:     balance,
		Config:      config,
		Calendar:    r.Calendars[economyBalanceKey(command.GuildID, command.UserID)],
		Reward:      25,
		ConfigFound: ok,
		SignedAt:    command.Now,
	}
	return result, nil
}

func (r *EconomyRepository) GetSignCalendar(ctx context.Context, guildID string, userID string, year string, month string) (domain.SignCalendar, error) {
	if err := r.ready(ctx); err != nil {
		return domain.SignCalendar{}, err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	calendar, ok := r.Calendars[economyBalanceKey(guildID, userID)]
	if !ok {
		return domain.SignCalendar{GuildID: guildID, UserID: userID, Date: map[string]map[string][]string{}}, nil
	}
	return cloneSignCalendar(calendar), nil
}

func (r *EconomyRepository) PutCalendar(calendar domain.SignCalendar) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.Calendars[economyBalanceKey(calendar.GuildID, calendar.UserID)] = cloneSignCalendar(calendar)
}

func (r *EconomyRepository) PutSignInResult(result domain.SignInResult) {
	r.mu.Lock()
	defer r.mu.Unlock()
	copied := result
	r.SignInResult = &copied
}

func (r *EconomyRepository) ready(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return r.Err
}

func economyBalanceKey(guildID string, userID string) string {
	return guildID + "\x00" + userID
}

var _ ports.EconomyRepository = (*EconomyRepository)(nil)
var _ ports.EconomySignInRepository = (*EconomyRepository)(nil)
var _ ports.EconomySettingsRepository = (*EconomyRepository)(nil)

func cloneSignCalendar(calendar domain.SignCalendar) domain.SignCalendar {
	cloned := domain.SignCalendar{
		GuildID: calendar.GuildID,
		UserID:  calendar.UserID,
		Date:    map[string]map[string][]string{},
	}
	for year, months := range calendar.Date {
		cloned.Date[year] = map[string][]string{}
		for month, days := range months {
			cloned.Date[year][month] = append([]string(nil), days...)
		}
	}
	return cloned
}
