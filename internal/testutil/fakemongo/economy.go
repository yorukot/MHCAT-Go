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
	balanceOrder   []string
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
	key := economyBalanceKey(balance.GuildID, balance.UserID)
	if _, ok := r.Balances[key]; !ok {
		r.balanceOrder = append(r.balanceOrder, key)
	}
	r.Balances[key] = balance
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

func (r *EconomyRepository) AdjustCoinBalance(ctx context.Context, command domain.CoinAdminCommand) (domain.CoinAdminResult, error) {
	if err := r.ready(ctx); err != nil {
		return domain.CoinAdminResult{}, err
	}
	command = command.Normalize()
	if err := command.Validate(); err != nil {
		return domain.CoinAdminResult{}, err
	}
	delta, err := command.SignedDelta()
	if err != nil {
		return domain.CoinAdminResult{}, err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	key := economyBalanceKey(command.GuildID, command.UserID)
	balance, ok := r.Balances[key]
	if !ok {
		if command.Operation == domain.CoinAdminOperationReduce {
			return domain.CoinAdminResult{}, ports.ErrCoinNegativeBalance
		}
		balance = domain.CoinBalance{GuildID: command.GuildID, UserID: command.UserID, Coins: command.Amount, Today: 1}
		r.Balances[key] = balance
		r.balanceOrder = append(r.balanceOrder, key)
		return domain.CoinAdminResult{Balance: balance, Delta: delta, Created: true}, nil
	}
	next := balance.Coins + delta
	if next < 0 {
		return domain.CoinAdminResult{}, ports.ErrCoinNegativeBalance
	}
	if next > 999999999 {
		return domain.CoinAdminResult{}, ports.ErrCoinLimitExceeded
	}
	balance.Coins = next
	r.Balances[key] = balance
	return domain.CoinAdminResult{Balance: balance, Delta: delta}, nil
}

func (r *EconomyRepository) ApplyRockPaperScissors(ctx context.Context, command domain.RockPaperScissorsCommand) (domain.RockPaperScissorsResult, error) {
	if err := r.ready(ctx); err != nil {
		return domain.RockPaperScissorsResult{}, err
	}
	command = command.Normalize()
	if err := command.Validate(); err != nil {
		return domain.RockPaperScissorsResult{}, err
	}
	outcome, delta, err := domain.ResolveRockPaperScissors(command)
	if err != nil {
		return domain.RockPaperScissorsResult{}, err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	key := economyBalanceKey(command.GuildID, command.UserID)
	balance, ok := r.Balances[key]
	if !ok {
		return domain.RockPaperScissorsResult{}, ports.ErrCoinBalanceNotFound
	}
	if balance.Coins < command.Wager {
		return domain.RockPaperScissorsResult{}, ports.ErrCoinNegativeBalance
	}
	next := balance.Coins + delta
	if next < 0 {
		return domain.RockPaperScissorsResult{}, ports.ErrCoinNegativeBalance
	}
	previous := balance.Coins
	balance.Coins = next
	r.Balances[key] = balance
	return domain.RockPaperScissorsResult{
		Balance:         balance,
		PreviousBalance: previous,
		Delta:           delta,
		Outcome:         outcome,
		PlayerChoice:    command.PlayerChoice,
		ComputerChoice:  command.ComputerChoice,
	}, nil
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

func (r *EconomyRepository) ListCoinBalances(ctx context.Context, guildID string) ([]domain.CoinBalance, error) {
	if err := r.ready(ctx); err != nil {
		return nil, err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	balances := []domain.CoinBalance{}
	for _, key := range r.balanceOrder {
		balance := r.Balances[key]
		if balance.GuildID == guildID {
			balances = append(balances, balance)
		}
	}
	return balances, nil
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
var _ ports.EconomyCoinRankRepository = (*EconomyRepository)(nil)
var _ ports.EconomySettingsRepository = (*EconomyRepository)(nil)
var _ ports.EconomyCoinAdminRepository = (*EconomyRepository)(nil)
var _ ports.EconomyRockPaperScissorsRepository = (*EconomyRepository)(nil)

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
