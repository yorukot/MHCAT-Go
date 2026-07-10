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
	mu      sync.Mutex
	Locks   map[string]domain.VoiceRoomLock
	Saved   []domain.VoiceRoomLock
	Deleted []domain.VoiceRoomLock
	Err     error
}

type VoiceRoomStateRepository struct {
	mu      sync.Mutex
	States  map[string]domain.VoiceRoomState
	Saved   []domain.VoiceRoomState
	Deleted []domain.VoiceRoomState
	Err     error
}

func NewVoiceRoomConfigRepository() *VoiceRoomConfigRepository {
	return &VoiceRoomConfigRepository{Configs: map[string]domain.VoiceRoomConfig{}}
}

func NewVoiceRoomLockRepository() *VoiceRoomLockRepository {
	return &VoiceRoomLockRepository{Locks: map[string]domain.VoiceRoomLock{}}
}

func NewVoiceRoomStateRepository() *VoiceRoomStateRepository {
	return &VoiceRoomStateRepository{States: map[string]domain.VoiceRoomState{}}
}

func (r *VoiceRoomConfigRepository) GetVoiceRoomConfigByTrigger(ctx context.Context, guildID string, triggerChannelID string) (domain.VoiceRoomConfig, error) {
	if err := r.ready(ctx); err != nil {
		return domain.VoiceRoomConfig{}, err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.Configs == nil {
		r.Configs = map[string]domain.VoiceRoomConfig{}
	}
	config, ok := r.Configs[voiceRoomKey(guildID, triggerChannelID)]
	if !ok {
		return domain.VoiceRoomConfig{}, ports.ErrVoiceRoomConfigMissing
	}
	return config, nil
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

func (r *VoiceRoomLockRepository) DeleteVoiceRoomLock(ctx context.Context, guildID string, channelID string) error {
	if err := r.ready(ctx); err != nil {
		return err
	}
	guildID = strings.TrimSpace(guildID)
	channelID = strings.TrimSpace(channelID)
	if guildID == "" || channelID == "" {
		return domain.ErrInvalidVoiceRoomLock
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.Locks == nil {
		r.Locks = map[string]domain.VoiceRoomLock{}
	}
	key := voiceRoomKey(guildID, channelID)
	lock, ok := r.Locks[key]
	if !ok {
		return ports.ErrVoiceRoomLockMissing
	}
	delete(r.Locks, key)
	r.Deleted = append(r.Deleted, lock)
	return nil
}

func (r *VoiceRoomLockRepository) AllowVoiceRoomLockUser(ctx context.Context, guildID string, channelID string, userID string) error {
	if err := r.ready(ctx); err != nil {
		return err
	}
	guildID = strings.TrimSpace(guildID)
	channelID = strings.TrimSpace(channelID)
	userID = strings.TrimSpace(userID)
	if guildID == "" || channelID == "" || userID == "" {
		return domain.ErrInvalidVoiceRoomLock
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.Locks == nil {
		r.Locks = map[string]domain.VoiceRoomLock{}
	}
	key := voiceRoomKey(guildID, channelID)
	lock, ok := r.Locks[key]
	if !ok {
		return ports.ErrVoiceRoomLockMissing
	}
	for _, allowed := range lock.AllowedUserIDs {
		if allowed == userID {
			r.Locks[key] = lock
			return nil
		}
	}
	lock.AllowedUserIDs = append(lock.AllowedUserIDs, userID)
	r.Locks[key] = lock
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

func (r *VoiceRoomStateRepository) GetVoiceRoomState(ctx context.Context, guildID string, channelID string) (domain.VoiceRoomState, error) {
	if err := r.ready(ctx); err != nil {
		return domain.VoiceRoomState{}, err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.States == nil {
		r.States = map[string]domain.VoiceRoomState{}
	}
	state, ok := r.States[voiceRoomKey(guildID, channelID)]
	if !ok {
		return domain.VoiceRoomState{}, ports.ErrVoiceRoomStateMissing
	}
	return state, nil
}

func (r *VoiceRoomStateRepository) SaveVoiceRoomState(ctx context.Context, state domain.VoiceRoomState) error {
	if err := r.ready(ctx); err != nil {
		return err
	}
	state.GuildID = strings.TrimSpace(state.GuildID)
	state.ChannelID = strings.TrimSpace(state.ChannelID)
	if err := state.Validate(); err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.States == nil {
		r.States = map[string]domain.VoiceRoomState{}
	}
	key := voiceRoomKey(state.GuildID, state.ChannelID)
	if _, exists := r.States[key]; !exists {
		r.States[key] = state
		r.Saved = append(r.Saved, state)
	}
	return nil
}

func (r *VoiceRoomStateRepository) DeleteVoiceRoomState(ctx context.Context, guildID string, channelID string) error {
	if err := r.ready(ctx); err != nil {
		return err
	}
	guildID = strings.TrimSpace(guildID)
	channelID = strings.TrimSpace(channelID)
	if guildID == "" || channelID == "" {
		return domain.ErrInvalidVoiceRoomConfig
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.States == nil {
		r.States = map[string]domain.VoiceRoomState{}
	}
	key := voiceRoomKey(guildID, channelID)
	state, ok := r.States[key]
	if !ok {
		return ports.ErrVoiceRoomStateMissing
	}
	delete(r.States, key)
	r.Deleted = append(r.Deleted, state)
	return nil
}

func (r *VoiceRoomStateRepository) ready(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return r.Err
}

var _ ports.VoiceRoomStateRepository = (*VoiceRoomStateRepository)(nil)
