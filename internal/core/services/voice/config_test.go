package voice_test

import (
	"context"
	"errors"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
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

func TestLockServiceSetPasswordSavesReplacement(t *testing.T) {
	repo := fakemongo.NewVoiceRoomLockRepository()
	repo.Locks["guild-1\x00voice-1"] = domain.VoiceRoomLock{
		GuildID:        "guild-1",
		ChannelID:      "voice-1",
		Password:       "old",
		OwnerID:        "owner-1",
		TextChannelID:  "old-text",
		AllowedUserIDs: []string{"user-2"},
	}
	service := coreservice.NewLockService(repo)
	if err := service.SetPassword(context.Background(), " guild-1 ", " voice-1 ", " owner-1 ", " text-1 ", " secret "); err != nil {
		t.Fatalf("set password: %v", err)
	}
	saved, ok := repo.Last()
	if !ok {
		t.Fatal("expected saved lock")
	}
	if saved.GuildID != "guild-1" || saved.ChannelID != "voice-1" || saved.OwnerID != "owner-1" || saved.TextChannelID != "text-1" || saved.Password != "secret" {
		t.Fatalf("saved lock = %#v", saved)
	}
	if len(saved.AllowedUserIDs) != 0 {
		t.Fatalf("allowed users should be reset, got %#v", saved.AllowedUserIDs)
	}
}

func TestLockServiceSetPasswordPropagatesMissingLock(t *testing.T) {
	service := coreservice.NewLockService(fakemongo.NewVoiceRoomLockRepository())
	err := service.SetPassword(context.Background(), "guild-1", "missing", "owner-1", "text-1", "secret")
	if !errors.Is(err, ports.ErrVoiceRoomLockMissing) {
		t.Fatalf("expected missing lock error, got %v", err)
	}
}

func TestLockServiceSetPasswordRejectsNonOwner(t *testing.T) {
	repo := fakemongo.NewVoiceRoomLockRepository()
	repo.Locks["guild-1\x00voice-1"] = domain.VoiceRoomLock{
		GuildID:       "guild-1",
		ChannelID:     "voice-1",
		OwnerID:       "other-user",
		TextChannelID: "text-1",
	}
	service := coreservice.NewLockService(repo)
	err := service.SetPassword(context.Background(), "guild-1", "voice-1", "owner-1", "text-1", "secret")
	if !errors.Is(err, ports.ErrVoiceRoomLockNotOwner) {
		t.Fatalf("expected owner mismatch error, got %v", err)
	}
}

func TestLockServiceSetPasswordRejectsInvalidInput(t *testing.T) {
	service := coreservice.NewLockService(fakemongo.NewVoiceRoomLockRepository())
	err := service.SetPassword(context.Background(), "", "voice-1", "owner-1", "text-1", "secret")
	if !errors.Is(err, domain.ErrInvalidVoiceRoomLock) {
		t.Fatalf("expected invalid lock error, got %v", err)
	}

	nilService := coreservice.NewLockService(nil)
	err = nilService.SetPassword(context.Background(), "guild-1", "voice-1", "owner-1", "text-1", "secret")
	if !errors.Is(err, domain.ErrInvalidVoiceRoomLock) {
		t.Fatalf("expected nil repository invalid lock error, got %v", err)
	}
}
