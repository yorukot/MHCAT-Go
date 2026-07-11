package repositories

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

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
	boundAnnouncementCacheTTL             = 30 * time.Second
	boundAnnouncementCacheRetryTTL        = 5 * time.Second
)

type AnnouncementConfigRepository struct {
	guilds     *drivermongo.Collection
	annAllSets *drivermongo.Collection
	now        func() time.Time
	boundCache sync.Map
}

type boundAnnouncementCacheKey struct {
	guildID   string
	channelID string
}

type boundAnnouncementCacheEntry struct {
	mu      sync.Mutex
	config  domain.BoundAnnouncementConfig
	exists  bool
	expires time.Time
}

func NewAnnouncementConfigRepository(guilds *drivermongo.Collection, annAllSets *drivermongo.Collection) (*AnnouncementConfigRepository, error) {
	if guilds == nil {
		return nil, errors.New("mongo guild collection is required")
	}
	if annAllSets == nil {
		return nil, errors.New("mongo ann_all_sets collection is required")
	}
	return &AnnouncementConfigRepository{guilds: guilds, annAllSets: annAllSets, now: time.Now}, nil
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
	key := boundAnnouncementCacheKey{guildID: guildID, channelID: channelID}
	entryValue, _ := r.boundCache.LoadOrStore(key, &boundAnnouncementCacheEntry{})
	entry := entryValue.(*boundAnnouncementCacheEntry)
	entry.mu.Lock()
	defer entry.mu.Unlock()
	now := r.currentTime()
	if now.Before(entry.expires) {
		if !entry.exists {
			return domain.BoundAnnouncementConfig{}, ports.ErrBoundAnnouncementConfigMissing
		}
		return entry.config, ctx.Err()
	}
	var document documents.BoundAnnouncementReadDocument
	err := r.annAllSets.FindOne(ctx, bson.D{{Key: "guild", Value: guildID}, {Key: "announcement_id", Value: channelID}}).Decode(&document)
	if errors.Is(err, drivermongo.ErrNoDocuments) {
		entry.config = domain.BoundAnnouncementConfig{}
		entry.exists = false
		entry.expires = now.Add(boundAnnouncementCacheTTL)
		return domain.BoundAnnouncementConfig{}, ports.ErrBoundAnnouncementConfigMissing
	}
	if err != nil {
		if ctx.Err() == nil && !entry.expires.IsZero() {
			entry.expires = now.Add(boundAnnouncementCacheRetryTTL)
			if !entry.exists {
				return domain.BoundAnnouncementConfig{}, ports.ErrBoundAnnouncementConfigMissing
			}
			return entry.config, nil
		}
		return domain.BoundAnnouncementConfig{}, mhcatmongo.MapError(fmt.Errorf("get bound announcement: %w", err))
	}
	config := document.ToDomain()
	if config.GuildID == "" {
		config.GuildID = guildID
	}
	if config.ChannelID == "" {
		config.ChannelID = channelID
	}
	entry.config = config
	entry.exists = true
	entry.expires = now.Add(boundAnnouncementCacheTTL)
	return entry.config, ctx.Err()
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
		r.storeCachedBoundAnnouncement(config.GuildID, config.ChannelID, config, true)
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
	r.storeCachedBoundAnnouncement(config.GuildID, config.ChannelID, config, true)
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
		r.storeCachedBoundAnnouncement(guildID, channelID, domain.BoundAnnouncementConfig{}, false)
		return ports.ErrBoundAnnouncementConfigMissing
	}
	r.storeCachedBoundAnnouncement(guildID, channelID, domain.BoundAnnouncementConfig{}, false)
	return ctx.Err()
}

func (r *AnnouncementConfigRepository) storeCachedBoundAnnouncement(guildID string, channelID string, config domain.BoundAnnouncementConfig, exists bool) {
	guildID = strings.TrimSpace(guildID)
	channelID = strings.TrimSpace(channelID)
	if guildID == "" || channelID == "" {
		return
	}
	key := boundAnnouncementCacheKey{guildID: guildID, channelID: channelID}
	entryValue, _ := r.boundCache.LoadOrStore(key, &boundAnnouncementCacheEntry{})
	entry := entryValue.(*boundAnnouncementCacheEntry)
	entry.mu.Lock()
	entry.config = config
	entry.exists = exists
	entry.expires = r.currentTime().Add(boundAnnouncementCacheTTL)
	entry.mu.Unlock()
}

func (r *AnnouncementConfigRepository) currentTime() time.Time {
	if r.now != nil {
		return r.now()
	}
	return time.Now()
}

var _ ports.AnnouncementConfigRepository = (*AnnouncementConfigRepository)(nil)
