package repositories

import (
	"context"
	"errors"
	"fmt"

	mhcatmongo "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/adapters/mongo"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/adapters/mongo/documents"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	"go.mongodb.org/mongo-driver/v2/bson"
	drivermongo "go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const BirthdayConfigCollectionName = "birthday_sets"

type BirthdayConfigRepository struct {
	collection *drivermongo.Collection
}

func NewBirthdayConfigRepository(collection *drivermongo.Collection) (*BirthdayConfigRepository, error) {
	if collection == nil {
		return nil, errors.New("mongo birthday config collection is required")
	}
	return &BirthdayConfigRepository{collection: collection}, nil
}

func NewBirthdayConfigRepositoryFromDatabase(database *drivermongo.Database) (*BirthdayConfigRepository, error) {
	if database == nil {
		return nil, errors.New("mongo database is required")
	}
	return NewBirthdayConfigRepository(database.Collection(BirthdayConfigCollectionName))
}

func (r *BirthdayConfigRepository) SaveBirthdayConfig(ctx context.Context, config domain.BirthdayConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := config.Validate(); err != nil {
		return err
	}
	document := documents.BirthdayConfigDocumentFromDomain(config)
	update, err := birthdayConfigUpdate(document)
	if err != nil {
		return err
	}
	result, err := r.collection.UpdateMany(ctx, bson.D{{Key: "guild", Value: document.Guild}}, update)
	if err != nil {
		return mhcatmongo.MapError(fmt.Errorf("save birthday config: %w", err))
	}
	if result.MatchedCount > 0 {
		return ctx.Err()
	}
	insertUpdate, err := birthdayConfigInsertUpdate(document)
	if err != nil {
		return err
	}
	_, err = r.collection.UpdateOne(
		ctx,
		bson.D{{Key: "guild", Value: document.Guild}},
		insertUpdate,
		options.UpdateOne().SetUpsert(true),
	)
	if err != nil {
		return mhcatmongo.MapError(fmt.Errorf("upsert birthday config: %w", err))
	}
	return ctx.Err()
}

func birthdayConfigUpdate(document documents.BirthdayConfigDocument) (bson.D, error) {
	return mhcatmongo.NewUpdate().
		Set("msg", document.Message).
		Set("utc", document.UTCOffset).
		Set("channel", document.Channel).
		Set("everyone_can_set_birthday_date", document.EveryoneCanSetBirthdayDate).
		Set("role", birthdayRoleValue(document)).
		Build()
}

func birthdayConfigInsertUpdate(document documents.BirthdayConfigDocument) (bson.D, error) {
	return mhcatmongo.NewUpdate().
		Set("msg", document.Message).
		Set("utc", document.UTCOffset).
		Set("channel", document.Channel).
		Set("everyone_can_set_birthday_date", document.EveryoneCanSetBirthdayDate).
		Set("role", birthdayRoleValue(document)).
		SetOnInsert("guild", document.Guild).
		Build()
}

func birthdayRoleValue(document documents.BirthdayConfigDocument) any {
	if document.Role == nil {
		return nil
	}
	return *document.Role
}

var _ ports.BirthdayConfigRepository = (*BirthdayConfigRepository)(nil)
