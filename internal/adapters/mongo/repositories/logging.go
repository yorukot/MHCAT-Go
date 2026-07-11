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
	LoggingConfigCollectionName = "loggings"
	loggingConfigCacheTTL       = 30 * time.Second
	loggingConfigCacheRetryTTL  = 5 * time.Second
)

type LoggingConfigRepository struct {
	collection *drivermongo.Collection
	now        func() time.Time
	cache      sync.Map
}

type loggingConfigCacheEntry struct {
	mu      sync.Mutex
	config  domain.LoggingConfig
	exists  bool
	expires time.Time
}

func NewLoggingConfigRepository(collection *drivermongo.Collection) (*LoggingConfigRepository, error) {
	if collection == nil {
		return nil, errors.New("mongo logging collection is required")
	}
	return &LoggingConfigRepository{collection: collection, now: time.Now}, nil
}

func NewLoggingConfigRepositoryFromDatabase(database *drivermongo.Database) (*LoggingConfigRepository, error) {
	if database == nil {
		return nil, errors.New("mongo database is required")
	}
	return NewLoggingConfigRepository(database.Collection(LoggingConfigCollectionName))
}

func (r *LoggingConfigRepository) SaveLoggingConfig(ctx context.Context, config domain.LoggingConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	config.GuildID = strings.TrimSpace(config.GuildID)
	config.ChannelID = strings.TrimSpace(config.ChannelID)
	if err := config.Validate(); err != nil {
		return err
	}
	document := documents.LoggingConfigDocumentFromDomain(config)
	update, err := mhcatmongo.NewUpdate().
		Set("channel_id", document.ChannelID).
		Set("message_update", document.MessageUpdate).
		Set("message_delete", document.MessageDelete).
		Set("channel_update", document.ChannelUpdate).
		Set("member_voice_update", document.MemberVoiceUpdate).
		SetOnInsert("guild", document.Guild).
		Build()
	if err != nil {
		return err
	}
	_, err = r.collection.UpdateMany(
		ctx,
		bson.D{{Key: "guild", Value: document.Guild}},
		update,
		options.UpdateMany().SetUpsert(true),
	)
	if err != nil {
		return mhcatmongo.MapError(fmt.Errorf("save logging config: %w", err))
	}
	r.storeCachedLoggingConfig(config.GuildID, config, true)
	return ctx.Err()
}

func (r *LoggingConfigRepository) GetLoggingConfig(ctx context.Context, guildID string) (domain.LoggingConfig, error) {
	if err := ctx.Err(); err != nil {
		return domain.LoggingConfig{}, err
	}
	guildID = strings.TrimSpace(guildID)
	if guildID == "" {
		return domain.LoggingConfig{}, domain.ErrInvalidLoggingConfig
	}
	entryValue, _ := r.cache.LoadOrStore(guildID, &loggingConfigCacheEntry{})
	entry := entryValue.(*loggingConfigCacheEntry)
	entry.mu.Lock()
	defer entry.mu.Unlock()
	now := r.currentTime()
	if now.Before(entry.expires) {
		if !entry.exists {
			return domain.LoggingConfig{}, ports.ErrLoggingConfigMissing
		}
		return entry.config, ctx.Err()
	}
	var document documents.LoggingConfigReadDocument
	err := r.collection.FindOne(ctx, bson.D{{Key: "guild", Value: guildID}}).Decode(&document)
	if err != nil {
		if err == drivermongo.ErrNoDocuments {
			entry.config = domain.LoggingConfig{}
			entry.exists = false
			entry.expires = now.Add(loggingConfigCacheTTL)
			return domain.LoggingConfig{}, ports.ErrLoggingConfigMissing
		}
		if ctx.Err() == nil && !entry.expires.IsZero() {
			entry.expires = now.Add(loggingConfigCacheRetryTTL)
			if !entry.exists {
				return domain.LoggingConfig{}, ports.ErrLoggingConfigMissing
			}
			return entry.config, nil
		}
		return domain.LoggingConfig{}, mhcatmongo.MapError(fmt.Errorf("get logging config: %w", err))
	}
	entry.config = document.ToDomain()
	entry.exists = true
	entry.expires = now.Add(loggingConfigCacheTTL)
	return entry.config, ctx.Err()
}

func (r *LoggingConfigRepository) storeCachedLoggingConfig(guildID string, config domain.LoggingConfig, exists bool) {
	guildID = strings.TrimSpace(guildID)
	if guildID == "" {
		return
	}
	entryValue, _ := r.cache.LoadOrStore(guildID, &loggingConfigCacheEntry{})
	entry := entryValue.(*loggingConfigCacheEntry)
	entry.mu.Lock()
	entry.config = config
	entry.exists = exists
	entry.expires = r.currentTime().Add(loggingConfigCacheTTL)
	entry.mu.Unlock()
}

func (r *LoggingConfigRepository) currentTime() time.Time {
	if r.now != nil {
		return r.now()
	}
	return time.Now()
}

var _ ports.LoggingConfigRepository = (*LoggingConfigRepository)(nil)
var _ ports.LoggingConfigReader = (*LoggingConfigRepository)(nil)
