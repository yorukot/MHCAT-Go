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
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const WarningCollectionName = "warndbs"
const WarningSettingsCollectionName = "errors_sets"

type WarningHistoryRepository struct {
	collection *drivermongo.Collection
}

type WarningSettingsRepository struct {
	collection *drivermongo.Collection
}

func NewWarningHistoryRepository(collection *drivermongo.Collection) (*WarningHistoryRepository, error) {
	if collection == nil {
		return nil, errors.New("mongo warning collection is required")
	}
	return &WarningHistoryRepository{collection: collection}, nil
}

func NewWarningHistoryRepositoryFromDatabase(database *drivermongo.Database) (*WarningHistoryRepository, error) {
	if database == nil {
		return nil, errors.New("mongo database is required")
	}
	return NewWarningHistoryRepository(database.Collection(WarningCollectionName))
}

func NewWarningSettingsRepository(collection *drivermongo.Collection) (*WarningSettingsRepository, error) {
	if collection == nil {
		return nil, errors.New("mongo warning settings collection is required")
	}
	return &WarningSettingsRepository{collection: collection}, nil
}

func NewWarningSettingsRepositoryFromDatabase(database *drivermongo.Database) (*WarningSettingsRepository, error) {
	if database == nil {
		return nil, errors.New("mongo database is required")
	}
	return NewWarningSettingsRepository(database.Collection(WarningSettingsCollectionName))
}

func (r *WarningHistoryRepository) GetWarningHistory(ctx context.Context, guildID string, userID string) (domain.WarningHistory, error) {
	if err := ctx.Err(); err != nil {
		return domain.WarningHistory{}, err
	}
	guildID = strings.TrimSpace(guildID)
	userID = strings.TrimSpace(userID)
	if guildID == "" || userID == "" {
		return domain.WarningHistory{}, domain.ErrInvalidWarningQuery
	}
	var document documents.WarningDocument
	err := r.collection.FindOne(ctx, bson.D{{Key: "guild", Value: guildID}, {Key: "user", Value: userID}}).Decode(&document)
	if err != nil {
		if err == drivermongo.ErrNoDocuments {
			return domain.WarningHistory{}, ports.ErrWarningsNotFound
		}
		return domain.WarningHistory{}, mhcatmongo.MapError(fmt.Errorf("get warning history: %w", err))
	}
	history := document.ToDomain()
	if len(history.Entries) == 0 {
		return domain.WarningHistory{}, ports.ErrWarningsNotFound
	}
	return history, ctx.Err()
}

func (r *WarningHistoryRepository) RemoveWarning(ctx context.Context, removal domain.WarningRemoval) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	removal.GuildID = strings.TrimSpace(removal.GuildID)
	removal.UserID = strings.TrimSpace(removal.UserID)
	if err := removal.ValidateSingle(); err != nil {
		return err
	}
	filter := bson.D{{Key: "guild", Value: removal.GuildID}, {Key: "user", Value: removal.UserID}}
	var document documents.WarningDocument
	err := r.collection.FindOne(ctx, filter).Decode(&document)
	if err != nil {
		if err == drivermongo.ErrNoDocuments {
			return ports.ErrWarningsNotFound
		}
		return mhcatmongo.MapError(fmt.Errorf("find warning for removal: %w", err))
	}
	index := int(removal.Index - 1)
	if index < 0 || index >= len(document.Content) {
		return ports.ErrWarningsNotFound
	}
	next := append([]documents.WarningEntryDocument(nil), document.Content[:index]...)
	next = append(next, document.Content[index+1:]...)
	result, err := r.collection.UpdateOne(ctx, filter, bson.D{{Key: "$set", Value: bson.D{{Key: "content", Value: next}}}})
	if err != nil {
		return mhcatmongo.MapError(fmt.Errorf("remove warning: %w", err))
	}
	if result.MatchedCount == 0 {
		return ports.ErrWarningsNotFound
	}
	return ctx.Err()
}

func (r *WarningHistoryRepository) RemoveAllWarnings(ctx context.Context, removal domain.WarningRemoval) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	removal.GuildID = strings.TrimSpace(removal.GuildID)
	removal.UserID = strings.TrimSpace(removal.UserID)
	if err := removal.ValidateAll(); err != nil {
		return err
	}
	result, err := r.collection.DeleteMany(ctx, bson.D{{Key: "guild", Value: removal.GuildID}, {Key: "user", Value: removal.UserID}})
	if err != nil {
		return mhcatmongo.MapError(fmt.Errorf("remove all warnings: %w", err))
	}
	if result.DeletedCount == 0 {
		return ports.ErrWarningsNotFound
	}
	return ctx.Err()
}

func (r *WarningSettingsRepository) SaveWarningSettings(ctx context.Context, settings domain.WarningSettings) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := settings.Validate(); err != nil {
		return err
	}
	document := documents.WarningSettingsDocumentFromDomain(settings)
	update, err := warningSettingsUpdate(document, false)
	if err != nil {
		return err
	}
	result, err := r.collection.UpdateMany(ctx, bson.D{{Key: "guild", Value: document.Guild}}, update)
	if err != nil {
		return mhcatmongo.MapError(fmt.Errorf("save warning settings: %w", err))
	}
	if result.MatchedCount > 0 {
		return ctx.Err()
	}
	insertUpdate, err := warningSettingsUpdate(document, true)
	if err != nil {
		return err
	}
	_, err = r.collection.UpdateOne(ctx, bson.D{{Key: "guild", Value: document.Guild}}, insertUpdate, options.UpdateOne().SetUpsert(true))
	if err != nil {
		return mhcatmongo.MapError(fmt.Errorf("upsert warning settings: %w", err))
	}
	return ctx.Err()
}

func warningSettingsUpdate(document documents.WarningSettingsDocument, upsert bool) (bson.D, error) {
	builder := mhcatmongo.NewUpdate().
		Set("ban_count", document.BanCount).
		Set("move", document.Move)
	if upsert {
		builder.SetOnInsert("guild", document.Guild)
	}
	return builder.Build()
}

var _ ports.WarningHistoryRepository = (*WarningHistoryRepository)(nil)
var _ ports.WarningRemovalRepository = (*WarningHistoryRepository)(nil)
var _ ports.WarningSettingsRepository = (*WarningSettingsRepository)(nil)
