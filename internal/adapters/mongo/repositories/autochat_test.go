package repositories

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
)

func TestAutoChatConfigCollectionName(t *testing.T) {
	if AutoChatConfigCollectionName != "chats" {
		t.Fatalf("autochat collection = %s, want chats", AutoChatConfigCollectionName)
	}
}

func TestNewAutoChatConfigRepositoryRequiresCollection(t *testing.T) {
	if _, err := NewAutoChatConfigRepository(nil); err == nil {
		t.Fatal("expected nil collection error")
	}
}

func TestNewAutoChatConfigRepositoryFromDatabaseRequiresDatabase(t *testing.T) {
	if _, err := NewAutoChatConfigRepositoryFromDatabase(nil); err == nil {
		t.Fatal("expected nil database error")
	}
}

func TestAutoChatConfigGetRejectsBlankGuild(t *testing.T) {
	repo := &AutoChatConfigRepository{}
	if _, err := repo.GetAutoChatConfig(context.Background(), " "); !errors.Is(err, domain.ErrInvalidAutoChatConfig) {
		t.Fatalf("expected ErrInvalidAutoChatConfig, got %v", err)
	}
}

func TestAutoChatConfigCacheServesConfiguredAndMissingGuilds(t *testing.T) {
	now := time.Date(2026, 7, 11, 0, 0, 0, 0, time.UTC)
	repo := &AutoChatConfigRepository{now: func() time.Time { return now }}
	config := domain.AutoChatConfig{GuildID: "guild-1", ChannelID: "channel-1"}
	repo.storeCachedConfig("guild-1", config, true)
	got, err := repo.GetAutoChatConfig(context.Background(), "guild-1")
	if err != nil || got != config {
		t.Fatalf("config=%#v err=%v", got, err)
	}
	repo.storeCachedConfig("guild-2", domain.AutoChatConfig{}, false)
	if _, err := repo.GetAutoChatConfig(context.Background(), "guild-2"); !errors.Is(err, ports.ErrAutoChatConfigMissing) {
		t.Fatalf("missing error = %v", err)
	}
}
