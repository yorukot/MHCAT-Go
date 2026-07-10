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

const (
	GuildConfigCollectionName             = "guilds"
	BoundAnnouncementConfigCollectionName = "ann_all_sets"
)

type AnnouncementConfigRepository struct {
	guilds     *drivermongo.Collection
	annAllSets *drivermongo.Collection
}

func NewAnnouncementConfigRepository(guilds *drivermongo.Collection, annAllSets *drivermongo.Collection) (*AnnouncementConfigRepository, error) {
	if guilds == nil {
		return nil, errors.New("mongo guild collection is required")
	}
	if annAllSets == nil {
		return nil, errors.New("mongo ann_all_sets collection is required")
	}
	return &AnnouncementConfigRepository{guilds: guilds, annAllSets: annAllSets}, nil
}

func NewAnnouncementConfigRepositoryFromDatabase(database *drivermongo.Database) (*AnnouncementConfigRepository, error) {
	if database == nil {
		return nil, errors.New("mongo database is required")
	}
	return NewAnnouncementConfigRepository(
		database.Collection(GuildConfigCollectionName),
		database.Collection(BoundAnnouncementConfigCollectionName),
	)
}

func (r *AnnouncementConfigRepository) GetAnnouncementChannel(ctx context.Context, guildID string) (domain.AnnouncementChannelConfig, error) {
	if err := ctx.Err(); err != nil {
		return domain.AnnouncementChannelConfig{}, err
	}
	guildID = strings.TrimSpace(guildID)
	if guildID == "" {
		return domain.AnnouncementChannelConfig{}, domain.ErrInvalidAnnouncementConfig
	}
	var document documents.GuildAnnouncementReadDocument
	err := r.guilds.FindOne(ctx, bson.D{{Key: "guild", Value: guildID}}).Decode(&document)
	if errors.Is(err, drivermongo.ErrNoDocuments) {
		return domain.AnnouncementChannelConfig{}, ports.ErrAnnouncementChannelMissing
	}
	if err != nil {
		return domain.AnnouncementChannelConfig{}, mhcatmongo.MapError(fmt.Errorf("get announcement channel: %w", err))
	}
	config := document.ToDomain()
	if config.GuildID == "" {
		config.GuildID = guildID
	}
	if config.ChannelID == "" || config.ChannelID == "0" {
		return domain.AnnouncementChannelConfig{}, ports.ErrAnnouncementChannelMissing
	}
	return config, ctx.Err()
}

func (r *AnnouncementConfigRepository) GetBoundAnnouncement(ctx context.Context, guildID string, channelID string) (domain.BoundAnnouncementConfig, error) {
	if err := ctx.Err(); err != nil {
		return domain.BoundAnnouncementConfig{}, err
	}
	guildID = strings.TrimSpace(guildID)
	channelID = strings.TrimSpace(channelID)
	if guildID == "" || channelID == "" {
		return domain.BoundAnnouncementConfig{}, domain.ErrInvalidAnnouncementConfig
	}
	var document documents.BoundAnnouncementReadDocument
	err := r.annAllSets.FindOne(ctx, bson.D{{Key: "guild", Value: guildID}, {Key: "announcement_id", Value: channelID}}).Decode(&document)
	if errors.Is(err, drivermongo.ErrNoDocuments) {
		return domain.BoundAnnouncementConfig{}, ports.ErrBoundAnnouncementConfigMissing
	}
	if err != nil {
		return domain.BoundAnnouncementConfig{}, mhcatmongo.MapError(fmt.Errorf("get bound announcement: %w", err))
	}
	config := document.ToDomain()
	if config.GuildID == "" {
		config.GuildID = guildID
	}
	if config.ChannelID == "" {
		config.ChannelID = channelID
	}
	return config, ctx.Err()
}

func (r *AnnouncementConfigRepository) SetAnnouncementChannel(ctx context.Context, config domain.AnnouncementChannelConfig) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	if err := config.Validate(); err != nil {
		return false, err
	}
	document := documents.GuildAnnouncementDocumentFromDomain(config)
	update, err := mhcatmongo.NewUpdate().Set("announcement_id", document.AnnouncementID).Build()
	if err != nil {
		return false, err
	}
	result, err := r.guilds.UpdateMany(ctx, bson.D{{Key: "guild", Value: document.Guild}}, update)
	if err != nil {
		return false, mhcatmongo.MapError(fmt.Errorf("update announcement channel: %w", err))
	}
	if result.MatchedCount > 0 {
		return false, ctx.Err()
	}
	insertUpdate, err := mhcatmongo.NewUpdate().
		Set("announcement_id", document.AnnouncementID).
		SetOnInsert("guild", document.Guild).
		Build()
	if err != nil {
		return false, err
	}
	_, err = r.guilds.UpdateOne(ctx, bson.D{{Key: "guild", Value: document.Guild}}, insertUpdate, options.UpdateOne().SetUpsert(true))
	if err != nil {
		return false, mhcatmongo.MapError(fmt.Errorf("upsert announcement channel: %w", err))
	}
	return true, ctx.Err()
}

func (r *AnnouncementConfigRepository) SetBoundAnnouncement(ctx context.Context, config domain.BoundAnnouncementConfig) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	if err := config.Validate(); err != nil {
		return false, err
	}
	document := documents.BoundAnnouncementDocumentFromDomain(config)
	filter := bson.D{{Key: "guild", Value: document.Guild}, {Key: "announcement_id", Value: document.AnnouncementID}}
	update, err := mhcatmongo.NewUpdate().
		Set("tag", document.Tag).
		Set("color", document.Color).
		Set("title", document.Title).
		Build()
	if err != nil {
		return false, err
	}
	result, err := r.annAllSets.UpdateMany(ctx, filter, update)
	if err != nil {
		return false, mhcatmongo.MapError(fmt.Errorf("update bound announcement: %w", err))
	}
	if result.MatchedCount > 0 {
		return false, ctx.Err()
	}
	insertUpdate, err := mhcatmongo.NewUpdate().
		Set("tag", document.Tag).
		Set("color", document.Color).
		Set("title", document.Title).
		SetOnInsert("guild", document.Guild).
		SetOnInsert("announcement_id", document.AnnouncementID).
		Build()
	if err != nil {
		return false, err
	}
	_, err = r.annAllSets.UpdateOne(ctx, filter, insertUpdate, options.UpdateOne().SetUpsert(true))
	if err != nil {
		return false, mhcatmongo.MapError(fmt.Errorf("upsert bound announcement: %w", err))
	}
	return true, ctx.Err()
}

func (r *AnnouncementConfigRepository) DeleteBoundAnnouncement(ctx context.Context, guildID string, channelID string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	guildID = strings.TrimSpace(guildID)
	channelID = strings.TrimSpace(channelID)
	if guildID == "" || channelID == "" {
		return domain.ErrInvalidAnnouncementConfig
	}
	result, err := r.annAllSets.DeleteMany(ctx, bson.D{{Key: "guild", Value: guildID}, {Key: "announcement_id", Value: channelID}})
	if err != nil {
		return mhcatmongo.MapError(fmt.Errorf("delete bound announcement: %w", err))
	}
	if result.DeletedCount == 0 {
		return ports.ErrBoundAnnouncementConfigMissing
	}
	return ctx.Err()
}

var _ ports.AnnouncementConfigRepository = (*AnnouncementConfigRepository)(nil)
