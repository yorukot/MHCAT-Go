package repositories

import (
	"context"
	"errors"
	"fmt"
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
)

type EconomyRepository struct {
	coins       *drivermongo.Collection
	giftChanges *drivermongo.Collection
	signLists   *drivermongo.Collection
}

func NewEconomyRepository(coins *drivermongo.Collection, giftChanges *drivermongo.Collection, signLists *drivermongo.Collection) (*EconomyRepository, error) {
	if coins == nil {
		return nil, fmt.Errorf("coins collection is required")
	}
	if giftChanges == nil {
		return nil, fmt.Errorf("gift_changes collection is required")
	}
	if signLists == nil {
		return nil, fmt.Errorf("sign_lists collection is required")
	}
	return &EconomyRepository{coins: coins, giftChanges: giftChanges, signLists: signLists}, nil
}

func NewEconomyRepositoryFromDatabase(database *drivermongo.Database) (*EconomyRepository, error) {
	if database == nil {
		return nil, fmt.Errorf("mongo database is required")
	}
	return NewEconomyRepository(database.Collection(CoinCollectionName), database.Collection(GiftChangeCollectionName), database.Collection(SignListCollectionName))
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
	update := bson.D{{Key: "$set", Value: documents.GiftChangeUpdateFromDomain(config)}}
	result, err := r.giftChanges.UpdateMany(ctx, filter, update)
	if err != nil {
		return domain.EconomyConfig{}, mhcatmongo.MapError(fmt.Errorf("update economy config: %w", err))
	}
	if result.MatchedCount == 0 {
		insertUpdate := bson.D{
			{Key: "$set", Value: documents.GiftChangeUpdateFromDomain(config)},
			{Key: "$setOnInsert", Value: bson.D{{Key: "guild", Value: config.GuildID}}},
		}
		if _, err := r.giftChanges.UpdateOne(ctx, filter, insertUpdate, options.UpdateOne().SetUpsert(true)); err != nil {
			return domain.EconomyConfig{}, mhcatmongo.MapError(fmt.Errorf("create economy config: %w", err))
		}
	}
	return config, ctx.Err()
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
	reward := config.SignCoins
	if !configFound {
		reward = coreeconomy.DefaultSignCoins
	}
	nowUnix := command.Now.Unix()
	todayValue := int64(1)
	filter := signInDailyFilter(command.GuildID, command.UserID, reward)
	if configFound && config.ResetMarker != 0 {
		cooldown := config.ResetMarker
		if cooldown < 0 {
			cooldown = coreeconomy.DefaultSignCooldownSec
		}
		todayValue = nowUnix
		filter = signInRollingFilter(command.GuildID, command.UserID, reward, nowUnix-cooldown)
	}
	update := bson.D{
		{Key: "$inc", Value: bson.D{{Key: "coin", Value: reward}}},
		{Key: "$set", Value: bson.D{{Key: "today", Value: todayValue}}},
	}
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var updated documents.CoinDocument
	var balance domain.CoinBalance
	err = r.coins.FindOneAndUpdate(ctx, filter, update, opts).Decode(&updated)
	if err != nil {
		if err == drivermongo.ErrNoDocuments {
			created, createErr := r.createFirstSignInBalance(ctx, command, reward, todayValue)
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

func signInDailyFilter(guildID string, userID string, reward int64) bson.D {
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

func signInRollingFilter(guildID string, userID string, reward int64, eligibleBefore int64) bson.D {
	return bson.D{
		{Key: "guild", Value: guildID},
		{Key: "member", Value: userID},
		{Key: "$and", Value: bson.A{
			bson.D{{Key: "$or", Value: bson.A{
				bson.D{{Key: "today", Value: bson.D{{Key: "$lte", Value: eligibleBefore}}}},
				bson.D{{Key: "today", Value: bson.D{{Key: "$exists", Value: false}}}},
				bson.D{{Key: "today", Value: 0}},
			}}},
			coinLimitFilter(reward),
		}},
	}
}

func coinLimitFilter(reward int64) bson.D {
	return bson.D{{Key: "$or", Value: bson.A{
		bson.D{{Key: "coin", Value: bson.D{{Key: "$lte", Value: coreeconomy.MaxLegacyCoinBalance - reward}}}},
		bson.D{{Key: "coin", Value: bson.D{{Key: "$exists", Value: false}}}},
	}}}
}

func (r *EconomyRepository) signInMissReason(ctx context.Context, guildID string, userID string, reward int64) error {
	balance, err := r.GetCoinBalance(ctx, guildID, userID)
	if err != nil {
		if errors.Is(err, ports.ErrCoinBalanceNotFound) {
			return err
		}
		return err
	}
	if balance.Coins+reward > coreeconomy.MaxLegacyCoinBalance {
		return ports.ErrCoinLimitExceeded
	}
	return ports.ErrAlreadySignedIn
}

func (r *EconomyRepository) createFirstSignInBalance(ctx context.Context, command domain.SignInCommand, reward int64, todayValue int64) (domain.CoinBalance, error) {
	if reward > coreeconomy.MaxLegacyCoinBalance {
		return domain.CoinBalance{}, ports.ErrCoinLimitExceeded
	}
	if err := r.signInMissReason(ctx, command.GuildID, command.UserID, reward); err != nil && !errors.Is(err, ports.ErrCoinBalanceNotFound) {
		return domain.CoinBalance{}, err
	}
	balance := domain.CoinBalance{
		GuildID: command.GuildID,
		UserID:  command.UserID,
		Coins:   reward,
		Today:   todayValue,
	}
	document := bson.D{
		{Key: "guild", Value: balance.GuildID},
		{Key: "member", Value: balance.UserID},
		{Key: "coin", Value: balance.Coins},
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
