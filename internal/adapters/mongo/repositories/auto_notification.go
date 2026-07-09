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
	"go.mongodb.org/mongo-driver/v2/bson"
	drivermongo "go.mongodb.org/mongo-driver/v2/mongo"
)

const AutoNotificationScheduleCollectionName = "cron_sets"

type AutoNotificationScheduleRepository struct {
	collection *drivermongo.Collection
}

func NewAutoNotificationScheduleRepository(collection *drivermongo.Collection) (*AutoNotificationScheduleRepository, error) {
	if collection == nil {
		return nil, errors.New("mongo auto-notification schedule collection is required")
	}
	return &AutoNotificationScheduleRepository{collection: collection}, nil
}

func NewAutoNotificationScheduleRepositoryFromDatabase(database *drivermongo.Database) (*AutoNotificationScheduleRepository, error) {
	if database == nil {
		return nil, errors.New("mongo database is required")
	}
	return NewAutoNotificationScheduleRepository(database.Collection(AutoNotificationScheduleCollectionName))
}

func (r *AutoNotificationScheduleRepository) ListAutoNotificationSchedules(ctx context.Context, guildID string) ([]domain.AutoNotificationSchedule, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	guildID = strings.TrimSpace(guildID)
	if err := domain.ValidateAutoNotificationGuildID(guildID); err != nil {
		return nil, err
	}
	cursor, err := r.collection.Find(ctx, bson.D{{Key: "guild", Value: guildID}})
	if err != nil {
		return nil, mhcatmongo.MapError(fmt.Errorf("list auto-notification schedules: %w", err))
	}
	defer cursor.Close(ctx)
	var schedules []domain.AutoNotificationSchedule
	for cursor.Next(ctx) {
		var document documents.AutoNotificationScheduleDocument
		if err := cursor.Decode(&document); err != nil {
			return nil, mhcatmongo.MapError(fmt.Errorf("decode auto-notification schedule: %w", err))
		}
		schedules = append(schedules, document.ToDomain())
	}
	if err := cursor.Err(); err != nil {
		return nil, mhcatmongo.MapError(fmt.Errorf("iterate auto-notification schedules: %w", err))
	}
	return schedules, ctx.Err()
}

func (r *AutoNotificationScheduleRepository) DeleteAutoNotificationSchedule(ctx context.Context, guildID string, id string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	guildID = strings.TrimSpace(guildID)
	id = strings.TrimSpace(id)
	if err := domain.ValidateAutoNotificationDelete(guildID, id); err != nil {
		return err
	}
	result, err := r.collection.DeleteOne(ctx, bson.D{{Key: "guild", Value: guildID}, {Key: "id", Value: id}})
	if err != nil {
		return mhcatmongo.MapError(fmt.Errorf("delete auto-notification schedule: %w", err))
	}
	if result.DeletedCount == 0 {
		return ports.ErrAutoNotificationScheduleMissing
	}
	return ctx.Err()
}

func (r *AutoNotificationScheduleRepository) CreatePendingAutoNotificationSchedule(ctx context.Context, draft domain.AutoNotificationSetupDraft) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	draft = draft.Normalized()
	if err := domain.ValidateAutoNotificationSetupDraft(draft); err != nil {
		return err
	}
	_, err := r.collection.InsertOne(ctx, documents.AutoNotificationPendingWriteDocumentFromDomain(draft))
	if err != nil {
		mapped := mhcatmongo.MapError(fmt.Errorf("create pending auto-notification schedule: %w", err))
		if mhcatmongo.ErrorIs(mapped, mhcatmongo.ErrorKindConflict) {
			return ports.ErrAutoNotificationScheduleExists
		}
		return mapped
	}
	return ctx.Err()
}

func (r *AutoNotificationScheduleRepository) CompleteAutoNotificationSchedule(ctx context.Context, setup domain.AutoNotificationSetup) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	setup = setup.Normalized()
	if err := domain.ValidateAutoNotificationSetup(setup); err != nil {
		return err
	}
	result, err := r.collection.UpdateOne(
		ctx,
		bson.D{{Key: "guild", Value: setup.GuildID}, {Key: "id", Value: setup.ID}},
		bson.D{{Key: "$set", Value: bson.D{
			{Key: "cron", Value: setup.Cron},
			{Key: "message", Value: documents.AutoNotificationMessageBSON(setup.Message)},
		}}},
	)
	if err != nil {
		return mhcatmongo.MapError(fmt.Errorf("complete auto-notification schedule: %w", err))
	}
	if result.MatchedCount == 0 {
		return ports.ErrAutoNotificationScheduleMissing
	}
	return ctx.Err()
}

func (r *AutoNotificationScheduleRepository) DeletePendingAutoNotificationSchedules(ctx context.Context, guildID string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	guildID = strings.TrimSpace(guildID)
	if err := domain.ValidateAutoNotificationGuildID(guildID); err != nil {
		return err
	}
	_, err := r.collection.DeleteMany(ctx, bson.D{
		{Key: "guild", Value: guildID},
		// Mongo null equality matches null and missing fields, both of which
		// decode as abandoned setup drafts.
		{Key: "cron", Value: nil},
	})
	if err != nil {
		return mhcatmongo.MapError(fmt.Errorf("delete pending auto-notification schedules: %w", err))
	}
	return ctx.Err()
}

var _ ports.AutoNotificationScheduleRepository = (*AutoNotificationScheduleRepository)(nil)
