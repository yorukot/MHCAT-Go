package fakemongo

import (
	"context"
	"strings"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
)

type AnnouncementConfigRepository struct {
	AnnouncementChannels map[string]string
	BoundAnnouncements   map[string]domain.BoundAnnouncementConfig
	Err                  error
}

func NewAnnouncementConfigRepository() *AnnouncementConfigRepository {
	return &AnnouncementConfigRepository{
		AnnouncementChannels: map[string]string{},
		BoundAnnouncements:   map[string]domain.BoundAnnouncementConfig{},
	}
}

func (r *AnnouncementConfigRepository) GetAnnouncementChannel(ctx context.Context, guildID string) (domain.AnnouncementChannelConfig, error) {
	if err := ctx.Err(); err != nil {
		return domain.AnnouncementChannelConfig{}, err
	}
	if r.Err != nil {
		return domain.AnnouncementChannelConfig{}, r.Err
	}
	guildID = strings.TrimSpace(guildID)
	channelID := strings.TrimSpace(r.AnnouncementChannels[guildID])
	if guildID == "" || channelID == "" || channelID == "0" {
		return domain.AnnouncementChannelConfig{}, ports.ErrAnnouncementChannelMissing
	}
	return domain.AnnouncementChannelConfig{GuildID: guildID, ChannelID: channelID}, nil
}

func (r *AnnouncementConfigRepository) GetBoundAnnouncement(ctx context.Context, guildID string, channelID string) (domain.BoundAnnouncementConfig, error) {
	if err := ctx.Err(); err != nil {
		return domain.BoundAnnouncementConfig{}, err
	}
	if r.Err != nil {
		return domain.BoundAnnouncementConfig{}, r.Err
	}
	key := boundAnnouncementKey(guildID, channelID)
	config, exists := r.BoundAnnouncements[key]
	if !exists {
		return domain.BoundAnnouncementConfig{}, ports.ErrBoundAnnouncementConfigMissing
	}
	return config, nil
}

func (r *AnnouncementConfigRepository) SetAnnouncementChannel(ctx context.Context, config domain.AnnouncementChannelConfig) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	if r.Err != nil {
		return false, r.Err
	}
	if err := config.Validate(); err != nil {
		return false, err
	}
	_, exists := r.AnnouncementChannels[strings.TrimSpace(config.GuildID)]
	r.AnnouncementChannels[strings.TrimSpace(config.GuildID)] = strings.TrimSpace(config.ChannelID)
	return !exists, nil
}

func (r *AnnouncementConfigRepository) SetBoundAnnouncement(ctx context.Context, config domain.BoundAnnouncementConfig) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	if r.Err != nil {
		return false, r.Err
	}
	if err := config.Validate(); err != nil {
		return false, err
	}
	key := boundAnnouncementKey(config.GuildID, config.ChannelID)
	_, exists := r.BoundAnnouncements[key]
	r.BoundAnnouncements[key] = config
	return !exists, nil
}

func (r *AnnouncementConfigRepository) DeleteBoundAnnouncement(ctx context.Context, guildID string, channelID string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if r.Err != nil {
		return r.Err
	}
	key := boundAnnouncementKey(guildID, channelID)
	if _, exists := r.BoundAnnouncements[key]; !exists {
		return ports.ErrBoundAnnouncementConfigMissing
	}
	delete(r.BoundAnnouncements, key)
	return nil
}

func boundAnnouncementKey(guildID string, channelID string) string {
	return strings.TrimSpace(guildID) + ":" + strings.TrimSpace(channelID)
}
