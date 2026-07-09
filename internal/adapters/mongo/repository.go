package mongo

import (
	"context"
	"errors"

	drivermongo "go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
)

type BaseRepository struct {
	collection *drivermongo.Collection
}

func NewBaseRepository(collection *drivermongo.Collection) (*BaseRepository, error) {
	if collection == nil {
		return nil, errors.New("mongo collection is required")
	}
	return &BaseRepository{collection: collection}, nil
}

func (r *BaseRepository) CollectionName() string {
	if r == nil || r.collection == nil {
		return ""
	}
	return r.collection.Name()
}

func (r *BaseRepository) Ping(ctx context.Context) error {
	if r == nil || r.collection == nil {
		return errors.New("mongo repository is not configured")
	}
	return MapError(r.collection.Database().Client().Ping(ctx, readpref.Primary()))
}
