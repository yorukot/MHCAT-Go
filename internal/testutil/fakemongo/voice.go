package fakemongo

import (
	"context"
	"strings"
	"sync"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
)

type VoiceRoomConfigRepository struct {
	mu      sync.Mutex
	Configs map[string]domain.VoiceRoomConfig
	Saved   []domain.VoiceRoomConfig
	Err     error
}

type VoiceRoomLockRepository struct {
	mu    sync.Mutex
	Locks map[string]domain.VoiceRoomLock
	Saved []domain.VoiceRoomLock
	Err   error
}

func NewVoiceRoomConfigRepository() *VoiceRoomConfigRepository {
	return &VoiceRoomConfigRepository{Configs: map[string]domain.VoiceRoomConfig{}}
}

func NewVoiceRoomLockRepository() *VoiceRoomLockRepository {
	return &VoiceRoomLockRepository{Locks: map[string]domain.VoiceRoomLock{}}
}

func (r *VoiceRoomConfigRepository) SaveVoiceRoomConfig(ctx context.Context, config domain.VoiceRoomConfig) error {
	if err := r.ready(ctx); err != nil {
		return err
	}
	if err := config.Validate(); err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.Configs == nil {
		r.Configs = map[string]domain.VoiceRoomConfig{}
	}
	key := voiceRoomKey(config.GuildID, config.TriggerChannelID)
	r.Configs[key] = config
	r.Saved = append(r.Saved, config)
	return nil
}

func (r *VoiceRoomConfigRepository) DeleteVoiceRoomConfigByTrigger(ctx context.Context, guildID string, triggerChannelID string) error {
	if err := r.ready(ctx); err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.Configs == nil {
		r.Configs = map[string]domain.VoiceRoomConfig{}
	}
	key := voiceRoomKey(guildID, triggerChannelID)
	if _, ok := r.Configs[key]; !ok {
		return ports.ErrVoiceRoomConfigMissing
	}
	delete(r.Configs, key)
	return nil
}

func (r *VoiceRoomConfigRepository) DeleteVoiceRoomConfigsByParent(ctx context.Context, guildID string, parentID string) error {
	if err := r.ready(ctx); err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.Configs == nil {
		r.Configs = map[string]domain.VoiceRoomConfig{}
	}
	guildID = strings.TrimSpace(guildID)
	parentID = strings.TrimSpace(parentID)
	deleted := false
	for key, config := range r.Configs {
		if strings.TrimSpace(config.GuildID) == guildID && strings.TrimSpace(config.ParentID) == parentID {
			delete(r.Configs, key)
			deleted = true
		}
	}
	if !deleted {
		return ports.ErrVoiceRoomConfigMissing
	}
	return nil
}

func (r *VoiceRoomConfigRepository) Last() (domain.VoiceRoomConfig, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.Saved) == 0 {
		return domain.VoiceRoomConfig{}, false
	}
	return r.Saved[len(r.Saved)-1], true
}

func (r *VoiceRoomConfigRepository) ready(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return r.Err
}

func voiceRoomKey(guildID string, triggerChannelID string) string {
	return strings.TrimSpace(guildID) + "\x00" + strings.TrimSpace(triggerChannelID)
}

var _ ports.VoiceRoomConfigRepository = (*VoiceRoomConfigRepository)(nil)

func (r *VoiceRoomLockRepository) GetVoiceRoomLock(ctx context.Context, guildID string, channelID string) (domain.VoiceRoomLock, error) {
	if err := r.ready(ctx); err != nil {
		return domain.VoiceRoomLock{}, err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.Locks == nil {
		r.Locks = map[string]domain.VoiceRoomLock{}
	}
	lock, ok := r.Locks[voiceRoomKey(guildID, channelID)]
	if !ok {
		return domain.VoiceRoomLock{}, ports.ErrVoiceRoomLockMissing
	}
	return lock, nil
}

func (r *VoiceRoomLockRepository) SaveVoiceRoomLock(ctx context.Context, lock domain.VoiceRoomLock) error {
	if err := r.ready(ctx); err != nil {
		return err
	}
	lock = lock.Normalize()
	if err := lock.Validate(); err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.Locks == nil {
		r.Locks = map[string]domain.VoiceRoomLock{}
	}
	r.Locks[voiceRoomKey(lock.GuildID, lock.ChannelID)] = lock
	r.Saved = append(r.Saved, lock)
	return nil
}

func (r *VoiceRoomLockRepository) Last() (domain.VoiceRoomLock, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.Saved) == 0 {
		return domain.VoiceRoomLock{}, false
	}
	return r.Saved[len(r.Saved)-1], true
}

func (r *VoiceRoomLockRepository) ready(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return r.Err
}

var _ ports.VoiceRoomLockRepository = (*VoiceRoomLockRepository)(nil)
