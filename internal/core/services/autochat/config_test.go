package autochat

import (
	"context"
	"errors"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakemongo"
)

func TestConfigServiceSaveTrimsAndStores(t *testing.T) {
	repo := fakemongo.NewAutoChatConfigRepository()
	service := NewConfigService(repo)

	if err := service.Save(context.Background(), domain.AutoChatConfig{GuildID: " guild-1 ", ChannelID: " channel-1 "}); err != nil {
		t.Fatalf("save: %v", err)
	}
	saved, ok := repo.Configs["guild-1"]
	if !ok || saved.ChannelID != "channel-1" {
		t.Fatalf("saved = %#v ok=%v", saved, ok)
	}
}

func TestConfigServiceSaveRejectsInvalidConfig(t *testing.T) {
	err := NewConfigService(fakemongo.NewAutoChatConfigRepository()).Save(context.Background(), domain.AutoChatConfig{GuildID: "guild-1"})
	if !errors.Is(err, domain.ErrInvalidAutoChatConfig) {
		t.Fatalf("expected ErrInvalidAutoChatConfig, got %v", err)
	}
}

func TestConfigServiceDeleteMissing(t *testing.T) {
	err := NewConfigService(fakemongo.NewAutoChatConfigRepository()).Delete(context.Background(), "guild-1")
	if !errors.Is(err, ports.ErrAutoChatConfigMissing) {
		t.Fatalf("expected ErrAutoChatConfigMissing, got %v", err)
	}
}

func TestConfigServiceRejectsNilRepository(t *testing.T) {
	err := NewConfigService(nil).Save(context.Background(), domain.AutoChatConfig{GuildID: "guild-1", ChannelID: "channel-1"})
	if !errors.Is(err, domain.ErrInvalidAutoChatConfig) {
		t.Fatalf("expected ErrInvalidAutoChatConfig, got %v", err)
	}
}
