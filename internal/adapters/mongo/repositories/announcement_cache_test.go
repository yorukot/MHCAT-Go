package repositories

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
)

func TestBoundAnnouncementCacheServesPositiveAndNegativeEntries(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	repository := &AnnouncementConfigRepository{now: func() time.Time { return now }}
	config := domain.BoundAnnouncementConfig{
		GuildID: "guild-1", ChannelID: "channel-1", Tag: "@here", Color: "Random", Title: "Notice",
	}
	repository.storeCachedBoundAnnouncement(config.GuildID, config.ChannelID, config, true)
	repository.storeCachedBoundAnnouncement("guild-1", "channel-2", domain.BoundAnnouncementConfig{}, false)

	got, err := repository.GetBoundAnnouncement(context.Background(), " guild-1 ", " channel-1 ")
	if err != nil || got != config {
		t.Fatalf("cached config = %#v, err=%v", got, err)
	}
	if _, err := repository.GetBoundAnnouncement(context.Background(), "guild-1", "channel-2"); !errors.Is(err, ports.ErrBoundAnnouncementConfigMissing) {
		t.Fatalf("negative cache error = %v", err)
	}
}

func TestBoundAnnouncementCacheSeparatesGuildsAndChannels(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	repository := &AnnouncementConfigRepository{now: func() time.Time { return now }}
	first := domain.BoundAnnouncementConfig{GuildID: "ab", ChannelID: "c", Tag: "one", Color: "Random", Title: "One"}
	second := domain.BoundAnnouncementConfig{GuildID: "a", ChannelID: "bc", Tag: "two", Color: "Random", Title: "Two"}
	repository.storeCachedBoundAnnouncement(first.GuildID, first.ChannelID, first, true)
	repository.storeCachedBoundAnnouncement(second.GuildID, second.ChannelID, second, true)

	gotFirst, err := repository.GetBoundAnnouncement(context.Background(), first.GuildID, first.ChannelID)
	if err != nil || gotFirst != first {
		t.Fatalf("first config = %#v, err=%v", gotFirst, err)
	}
	gotSecond, err := repository.GetBoundAnnouncement(context.Background(), second.GuildID, second.ChannelID)
	if err != nil || gotSecond != second {
		t.Fatalf("second config = %#v, err=%v", gotSecond, err)
	}
}
