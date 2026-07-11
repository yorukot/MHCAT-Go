package repositories

import (
	"context"
	"fmt"
	"strings"
	"time"

	mhcatmongo "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/adapters/mongo"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/adapters/mongo/documents"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
	drivermongo "go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const (
	WorkSetCollectionName       = "work_sets"
	WorkUserCollectionName      = "work_users"
	dailyResetWorkBatchSize     = 25
	dailyResetWorkBatchInterval = 100 * time.Millisecond
)

type DailyResetRepository struct {
	coins       *drivermongo.Collection
	giftChanges *drivermongo.Collection
	workSets    *drivermongo.Collection
	workUsers   *drivermongo.Collection
}

type workSetDocument struct {
	Guild     string        `bson:"guild"`
	GetEnergy bson.RawValue `bson:"get_energy"`
	MaxEnergy bson.RawValue `bson:"max_energy"`
}

func NewDailyResetRepository(coins *drivermongo.Collection, giftChanges *drivermongo.Collection, workSets *drivermongo.Collection, workUsers *drivermongo.Collection) (*DailyResetRepository, error) {
	if coins == nil {
		return nil, fmt.Errorf("coins collection is required")
	}
	if giftChanges == nil {
		return nil, fmt.Errorf("gift_changes collection is required")
	}
	if workSets == nil {
		return nil, fmt.Errorf("work_sets collection is required")
	}
	if workUsers == nil {
		return nil, fmt.Errorf("work_users collection is required")
	}
	return &DailyResetRepository{coins: coins, giftChanges: giftChanges, workSets: workSets, workUsers: workUsers}, nil
}

func NewDailyResetRepositoryFromDatabase(database *drivermongo.Database) (*DailyResetRepository, error) {
	if database == nil {
		return nil, fmt.Errorf("mongo database is required")
	}
	return NewDailyResetRepository(
		database.Collection(CoinCollectionName),
		database.Collection(GiftChangeCollectionName),
		database.Collection(WorkSetCollectionName),
		database.Collection(WorkUserCollectionName),
	)
}

func (r *DailyResetRepository) PreviewDailyReset(ctx context.Context) (domain.DailyResetResult, error) {
	if err := ctx.Err(); err != nil {
		return domain.DailyResetResult{}, err
	}
	excludedGuilds, err := r.guildsWithRollingSignCooldown(ctx)
	if err != nil {
		return domain.DailyResetResult{}, err
	}
	filter := dailyCoinResetFilter(excludedGuilds)
	coinsMatched, err := r.coins.CountDocuments(ctx, filter)
	if err != nil {
		return domain.DailyResetResult{}, mhcatmongo.MapError(fmt.Errorf("preview daily reset coin today: %w", err))
	}
	result := domain.DailyResetResult{
		ExcludedGuilds: len(excludedGuilds),
		CoinsMatched:   coinsMatched,
	}
	workResult, err := r.previewWorkEnergy(ctx)
	if err != nil {
		return domain.DailyResetResult{}, err
	}
	result.WorkGuilds = workResult.WorkGuilds
	result.WorkEnergyIncrements = workResult.WorkEnergyIncrements
	result.WorkEnergyClamps = workResult.WorkEnergyClamps
	return result, ctx.Err()
}

func (r *DailyResetRepository) RunDailyReset(ctx context.Context) (domain.DailyResetResult, error) {
	if err := ctx.Err(); err != nil {
		return domain.DailyResetResult{}, err
	}
	excludedGuilds, err := r.guildsWithRollingSignCooldown(ctx)
	if err != nil {
		return domain.DailyResetResult{}, err
	}
	filter := dailyCoinResetFilter(excludedGuilds)
	coinResult, err := r.coins.UpdateMany(ctx, filter, bson.D{{Key: "$set", Value: bson.D{{Key: "today", Value: int64(0)}}}})
	if err != nil {
		return domain.DailyResetResult{}, mhcatmongo.MapError(fmt.Errorf("daily reset coin today: %w", err))
	}
	result := domain.DailyResetResult{
		ExcludedGuilds: len(excludedGuilds),
		CoinsMatched:   coinResult.MatchedCount,
		CoinsModified:  coinResult.ModifiedCount,
	}
	workResult, err := r.refillWorkEnergy(ctx)
	if err != nil {
		return domain.DailyResetResult{}, err
	}
	result.WorkGuilds = workResult.WorkGuilds
	result.WorkEnergyIncrements = workResult.WorkEnergyIncrements
	result.WorkEnergyClamps = workResult.WorkEnergyClamps
	return result, ctx.Err()
}

func (r *DailyResetRepository) guildsWithRollingSignCooldown(ctx context.Context) ([]string, error) {
	cursor, err := r.giftChanges.Find(
		ctx,
		bson.D{},
		options.Find().SetProjection(bson.D{{Key: "guild", Value: 1}, {Key: "time", Value: 1}}),
	)
	if err != nil {
		return nil, mhcatmongo.MapError(fmt.Errorf("daily reset list economy configs: %w", err))
	}
	defer cursor.Close(ctx)
	guilds := make([]string, 0)
	seen := map[string]struct{}{}
	for cursor.Next(ctx) {
		var document documents.GiftChangeDocument
		if err := cursor.Decode(&document); err != nil {
			return nil, mhcatmongo.MapError(fmt.Errorf("daily reset decode economy config: %w", err))
		}
		config := document.ToDomain()
		if config.ResetMarker == 0 {
			continue
		}
		guild := strings.TrimSpace(config.GuildID)
		if guild == "" {
			continue
		}
		if _, ok := seen[guild]; ok {
			continue
		}
		seen[guild] = struct{}{}
		guilds = append(guilds, guild)
	}
	if err := cursor.Err(); err != nil {
		return nil, mhcatmongo.MapError(fmt.Errorf("daily reset iterate economy configs: %w", err))
	}
	return guilds, ctx.Err()
}

func (r *DailyResetRepository) refillWorkEnergy(ctx context.Context) (domain.DailyResetResult, error) {
	cursor, err := r.workSets.Find(ctx, bson.D{})
	if err != nil {
		return domain.DailyResetResult{}, mhcatmongo.MapError(fmt.Errorf("daily reset list work settings: %w", err))
	}
	defer cursor.Close(ctx)
	var result domain.DailyResetResult
	for cursor.Next(ctx) {
		var config workSetDocument
		if err := cursor.Decode(&config); err != nil {
			return domain.DailyResetResult{}, mhcatmongo.MapError(fmt.Errorf("daily reset decode work setting: %w", err))
		}
		guild := strings.TrimSpace(config.Guild)
		if guild == "" {
			continue
		}
		if shouldPauseDailyResetWork(result.WorkGuilds) {
			if err := waitForDailyResetBatch(ctx, dailyResetWorkBatchInterval); err != nil {
				return domain.DailyResetResult{}, err
			}
		}
		result.WorkGuilds++
		getEnergy := rawInt64(config.GetEnergy)
		maxEnergy := rawInt64(config.MaxEnergy)
		inc, err := r.workUsers.UpdateMany(ctx,
			bson.D{{Key: "guild", Value: guild}, {Key: "energi", Value: bson.D{{Key: "$lt", Value: maxEnergy}}}},
			bson.D{{Key: "$inc", Value: bson.D{{Key: "energi", Value: getEnergy}}}},
		)
		if err != nil {
			return domain.DailyResetResult{}, mhcatmongo.MapError(fmt.Errorf("daily reset refill work energy for guild %s: %w", guild, err))
		}
		clamp, err := r.workUsers.UpdateMany(ctx,
			bson.D{{Key: "guild", Value: guild}, {Key: "energi", Value: bson.D{{Key: "$gt", Value: maxEnergy}}}},
			bson.D{{Key: "$set", Value: bson.D{{Key: "energi", Value: maxEnergy}}}},
		)
		if err != nil {
			return domain.DailyResetResult{}, mhcatmongo.MapError(fmt.Errorf("daily reset clamp work energy for guild %s: %w", guild, err))
		}
		result.WorkEnergyIncrements += inc.ModifiedCount
		result.WorkEnergyClamps += clamp.ModifiedCount
	}
	if err := cursor.Err(); err != nil {
		return domain.DailyResetResult{}, mhcatmongo.MapError(fmt.Errorf("daily reset iterate work settings: %w", err))
	}
	return result, ctx.Err()
}

func (r *DailyResetRepository) previewWorkEnergy(ctx context.Context) (domain.DailyResetResult, error) {
	cursor, err := r.workSets.Find(ctx, bson.D{})
	if err != nil {
		return domain.DailyResetResult{}, mhcatmongo.MapError(fmt.Errorf("daily reset preview list work settings: %w", err))
	}
	defer cursor.Close(ctx)
	var result domain.DailyResetResult
	for cursor.Next(ctx) {
		var config workSetDocument
		if err := cursor.Decode(&config); err != nil {
			return domain.DailyResetResult{}, mhcatmongo.MapError(fmt.Errorf("daily reset preview decode work setting: %w", err))
		}
		guild := strings.TrimSpace(config.Guild)
		if guild == "" {
			continue
		}
		result.WorkGuilds++
		maxEnergy := rawInt64(config.MaxEnergy)
		increments, err := r.workUsers.CountDocuments(ctx, bson.D{{Key: "guild", Value: guild}, {Key: "energi", Value: bson.D{{Key: "$lt", Value: maxEnergy}}}})
		if err != nil {
			return domain.DailyResetResult{}, mhcatmongo.MapError(fmt.Errorf("daily reset preview refill work energy for guild %s: %w", guild, err))
		}
		clamps, err := r.workUsers.CountDocuments(ctx, bson.D{{Key: "guild", Value: guild}, {Key: "energi", Value: bson.D{{Key: "$gt", Value: maxEnergy}}}})
		if err != nil {
			return domain.DailyResetResult{}, mhcatmongo.MapError(fmt.Errorf("daily reset preview clamp work energy for guild %s: %w", guild, err))
		}
		result.WorkEnergyIncrements += increments
		result.WorkEnergyClamps += clamps
	}
	if err := cursor.Err(); err != nil {
		return domain.DailyResetResult{}, mhcatmongo.MapError(fmt.Errorf("daily reset preview iterate work settings: %w", err))
	}
	return result, ctx.Err()
}

func dailyCoinResetFilter(excludedGuilds []string) bson.D {
	filter := bson.D{{Key: "today", Value: bson.D{{Key: "$ne", Value: int64(0)}}}}
	if len(excludedGuilds) == 0 {
		return filter
	}
	return append(filter, bson.E{Key: "guild", Value: bson.D{{Key: "$nin", Value: excludedGuilds}}})
}

func waitForDailyResetBatch(ctx context.Context, pause time.Duration) error {
	if pause <= 0 {
		return ctx.Err()
	}
	timer := time.NewTimer(pause)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func shouldPauseDailyResetWork(processed int) bool {
	return processed > 0 && processed%dailyResetWorkBatchSize == 0
}

func rawInt64(value bson.RawValue) int64 {
	if value.Type == 0 || value.Type == bson.TypeNull || value.Type == bson.TypeUndefined {
		return 0
	}
	if parsed, ok := value.AsInt64OK(); ok {
		return parsed
	}
	if parsed, ok := value.DoubleOK(); ok {
		return int64(parsed)
	}
	if text, ok := value.StringValueOK(); ok {
		var out int64
		if _, err := fmt.Sscan(strings.TrimSpace(text), &out); err == nil {
			return out
		}
	}
	return 0
}
