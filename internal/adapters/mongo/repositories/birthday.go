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

const BirthdayConfigCollectionName = "birthday_sets"
const BirthdayProfileCollectionName = "birthdays"

type BirthdayConfigRepository struct {
	collection        *drivermongo.Collection
	profileCollection *drivermongo.Collection
}

func NewBirthdayConfigRepository(collection *drivermongo.Collection) (*BirthdayConfigRepository, error) {
	return NewBirthdayConfigRepositoryWithProfiles(collection, nil)
}

func NewBirthdayConfigRepositoryWithProfiles(collection *drivermongo.Collection, profileCollection *drivermongo.Collection) (*BirthdayConfigRepository, error) {
	if collection == nil {
		return nil, errors.New("mongo birthday config collection is required")
	}
	return &BirthdayConfigRepository{collection: collection, profileCollection: profileCollection}, nil
}

func NewBirthdayConfigRepositoryFromDatabase(database *drivermongo.Database) (*BirthdayConfigRepository, error) {
	if database == nil {
		return nil, errors.New("mongo database is required")
	}
	return NewBirthdayConfigRepositoryWithProfiles(
		database.Collection(BirthdayConfigCollectionName),
		database.Collection(BirthdayProfileCollectionName),
	)
}

func (r *BirthdayConfigRepository) FindBirthdayConfig(ctx context.Context, guildID string) (domain.BirthdayConfig, error) {
	if err := ctx.Err(); err != nil {
		return domain.BirthdayConfig{}, err
	}
	guildID = strings.TrimSpace(guildID)
	if guildID == "" {
		return domain.BirthdayConfig{}, domain.ErrInvalidBirthdayConfig
	}
	var document documents.BirthdayConfigDocument
	if err := r.collection.FindOne(ctx, bson.D{{Key: "guild", Value: guildID}}).Decode(&document); err != nil {
		if mhcatmongo.ErrorIs(mhcatmongo.MapError(err), mhcatmongo.ErrorKindNotFound) {
			return domain.BirthdayConfig{}, ports.ErrBirthdayConfigMissing
		}
		return domain.BirthdayConfig{}, mhcatmongo.MapError(fmt.Errorf("find birthday config: %w", err))
	}
	return document.ToDomain(), ctx.Err()
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

func (r *BirthdayConfigRepository) FindBirthdayProfile(ctx context.Context, guildID string, userID string) (domain.BirthdayProfile, error) {
	if err := ctx.Err(); err != nil {
		return domain.BirthdayProfile{}, err
	}
	collection, err := r.requireProfileCollection()
	if err != nil {
		return domain.BirthdayProfile{}, err
	}
	filter, err := birthdayProfileFilter(guildID, userID)
	if err != nil {
		return domain.BirthdayProfile{}, err
	}
	var document documents.BirthdayProfileDocument
	if err := collection.FindOne(ctx, filter).Decode(&document); err != nil {
		if mhcatmongo.ErrorIs(mhcatmongo.MapError(err), mhcatmongo.ErrorKindNotFound) {
			return domain.BirthdayProfile{}, ports.ErrBirthdayProfileMissing
		}
		return domain.BirthdayProfile{}, mhcatmongo.MapError(fmt.Errorf("find birthday profile: %w", err))
	}
	return document.ToDomain(), ctx.Err()
}

func (r *BirthdayConfigRepository) SaveBirthdayProfile(ctx context.Context, profile domain.BirthdayProfile) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	collection, err := r.requireProfileCollection()
	if err != nil {
		return err
	}
	profile = trimBirthdayProfile(profile)
	if err := profile.ValidateIdentity(); err != nil {
		return err
	}
	document := documents.BirthdayProfileDocumentFromDomain(profile)
	update, err := birthdayProfileUpdate(document)
	if err != nil {
		return err
	}
	filter := bson.D{{Key: "guild", Value: document.Guild}, {Key: "user", Value: document.User}}
	result, err := collection.UpdateMany(ctx, filter, update)
	if err != nil {
		return mhcatmongo.MapError(fmt.Errorf("save birthday profile: %w", err))
	}
	if result.MatchedCount > 0 {
		return ctx.Err()
	}
	insertUpdate, err := birthdayProfileInsertUpdate(document)
	if err != nil {
		return err
	}
	_, err = collection.UpdateOne(ctx, filter, insertUpdate, options.UpdateOne().SetUpsert(true))
	if err != nil {
		return mhcatmongo.MapError(fmt.Errorf("upsert birthday profile: %w", err))
	}
	return ctx.Err()
}

func (r *BirthdayConfigRepository) DeleteBirthdayProfile(ctx context.Context, guildID string, userID string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	collection, err := r.requireProfileCollection()
	if err != nil {
		return err
	}
	filter, err := birthdayProfileFilter(guildID, userID)
	if err != nil {
		return err
	}
	result, err := collection.DeleteMany(ctx, filter)
	if err != nil {
		return mhcatmongo.MapError(fmt.Errorf("delete birthday profile: %w", err))
	}
	if result.DeletedCount == 0 {
		return ports.ErrBirthdayProfileMissing
	}
	return ctx.Err()
}

func (r *BirthdayConfigRepository) ListBirthdayProfiles(ctx context.Context, guildID string) ([]domain.BirthdayProfile, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	collection, err := r.requireProfileCollection()
	if err != nil {
		return nil, err
	}
	guildID = strings.TrimSpace(guildID)
	if guildID == "" {
		return nil, domain.ErrInvalidBirthdayProfile
	}
	cursor, err := collection.Find(ctx, bson.D{{Key: "guild", Value: guildID}})
	if err != nil {
		return nil, mhcatmongo.MapError(fmt.Errorf("list birthday profiles: %w", err))
	}
	defer cursor.Close(ctx)
	profiles := []domain.BirthdayProfile{}
	for cursor.Next(ctx) {
		var document documents.BirthdayProfileDocument
		if err := cursor.Decode(&document); err != nil {
			return nil, mhcatmongo.MapError(fmt.Errorf("decode birthday profile: %w", err))
		}
		profiles = append(profiles, document.ToDomain())
	}
	if err := cursor.Err(); err != nil {
		return nil, mhcatmongo.MapError(fmt.Errorf("iterate birthday profiles: %w", err))
	}
	return profiles, ctx.Err()
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

func (r *BirthdayConfigRepository) requireProfileCollection() (*drivermongo.Collection, error) {
	if r.profileCollection == nil {
		return nil, errors.New("mongo birthday profile collection is required")
	}
	return r.profileCollection, nil
}

func birthdayProfileFilter(guildID string, userID string) (bson.D, error) {
	guildID = strings.TrimSpace(guildID)
	userID = strings.TrimSpace(userID)
	if guildID == "" || userID == "" {
		return nil, domain.ErrInvalidBirthdayProfile
	}
	return bson.D{{Key: "guild", Value: guildID}, {Key: "user", Value: userID}}, nil
}

func trimBirthdayProfile(profile domain.BirthdayProfile) domain.BirthdayProfile {
	profile.GuildID = strings.TrimSpace(profile.GuildID)
	profile.UserID = strings.TrimSpace(profile.UserID)
	return profile
}

func birthdayProfileUpdate(document documents.BirthdayProfileDocument) (bson.D, error) {
	return birthdayProfileUpdateBuilder(document).Build()
}

func birthdayProfileInsertUpdate(document documents.BirthdayProfileDocument) (bson.D, error) {
	return birthdayProfileUpdateBuilder(document).
		SetOnInsert("guild", document.Guild).
		SetOnInsert("user", document.User).
		Build()
}

func birthdayProfileUpdateBuilder(document documents.BirthdayProfileDocument) *mhcatmongo.UpdateBuilder {
	return mhcatmongo.NewUpdate().
		Set("birthday_year", birthdayIntValue(document.BirthdayYear)).
		Set("birthday_month", birthdayIntValue(document.BirthdayMonth)).
		Set("birthday_day", birthdayIntValue(document.BirthdayDay)).
		Set("send_msg_hour", birthdayIntValue(document.SendHour)).
		Set("send_msg_min", birthdayIntValue(document.SendMinute)).
		Set("allow", document.Allow)
}

func birthdayIntValue(value *int) any {
	if value == nil {
		return nil
	}
	return *value
}

var _ ports.BirthdayConfigRepository = (*BirthdayConfigRepository)(nil)
