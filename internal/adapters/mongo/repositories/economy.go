package repositories

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	mhcatmongo "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/adapters/mongo"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/adapters/mongo/documents"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	coreeconomy "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/services/economy"
	"go.mongodb.org/mongo-driver/v2/bson"
	drivermongo "go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const (
	CoinCollectionName       = "coins"
	GiftChangeCollectionName = "gift_changes"
	SignListCollectionName   = "sign_lists"
	ShopItemCollectionName   = "ghps"
)

var ErrCoinGameTransactionsRequired = errors.New("coin game transaction runner is required")

type EconomyRepository struct {
	coins                *drivermongo.Collection
	giftChanges          *drivermongo.Collection
	signLists            *drivermongo.Collection
	shopItems            *drivermongo.Collection
	coinGameTransactions ports.TransactionRunner
}

func NewEconomyRepository(coins *drivermongo.Collection, giftChanges *drivermongo.Collection, signLists *drivermongo.Collection, shopItems ...*drivermongo.Collection) (*EconomyRepository, error) {
	if coins == nil {
		return nil, fmt.Errorf("coins collection is required")
	}
	if giftChanges == nil {
		return nil, fmt.Errorf("gift_changes collection is required")
	}
	if signLists == nil {
		return nil, fmt.Errorf("sign_lists collection is required")
	}
	repo := &EconomyRepository{coins: coins, giftChanges: giftChanges, signLists: signLists}
	if len(shopItems) > 0 {
		repo.shopItems = shopItems[0]
	}
	return repo, nil
}

func NewEconomyRepositoryFromDatabase(database *drivermongo.Database) (*EconomyRepository, error) {
	if database == nil {
		return nil, fmt.Errorf("mongo database is required")
	}
	return NewEconomyRepository(database.Collection(CoinCollectionName), database.Collection(GiftChangeCollectionName), database.Collection(SignListCollectionName), database.Collection(ShopItemCollectionName))
}

func (r *EconomyRepository) SetCoinGameTransactionRunner(transactions ports.TransactionRunner) error {
	if r == nil || transactions == nil {
		return ErrCoinGameTransactionsRequired
	}
	r.coinGameTransactions = transactions
	return nil
}

func (r *EconomyRepository) GetCoinBalance(ctx context.Context, guildID string, userID string) (domain.CoinBalance, error) {
	guildID = strings.TrimSpace(guildID)
	userID = strings.TrimSpace(userID)
	if guildID == "" || userID == "" {
		return domain.CoinBalance{}, domain.ErrInvalidEconomyQuery
	}
	var document documents.CoinDocument
	filter := bson.D{{Key: "guild", Value: guildID}, {Key: "member", Value: userID}}
	if err := r.coins.FindOne(ctx, filter).Decode(&document); err != nil {
		if err == drivermongo.ErrNoDocuments {
			return domain.CoinBalance{}, ports.ErrCoinBalanceNotFound
		}
		return domain.CoinBalance{}, mhcatmongo.MapError(fmt.Errorf("get coin balance: %w", err))
	}
	return document.ToDomain(), ctx.Err()
}

func (r *EconomyRepository) ListCoinBalances(ctx context.Context, guildID string) ([]domain.CoinBalance, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	guildID = strings.TrimSpace(guildID)
	if guildID == "" {
		return nil, domain.ErrInvalidEconomyQuery
	}
	cursor, err := r.coins.Find(ctx, bson.D{{Key: "guild", Value: guildID}})
	if err != nil {
		return nil, mhcatmongo.MapError(fmt.Errorf("list coin balances: %w", err))
	}
	defer cursor.Close(ctx)
	balances := []domain.CoinBalance{}
	for cursor.Next(ctx) {
		var document documents.CoinDocument
		if err := cursor.Decode(&document); err != nil {
			return nil, mhcatmongo.MapError(fmt.Errorf("decode coin balance: %w", err))
		}
		balances = append(balances, document.ToDomain())
	}
	if err := cursor.Err(); err != nil {
		return nil, mhcatmongo.MapError(fmt.Errorf("iterate coin balances: %w", err))
	}
	return balances, ctx.Err()
}

func (r *EconomyRepository) GetEconomyConfig(ctx context.Context, guildID string) (domain.EconomyConfig, error) {
	guildID = strings.TrimSpace(guildID)
	if guildID == "" {
		return domain.EconomyConfig{}, domain.ErrInvalidEconomyQuery
	}
	var document documents.GiftChangeDocument
	filter := bson.D{{Key: "guild", Value: guildID}}
	if err := r.giftChanges.FindOne(ctx, filter).Decode(&document); err != nil {
		if err == drivermongo.ErrNoDocuments {
			return domain.EconomyConfig{GuildID: guildID}, ports.ErrEconomyConfigMissing
		}
		return domain.EconomyConfig{}, mhcatmongo.MapError(fmt.Errorf("get economy config: %w", err))
	}
	return document.ToDomain(), ctx.Err()
}

func (r *EconomyRepository) SaveEconomyConfig(ctx context.Context, config domain.EconomyConfig) (domain.EconomyConfig, error) {
	if err := ctx.Err(); err != nil {
		return domain.EconomyConfig{}, err
	}
	config.GuildID = strings.TrimSpace(config.GuildID)
	config.ChannelID = strings.TrimSpace(config.ChannelID)
	if config.GuildID == "" || config.ChannelID == "" {
		return domain.EconomyConfig{}, domain.ErrInvalidEconomySettings
	}
	filter := bson.D{{Key: "guild", Value: config.GuildID}}
	if err := r.giftChanges.FindOneAndDelete(ctx, filter).Err(); err != nil && err != drivermongo.ErrNoDocuments {
		return domain.EconomyConfig{}, mhcatmongo.MapError(fmt.Errorf("replace economy config: %w", err))
	}
	document := bson.D{{Key: "guild", Value: config.GuildID}}
	document = append(document, documents.GiftChangeUpdateFromDomain(config)...)
	if _, err := r.giftChanges.InsertOne(ctx, document); err != nil {
		return domain.EconomyConfig{}, mhcatmongo.MapError(fmt.Errorf("create economy config: %w", err))
	}
	return config, ctx.Err()
}

func (r *EconomyRepository) ApplyTextXPCoinReward(ctx context.Context, guildID string, userID string, level int64) (domain.CoinBalance, error) {
	return r.applyXPCoinReward(ctx, guildID, userID, level, "text", domain.LegacyTextXPCoinReward)
}

func (r *EconomyRepository) ApplyVoiceXPCoinReward(ctx context.Context, guildID string, userID string, level int64) (domain.CoinBalance, error) {
	return r.applyXPCoinReward(ctx, guildID, userID, level, "voice", domain.LegacyVoiceXPCoinReward)
}

func (r *EconomyRepository) applyXPCoinReward(ctx context.Context, guildID string, userID string, level int64, label string, rewardForLevel func(int64, float64) int64) (domain.CoinBalance, error) {
	if err := ctx.Err(); err != nil {
		return domain.CoinBalance{}, err
	}
	guildID = strings.TrimSpace(guildID)
	userID = strings.TrimSpace(userID)
	if guildID == "" || userID == "" {
		return domain.CoinBalance{}, domain.ErrInvalidEconomyQuery
	}
	config, err := r.GetEconomyConfig(ctx, guildID)
	if err != nil {
		if !errors.Is(err, ports.ErrEconomyConfigMissing) {
			return domain.CoinBalance{}, err
		}
		config = domain.EconomyConfig{GuildID: guildID}
	}
	reward := rewardForLevel(level, config.XPMultiple)
	current, err := r.GetCoinBalance(ctx, guildID, userID)
	if err != nil {
		if !errors.Is(err, ports.ErrCoinBalanceNotFound) {
			return domain.CoinBalance{}, err
		}
		balance := domain.CoinBalance{GuildID: guildID, UserID: userID, Coins: reward, Today: 0}
		if _, err := r.coins.InsertOne(ctx, bson.D{
			{Key: "guild", Value: balance.GuildID},
			{Key: "member", Value: balance.UserID},
			{Key: "coin", Value: balance.Coins},
			{Key: "today", Value: balance.Today},
		}); err != nil {
			return domain.CoinBalance{}, mhcatmongo.MapError(fmt.Errorf("create %s xp coin reward balance: %w", label, err))
		}
		return balance, ctx.Err()
	}
	current.Coins += reward
	result, err := r.coins.UpdateMany(
		ctx,
		bson.D{{Key: "guild", Value: guildID}, {Key: "member", Value: userID}},
		bson.D{{Key: "$set", Value: bson.D{{Key: "coin", Value: current.Coins}}}},
	)
	if err != nil {
		return domain.CoinBalance{}, mhcatmongo.MapError(fmt.Errorf("apply %s xp coin reward: %w", label, err))
	}
	if result.MatchedCount == 0 {
		return domain.CoinBalance{}, ports.ErrCoinBalanceNotFound
	}
	return current, ctx.Err()
}

func (r *EconomyRepository) AdjustCoinBalance(ctx context.Context, command domain.CoinAdminCommand) (domain.CoinAdminResult, error) {
	if err := ctx.Err(); err != nil {
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
	current, err := r.GetCoinBalance(ctx, command.GuildID, command.UserID)
	if err != nil {
		if errors.Is(err, ports.ErrCoinBalanceNotFound) && command.Operation == domain.CoinAdminOperationAdd {
			balance := domain.CoinBalance{
				GuildID: command.GuildID,
				UserID:  command.UserID,
				Coins:   command.Amount,
				Today:   1,
			}
			if _, err := r.coins.InsertOne(ctx, bson.D{
				{Key: "guild", Value: balance.GuildID},
				{Key: "member", Value: balance.UserID},
				{Key: "coin", Value: balance.Coins},
				{Key: "today", Value: balance.Today},
			}); err != nil {
				return domain.CoinAdminResult{}, mhcatmongo.MapError(fmt.Errorf("create coin admin balance: %w", err))
			}
			return domain.CoinAdminResult{Balance: balance, Delta: delta, Created: true}, ctx.Err()
		}
		if errors.Is(err, ports.ErrCoinBalanceNotFound) && command.Operation == domain.CoinAdminOperationReduce {
			return domain.CoinAdminResult{}, ports.ErrCoinNegativeBalance
		}
		return domain.CoinAdminResult{}, err
	}
	currentCoins, numeric := coreeconomy.LegacyEconomyNumber(current.CoinsText)
	if strings.TrimSpace(current.CoinsText) == "" {
		currentCoins = float64(current.Coins)
		numeric = true
	}
	if !numeric {
		return domain.CoinAdminResult{}, domain.ErrInvalidCoinAdminCommand
	}
	next := currentCoins + float64(delta)
	if command.Operation == domain.CoinAdminOperationReduce && next < 0 {
		return domain.CoinAdminResult{}, ports.ErrCoinNegativeBalance
	}
	if command.Operation == domain.CoinAdminOperationAdd && next > float64(coreeconomy.MaxLegacyCoinBalance) {
		return domain.CoinAdminResult{}, ports.ErrCoinLimitExceeded
	}
	result, err := r.coins.UpdateOne(
		ctx,
		bson.D{{Key: "guild", Value: command.GuildID}, {Key: "member", Value: command.UserID}},
		bson.D{{Key: "$set", Value: bson.D{{Key: "coin", Value: next}}}},
	)
	if err != nil {
		return domain.CoinAdminResult{}, mhcatmongo.MapError(fmt.Errorf("adjust coin admin balance: %w", err))
	}
	if result.MatchedCount == 0 {
		return domain.CoinAdminResult{}, ports.ErrCoinBalanceNotFound
	}
	current.Coins = int64(next)
	current.CoinsText = coreeconomy.LegacyEconomyNumberText(next)
	return domain.CoinAdminResult{Balance: current, Delta: delta}, ctx.Err()
}

func (r *EconomyRepository) ResetCoinBalances(ctx context.Context, command domain.CoinResetCommand) (domain.CoinResetResult, error) {
	if err := ctx.Err(); err != nil {
		return domain.CoinResetResult{}, err
	}
	command = command.Normalize()
	if err := command.Validate(); err != nil {
		return domain.CoinResetResult{}, err
	}
	filter := bson.D{{Key: "guild", Value: command.GuildID}}
	if command.Divisor == 0 {
		result, err := r.coins.DeleteMany(ctx, filter)
		if err != nil {
			return domain.CoinResetResult{}, mhcatmongo.MapError(fmt.Errorf("reset coin balances: %w", err))
		}
		if result.DeletedCount == 0 {
			return domain.CoinResetResult{}, ports.ErrCoinBalanceNotFound
		}
		return domain.CoinResetResult{GuildID: command.GuildID, AffectedCount: result.DeletedCount, Deleted: true}, ctx.Err()
	}
	cursor, err := r.coins.Find(ctx, filter)
	if err != nil {
		return domain.CoinResetResult{}, mhcatmongo.MapError(fmt.Errorf("list coin balances for reset: %w", err))
	}
	defer cursor.Close(ctx)
	affected := int64(0)
	for cursor.Next(ctx) {
		var document documents.CoinDocument
		if err := cursor.Decode(&document); err != nil {
			return domain.CoinResetResult{}, mhcatmongo.MapError(fmt.Errorf("decode coin balance for reset: %w", err))
		}
		balance := document.ToDomain()
		coins, numeric := coreeconomy.LegacyEconomyNumber(balance.CoinsText)
		if !numeric {
			return domain.CoinResetResult{}, domain.ErrInvalidCoinResetCommand
		}
		next := domain.LegacyJavaScriptRoundNumber(coins / float64(command.Divisor))
		updateFilter := bson.D{{Key: "guild", Value: document.Guild}, {Key: "member", Value: document.Member}}
		result, err := r.coins.UpdateOne(ctx, updateFilter, bson.D{{Key: "$set", Value: bson.D{{Key: "coin", Value: next}}}})
		if err != nil {
			return domain.CoinResetResult{}, mhcatmongo.MapError(fmt.Errorf("divide coin balance: %w", err))
		}
		if result.MatchedCount > 0 {
			affected++
		}
	}
	if err := cursor.Err(); err != nil {
		return domain.CoinResetResult{}, mhcatmongo.MapError(fmt.Errorf("iterate coin balances for reset: %w", err))
	}
	if affected == 0 {
		return domain.CoinResetResult{}, ports.ErrCoinBalanceNotFound
	}
	return domain.CoinResetResult{
		GuildID:       command.GuildID,
		Divisor:       command.Divisor,
		AffectedCount: affected,
		Deleted:       false,
	}, ctx.Err()
}

func (r *EconomyRepository) ApplyRockPaperScissors(ctx context.Context, command domain.RockPaperScissorsCommand) (domain.RockPaperScissorsResult, error) {
	if err := ctx.Err(); err != nil {
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
	current, err := r.GetCoinBalance(ctx, command.GuildID, command.UserID)
	if err != nil {
		return domain.RockPaperScissorsResult{}, err
	}
	currentCoins, numeric := coreeconomy.LegacyEconomyNumber(current.CoinsText)
	if strings.TrimSpace(current.CoinsText) == "" {
		currentCoins = float64(current.Coins)
		numeric = true
	}
	if !numeric {
		return domain.RockPaperScissorsResult{}, domain.ErrInvalidRockPaperScissorsCommand
	}
	if currentCoins-float64(command.Wager) < 0 {
		return domain.RockPaperScissorsResult{}, ports.ErrCoinNegativeBalance
	}
	next := currentCoins + float64(delta)
	result, err := r.coins.UpdateOne(
		ctx,
		bson.D{{Key: "guild", Value: command.GuildID}, {Key: "member", Value: command.UserID}},
		bson.D{{Key: "$set", Value: bson.D{{Key: "coin", Value: next}}}},
	)
	if err != nil {
		return domain.RockPaperScissorsResult{}, mhcatmongo.MapError(fmt.Errorf("apply rock paper scissors balance: %w", err))
	}
	if result.MatchedCount == 0 {
		return domain.RockPaperScissorsResult{}, ports.ErrCoinBalanceNotFound
	}
	previous := currentCoins
	current.Coins = int64(next)
	current.CoinsText = coreeconomy.LegacyEconomyNumberText(next)
	return domain.RockPaperScissorsResult{
		Balance:         current,
		PreviousBalance: previous,
		Delta:           delta,
		Outcome:         outcome,
		PlayerChoice:    command.PlayerChoice,
		ComputerChoice:  command.ComputerChoice,
	}, ctx.Err()
}

func (r *EconomyRepository) CheckCoinGameBalances(ctx context.Context, command domain.CoinGameCommand) (domain.CoinGameBalanceResult, error) {
	if err := ctx.Err(); err != nil {
		return domain.CoinGameBalanceResult{}, err
	}
	command = command.Normalize()
	if err := command.Validate(); err != nil {
		return domain.CoinGameBalanceResult{}, err
	}
	challenger, opponent, err := r.coinGameBalances(ctx, command)
	if err != nil {
		return domain.CoinGameBalanceResult{}, err
	}
	return domain.CoinGameBalanceResult{Challenger: challenger, Opponent: opponent, Wager: command.Wager}, ctx.Err()
}

func (r *EconomyRepository) ReserveCoinGameWager(ctx context.Context, command domain.CoinGameCommand) (domain.CoinGameBalanceResult, error) {
	if err := ctx.Err(); err != nil {
		return domain.CoinGameBalanceResult{}, err
	}
	command = command.Normalize()
	if err := command.Validate(); err != nil {
		return domain.CoinGameBalanceResult{}, err
	}
	if r.coinGameTransactions == nil {
		return domain.CoinGameBalanceResult{}, ErrCoinGameTransactionsRequired
	}
	var reserved domain.CoinGameBalanceResult
	err := r.coinGameTransactions.WithinTransaction(ctx, func(txCtx context.Context) error {
		challenger, opponent, err := r.coinGameBalances(txCtx, command)
		if err != nil {
			return err
		}
		if err := applyCoinGameBalanceDelta(&challenger, -command.Wager); err != nil {
			return err
		}
		if err := applyCoinGameBalanceDelta(&opponent, -command.Wager); err != nil {
			return err
		}
		if err := r.setCoinGameBalance(txCtx, challenger); err != nil {
			return err
		}
		if err := r.setCoinGameBalance(txCtx, opponent); err != nil {
			return err
		}
		reserved = domain.CoinGameBalanceResult{Challenger: challenger, Opponent: opponent, Wager: command.Wager}
		return txCtx.Err()
	})
	if err != nil {
		return domain.CoinGameBalanceResult{}, err
	}
	return reserved, ctx.Err()
}

func (r *EconomyRepository) SettleCoinGameWager(ctx context.Context, command domain.CoinGameSettlementCommand) (domain.CoinGameSettlementResult, error) {
	if err := ctx.Err(); err != nil {
		return domain.CoinGameSettlementResult{}, err
	}
	command = command.Normalize()
	if err := command.Validate(); err != nil {
		return domain.CoinGameSettlementResult{}, err
	}
	if r.coinGameTransactions == nil {
		return domain.CoinGameSettlementResult{}, ErrCoinGameTransactionsRequired
	}
	var settled domain.CoinGameSettlementResult
	err := r.coinGameTransactions.WithinTransaction(ctx, func(txCtx context.Context) error {
		challenger, err := r.GetCoinBalance(txCtx, command.GuildID, command.ChallengerID)
		if err != nil {
			return err
		}
		opponent, err := r.GetCoinBalance(txCtx, command.GuildID, command.OpponentID)
		if err != nil {
			return err
		}
		if err := applyCoinGameBalanceDelta(&challenger, command.ChallengerReturn); err != nil {
			return err
		}
		if err := applyCoinGameBalanceDelta(&opponent, command.OpponentReturn); err != nil {
			return err
		}
		if err := r.setCoinGameBalance(txCtx, challenger); err != nil {
			return err
		}
		if err := r.setCoinGameBalance(txCtx, opponent); err != nil {
			return err
		}
		settled = domain.CoinGameSettlementResult{Challenger: challenger, Opponent: opponent}
		return txCtx.Err()
	})
	if err != nil {
		return domain.CoinGameSettlementResult{}, err
	}
	return settled, ctx.Err()
}

func (r *EconomyRepository) coinGameBalances(ctx context.Context, command domain.CoinGameCommand) (domain.CoinBalance, domain.CoinBalance, error) {
	opponent, err := r.GetCoinBalance(ctx, command.GuildID, command.OpponentID)
	if err != nil {
		if errors.Is(err, ports.ErrCoinBalanceNotFound) {
			return domain.CoinBalance{}, domain.CoinBalance{}, ports.ErrCoinGameOpponent
		}
		return domain.CoinBalance{}, domain.CoinBalance{}, err
	}
	opponentCoins, ok := coinGameBalanceNumber(opponent)
	if !ok {
		return domain.CoinBalance{}, domain.CoinBalance{}, domain.ErrInvalidCoinGameCommand
	}
	if opponentCoins < float64(command.Wager) {
		return domain.CoinBalance{}, domain.CoinBalance{}, ports.ErrCoinGameOpponent
	}
	challenger, err := r.GetCoinBalance(ctx, command.GuildID, command.ChallengerID)
	if err != nil {
		if errors.Is(err, ports.ErrCoinBalanceNotFound) {
			return domain.CoinBalance{}, domain.CoinBalance{}, ports.ErrCoinGameChallenger
		}
		return domain.CoinBalance{}, domain.CoinBalance{}, err
	}
	challengerCoins, ok := coinGameBalanceNumber(challenger)
	if !ok {
		return domain.CoinBalance{}, domain.CoinBalance{}, domain.ErrInvalidCoinGameCommand
	}
	if challengerCoins < float64(command.Wager) {
		return domain.CoinBalance{}, domain.CoinBalance{}, ports.ErrCoinGameChallenger
	}
	return challenger, opponent, nil
}

func (r *EconomyRepository) setCoinGameBalance(ctx context.Context, balance domain.CoinBalance) error {
	coins, ok := coinGameBalanceNumber(balance)
	if !ok {
		return domain.ErrInvalidCoinGameCommand
	}
	result, err := r.coins.UpdateOne(
		ctx,
		bson.D{{Key: "guild", Value: balance.GuildID}, {Key: "member", Value: balance.UserID}},
		bson.D{{Key: "$set", Value: bson.D{{Key: "coin", Value: coins}}}},
	)
	if err != nil {
		return mhcatmongo.MapError(fmt.Errorf("update coin game balance: %w", err))
	}
	if result.MatchedCount == 0 {
		return ports.ErrCoinBalanceNotFound
	}
	return ctx.Err()
}

func coinGameBalanceNumber(balance domain.CoinBalance) (float64, bool) {
	if strings.TrimSpace(balance.CoinsText) == "" {
		return float64(balance.Coins), true
	}
	return coreeconomy.LegacyEconomyNumber(balance.CoinsText)
}

func applyCoinGameBalanceDelta(balance *domain.CoinBalance, delta int64) error {
	coins, ok := coinGameBalanceNumber(*balance)
	if !ok {
		return domain.ErrInvalidCoinGameCommand
	}
	next := coins + float64(delta)
	balance.Coins = int64(next)
	balance.CoinsText = coreeconomy.LegacyEconomyNumberText(next)
	return nil
}

func (r *EconomyRepository) SignIn(ctx context.Context, command domain.SignInCommand) (domain.SignInResult, error) {
	if err := ctx.Err(); err != nil {
		return domain.SignInResult{}, err
	}
	command.GuildID = strings.TrimSpace(command.GuildID)
	command.UserID = strings.TrimSpace(command.UserID)
	command.Year = strings.TrimSpace(command.Year)
	command.Month = strings.TrimSpace(command.Month)
	command.Day = strings.TrimSpace(command.Day)
	if command.GuildID == "" || command.UserID == "" || command.Year == "" || command.Month == "" || command.Day == "" || command.Now.IsZero() {
		return domain.SignInResult{}, domain.ErrInvalidSignIn
	}
	config, err := r.GetEconomyConfig(ctx, command.GuildID)
	configFound := true
	if err != nil {
		if !errors.Is(err, ports.ErrEconomyConfigMissing) {
			return domain.SignInResult{}, err
		}
		config = domain.EconomyConfig{
			GuildID:     command.GuildID,
			SignCoins:   coreeconomy.DefaultSignCoins,
			ResetMarker: 0,
		}
		configFound = false
	}
	reward, rewardValid := coreeconomy.LegacySignReward(config, configFound)
	if !rewardValid {
		return domain.SignInResult{}, domain.ErrInvalidSignIn
	}
	nowUnix := coreeconomy.LegacyRoundedUnixSeconds(command.Now)
	todayValue := int64(1)
	filter := signInDailyFilter(command.GuildID, command.UserID, reward)
	rollingWindow, cooldown := coreeconomy.LegacySignWindow(config, configFound)
	if rollingWindow {
		todayValue = nowUnix
		filter = signInRollingFilter(command.GuildID, command.UserID, reward, float64(nowUnix)-cooldown)
	}
	update := signInUpdate(reward, todayValue)
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var updated documents.CoinDocument
	var balance domain.CoinBalance
	err = r.coins.FindOneAndUpdate(ctx, filter, update, opts).Decode(&updated)
	if err != nil {
		if err == drivermongo.ErrNoDocuments {
			firstTodayValue := legacyFirstSignToday(configFound, nowUnix)
			created, createErr := r.createFirstSignInBalance(ctx, command, reward, firstTodayValue)
			if createErr != nil {
				return domain.SignInResult{}, createErr
			}
			balance = created
		} else {
			return domain.SignInResult{}, mhcatmongo.MapError(fmt.Errorf("sign in coin update: %w", err))
		}
	} else {
		balance = updated.ToDomain()
	}
	if err := r.addSignCalendarDay(ctx, command); err != nil {
		return domain.SignInResult{}, err
	}
	calendar, err := r.GetSignCalendar(ctx, command.GuildID, command.UserID, command.Year, command.Month)
	if err != nil {
		return domain.SignInResult{}, err
	}
	return domain.SignInResult{
		Balance:     balance,
		Config:      config,
		Calendar:    calendar,
		Reward:      reward,
		ConfigFound: configFound,
		SignedAt:    command.Now,
	}, ctx.Err()
}

func legacyFirstSignToday(configFound bool, nowUnix int64) int64 {
	if configFound {
		return nowUnix
	}
	return 1
}

func signInUpdate(reward float64, todayValue int64) drivermongo.Pipeline {
	return drivermongo.Pipeline{{{Key: "$set", Value: bson.D{
		{Key: "coin", Value: bson.D{{Key: "$add", Value: bson.A{
			bson.D{{Key: "$ifNull", Value: bson.A{"$coin", 0}}},
			reward,
		}}}},
		{Key: "today", Value: todayValue},
	}}}}
}

func (r *EconomyRepository) GetSignCalendar(ctx context.Context, guildID string, userID string, year string, month string) (domain.SignCalendar, error) {
	if err := ctx.Err(); err != nil {
		return domain.SignCalendar{}, err
	}
	guildID = strings.TrimSpace(guildID)
	userID = strings.TrimSpace(userID)
	if guildID == "" || userID == "" {
		return domain.SignCalendar{}, domain.ErrInvalidSignIn
	}
	var document documents.SignListDocument
	err := r.signLists.FindOne(ctx, bson.D{{Key: "guild", Value: guildID}, {Key: "member", Value: userID}}).Decode(&document)
	if err != nil {
		if err == drivermongo.ErrNoDocuments {
			return domain.SignCalendar{GuildID: guildID, UserID: userID, Date: map[string]map[string][]string{}}, nil
		}
		return domain.SignCalendar{}, mhcatmongo.MapError(fmt.Errorf("get sign calendar: %w", err))
	}
	return document.ToDomain(), ctx.Err()
}

func signInDailyFilter(guildID string, userID string, reward float64) bson.D {
	return bson.D{
		{Key: "guild", Value: guildID},
		{Key: "member", Value: userID},
		{Key: "$and", Value: bson.A{
			bson.D{{Key: "$or", Value: bson.A{
				bson.D{{Key: "today", Value: bson.D{{Key: "$ne", Value: int64(1)}}}},
				bson.D{{Key: "today", Value: bson.D{{Key: "$exists", Value: false}}}},
			}}},
			coinLimitFilter(reward),
		}},
	}
}

func signInRollingFilter(guildID string, userID string, reward float64, eligibleBefore float64) bson.D {
	return bson.D{
		{Key: "guild", Value: guildID},
		{Key: "member", Value: userID},
		{Key: "$and", Value: bson.A{
			bson.D{{Key: "$or", Value: bson.A{
				bson.D{{Key: "today", Value: bson.D{{Key: "$lte", Value: eligibleBefore}}}},
				bson.D{{Key: "today", Value: bson.D{{Key: "$exists", Value: false}}}},
				bson.D{{Key: "today", Value: bson.D{{Key: "$type", Value: "null"}}}},
				bson.D{{Key: "today", Value: 0}},
			}}},
			coinLimitFilter(reward),
		}},
	}
}

func coinLimitFilter(reward float64) bson.D {
	return bson.D{{Key: "$or", Value: bson.A{
		bson.D{{Key: "coin", Value: bson.D{{Key: "$lte", Value: float64(coreeconomy.MaxLegacyCoinBalance) - reward}}}},
		bson.D{{Key: "coin", Value: bson.D{{Key: "$type", Value: "null"}}}},
	}}}
}

func (r *EconomyRepository) signInMissReason(ctx context.Context, guildID string, userID string, reward float64) error {
	balance, err := r.GetCoinBalance(ctx, guildID, userID)
	if err != nil {
		if errors.Is(err, ports.ErrCoinBalanceNotFound) {
			return err
		}
		return err
	}
	coins, numeric := coreeconomy.LegacyEconomyNumber(balance.CoinsText)
	if strings.TrimSpace(balance.CoinsText) == "" {
		coins = float64(balance.Coins)
		numeric = true
	}
	if !numeric {
		return domain.ErrInvalidSignIn
	}
	if numeric && coins+reward > float64(coreeconomy.MaxLegacyCoinBalance) {
		return ports.ErrCoinLimitExceeded
	}
	return ports.ErrAlreadySignedIn
}

func (r *EconomyRepository) createFirstSignInBalance(ctx context.Context, command domain.SignInCommand, reward float64, todayValue int64) (domain.CoinBalance, error) {
	if err := r.signInMissReason(ctx, command.GuildID, command.UserID, reward); err != nil && !errors.Is(err, ports.ErrCoinBalanceNotFound) {
		return domain.CoinBalance{}, err
	}
	balance := domain.CoinBalance{
		GuildID:   command.GuildID,
		UserID:    command.UserID,
		Coins:     int64(reward),
		CoinsText: strconv.FormatFloat(reward, 'f', -1, 64),
		Today:     todayValue,
	}
	document := bson.D{
		{Key: "guild", Value: balance.GuildID},
		{Key: "member", Value: balance.UserID},
		{Key: "coin", Value: reward},
		{Key: "today", Value: balance.Today},
	}
	if _, err := r.coins.InsertOne(ctx, document); err != nil {
		return domain.CoinBalance{}, mhcatmongo.MapError(fmt.Errorf("create sign in coin balance: %w", err))
	}
	return balance, ctx.Err()
}

func (r *EconomyRepository) addSignCalendarDay(ctx context.Context, command domain.SignInCommand) error {
	field := "date." + command.Year + "." + command.Month
	update := bson.D{
		{Key: "$setOnInsert", Value: bson.D{{Key: "guild", Value: command.GuildID}, {Key: "member", Value: command.UserID}}},
		{Key: "$addToSet", Value: bson.D{{Key: field, Value: command.Day}}},
	}
	_, err := r.signLists.UpdateOne(ctx, bson.D{{Key: "guild", Value: command.GuildID}, {Key: "member", Value: command.UserID}}, update, options.UpdateOne().SetUpsert(true))
	if err != nil {
		return mhcatmongo.MapError(fmt.Errorf("update sign calendar: %w", err))
	}
	return ctx.Err()
}

func (r *EconomyRepository) ListShopItems(ctx context.Context, guildID string) ([]domain.ShopItem, error) {
	if err := r.shopReady(ctx); err != nil {
		return nil, err
	}
	guildID = strings.TrimSpace(guildID)
	if guildID == "" {
		return nil, domain.ErrInvalidShopItem
	}
	cursor, err := r.shopItems.Find(ctx, bson.D{{Key: "guild", Value: guildID}})
	if err != nil {
		return nil, mhcatmongo.MapError(fmt.Errorf("list shop items: %w", err))
	}
	defer cursor.Close(ctx)
	items := []domain.ShopItem{}
	for cursor.Next(ctx) {
		var document documents.ShopItemDocument
		if err := cursor.Decode(&document); err != nil {
			return nil, mhcatmongo.MapError(fmt.Errorf("decode shop item: %w", err))
		}
		items = append(items, document.ToDomain())
	}
	if err := cursor.Err(); err != nil {
		return nil, mhcatmongo.MapError(fmt.Errorf("iterate shop items: %w", err))
	}
	return items, ctx.Err()
}

func (r *EconomyRepository) GetShopItem(ctx context.Context, guildID string, commodityID int64) (domain.ShopItem, error) {
	if err := r.shopReady(ctx); err != nil {
		return domain.ShopItem{}, err
	}
	guildID = strings.TrimSpace(guildID)
	if guildID == "" || commodityID <= 0 {
		return domain.ShopItem{}, domain.ErrInvalidShopItem
	}
	var document documents.ShopItemDocument
	err := r.shopItems.FindOne(ctx, bson.D{{Key: "guild", Value: guildID}, {Key: "commodity_id", Value: commodityID}}).Decode(&document)
	if err != nil {
		if err == drivermongo.ErrNoDocuments {
			return domain.ShopItem{}, ports.ErrShopItemMissing
		}
		return domain.ShopItem{}, mhcatmongo.MapError(fmt.Errorf("get shop item: %w", err))
	}
	return document.ToDomain(), ctx.Err()
}

func (r *EconomyRepository) CreateShopItem(ctx context.Context, item domain.ShopItem) (domain.ShopItem, error) {
	if err := r.shopReady(ctx); err != nil {
		return domain.ShopItem{}, err
	}
	item = item.Normalize()
	if err := item.Validate(); err != nil {
		return domain.ShopItem{}, err
	}
	filter := bson.D{{Key: "guild", Value: item.GuildID}, {Key: "commodity_id", Value: item.CommodityID}}
	err := r.shopItems.FindOne(ctx, filter).Err()
	if err == nil {
		return domain.ShopItem{}, ports.ErrShopItemExists
	}
	if err != drivermongo.ErrNoDocuments {
		return domain.ShopItem{}, mhcatmongo.MapError(fmt.Errorf("find shop item before create: %w", err))
	}
	if _, err := r.shopItems.InsertOne(ctx, documents.ShopItemWriteDocumentFromDomain(item)); err != nil {
		mapped := mhcatmongo.MapError(fmt.Errorf("create shop item: %w", err))
		if mhcatmongo.ErrorIs(mapped, mhcatmongo.ErrorKindConflict) {
			return domain.ShopItem{}, ports.ErrShopItemExists
		}
		return domain.ShopItem{}, mapped
	}
	return item, ctx.Err()
}

func (r *EconomyRepository) DeleteShopItem(ctx context.Context, guildID string, commodityID int64) (domain.ShopItem, error) {
	if err := r.shopReady(ctx); err != nil {
		return domain.ShopItem{}, err
	}
	guildID = strings.TrimSpace(guildID)
	if guildID == "" || commodityID <= 0 {
		return domain.ShopItem{}, domain.ErrInvalidShopItem
	}
	var document documents.ShopItemDocument
	err := r.shopItems.FindOneAndDelete(ctx, bson.D{{Key: "guild", Value: guildID}, {Key: "commodity_id", Value: commodityID}}).Decode(&document)
	if err != nil {
		if err == drivermongo.ErrNoDocuments {
			return domain.ShopItem{}, ports.ErrShopItemMissing
		}
		return domain.ShopItem{}, mhcatmongo.MapError(fmt.Errorf("delete shop item: %w", err))
	}
	return document.ToDomain(), ctx.Err()
}

func (r *EconomyRepository) PurchaseShopItem(ctx context.Context, command domain.ShopPurchaseCommand) (domain.ShopPurchaseResult, error) {
	if err := r.shopReady(ctx); err != nil {
		return domain.ShopPurchaseResult{}, err
	}
	command = command.Normalize()
	if err := command.Validate(); err != nil {
		return domain.ShopPurchaseResult{}, err
	}
	item, err := r.GetShopItem(ctx, command.GuildID, command.CommodityID)
	if err != nil {
		return domain.ShopPurchaseResult{}, err
	}
	if command.Quantity > item.Count {
		return domain.ShopPurchaseResult{}, ports.ErrShopQuantityInvalid
	}
	balance, err := r.GetCoinBalance(ctx, command.GuildID, command.UserID)
	if err != nil {
		if errors.Is(err, ports.ErrCoinBalanceNotFound) {
			return domain.ShopPurchaseResult{}, ports.ErrShopInsufficientCoin
		}
		return domain.ShopPurchaseResult{}, err
	}
	totalCost, ok := item.PurchaseCost(command.Quantity)
	if !ok {
		return domain.ShopPurchaseResult{}, ports.ErrShopInsufficientCoin
	}
	if balance.Coins < totalCost {
		return domain.ShopPurchaseResult{}, ports.ErrShopInsufficientCoin
	}
	if item.AutoDelete {
		if item.Count == 1 {
			if _, err := r.shopItems.DeleteOne(ctx, bson.D{{Key: "guild", Value: command.GuildID}, {Key: "commodity_id", Value: command.CommodityID}}); err != nil {
				return domain.ShopPurchaseResult{}, mhcatmongo.MapError(fmt.Errorf("delete purchased shop item: %w", err))
			}
		} else {
			nextCount := item.Count - command.Quantity
			if _, err := r.shopItems.UpdateOne(ctx, bson.D{{Key: "guild", Value: command.GuildID}, {Key: "commodity_id", Value: command.CommodityID}}, bson.D{{Key: "$set", Value: bson.D{{Key: "commodity_count", Value: nextCount}}}}); err != nil {
				return domain.ShopPurchaseResult{}, mhcatmongo.MapError(fmt.Errorf("decrement shop item count: %w", err))
			}
		}
	}
	nextBalance := balance.Coins - totalCost
	result, err := r.coins.UpdateMany(ctx, bson.D{{Key: "guild", Value: command.GuildID}, {Key: "member", Value: command.UserID}}, bson.D{{Key: "$set", Value: bson.D{{Key: "coin", Value: nextBalance}}}})
	if err != nil {
		return domain.ShopPurchaseResult{}, mhcatmongo.MapError(fmt.Errorf("subtract shop purchase coins: %w", err))
	}
	if result.MatchedCount == 0 {
		return domain.ShopPurchaseResult{}, ports.ErrCoinBalanceNotFound
	}
	previous := balance.Coins
	balance.Coins = nextBalance
	return domain.ShopPurchaseResult{
		Item:            item,
		Quantity:        command.Quantity,
		TotalCost:       totalCost,
		PreviousBalance: previous,
		Balance:         balance,
	}, ctx.Err()
}

func (r *EconomyRepository) shopReady(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if r.shopItems == nil {
		return fmt.Errorf("shop items collection is required")
	}
	return nil
}

var _ ports.TextXPCoinRewardRepository = (*EconomyRepository)(nil)
var _ ports.VoiceXPCoinRewardRepository = (*EconomyRepository)(nil)
