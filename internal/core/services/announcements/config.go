package announcements

import (
	"context"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
)

type ConfigService struct {
	repo ports.AnnouncementConfigRepository
}

func NewConfigService(repo ports.AnnouncementConfigRepository) ConfigService {
	return ConfigService{repo: repo}
}

func (s ConfigService) SetAnnouncementChannel(ctx context.Context, config domain.AnnouncementChannelConfig) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	if s.repo == nil {
		return false, domain.ErrInvalidAnnouncementConfig
	}
	if err := config.Validate(); err != nil {
		return false, err
	}
	return s.repo.SetAnnouncementChannel(ctx, config)
}

func (s ConfigService) SetBoundAnnouncement(ctx context.Context, config domain.BoundAnnouncementConfig) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	if s.repo == nil {
		return false, domain.ErrInvalidAnnouncementConfig
	}
	if err := config.Validate(); err != nil {
		return false, err
	}
	return s.repo.SetBoundAnnouncement(ctx, config)
}

func (s ConfigService) DeleteBoundAnnouncement(ctx context.Context, guildID string, channelID string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if s.repo == nil {
		return domain.ErrInvalidAnnouncementConfig
	}
	return s.repo.DeleteBoundAnnouncement(ctx, guildID, channelID)
}
