package repositories

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	mhcatmongo "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/adapters/mongo"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	"go.mongodb.org/mongo-driver/v2/bson"
	drivermongo "go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const (
	UsageCollectionName      = "all_use_counts"
	DefaultUsageTrackTimeout = 500 * time.Millisecond
)

type UsageTracker struct {
	collection *drivermongo.Collection
	timeout    time.Duration
}

func NewUsageTracker(collection *drivermongo.Collection, timeout time.Duration) (*UsageTracker, error) {
	if collection == nil {
		return nil, errors.New("mongo usage collection is required")
	}
	if timeout <= 0 {
		return nil, errors.New("mongo usage timeout must be positive")
	}
	return &UsageTracker{collection: collection, timeout: timeout}, nil
}

func NewUsageTrackerFromDatabase(database *drivermongo.Database, timeout time.Duration) (*UsageTracker, error) {
	if database == nil {
		return nil, errors.New("mongo database is required")
	}
	return NewUsageTracker(database.Collection(UsageCollectionName), timeout)
}

func (t *UsageTracker) TrackCommand(ctx context.Context, event ports.UsageEvent) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	event.CommandName = strings.TrimSpace(event.CommandName)
	if err := event.Validate(); err != nil {
		return err
	}
	trackCtx, cancel := context.WithTimeout(ctx, t.timeout)
	defer cancel()
	_, err := t.collection.UpdateOne(
		trackCtx,
		bson.D{{Key: "slashcommand_name", Value: event.CommandName}},
		usageIncrementPipeline(event.CommandName),
		options.UpdateOne().SetUpsert(true),
	)
	if err != nil {
		return mhcatmongo.MapError(fmt.Errorf("track command usage: %w", err))
	}
	return trackCtx.Err()
}

func usageIncrementPipeline(commandName string) drivermongo.Pipeline {
	convertedCount := bson.D{{Key: "$convert", Value: bson.D{
		{Key: "input", Value: "$count"},
		{Key: "to", Value: "double"},
		{Key: "onError", Value: float64(0)},
		{Key: "onNull", Value: float64(0)},
	}}}
	return drivermongo.Pipeline{bson.D{{Key: "$set", Value: bson.D{
		{Key: "slashcommand_name", Value: commandName},
		{Key: "count", Value: bson.D{{Key: "$add", Value: bson.A{convertedCount, float64(1)}}}},
	}}}}
}

var _ ports.UsageTracker = (*UsageTracker)(nil)
