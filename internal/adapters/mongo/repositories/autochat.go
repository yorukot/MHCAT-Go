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
)

const AutoChatConfigCollectionName = "chats"
const autoChatConfigCacheTTL = 5 * time.Second

type AutoChatConfigRepository struct {
	collection *drivermongo.Collection
	now        func() time.Time
	cache      sync.Map
}

type autoChatConfigCacheEntry struct {
	mu      sync.Mutex
	config  domain.AutoChatConfig
	exists  bool
	expires time.Time
}

func NewAutoChatConfigRepository(collection *drivermongo.Collection) (*AutoChatConfigRepository, error) {
	if collection == nil {
		return nil, errors.New("mongo autochat collection is required")
	}
	return &AutoChatConfigRepository{collection: collection, now: time.Now}, nil
}

func NewAutoChatConfigRepositoryFromDatabase(database *drivermongo.Database) (*AutoChatConfigRepository, error) {
	if database == nil {
		return nil, errors.New("mongo database is required")
	}
	return NewAutoChatConfigRepository(database.Collection(AutoChatConfigCollectionName))
}

func (r *AutoChatConfigRepository) GetAutoChatConfig(ctx context.Context, guildID string) (domain.AutoChatConfig, error) {
	if err := ctx.Err(); err != nil {
		return domain.AutoChatConfig{}, err
	}
	guildID = strings.TrimSpace(guildID)
	if guildID == "" {
		return domain.AutoChatConfig{}, domain.ErrInvalidAutoChatConfig
	}
	entryValue, _ := r.cache.LoadOrStore(guildID, &autoChatConfigCacheEntry{})
	entry := entryValue.(*autoChatConfigCacheEntry)
	entry.mu.Lock()
	defer entry.mu.Unlock()
	now := r.currentTime()
	if now.Before(entry.expires) {
		if !entry.exists {
			return domain.AutoChatConfig{}, ports.ErrAutoChatConfigMissing
		}
		return entry.config, ctx.Err()
	}
	var document documents.AutoChatConfigReadDocument
	if err := r.collection.FindOne(ctx, bson.D{{Key: "guild", Value: guildID}}).Decode(&document); err != nil {
		if err == drivermongo.ErrNoDocuments {
			entry.config = domain.AutoChatConfig{}
			entry.exists = false
			entry.expires = now.Add(autoChatConfigCacheTTL)
			return domain.AutoChatConfig{}, ports.ErrAutoChatConfigMissing
		}
		return domain.AutoChatConfig{}, mhcatmongo.MapError(fmt.Errorf("get autochat config: %w", err))
	}
	entry.config = document.ToDomain()
	entry.exists = true
	entry.expires = now.Add(autoChatConfigCacheTTL)
	return entry.config, ctx.Err()
}

func (r *AutoChatConfigRepository) SaveAutoChatConfig(ctx context.Context, config domain.AutoChatConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := config.Validate(); err != nil {
		return err
	}
	type configRow struct {
		ID any `bson:"_id"`
	}
	var row configRow
	err := r.collection.FindOne(ctx, bson.D{{Key: "guild", Value: config.GuildID}}).Decode(&row)
	if err != nil && err != drivermongo.ErrNoDocuments {
		return mhcatmongo.MapError(fmt.Errorf("get autochat config for replacement: %w", err))
	}
	if err == nil {
		result, deleteErr := r.collection.DeleteOne(ctx, bson.D{{Key: "_id", Value: row.ID}})
		if deleteErr != nil {
			return mhcatmongo.MapError(fmt.Errorf("delete autochat config for replacement: %w", deleteErr))
		}
		if result.DeletedCount == 0 {
			return mhcatmongo.MapError(errors.New("autochat config changed before replacement"))
		}
		r.cache.Delete(config.GuildID)
	}
	if _, err := r.collection.InsertOne(ctx, bson.D{
		{Key: "guild", Value: config.GuildID},
		{Key: "channel", Value: config.ChannelID},
	}); err != nil {
		return mhcatmongo.MapError(fmt.Errorf("insert autochat config replacement: %w", err))
	}
	r.storeCachedConfig(config.GuildID, config, true)
	return ctx.Err()
}

func (r *AutoChatConfigRepository) DeleteAutoChatConfig(ctx context.Context, guildID string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	guildID = strings.TrimSpace(guildID)
	if guildID == "" {
		return domain.ErrInvalidAutoChatConfig
	}
	type configRow struct {
		ID any `bson:"_id"`
	}
	var row configRow
	if err := r.collection.FindOne(ctx, bson.D{{Key: "guild", Value: guildID}}).Decode(&row); err != nil {
		if err == drivermongo.ErrNoDocuments {
			return ports.ErrAutoChatConfigMissing
		}
		return mhcatmongo.MapError(fmt.Errorf("get autochat config for delete: %w", err))
	}
	result, err := r.collection.DeleteOne(ctx, bson.D{{Key: "_id", Value: row.ID}})
	if err != nil {
		return mhcatmongo.MapError(fmt.Errorf("delete autochat config: %w", err))
	}
	if result.DeletedCount == 0 {
		r.storeCachedConfig(guildID, domain.AutoChatConfig{}, false)
		return ports.ErrAutoChatConfigMissing
	}
	r.storeCachedConfig(guildID, domain.AutoChatConfig{}, false)
	return ctx.Err()
}

func (r *AutoChatConfigRepository) storeCachedConfig(guildID string, config domain.AutoChatConfig, exists bool) {
	guildID = strings.TrimSpace(guildID)
	if guildID == "" {
		return
	}
	entryValue, _ := r.cache.LoadOrStore(guildID, &autoChatConfigCacheEntry{})
	entry := entryValue.(*autoChatConfigCacheEntry)
	entry.mu.Lock()
	entry.config = config
	entry.exists = exists
	entry.expires = r.currentTime().Add(autoChatConfigCacheTTL)
	entry.mu.Unlock()
}

func (r *AutoChatConfigRepository) currentTime() time.Time {
	if r.now != nil {
		return r.now()
	}
	return time.Now()
}

var _ ports.AutoChatConfigRepository = (*AutoChatConfigRepository)(nil)
var _ ports.AutoChatConfigReader = (*AutoChatConfigRepository)(nil)
