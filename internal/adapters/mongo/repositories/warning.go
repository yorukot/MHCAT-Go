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
	var document documents.WarningReadDocument
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

func (r *WarningHistoryRepository) AddWarning(ctx context.Context, issue domain.WarningIssue) (domain.WarningIssueResult, error) {
	if err := ctx.Err(); err != nil {
		return domain.WarningIssueResult{}, err
	}
	issue.GuildID = strings.TrimSpace(issue.GuildID)
	issue.UserID = strings.TrimSpace(issue.UserID)
	issue.ModeratorID = strings.TrimSpace(issue.ModeratorID)
	issue.Time = strings.TrimSpace(issue.Time)
	if err := issue.Validate(); err != nil {
		return domain.WarningIssueResult{}, err
	}
	filter := bson.D{{Key: "guild", Value: issue.GuildID}, {Key: "user", Value: issue.UserID}}
	entry := documents.WarningEntryDocumentFromIssue(issue)
	var document documents.WarningReadDocument
	err := r.collection.FindOne(ctx, filter).Decode(&document)
	if err != nil {
		if err != drivermongo.ErrNoDocuments {
			return domain.WarningIssueResult{}, mhcatmongo.MapError(fmt.Errorf("find warning for append: %w", err))
		}
		created := documents.WarningDocument{
			Guild:   issue.GuildID,
			User:    issue.UserID,
			Content: []documents.WarningEntryDocument{entry},
		}
		if _, err := r.collection.InsertOne(ctx, created); err != nil {
			return domain.WarningIssueResult{}, mhcatmongo.MapError(fmt.Errorf("insert warning: %w", err))
		}
		return domain.WarningIssueResult{History: created.ToDomain(), Created: true}, ctx.Err()
	}
	content, isArray, err := document.ContentValues()
	if err != nil {
		return domain.WarningIssueResult{}, mhcatmongo.MapError(fmt.Errorf("decode warning content for append: %w", err))
	}
	update := bson.D{{Key: "$push", Value: bson.D{{Key: "content", Value: entry}}}}
	if !isArray {
		content = append(content, entry)
		update = bson.D{{Key: "$set", Value: bson.D{{Key: "content", Value: content}}}}
	}
	var updated documents.WarningReadDocument
	err = r.collection.FindOneAndUpdate(
		ctx,
		bson.D{{Key: "_id", Value: document.ID}},
		update,
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	).Decode(&updated)
	if err != nil {
		if err == drivermongo.ErrNoDocuments {
			return domain.WarningIssueResult{}, ports.ErrWarningsNotFound
		}
		return domain.WarningIssueResult{}, mhcatmongo.MapError(fmt.Errorf("append warning: %w", err))
	}
	return domain.WarningIssueResult{History: updated.ToDomain()}, ctx.Err()
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
	var document documents.WarningReadDocument
	err := r.collection.FindOne(ctx, filter).Decode(&document)
	if err != nil {
		if err == drivermongo.ErrNoDocuments {
			return ports.ErrWarningsNotFound
		}
		return mhcatmongo.MapError(fmt.Errorf("find warning for removal: %w", err))
	}
	content, _, err := document.ContentValues()
	if err != nil {
		return mhcatmongo.MapError(fmt.Errorf("decode warning content for removal: %w", err))
	}
	index := int(removal.Index - 1)
	if index < 0 || index >= len(content) {
		return ports.ErrWarningsNotFound
	}
	next := append([]any(nil), content[:index]...)
	next = append(next, content[index+1:]...)
	result, err := r.collection.UpdateOne(ctx, bson.D{{Key: "_id", Value: document.ID}}, bson.D{{Key: "$set", Value: bson.D{{Key: "content", Value: next}}}})
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

func (r *WarningSettingsRepository) GetWarningSettings(ctx context.Context, guildID string) (domain.WarningSettings, error) {
	if err := ctx.Err(); err != nil {
		return domain.WarningSettings{}, err
	}
	guildID = strings.TrimSpace(guildID)
	if guildID == "" {
		return domain.WarningSettings{}, domain.ErrInvalidWarningSettings
	}
	var document documents.WarningSettingsReadDocument
	err := r.collection.FindOne(ctx, bson.D{{Key: "guild", Value: guildID}}).Decode(&document)
	if err != nil {
		if err == drivermongo.ErrNoDocuments {
			return domain.WarningSettings{}, ports.ErrWarningSettingsNotFound
		}
		return domain.WarningSettings{}, mhcatmongo.MapError(fmt.Errorf("get warning settings: %w", err))
	}
	settings, err := document.ToDomain()
	if err != nil {
		return domain.WarningSettings{}, err
	}
	return settings, ctx.Err()
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
var _ ports.WarningIssueRepository = (*WarningHistoryRepository)(nil)
var _ ports.WarningRemovalRepository = (*WarningHistoryRepository)(nil)
var _ ports.WarningSettingsRepository = (*WarningSettingsRepository)(nil)
