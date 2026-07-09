package fakemongo

import (
	"context"
	"sync"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
)

type LoggingConfigRepository struct {
	mu      sync.Mutex
	Err     error
	Saved   []domain.LoggingConfig
	Configs map[string]domain.LoggingConfig
}

func (r *LoggingConfigRepository) SaveLoggingConfig(ctx context.Context, config domain.LoggingConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if r.Err != nil {
		return r.Err
	}
	if err := config.Validate(); err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.Configs == nil {
		r.Configs = map[string]domain.LoggingConfig{}
	}
	r.Saved = append(r.Saved, config)
	r.Configs[config.GuildID] = config
	return nil
}

func (r *LoggingConfigRepository) Last() (domain.LoggingConfig, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.Saved) == 0 {
		return domain.LoggingConfig{}, false
	}
	return r.Saved[len(r.Saved)-1], true
}

var _ ports.LoggingConfigRepository = (*LoggingConfigRepository)(nil)
