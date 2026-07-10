package ports

import (
	"context"
	"errors"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
)

var ErrAutoChatConfigMissing = errors.New("autochat config is missing")

type AutoChatConfigRepository interface {
	SaveAutoChatConfig(ctx context.Context, config domain.AutoChatConfig) error
	DeleteAutoChatConfig(ctx context.Context, guildID string) error
}

type AutoChatConfigReader interface {
	GetAutoChatConfig(ctx context.Context, guildID string) (domain.AutoChatConfig, error)
}
