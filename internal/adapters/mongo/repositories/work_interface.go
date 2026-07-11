package repositories

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"

	mhcatmongo "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/adapters/mongo"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/adapters/mongo/documents"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	"go.mongodb.org/mongo-driver/v2/bson"
	drivermongo "go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const WorkSomethingCollectionName = "work_somethings"

type WorkInterfaceRepository struct {
	workSets       *drivermongo.Collection
	workSomethings *drivermongo.Collection
	workUsers      *drivermongo.Collection
}

func NewWorkInterfaceRepository(workSets *drivermongo.Collection, workSomethings *drivermongo.Collection, workUsers *drivermongo.Collection) (*WorkInterfaceRepository, error) {
	if workSets == nil {
		return nil, fmt.Errorf("work_sets collection is required")
	}
	if workSomethings == nil {
		return nil, fmt.Errorf("work_somethings collection is required")
	}
	if workUsers == nil {
		return nil, fmt.Errorf("work_users collection is required")
	}
	return &WorkInterfaceRepository{workSets: workSets, workSomethings: workSomethings, workUsers: workUsers}, nil
}

func NewWorkInterfaceRepositoryFromDatabase(database *drivermongo.Database) (*WorkInterfaceRepository, error) {
	if database == nil {
		return nil, fmt.Errorf("mongo database is required")
	}
	return NewWorkInterfaceRepository(
		database.Collection(WorkSetCollectionName),
		database.Collection(WorkSomethingCollectionName),
		database.Collection(WorkUserCollectionName),
	)
}

func (r *WorkInterfaceRepository) GetWorkConfig(ctx context.Context, guildID string) (domain.WorkConfig, error) {
	guildID = strings.TrimSpace(guildID)
	if guildID == "" {
		return domain.WorkConfig{}, domain.ErrInvalidWorkQuery
	}
	var document documents.WorkConfigDocument
	err := r.workSets.FindOne(ctx, bson.D{{Key: "guild", Value: guildID}}).Decode(&document)
	if err != nil {
		if err == drivermongo.ErrNoDocuments {
			return domain.WorkConfig{}, ports.ErrWorkConfigMissing
		}
		return domain.WorkConfig{}, mhcatmongo.MapError(fmt.Errorf("get work config: %w", err))
	}
	return document.ToDomain(), ctx.Err()
}

func (r *WorkInterfaceRepository) ListWorkItems(ctx context.Context, guildID string) ([]domain.WorkItem, error) {
	guildID = strings.TrimSpace(guildID)
	if guildID == "" {
		return nil, domain.ErrInvalidWorkQuery
	}
	cursor, err := r.workSomethings.Find(ctx, bson.D{{Key: "guild", Value: guildID}}, options.Find().SetSort(bson.D{{Key: "_id", Value: 1}}))
	if err != nil {
		return nil, mhcatmongo.MapError(fmt.Errorf("list work items: %w", err))
	}
	defer cursor.Close(ctx)
	var items []domain.WorkItem
	for cursor.Next(ctx) {
		var document documents.WorkItemDocument
		if err := cursor.Decode(&document); err != nil {
			return nil, mhcatmongo.MapError(fmt.Errorf("decode work item: %w", err))
		}
		items = append(items, document.ToDomain())
	}
	if err := cursor.Err(); err != nil {
		return nil, mhcatmongo.MapError(fmt.Errorf("iterate work items: %w", err))
	}
	if len(items) == 0 {
		return nil, ports.ErrWorkItemsMissing
	}
	return items, ctx.Err()
}

func (r *WorkInterfaceRepository) GetWorkUser(ctx context.Context, guildID string, userID string) (domain.WorkUserState, error) {
	guildID = strings.TrimSpace(guildID)
	userID = strings.TrimSpace(userID)
	if guildID == "" || userID == "" {
		return domain.WorkUserState{}, domain.ErrInvalidWorkQuery
	}
	var document documents.WorkUserDocument
	err := r.workUsers.FindOne(ctx, bson.D{{Key: "guild", Value: guildID}, {Key: "user", Value: userID}}).Decode(&document)
	if err != nil {
		if err == drivermongo.ErrNoDocuments {
			return domain.WorkUserState{}, ports.ErrWorkUserMissing
		}
		return domain.WorkUserState{}, mhcatmongo.MapError(fmt.Errorf("get work user: %w", err))
	}
	return document.ToDomain(), ctx.Err()
}

func (r *WorkInterfaceRepository) StartWork(ctx context.Context, command domain.WorkStartCommand) (domain.WorkUserState, error) {
	if err := ctx.Err(); err != nil {
		return domain.WorkUserState{}, err
	}
	if err := validateWorkStartCommand(command); err != nil {
		return domain.WorkUserState{}, err
	}
	if err := r.ensureWorkUser(ctx, command); err != nil {
		return domain.WorkUserState{}, err
	}
	filter := workStartFilter(command)
	update := workStartUpdate(command)
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var updated documents.WorkUserDocument
	err := r.workUsers.FindOneAndUpdate(ctx, filter, update, opts).Decode(&updated)
	if err != nil {
		if err == drivermongo.ErrNoDocuments {
			return domain.WorkUserState{}, r.workStartMissReason(ctx, command)
		}
		return domain.WorkUserState{}, mhcatmongo.MapError(fmt.Errorf("start work: %w", err))
	}
	return updated.ToDomain(), ctx.Err()
}

func workStartUpdate(command domain.WorkStartCommand) bson.D {
	duration := workStartNumber(command.DurationText, command.DurationSec)
	energy := workStartNumber(command.EnergyCostText, command.EnergyCost)
	reward := workStartNumber(command.CoinRewardText, command.CoinReward)
	return bson.D{
		{Key: "$inc", Value: bson.D{{Key: "energi", Value: -energy}}},
		{Key: "$set", Value: bson.D{
			{Key: "state", Value: command.WorkName},
			{Key: "end_time", Value: float64(command.NowUnix) + duration},
			{Key: "get_coin", Value: reward},
		}},
	}
}

func (r *WorkInterfaceRepository) EnsureWorkUser(ctx context.Context, guildID string, userID string, maxEnergy int64) (domain.WorkUserState, error) {
	if strings.TrimSpace(guildID) == "" || strings.TrimSpace(userID) == "" || maxEnergy < 0 {
		return domain.WorkUserState{}, domain.ErrInvalidWorkQuery
	}
	if err := r.ensureWorkUserWithMaxEnergy(ctx, guildID, userID, maxEnergy); err != nil {
		return domain.WorkUserState{}, err
	}
	return r.GetWorkUser(ctx, guildID, userID)
}

func (r *WorkInterfaceRepository) SaveWorkConfig(ctx context.Context, command domain.WorkConfigCommand) (domain.WorkConfig, error) {
	if err := ctx.Err(); err != nil {
		return domain.WorkConfig{}, err
	}
	if err := validateWorkConfigCommand(command); err != nil {
		return domain.WorkConfig{}, err
	}
	config := domain.WorkConfig{
		GuildID:     strings.TrimSpace(command.GuildID),
		DailyEnergy: command.DailyEnergy,
		MaxEnergy:   command.MaxEnergy,
		Captcha:     command.Captcha,
	}
	update := bson.D{{Key: "$set", Value: bson.D{
		{Key: "guild", Value: config.GuildID},
		{Key: "get_energy", Value: config.DailyEnergy},
		{Key: "max_energy", Value: config.MaxEnergy},
		{Key: "captcha", Value: config.Captcha},
	}}}
	if _, err := r.workSets.UpdateMany(ctx, bson.D{{Key: "guild", Value: config.GuildID}}, update, options.UpdateMany().SetUpsert(true)); err != nil {
		return domain.WorkConfig{}, mhcatmongo.MapError(fmt.Errorf("save work config: %w", err))
	}
	return config, ctx.Err()
}

func (r *WorkInterfaceRepository) DeleteWorkItem(ctx context.Context, command domain.WorkDeleteItemCommand) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := validateWorkDeleteItemCommand(command); err != nil {
		return err
	}
	result, err := r.workSomethings.DeleteMany(ctx, bson.D{{Key: "guild", Value: strings.TrimSpace(command.GuildID)}, {Key: "name", Value: strings.TrimSpace(command.Name)}})
	if err != nil {
		return mhcatmongo.MapError(fmt.Errorf("delete work item: %w", err))
	}
	if result.DeletedCount == 0 {
		return ports.ErrWorkItemMissing
	}
	return ctx.Err()
}

func (r *WorkInterfaceRepository) GrantWorkEnergy(ctx context.Context, command domain.WorkEnergyGrantCommand) (domain.WorkUserState, error) {
	if err := ctx.Err(); err != nil {
		return domain.WorkUserState{}, err
	}
	if err := validateWorkEnergyGrantCommand(command.GuildID, command.UserID, command.Amount, command.MaxEnergy); err != nil {
		return domain.WorkUserState{}, err
	}
	if err := r.ensureWorkUserWithMaxEnergy(ctx, command.GuildID, command.UserID, command.MaxEnergy); err != nil {
		return domain.WorkUserState{}, err
	}
	var updated documents.WorkUserDocument
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	err := r.workUsers.FindOneAndUpdate(
		ctx,
		bson.D{{Key: "guild", Value: strings.TrimSpace(command.GuildID)}, {Key: "user", Value: strings.TrimSpace(command.UserID)}},
		workEnergyGrantPipeline(command.Amount, command.MaxEnergy),
		opts,
	).Decode(&updated)
	if err != nil {
		if err == drivermongo.ErrNoDocuments {
			return domain.WorkUserState{}, ports.ErrWorkUserMissing
		}
		return domain.WorkUserState{}, mhcatmongo.MapError(fmt.Errorf("grant work energy: %w", err))
	}
	return updated.ToDomain(), ctx.Err()
}

func (r *WorkInterfaceRepository) GrantWorkEnergyToAll(ctx context.Context, command domain.WorkEnergyGrantAllCommand) (domain.WorkEnergyGrantAllResult, error) {
	if err := ctx.Err(); err != nil {
		return domain.WorkEnergyGrantAllResult{}, err
	}
	if err := validateWorkEnergyGrantCommand(command.GuildID, "all", command.Amount, command.MaxEnergy); err != nil {
		return domain.WorkEnergyGrantAllResult{}, err
	}
	result, err := r.workUsers.UpdateMany(
		ctx,
		bson.D{{Key: "guild", Value: strings.TrimSpace(command.GuildID)}},
		workEnergyGrantPipeline(command.Amount, command.MaxEnergy),
	)
	if err != nil {
		return domain.WorkEnergyGrantAllResult{}, mhcatmongo.MapError(fmt.Errorf("grant work energy to all: %w", err))
	}
	return domain.WorkEnergyGrantAllResult{Matched: result.MatchedCount, Modified: result.ModifiedCount}, ctx.Err()
}

func (r *WorkInterfaceRepository) ensureWorkUser(ctx context.Context, command domain.WorkStartCommand) error {
	return r.ensureWorkUserWithMaxEnergy(ctx, command.GuildID, command.UserID, command.MaxEnergy)
}

func (r *WorkInterfaceRepository) ensureWorkUserWithMaxEnergy(ctx context.Context, guildID string, userID string, maxEnergy int64) error {
	_, err := r.workUsers.UpdateOne(
		ctx,
		bson.D{{Key: "guild", Value: strings.TrimSpace(guildID)}, {Key: "user", Value: strings.TrimSpace(userID)}},
		bson.D{{Key: "$setOnInsert", Value: bson.D{
			{Key: "guild", Value: strings.TrimSpace(guildID)},
			{Key: "user", Value: strings.TrimSpace(userID)},
			{Key: "state", Value: LegacyIdleWorkState},
			{Key: "end_time", Value: int64(0)},
			{Key: "energi", Value: maxEnergy},
			{Key: "get_coin", Value: int64(0)},
		}}},
		options.UpdateOne().SetUpsert(true),
	)
	if err != nil {
		return mhcatmongo.MapError(fmt.Errorf("ensure work user: %w", err))
	}
	return ctx.Err()
}

func (r *WorkInterfaceRepository) workStartMissReason(ctx context.Context, command domain.WorkStartCommand) error {
	current, err := r.GetWorkUser(ctx, command.GuildID, command.UserID)
	if err != nil {
		return err
	}
	if workStartNumber(current.EnergyText, current.Energy) < workStartNumber(command.EnergyCostText, command.EnergyCost) {
		return domain.ErrWorkEnergyInsufficient
	}
	if !command.Override && current.State != LegacyIdleWorkState {
		return domain.ErrWorkAlreadyBusy
	}
	return domain.ErrInvalidWorkQuery
}

func validateWorkStartCommand(command domain.WorkStartCommand) error {
	if strings.TrimSpace(command.GuildID) == "" ||
		strings.TrimSpace(command.UserID) == "" ||
		strings.TrimSpace(command.WorkName) == "" ||
		command.MaxEnergy < 0 ||
		command.NowUnix <= 0 {
		return domain.ErrInvalidWorkQuery
	}
	return nil
}

func validateWorkConfigCommand(command domain.WorkConfigCommand) error {
	if strings.TrimSpace(command.GuildID) == "" || command.DailyEnergy < 0 || command.MaxEnergy < 0 {
		return domain.ErrInvalidWorkQuery
	}
	return nil
}

func validateWorkDeleteItemCommand(command domain.WorkDeleteItemCommand) error {
	if strings.TrimSpace(command.GuildID) == "" || strings.TrimSpace(command.Name) == "" {
		return domain.ErrInvalidWorkQuery
	}
	return nil
}

func validateWorkEnergyGrantCommand(guildID string, userID string, amount int64, maxEnergy int64) error {
	if strings.TrimSpace(guildID) == "" || strings.TrimSpace(userID) == "" || amount <= 0 || maxEnergy < 0 {
		return domain.ErrInvalidWorkQuery
	}
	return nil
}

func workStartFilter(command domain.WorkStartCommand) bson.D {
	filter := bson.D{
		{Key: "guild", Value: command.GuildID},
		{Key: "user", Value: command.UserID},
	}
	energy := workStartNumber(command.EnergyCostText, command.EnergyCost)
	if !math.IsNaN(energy) {
		filter = append(filter, bson.E{Key: "energi", Value: bson.D{{Key: "$gte", Value: energy}}})
	}
	if !command.Override {
		filter = append(filter, bson.E{Key: "state", Value: LegacyIdleWorkState})
	}
	return filter
}

func workStartNumber(text string, fallback int64) float64 {
	text = strings.TrimSpace(text)
	if text == "" {
		return float64(fallback)
	}
	if text == "null" {
		return 0
	}
	value, err := strconv.ParseFloat(text, 64)
	if err != nil {
		return math.NaN()
	}
	return value
}

func workEnergyGrantPipeline(amount int64, maxEnergy int64) drivermongo.Pipeline {
	return drivermongo.Pipeline{
		bson.D{{Key: "$set", Value: bson.D{{
			Key: "energi",
			Value: bson.D{{Key: "$min", Value: bson.A{
				maxEnergy,
				bson.D{{Key: "$add", Value: bson.A{
					bson.D{{Key: "$ifNull", Value: bson.A{"$energi", int64(0)}}},
					amount,
				}}},
			}}},
		}}}},
	}
}

var _ ports.WorkInterfaceRepository = (*WorkInterfaceRepository)(nil)
var _ ports.WorkStartRepository = (*WorkInterfaceRepository)(nil)
var _ ports.WorkAdminRepository = (*WorkInterfaceRepository)(nil)
