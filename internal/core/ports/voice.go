package ports

import (
	"context"
	"errors"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
)

var ErrVoiceRoomConfigMissing = errors.New("voice room config is missing")

type VoiceRoomConfigRepository interface {
	SaveVoiceRoomConfig(ctx context.Context, config domain.VoiceRoomConfig) error
	DeleteVoiceRoomConfigByTrigger(ctx context.Context, guildID string, triggerChannelID string) error
	DeleteVoiceRoomConfigsByParent(ctx context.Context, guildID string, parentID string) error
}
