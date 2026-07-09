package announcements

import (
	"context"
	"errors"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakemongo"
)

func TestSetAnnouncementChannelDelegates(t *testing.T) {
	repo := fakemongo.NewAnnouncementConfigRepository()
	service := NewConfigService(repo)
	created, err := service.SetAnnouncementChannel(context.Background(), domain.AnnouncementChannelConfig{GuildID: "guild", ChannelID: "channel"})
	if err != nil {
		t.Fatalf("set channel: %v", err)
	}
	if !created || repo.AnnouncementChannels["guild"] != "channel" {
		t.Fatalf("repo state = %#v created=%v", repo.AnnouncementChannels, created)
	}
}

func TestSetBoundAnnouncementRejectsInvalidColor(t *testing.T) {
	service := NewConfigService(fakemongo.NewAnnouncementConfigRepository())
	_, err := service.SetBoundAnnouncement(context.Background(), domain.BoundAnnouncementConfig{
		GuildID:   "guild",
		ChannelID: "channel",
		Tag:       "@here",
		Color:     "not-a-color",
		Title:     "公告",
	})
	if !errors.Is(err, domain.ErrInvalidAnnouncementConfig) {
		t.Fatalf("expected invalid config, got %v", err)
	}
}

func TestDeleteBoundAnnouncementMapsMissing(t *testing.T) {
	service := NewConfigService(fakemongo.NewAnnouncementConfigRepository())
	err := service.DeleteBoundAnnouncement(context.Background(), "guild", "channel")
	if !errors.Is(err, ports.ErrBoundAnnouncementConfigMissing) {
		t.Fatalf("expected missing config, got %v", err)
	}
}

func TestNilRepositoryReturnsInvalidConfig(t *testing.T) {
	service := NewConfigService(nil)
	_, err := service.SetAnnouncementChannel(context.Background(), domain.AnnouncementChannelConfig{GuildID: "guild", ChannelID: "channel"})
	if !errors.Is(err, domain.ErrInvalidAnnouncementConfig) {
		t.Fatalf("expected invalid config, got %v", err)
	}
}
