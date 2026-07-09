package ports

import (
	"context"
	"errors"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
)

var ErrTextXPConfigMissing = errors.New("text xp config is missing")
var ErrVoiceXPConfigMissing = errors.New("voice xp config is missing")
var ErrTextXPProfileMissing = errors.New("text xp profile is missing")
var ErrVoiceXPProfileMissing = errors.New("voice xp profile is missing")

type TextXPConfigRepository interface {
	SaveTextXPConfig(ctx context.Context, config domain.TextXPConfig) error
	DeleteTextXPConfig(ctx context.Context, guildID string) error
}

type VoiceXPConfigRepository interface {
	SaveVoiceXPConfig(ctx context.Context, config domain.VoiceXPConfig) error
	DeleteVoiceXPConfig(ctx context.Context, guildID string) error
}
