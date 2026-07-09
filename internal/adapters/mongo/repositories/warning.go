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

const WarningCollectionName = "warndbs"

type WarningHistoryRepository struct {
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

var _ ports.WarningHistoryRepository = (*WarningHistoryRepository)(nil)
