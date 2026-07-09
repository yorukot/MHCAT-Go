package autochat

import (
	"context"
	"strings"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
)

type ConfigService struct {
	Repository ports.AutoChatConfigRepository
}

func NewConfigService(repository ports.AutoChatConfigRepository) ConfigService {
	return ConfigService{Repository: repository}
}

func (s ConfigService) Save(ctx context.Context, config domain.AutoChatConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if s.Repository == nil {
		return domain.ErrInvalidAutoChatConfig
	}
	config.GuildID = strings.TrimSpace(config.GuildID)
	config.ChannelID = strings.TrimSpace(config.ChannelID)
	if err := config.Validate(); err != nil {
		return err
	}
	return s.Repository.SaveAutoChatConfig(ctx, config)
}

func (s ConfigService) Delete(ctx context.Context, guildID string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if s.Repository == nil {
		return domain.ErrInvalidAutoChatConfig
	}
	guildID = strings.TrimSpace(guildID)
	if guildID == "" {
		return domain.ErrInvalidAutoChatConfig
	}
	return s.Repository.DeleteAutoChatConfig(ctx, guildID)
}
