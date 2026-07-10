package fakemongo

import (
	"context"
	"strconv"
	"sync"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
)

type EconomyRepository struct {
	mu             sync.Mutex
	Balances       map[string]domain.CoinBalance
	balanceOrder   []string
	ShopItems      map[string]domain.ShopItem
	shopItemOrder  []string
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
		ShopItems: map[string]domain.ShopItem{},
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

func (r *EconomyRepository) PutShopItem(item domain.ShopItem) {
	r.mu.Lock()
	defer r.mu.Unlock()
	key := economyShopItemKey(item.GuildID, item.CommodityID)
	if _, ok := r.ShopItems[key]; !ok {
		r.shopItemOrder = append(r.shopItemOrder, key)
	}
	r.ShopItems[key] = item
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

func (r *EconomyRepository) ApplyTextXPCoinReward(ctx context.Context, guildID string, userID string, level int64) (domain.CoinBalance, error) {
	return r.applyXPCoinReward(ctx, guildID, userID, level, domain.LegacyTextXPCoinReward)
}

func (r *EconomyRepository) ApplyVoiceXPCoinReward(ctx context.Context, guildID string, userID string, level int64) (domain.CoinBalance, error) {
	return r.applyXPCoinReward(ctx, guildID, userID, level, domain.LegacyVoiceXPCoinReward)
}

func (r *EconomyRepository) applyXPCoinReward(ctx context.Context, guildID string, userID string, level int64, rewardForLevel func(int64, float64) int64) (domain.CoinBalance, error) {
	if err := r.ready(ctx); err != nil {
		return domain.CoinBalance{}, err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	config := r.Configs[guildID]
	reward := rewardForLevel(level, config.XPMultiple)
	key := economyBalanceKey(guildID, userID)
	balance, ok := r.Balances[key]
	if !ok {
		balance = domain.CoinBalance{GuildID: guildID, UserID: userID, Coins: reward, Today: 0}
		r.Balances[key] = balance
		r.balanceOrder = append(r.balanceOrder, key)
		return balance, nil
	}
	balance.Coins += reward
	r.Balances[key] = balance
	return balance, nil
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

func (r *EconomyRepository) ResetCoinBalances(ctx context.Context, command domain.CoinResetCommand) (domain.CoinResetResult, error) {
	if err := r.ready(ctx); err != nil {
		return domain.CoinResetResult{}, err
	}
	command = command.Normalize()
	if err := command.Validate(); err != nil {
		return domain.CoinResetResult{}, err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	affected := int64(0)
	nextOrder := r.balanceOrder[:0]
	for _, key := range r.balanceOrder {
		balance := r.Balances[key]
		if balance.GuildID != command.GuildID {
			nextOrder = append(nextOrder, key)
			continue
		}
		affected++
		if command.Divisor == 0 {
			delete(r.Balances, key)
			continue
		}
		balance.Coins = domain.LegacyJavaScriptRound(float64(balance.Coins) / float64(command.Divisor))
		r.Balances[key] = balance
		nextOrder = append(nextOrder, key)
	}
	r.balanceOrder = nextOrder
	if affected == 0 {
		return domain.CoinResetResult{}, ports.ErrCoinBalanceNotFound
	}
	return domain.CoinResetResult{
		GuildID:       command.GuildID,
		Divisor:       command.Divisor,
		AffectedCount: affected,
		Deleted:       command.Divisor == 0,
	}, nil
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

func (r *EconomyRepository) CheckCoinGameBalances(ctx context.Context, command domain.CoinGameCommand) (domain.CoinGameBalanceResult, error) {
	if err := r.ready(ctx); err != nil {
		return domain.CoinGameBalanceResult{}, err
	}
	command = command.Normalize()
	if err := command.Validate(); err != nil {
		return domain.CoinGameBalanceResult{}, err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	challenger, opponent, err := r.coinGameBalancesLocked(command)
	if err != nil {
		return domain.CoinGameBalanceResult{}, err
	}
	return domain.CoinGameBalanceResult{Challenger: challenger, Opponent: opponent, Wager: command.Wager}, nil
}

func (r *EconomyRepository) ReserveCoinGameWager(ctx context.Context, command domain.CoinGameCommand) (domain.CoinGameBalanceResult, error) {
	if err := r.ready(ctx); err != nil {
		return domain.CoinGameBalanceResult{}, err
	}
	command = command.Normalize()
	if err := command.Validate(); err != nil {
		return domain.CoinGameBalanceResult{}, err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	challenger, opponent, err := r.coinGameBalancesLocked(command)
	if err != nil {
		return domain.CoinGameBalanceResult{}, err
	}
	challenger.Coins -= command.Wager
	opponent.Coins -= command.Wager
	r.Balances[economyBalanceKey(command.GuildID, command.ChallengerID)] = challenger
	r.Balances[economyBalanceKey(command.GuildID, command.OpponentID)] = opponent
	return domain.CoinGameBalanceResult{Challenger: challenger, Opponent: opponent, Wager: command.Wager}, nil
}

func (r *EconomyRepository) SettleCoinGameWager(ctx context.Context, command domain.CoinGameSettlementCommand) (domain.CoinGameSettlementResult, error) {
	if err := r.ready(ctx); err != nil {
		return domain.CoinGameSettlementResult{}, err
	}
	command = command.Normalize()
	if err := command.Validate(); err != nil {
		return domain.CoinGameSettlementResult{}, err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	challengerKey := economyBalanceKey(command.GuildID, command.ChallengerID)
	opponentKey := economyBalanceKey(command.GuildID, command.OpponentID)
	challenger, ok := r.Balances[challengerKey]
	if !ok {
		return domain.CoinGameSettlementResult{}, ports.ErrCoinBalanceNotFound
	}
	opponent, ok := r.Balances[opponentKey]
	if !ok {
		return domain.CoinGameSettlementResult{}, ports.ErrCoinBalanceNotFound
	}
	challenger.Coins += command.ChallengerReturn
	opponent.Coins += command.OpponentReturn
	r.Balances[challengerKey] = challenger
	r.Balances[opponentKey] = opponent
	return domain.CoinGameSettlementResult{Challenger: challenger, Opponent: opponent}, nil
}

func (r *EconomyRepository) coinGameBalancesLocked(command domain.CoinGameCommand) (domain.CoinBalance, domain.CoinBalance, error) {
	opponent, ok := r.Balances[economyBalanceKey(command.GuildID, command.OpponentID)]
	if !ok {
		return domain.CoinBalance{}, domain.CoinBalance{}, ports.ErrCoinGameOpponent
	}
	if opponent.Coins < command.Wager {
		return domain.CoinBalance{}, domain.CoinBalance{}, ports.ErrCoinGameOpponent
	}
	challenger, ok := r.Balances[economyBalanceKey(command.GuildID, command.ChallengerID)]
	if !ok {
		return domain.CoinBalance{}, domain.CoinBalance{}, ports.ErrCoinGameChallenger
	}
	if challenger.Coins < command.Wager {
		return domain.CoinBalance{}, domain.CoinBalance{}, ports.ErrCoinGameChallenger
	}
	return challenger, opponent, nil
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

func (r *EconomyRepository) ListShopItems(ctx context.Context, guildID string) ([]domain.ShopItem, error) {
	if err := r.ready(ctx); err != nil {
		return nil, err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	items := []domain.ShopItem{}
	for _, key := range r.shopItemOrder {
		item := r.ShopItems[key]
		if item.GuildID == guildID {
			items = append(items, item)
		}
	}
	return items, nil
}

func (r *EconomyRepository) GetShopItem(ctx context.Context, guildID string, commodityID int64) (domain.ShopItem, error) {
	if err := r.ready(ctx); err != nil {
		return domain.ShopItem{}, err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	item, ok := r.ShopItems[economyShopItemKey(guildID, commodityID)]
	if !ok {
		return domain.ShopItem{}, ports.ErrShopItemMissing
	}
	return item, nil
}

func (r *EconomyRepository) CreateShopItem(ctx context.Context, item domain.ShopItem) (domain.ShopItem, error) {
	if err := r.ready(ctx); err != nil {
		return domain.ShopItem{}, err
	}
	item = item.Normalize()
	if err := item.Validate(); err != nil {
		return domain.ShopItem{}, err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	key := economyShopItemKey(item.GuildID, item.CommodityID)
	if _, ok := r.ShopItems[key]; ok {
		return domain.ShopItem{}, ports.ErrShopItemExists
	}
	r.ShopItems[key] = item
	r.shopItemOrder = append(r.shopItemOrder, key)
	return item, nil
}

func (r *EconomyRepository) DeleteShopItem(ctx context.Context, guildID string, commodityID int64) (domain.ShopItem, error) {
	if err := r.ready(ctx); err != nil {
		return domain.ShopItem{}, err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	key := economyShopItemKey(guildID, commodityID)
	item, ok := r.ShopItems[key]
	if !ok {
		return domain.ShopItem{}, ports.ErrShopItemMissing
	}
	delete(r.ShopItems, key)
	for index, candidate := range r.shopItemOrder {
		if candidate == key {
			r.shopItemOrder = append(r.shopItemOrder[:index], r.shopItemOrder[index+1:]...)
			break
		}
	}
	return item, nil
}

func (r *EconomyRepository) PurchaseShopItem(ctx context.Context, command domain.ShopPurchaseCommand) (domain.ShopPurchaseResult, error) {
	if err := r.ready(ctx); err != nil {
		return domain.ShopPurchaseResult{}, err
	}
	command = command.Normalize()
	if err := command.Validate(); err != nil {
		return domain.ShopPurchaseResult{}, err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	itemKey := economyShopItemKey(command.GuildID, command.CommodityID)
	item, ok := r.ShopItems[itemKey]
	if !ok {
		return domain.ShopPurchaseResult{}, ports.ErrShopItemMissing
	}
	if command.Quantity > item.Count || (item.RoleID != "" && command.Quantity > 1) {
		return domain.ShopPurchaseResult{}, ports.ErrShopQuantityInvalid
	}
	balanceKey := economyBalanceKey(command.GuildID, command.UserID)
	balance, ok := r.Balances[balanceKey]
	if !ok {
		return domain.ShopPurchaseResult{}, ports.ErrShopInsufficientCoin
	}
	totalCost, ok := item.PurchaseCost(command.Quantity)
	if !ok {
		return domain.ShopPurchaseResult{}, ports.ErrShopInsufficientCoin
	}
	if balance.Coins < totalCost {
		return domain.ShopPurchaseResult{}, ports.ErrShopInsufficientCoin
	}
	purchased := item
	if item.AutoDelete {
		if item.Count == command.Quantity {
			delete(r.ShopItems, itemKey)
			for index, candidate := range r.shopItemOrder {
				if candidate == itemKey {
					r.shopItemOrder = append(r.shopItemOrder[:index], r.shopItemOrder[index+1:]...)
					break
				}
			}
		} else {
			item.Count -= command.Quantity
			r.ShopItems[itemKey] = item
		}
	}
	previous := balance.Coins
	balance.Coins -= totalCost
	r.Balances[balanceKey] = balance
	return domain.ShopPurchaseResult{
		Item:            purchased,
		Quantity:        command.Quantity,
		TotalCost:       totalCost,
		PreviousBalance: previous,
		Balance:         balance,
	}, nil
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

func economyShopItemKey(guildID string, commodityID int64) string {
	return guildID + "\x00" + strconv.FormatInt(commodityID, 10)
}

var _ ports.EconomyRepository = (*EconomyRepository)(nil)
var _ ports.EconomySignInRepository = (*EconomyRepository)(nil)
var _ ports.EconomyCoinRankRepository = (*EconomyRepository)(nil)
var _ ports.EconomySettingsRepository = (*EconomyRepository)(nil)
var _ ports.EconomyCoinAdminRepository = (*EconomyRepository)(nil)
var _ ports.EconomyCoinResetRepository = (*EconomyRepository)(nil)
var _ ports.EconomyRockPaperScissorsRepository = (*EconomyRepository)(nil)
var _ ports.EconomyCoinGameRepository = (*EconomyRepository)(nil)
var _ ports.EconomyShopRepository = (*EconomyRepository)(nil)
var _ ports.TextXPCoinRewardRepository = (*EconomyRepository)(nil)
var _ ports.VoiceXPCoinRewardRepository = (*EconomyRepository)(nil)

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
