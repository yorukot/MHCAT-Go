package ports

import (
	"context"
	"errors"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
)

var ErrBoundAnnouncementConfigMissing = errors.New("bound announcement config missing")
var ErrAnnouncementChannelMissing = errors.New("announcement channel missing")

type AnnouncementConfigRepository interface {
	AnnouncementChannelReader
	AnnouncementChannelWriter
	BoundAnnouncementReader
}

type AnnouncementChannelReader interface {
	GetAnnouncementChannel(ctx context.Context, guildID string) (domain.AnnouncementChannelConfig, error)
}

type BoundAnnouncementReader interface {
	GetBoundAnnouncement(ctx context.Context, guildID string, channelID string) (domain.BoundAnnouncementConfig, error)
}

type AnnouncementChannelWriter interface {
	SetAnnouncementChannel(ctx context.Context, config domain.AnnouncementChannelConfig) (created bool, err error)
	SetBoundAnnouncement(ctx context.Context, config domain.BoundAnnouncementConfig) (created bool, err error)
	DeleteBoundAnnouncement(ctx context.Context, guildID string, channelID string) error
}
