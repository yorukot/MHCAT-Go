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

type StatsConfigRepository struct {
	numbers *drivermongo.Collection
}

func NewStatsConfigRepository(numbers *drivermongo.Collection) (*StatsConfigRepository, error) {
	if numbers == nil {
		return nil, errors.New("numbers collection is required")
	}
	return &StatsConfigRepository{numbers: numbers}, nil
}

func NewStatsConfigRepositoryFromDatabase(database *drivermongo.Database) (*StatsConfigRepository, error) {
	if database == nil {
		return nil, errors.New("mongo database is required")
	}
	return NewStatsConfigRepository(database.Collection(StatsConfigCollectionName))
}

func (r *StatsConfigRepository) DeleteStatsConfig(ctx context.Context, guildID string) (domain.StatsConfig, error) {
	if err := ctx.Err(); err != nil {
		return domain.StatsConfig{}, err
	}
	guildID = strings.TrimSpace(guildID)
	if guildID == "" {
		return domain.StatsConfig{}, domain.ErrInvalidStatsConfigRequest
	}
	filter := bson.D{{Key: "guild", Value: guildID}}
	var document documents.StatsConfigDocument
	err := r.numbers.FindOne(ctx, filter).Decode(&document)
	if err != nil {
		if err == drivermongo.ErrNoDocuments {
			return domain.StatsConfig{}, ports.ErrStatsConfigMissing
		}
		return domain.StatsConfig{}, mhcatmongo.MapError(fmt.Errorf("find stats config: %w", err))
	}
	result, err := r.numbers.DeleteMany(ctx, filter)
	if err != nil {
		return domain.StatsConfig{}, mhcatmongo.MapError(fmt.Errorf("delete stats config: %w", err))
	}
	if result.DeletedCount == 0 {
		return domain.StatsConfig{}, ports.ErrStatsConfigMissing
	}
	return document.ToDomain().Normalize(), ctx.Err()
}

var _ ports.StatsConfigRepository = (*StatsConfigRepository)(nil)
