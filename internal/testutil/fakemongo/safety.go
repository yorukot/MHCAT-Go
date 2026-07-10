package fakemongo

import (
	"context"
	"strings"
	"sync"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
)

type AntiScamConfigRepository struct {
	mu      sync.Mutex
	Configs map[string]domain.AntiScamConfig
	Err     error
	Saved   []domain.AntiScamConfig
}

func NewAntiScamConfigRepository() *AntiScamConfigRepository {
	return &AntiScamConfigRepository{Configs: map[string]domain.AntiScamConfig{}}
}

func (r *AntiScamConfigRepository) FindAntiScamConfig(ctx context.Context, guildID string) (domain.AntiScamConfig, error) {
	if err := r.ready(ctx); err != nil {
		return domain.AntiScamConfig{}, err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	config, ok := r.Configs[strings.TrimSpace(guildID)]
	if !ok {
		return domain.AntiScamConfig{}, ports.ErrAntiScamConfigMissing
	}
	return config, nil
}

func (r *AntiScamConfigRepository) SaveAntiScamConfig(ctx context.Context, config domain.AntiScamConfig) error {
	if err := r.ready(ctx); err != nil {
		return err
	}
	config.GuildID = strings.TrimSpace(config.GuildID)
	if err := config.Validate(); err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.Configs == nil {
		r.Configs = map[string]domain.AntiScamConfig{}
	}
	r.Configs[config.GuildID] = config
	r.Saved = append(r.Saved, config)
	return nil
}

func (r *AntiScamConfigRepository) Last() (domain.AntiScamConfig, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.Saved) == 0 {
		return domain.AntiScamConfig{}, false
	}
	return r.Saved[len(r.Saved)-1], true
}

func (r *AntiScamConfigRepository) ready(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return r.Err
}

var _ ports.AntiScamConfigRepository = (*AntiScamConfigRepository)(nil)

type ScamURLCatalogRepository struct {
	mu      sync.Mutex
	Known   []string
	Checked []string
	Err     error
}

func NewScamURLCatalogRepository() *ScamURLCatalogRepository {
	return &ScamURLCatalogRepository{}
}

func (r *ScamURLCatalogRepository) ContainsScamURL(ctx context.Context, rawURL string) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	if r.Err != nil {
		return false, r.Err
	}
	rawURL = strings.TrimSpace(rawURL)
	r.mu.Lock()
	defer r.mu.Unlock()
	r.Checked = append(r.Checked, rawURL)
	for _, known := range r.Known {
		if strings.Contains(known, rawURL) {
			return true, nil
		}
	}
	return false, nil
}

func (r *ScamURLCatalogRepository) FindScamURLInContent(ctx context.Context, content string) (string, bool, error) {
	if err := ctx.Err(); err != nil {
		return "", false, err
	}
	if r.Err != nil {
		return "", false, r.Err
	}
	content = strings.TrimSpace(content)
	r.mu.Lock()
	defer r.mu.Unlock()
	r.Checked = append(r.Checked, content)
	for _, known := range r.Known {
		known = strings.TrimSpace(known)
		if known != "" && strings.Contains(content, known) {
			return known, true, nil
		}
	}
	return "", false, nil
}

var _ ports.ScamURLCatalog = (*ScamURLCatalogRepository)(nil)
