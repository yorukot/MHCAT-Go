package fakemongo

import (
	"context"
	"strings"
	"sync"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
)

type BirthdayConfigRepository struct {
	mu           sync.Mutex
	Err          error
	Saved        []domain.BirthdayConfig
	Configs      map[string]domain.BirthdayConfig
	Profiles     map[string]domain.BirthdayProfile
	ProfileSaved []domain.BirthdayProfile
}

func (r *BirthdayConfigRepository) FindBirthdayConfig(ctx context.Context, guildID string) (domain.BirthdayConfig, error) {
	if err := r.ready(ctx); err != nil {
		return domain.BirthdayConfig{}, err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	config, ok := r.Configs[strings.TrimSpace(guildID)]
	if !ok {
		return domain.BirthdayConfig{}, ports.ErrBirthdayConfigMissing
	}
	return config, nil
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

func (r *BirthdayConfigRepository) FindBirthdayProfile(ctx context.Context, guildID string, userID string) (domain.BirthdayProfile, error) {
	if err := r.ready(ctx); err != nil {
		return domain.BirthdayProfile{}, err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	profile, ok := r.Profiles[birthdayProfileKey(guildID, userID)]
	if !ok {
		return domain.BirthdayProfile{}, ports.ErrBirthdayProfileMissing
	}
	return profile, nil
}

func (r *BirthdayConfigRepository) SaveBirthdayProfile(ctx context.Context, profile domain.BirthdayProfile) error {
	if err := r.ready(ctx); err != nil {
		return err
	}
	profile.GuildID = strings.TrimSpace(profile.GuildID)
	profile.UserID = strings.TrimSpace(profile.UserID)
	if err := profile.ValidateIdentity(); err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.Profiles == nil {
		r.Profiles = map[string]domain.BirthdayProfile{}
	}
	r.Profiles[birthdayProfileKey(profile.GuildID, profile.UserID)] = profile
	r.ProfileSaved = append(r.ProfileSaved, profile)
	return nil
}

func (r *BirthdayConfigRepository) DeleteBirthdayProfile(ctx context.Context, guildID string, userID string) error {
	if err := r.ready(ctx); err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	key := birthdayProfileKey(guildID, userID)
	if _, ok := r.Profiles[key]; !ok {
		return ports.ErrBirthdayProfileMissing
	}
	delete(r.Profiles, key)
	return nil
}

func (r *BirthdayConfigRepository) ListBirthdayProfiles(ctx context.Context, guildID string) ([]domain.BirthdayProfile, error) {
	if err := r.ready(ctx); err != nil {
		return nil, err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	guildID = strings.TrimSpace(guildID)
	if guildID == "" {
		return nil, domain.ErrInvalidBirthdayProfile
	}
	profiles := []domain.BirthdayProfile{}
	for _, profile := range r.Profiles {
		if strings.TrimSpace(profile.GuildID) == guildID {
			profiles = append(profiles, profile)
		}
	}
	return profiles, nil
}

func (r *BirthdayConfigRepository) ready(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return r.Err
}

func birthdayProfileKey(guildID string, userID string) string {
	return strings.TrimSpace(guildID) + "/" + strings.TrimSpace(userID)
}

var _ ports.BirthdayConfigRepository = (*BirthdayConfigRepository)(nil)
