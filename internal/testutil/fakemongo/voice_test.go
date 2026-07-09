package fakemongo

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
)

func TestVoiceRoomLockRepositoryGetSaveAndLast(t *testing.T) {
	repo := NewVoiceRoomLockRepository()
	seed := domain.VoiceRoomLock{
		GuildID:       "guild-1",
		ChannelID:     "voice-1",
		Password:      "old",
		OwnerID:       "owner-1",
		TextChannelID: "text-1",
	}
	repo.Locks[voiceRoomKey(seed.GuildID, seed.ChannelID)] = seed

	got, err := repo.GetVoiceRoomLock(context.Background(), " guild-1 ", " voice-1 ")
	if err != nil {
		t.Fatalf("get lock: %v", err)
	}
	if !reflect.DeepEqual(got, seed) {
		t.Fatalf("got lock = %#v", got)
	}

	replacement := domain.VoiceRoomLock{
		GuildID:        " guild-1 ",
		ChannelID:      " voice-1 ",
		Password:       " new ",
		OwnerID:        " owner-1 ",
		TextChannelID:  " text-2 ",
		AllowedUserIDs: []string{" user-2 ", " "},
	}
	if err := repo.SaveVoiceRoomLock(context.Background(), replacement); err != nil {
		t.Fatalf("save lock: %v", err)
	}
	saved, ok := repo.Last()
	if !ok {
		t.Fatal("expected saved lock")
	}
	if saved.GuildID != "guild-1" ||
		saved.ChannelID != "voice-1" ||
		saved.Password != "new" ||
		saved.OwnerID != "owner-1" ||
		saved.TextChannelID != "text-2" ||
		!reflect.DeepEqual(saved.AllowedUserIDs, []string{"user-2"}) {
		t.Fatalf("saved lock = %#v", saved)
	}
}

func TestVoiceRoomLockRepositoryMissing(t *testing.T) {
	repo := NewVoiceRoomLockRepository()
	_, err := repo.GetVoiceRoomLock(context.Background(), "guild-1", "missing")
	if !errors.Is(err, ports.ErrVoiceRoomLockMissing) {
		t.Fatalf("expected missing lock error, got %v", err)
	}
}
