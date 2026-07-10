package xp

import (
	"context"
	"errors"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakemongo"
)

func TestTextConfigSaveValidatesAndStores(t *testing.T) {
	repo := fakemongo.NewTextXPConfigRepository()
	service := TextConfigService{Repository: repo}
	config := domain.TextXPConfig{GuildID: " guild-1 ", ChannelID: " channel-1 ", Color: "rgba(0, 0, 0, .45)", Message: "hello"}
	if err := service.Save(context.Background(), config); err != nil {
		t.Fatalf("save: %v", err)
	}
	saved, ok := repo.Configs["guild-1"]
	if !ok || saved.ChannelID != "channel-1" || saved.Color != "rgba(0, 0, 0, .45)" || saved.Message != "hello" {
		t.Fatalf("saved = %#v ok=%v", saved, ok)
	}
}

func TestTextConfigSaveRejectsInvalidColor(t *testing.T) {
	err := (TextConfigService{Repository: fakemongo.NewTextXPConfigRepository()}).Save(context.Background(), domain.TextXPConfig{
		GuildID:   "guild-1",
		ChannelID: "channel-1",
		Color:     "not-a-color",
	})
	if !errors.Is(err, domain.ErrInvalidTextXPConfig) {
		t.Fatalf("expected ErrInvalidTextXPConfig, got %v", err)
	}
}

func TestTextConfigDeleteMissing(t *testing.T) {
	err := (TextConfigService{Repository: fakemongo.NewTextXPConfigRepository()}).Delete(context.Background(), "guild-1")
	if !errors.Is(err, ports.ErrTextXPConfigMissing) {
		t.Fatalf("expected ErrTextXPConfigMissing, got %v", err)
	}
}

func TestVoiceConfigSaveValidatesAndStores(t *testing.T) {
	repo := fakemongo.NewVoiceXPConfigRepository()
	service := VoiceConfigService{Repository: repo}
	config := domain.VoiceXPConfig{GuildID: " guild-1 ", ChannelID: " channel-1 ", Color: "RebeccaPurple", Message: "hello"}
	if err := service.Save(context.Background(), config); err != nil {
		t.Fatalf("save: %v", err)
	}
	saved, ok := repo.Configs["guild-1"]
	if !ok || saved.ChannelID != "channel-1" || saved.Color != "RebeccaPurple" || saved.Message != "hello" {
		t.Fatalf("saved = %#v ok=%v", saved, ok)
	}
}

func TestVoiceConfigSaveRejectsInvalidColor(t *testing.T) {
	err := (VoiceConfigService{Repository: fakemongo.NewVoiceXPConfigRepository()}).Save(context.Background(), domain.VoiceXPConfig{
		GuildID:   "guild-1",
		ChannelID: "channel-1",
		Color:     "not-a-color",
	})
	if !errors.Is(err, domain.ErrInvalidVoiceXPConfig) {
		t.Fatalf("expected ErrInvalidVoiceXPConfig, got %v", err)
	}
}

func TestVoiceConfigDeleteMissing(t *testing.T) {
	err := (VoiceConfigService{Repository: fakemongo.NewVoiceXPConfigRepository()}).Delete(context.Background(), "guild-1")
	if !errors.Is(err, ports.ErrVoiceXPConfigMissing) {
		t.Fatalf("expected ErrVoiceXPConfigMissing, got %v", err)
	}
}
