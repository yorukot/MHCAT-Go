package voice

import (
	"context"
	"strings"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
)

type ConfigService struct {
	Repository ports.VoiceRoomConfigRepository
}

type LockService struct {
	Repository ports.VoiceRoomLockRepository
}

func NewConfigService(repo ports.VoiceRoomConfigRepository) ConfigService {
	return ConfigService{Repository: repo}
}

func NewLockService(repo ports.VoiceRoomLockRepository) LockService {
	return LockService{Repository: repo}
}

func (s ConfigService) Save(ctx context.Context, config domain.VoiceRoomConfig) error {
	if s.Repository == nil {
		return domain.ErrInvalidVoiceRoomConfig
	}
	config.GuildID = strings.TrimSpace(config.GuildID)
	config.TriggerChannelID = strings.TrimSpace(config.TriggerChannelID)
	config.ParentID = strings.TrimSpace(config.ParentID)
	config.Name = strings.TrimSpace(config.Name)
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
