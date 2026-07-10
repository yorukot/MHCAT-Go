package voice

import (
	"context"
	"errors"
	"strings"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
)

var ErrVoiceRoomLockPasswordMismatch = errors.New("voice room lock password mismatch")

type ConfigService struct {
	Repository ports.VoiceRoomConfigRepository
}

type LockService struct {
	Repository ports.VoiceRoomLockRepository
}

type RoomService struct {
	Configs ports.VoiceRoomConfigRepository
	States  ports.VoiceRoomStateRepository
	Locks   ports.VoiceRoomLockRepository
}

func NewConfigService(repo ports.VoiceRoomConfigRepository) ConfigService {
	return ConfigService{Repository: repo}
}

func NewLockService(repo ports.VoiceRoomLockRepository) LockService {
	return LockService{Repository: repo}
}

func NewRoomService(configs ports.VoiceRoomConfigRepository, states ports.VoiceRoomStateRepository, locks ports.VoiceRoomLockRepository) RoomService {
	return RoomService{Configs: configs, States: states, Locks: locks}
}

func (s ConfigService) Save(ctx context.Context, config domain.VoiceRoomConfig) error {
	if s.Repository == nil {
		return domain.ErrInvalidVoiceRoomConfig
	}
	config.GuildID = strings.TrimSpace(config.GuildID)
	config.TriggerChannelID = strings.TrimSpace(config.TriggerChannelID)
	config.ParentID = strings.TrimSpace(config.ParentID)
	if err := config.Validate(); err != nil {
		return err
	}
	return s.Repository.SaveVoiceRoomConfig(ctx, config)
}

func (s ConfigService) DeleteByTrigger(ctx context.Context, guildID string, triggerChannelID string) error {
	if s.Repository == nil {
		return domain.ErrInvalidVoiceRoomConfig
	}
	guildID = strings.TrimSpace(guildID)
	triggerChannelID = strings.TrimSpace(triggerChannelID)
	if guildID == "" || triggerChannelID == "" {
		return domain.ErrInvalidVoiceRoomConfig
	}
	return s.Repository.DeleteVoiceRoomConfigByTrigger(ctx, guildID, triggerChannelID)
}

func (s ConfigService) DeleteByParent(ctx context.Context, guildID string, parentID string) error {
	if s.Repository == nil {
		return domain.ErrInvalidVoiceRoomConfig
	}
	guildID = strings.TrimSpace(guildID)
	parentID = strings.TrimSpace(parentID)
	if guildID == "" || parentID == "" {
		return domain.ErrInvalidVoiceRoomConfig
	}
	return s.Repository.DeleteVoiceRoomConfigsByParent(ctx, guildID, parentID)
}

func (s RoomService) TriggerConfig(ctx context.Context, guildID string, channelID string) (domain.VoiceRoomConfig, bool, error) {
	if s.Configs == nil {
		return domain.VoiceRoomConfig{}, false, domain.ErrInvalidVoiceRoomConfig
	}
	guildID = strings.TrimSpace(guildID)
	channelID = strings.TrimSpace(channelID)
	if guildID == "" || channelID == "" {
		return domain.VoiceRoomConfig{}, false, domain.ErrInvalidVoiceRoomConfig
	}
	config, err := s.Configs.GetVoiceRoomConfigByTrigger(ctx, guildID, channelID)
	if err != nil {
		if errors.Is(err, ports.ErrVoiceRoomConfigMissing) {
			return domain.VoiceRoomConfig{}, false, nil
		}
		return domain.VoiceRoomConfig{}, false, err
	}
	config.GuildID = strings.TrimSpace(config.GuildID)
	config.TriggerChannelID = strings.TrimSpace(config.TriggerChannelID)
	config.ParentID = strings.TrimSpace(config.ParentID)
	return config, true, nil
}

func (s RoomService) TrackDynamicRoom(ctx context.Context, guildID string, channelID string, ownerID string, lockable bool) error {
	if s.States == nil {
		return domain.ErrInvalidVoiceRoomConfig
	}
	guildID = strings.TrimSpace(guildID)
	channelID = strings.TrimSpace(channelID)
	ownerID = strings.TrimSpace(ownerID)
	if guildID == "" || channelID == "" || ownerID == "" {
		return domain.ErrInvalidVoiceRoomConfig
	}
	if err := s.States.SaveVoiceRoomState(ctx, domain.VoiceRoomState{GuildID: guildID, ChannelID: channelID}); err != nil {
		return err
	}
	if !lockable {
		return ctx.Err()
	}
	if s.Locks == nil {
		return domain.ErrInvalidVoiceRoomLock
	}
	err := s.Locks.SaveVoiceRoomLock(ctx, domain.VoiceRoomLock{
		GuildID:   guildID,
		ChannelID: channelID,
		OwnerID:   ownerID,
	})
	if err != nil {
		_ = s.States.DeleteVoiceRoomState(context.Background(), guildID, channelID)
	}
	return err
}

func (s RoomService) IsDynamicRoom(ctx context.Context, guildID string, channelID string) (bool, error) {
	if s.States == nil {
		return false, domain.ErrInvalidVoiceRoomConfig
	}
	guildID = strings.TrimSpace(guildID)
	channelID = strings.TrimSpace(channelID)
	if guildID == "" || channelID == "" {
		return false, domain.ErrInvalidVoiceRoomConfig
	}
	if _, err := s.States.GetVoiceRoomState(ctx, guildID, channelID); err != nil {
		if errors.Is(err, ports.ErrVoiceRoomStateMissing) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (s RoomService) DeleteDynamicRoomLock(ctx context.Context, guildID string, channelID string) error {
	if s.Locks == nil {
		return nil
	}
	err := s.Locks.DeleteVoiceRoomLock(ctx, guildID, channelID)
	if errors.Is(err, ports.ErrVoiceRoomLockMissing) {
		return nil
	}
	return err
}

func (s RoomService) DeleteDynamicRoomState(ctx context.Context, guildID string, channelID string) error {
	if s.States == nil {
		return domain.ErrInvalidVoiceRoomConfig
	}
	err := s.States.DeleteVoiceRoomState(ctx, guildID, channelID)
	if errors.Is(err, ports.ErrVoiceRoomStateMissing) {
		return nil
	}
	return err
}

func (s LockService) SetPassword(ctx context.Context, guildID string, channelID string, ownerID string, textChannelID string, password string) error {
	if s.Repository == nil {
		return domain.ErrInvalidVoiceRoomLock
	}
	guildID = strings.TrimSpace(guildID)
	channelID = strings.TrimSpace(channelID)
	ownerID = strings.TrimSpace(ownerID)
	textChannelID = strings.TrimSpace(textChannelID)
	if guildID == "" || channelID == "" || ownerID == "" || textChannelID == "" {
		return domain.ErrInvalidVoiceRoomLock
	}
	existing, err := s.Repository.GetVoiceRoomLock(ctx, guildID, channelID)
	if err != nil {
		return err
	}
	if strings.TrimSpace(existing.OwnerID) != ownerID {
		return ports.ErrVoiceRoomLockNotOwner
	}
	lock := domain.VoiceRoomLock{
		GuildID:        guildID,
		ChannelID:      channelID,
		Password:       password,
		OwnerID:        ownerID,
		TextChannelID:  textChannelID,
		AllowedUserIDs: []string{},
	}
	return s.Repository.SaveVoiceRoomLock(ctx, lock)
}

func (s LockService) AnswerPassword(ctx context.Context, guildID string, channelID string, userID string, password string) error {
	if s.Repository == nil {
		return domain.ErrInvalidVoiceRoomLock
	}
	guildID = strings.TrimSpace(guildID)
	channelID = strings.TrimSpace(channelID)
	userID = strings.TrimSpace(userID)
	password = strings.TrimSpace(password)
	if guildID == "" || channelID == "" || userID == "" || password == "" {
		return domain.ErrInvalidVoiceRoomLock
	}
	lock, err := s.Repository.GetVoiceRoomLock(ctx, guildID, channelID)
	if err != nil {
		return err
	}
	if strings.TrimSpace(lock.Password) == "" || strings.TrimSpace(lock.Password) != password {
		return ErrVoiceRoomLockPasswordMismatch
	}
	return s.Repository.AllowVoiceRoomLockUser(ctx, guildID, channelID, userID)
}

func (s LockService) LockedJoinPrompt(ctx context.Context, guildID string, channelID string, userID string) (domain.VoiceRoomLock, bool, error) {
	if s.Repository == nil {
		return domain.VoiceRoomLock{}, false, domain.ErrInvalidVoiceRoomLock
	}
	guildID = strings.TrimSpace(guildID)
	channelID = strings.TrimSpace(channelID)
	userID = strings.TrimSpace(userID)
	if guildID == "" || channelID == "" || userID == "" {
		return domain.VoiceRoomLock{}, false, domain.ErrInvalidVoiceRoomLock
	}
	lock, err := s.Repository.GetVoiceRoomLock(ctx, guildID, channelID)
	if err != nil {
		if errors.Is(err, ports.ErrVoiceRoomLockMissing) {
			return domain.VoiceRoomLock{}, false, nil
		}
		return domain.VoiceRoomLock{}, false, err
	}
	lock = lock.Normalize()
	if lock.Password == "" || lock.TextChannelID == "" {
		return domain.VoiceRoomLock{}, false, nil
	}
	for _, allowed := range lock.AllowedUserIDs {
		if strings.TrimSpace(allowed) == userID {
			return domain.VoiceRoomLock{}, false, nil
		}
	}
	return lock, true, nil
}
