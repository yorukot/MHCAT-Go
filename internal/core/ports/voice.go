package ports

import (
	"context"
	"errors"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
)

var ErrVoiceRoomConfigMissing = errors.New("voice room config is missing")
var ErrVoiceRoomLockMissing = errors.New("voice room lock is missing")
var ErrVoiceRoomLockNotOwner = errors.New("voice room lock owner mismatch")
var ErrVoiceRoomStateMissing = errors.New("voice room state is missing")

type VoiceRoomConfigRepository interface {
	GetVoiceRoomConfigByTrigger(ctx context.Context, guildID string, triggerChannelID string) (domain.VoiceRoomConfig, error)
	SaveVoiceRoomConfig(ctx context.Context, config domain.VoiceRoomConfig) error
	DeleteVoiceRoomConfigByTrigger(ctx context.Context, guildID string, triggerChannelID string) error
	DeleteVoiceRoomConfigsByParent(ctx context.Context, guildID string, parentID string) error
}

type VoiceRoomLockRepository interface {
	GetVoiceRoomLock(ctx context.Context, guildID string, channelID string) (domain.VoiceRoomLock, error)
	SaveVoiceRoomLock(ctx context.Context, lock domain.VoiceRoomLock) error
	DeleteVoiceRoomLock(ctx context.Context, guildID string, channelID string) error
	AllowVoiceRoomLockUser(ctx context.Context, guildID string, channelID string, userID string) error
}

type VoiceRoomStateRepository interface {
	GetVoiceRoomState(ctx context.Context, guildID string, channelID string) (domain.VoiceRoomState, error)
	SaveVoiceRoomState(ctx context.Context, state domain.VoiceRoomState) error
	DeleteVoiceRoomState(ctx context.Context, guildID string, channelID string) error
}
