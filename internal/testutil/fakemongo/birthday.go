package fakemongo

import (
	"context"
	"sync"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
)

type BirthdayConfigRepository struct {
	mu      sync.Mutex
	Err     error
	Saved   []domain.BirthdayConfig
	Configs map[string]domain.BirthdayConfig
}

func (r *BirthdayConfigRepository) SaveBirthdayConfig(ctx context.Context, config domain.BirthdayConfig) error {
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
		r.Configs = map[string]domain.BirthdayConfig{}
	}
	r.Saved = append(r.Saved, config)
	r.Configs[config.GuildID] = config
	return nil
}

func (r *BirthdayConfigRepository) Last() (domain.BirthdayConfig, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.Saved) == 0 {
		return domain.BirthdayConfig{}, false
	}
	return r.Saved[len(r.Saved)-1], true
}

var _ ports.BirthdayConfigRepository = (*BirthdayConfigRepository)(nil)
