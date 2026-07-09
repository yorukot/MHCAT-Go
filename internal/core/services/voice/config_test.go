package voice_test

import (
	"context"
	"errors"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	coreservice "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/services/voice"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakemongo"
)

func TestConfigServiceSaveTrimsAndStoresVoiceRoomConfig(t *testing.T) {
	repo := fakemongo.NewVoiceRoomConfigRepository()
	service := coreservice.NewConfigService(repo)
	err := service.Save(context.Background(), domain.VoiceRoomConfig{
		GuildID:          " guild-1 ",
		TriggerChannelID: " voice-1 ",
		ParentID:         " category-1 ",
		Name:             " {name} 的包廂 ",
		Limit:            12,
		Lock:             true,
	})
	if err != nil {
		t.Fatalf("save config: %v", err)
	}
	saved, ok := repo.Last()
	if !ok {
		t.Fatal("expected saved config")
	}
	if saved.GuildID != "guild-1" || saved.TriggerChannelID != "voice-1" || saved.ParentID != "category-1" || saved.Name != "{name} 的包廂" || saved.Limit != 12 || !saved.Lock {
		t.Fatalf("saved config = %#v", saved)
	}
}

func TestConfigServiceRejectsInvalidVoiceRoomConfig(t *testing.T) {
	service := coreservice.NewConfigService(fakemongo.NewVoiceRoomConfigRepository())
	err := service.Save(context.Background(), domain.VoiceRoomConfig{
		GuildID:          "guild-1",
		TriggerChannelID: "voice-1",
		Name:             "{name}",
		Limit:            -1,
	})
	if !errors.Is(err, domain.ErrInvalidVoiceRoomConfig) {
		t.Fatalf("expected invalid config error, got %v", err)
	}
}
