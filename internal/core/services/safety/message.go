package safety

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
)

const (
	antiScamConfigCacheTTL = 5 * time.Second
	scamURLCatalogCacheTTL = time.Minute
	cacheRefreshRetryTTL   = 5 * time.Second
)

type MessageService struct {
	configs ports.AntiScamConfigRepository
	catalog ports.ScamURLCatalog
	cache   *messageScanCache
}

type messageScanCache struct {
	now     func() time.Time
	configs sync.Map
	catalog scamURLCatalogCache
}

type antiScamConfigCacheEntry struct {
	mu      sync.Mutex
	config  domain.AntiScamConfig
	exists  bool
	expires time.Time
}

type scamURLCatalogCache struct {
	mu      sync.Mutex
	urls    []string
	expires time.Time
}

type MessageScanResult struct {
	MatchedURL string
	Delete     bool
}

func NewMessageService(configs ports.AntiScamConfigRepository, catalog ports.ScamURLCatalog) MessageService {
	return newMessageService(configs, catalog, time.Now)
}

func newMessageService(configs ports.AntiScamConfigRepository, catalog ports.ScamURLCatalog, now func() time.Time) MessageService {
	return MessageService{
		configs: configs,
		catalog: catalog,
		cache:   &messageScanCache{now: now},
	}
}

func (s MessageService) Scan(ctx context.Context, guildID string, content string) (MessageScanResult, error) {
	if err := ctx.Err(); err != nil {
		return MessageScanResult{}, err
	}
	if s.configs == nil || s.catalog == nil {
		return MessageScanResult{}, domain.ErrInvalidAntiScamConfig
	}
	guildID = strings.TrimSpace(guildID)
	if guildID == "" {
		return MessageScanResult{}, domain.ErrInvalidAntiScamConfig
	}
	if strings.TrimSpace(content) == "" {
		return MessageScanResult{}, ctx.Err()
	}
	config, exists, err := s.cachedConfig(ctx, guildID)
	if err != nil {
		return MessageScanResult{}, err
	}
	if !exists || !config.Open {
		return MessageScanResult{}, ctx.Err()
	}
	matched, ok, err := s.findScamURL(ctx, content)
	if err != nil {
		return MessageScanResult{}, err
	}
	if !ok {
		return MessageScanResult{}, ctx.Err()
	}
	return MessageScanResult{MatchedURL: matched, Delete: true}, ctx.Err()
}

func (s MessageService) cachedConfig(ctx context.Context, guildID string) (domain.AntiScamConfig, bool, error) {
	if s.cache == nil || s.cache.now == nil {
		config, err := s.configs.FindAntiScamConfig(ctx, guildID)
		if errors.Is(err, ports.ErrAntiScamConfigMissing) {
			return domain.AntiScamConfig{}, false, nil
		}
		return config, err == nil, err
	}
	entryValue, _ := s.cache.configs.LoadOrStore(guildID, &antiScamConfigCacheEntry{})
	entry := entryValue.(*antiScamConfigCacheEntry)
	entry.mu.Lock()
	defer entry.mu.Unlock()
	now := s.cache.now()
	if now.Before(entry.expires) {
		return entry.config, entry.exists, ctx.Err()
	}
	config, err := s.configs.FindAntiScamConfig(ctx, guildID)
	exists := true
	if errors.Is(err, ports.ErrAntiScamConfigMissing) {
		err = nil
		exists = false
	}
	if err != nil {
		if ctx.Err() != nil {
			return domain.AntiScamConfig{}, false, err
		}
		if !entry.expires.IsZero() {
			entry.expires = now.Add(cacheRefreshRetryTTL)
			return entry.config, entry.exists, nil
		}
		return domain.AntiScamConfig{}, false, err
	}
	entry.config = config
	entry.exists = exists
	entry.expires = now.Add(antiScamConfigCacheTTL)
	return config, exists, ctx.Err()
}

func (s MessageService) findScamURL(ctx context.Context, content string) (string, bool, error) {
	lister, ok := s.catalog.(ports.ScamURLLister)
	if !ok || s.cache == nil || s.cache.now == nil {
		return s.catalog.FindScamURLInContent(ctx, content)
	}
	urls, err := s.cachedScamURLs(ctx, lister)
	if err != nil {
		return "", false, err
	}
	for _, url := range urls {
		if url != "" && strings.Contains(content, url) {
			return url, true, ctx.Err()
		}
	}
	return "", false, ctx.Err()
}

func (s MessageService) cachedScamURLs(ctx context.Context, lister ports.ScamURLLister) ([]string, error) {
	cache := &s.cache.catalog
	cache.mu.Lock()
	defer cache.mu.Unlock()
	now := s.cache.now()
	if now.Before(cache.expires) {
		return cache.urls, ctx.Err()
	}
	urls, err := lister.ListScamURLs(ctx)
	if err != nil {
		if ctx.Err() != nil {
			return nil, err
		}
		if cache.urls != nil {
			cache.expires = now.Add(cacheRefreshRetryTTL)
			return cache.urls, nil
		}
		return nil, err
	}
	// Publish a new immutable snapshot so readers from the previous TTL window
	// never share a backing array with a concurrent refresh.
	cache.urls = append([]string(nil), urls...)
	cache.expires = now.Add(scamURLCatalogCacheTTL)
	return cache.urls, ctx.Err()
}
