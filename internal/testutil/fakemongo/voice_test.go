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
		saved.Password != " new " ||
		!saved.PasswordPresent ||
		saved.OwnerID != "owner-1" ||
		saved.TextChannelID != "text-2" ||
		!reflect.DeepEqual(saved.AllowedUserIDs, []string{"user-2"}) {
		t.Fatalf("saved lock = %#v", saved)
	}
}

func TestVoiceRoomLockRepositoryDelete(t *testing.T) {
	repo := NewVoiceRoomLockRepository()
	repo.Locks["guild-1\x00voice-1"] = domain.VoiceRoomLock{
		GuildID:   "guild-1",
		ChannelID: "voice-1",
		OwnerID:   "owner-1",
	}
	if err := repo.DeleteVoiceRoomLock(context.Background(), " guild-1 ", " voice-1 "); err != nil {
		t.Fatalf("delete lock: %v", err)
	}
	if _, ok := repo.Locks["guild-1\x00voice-1"]; ok || len(repo.Deleted) != 1 {
		t.Fatalf("locks=%#v deleted=%#v", repo.Locks, repo.Deleted)
	}
	if err := repo.DeleteVoiceRoomLock(context.Background(), "guild-1", "voice-1"); !errors.Is(err, ports.ErrVoiceRoomLockMissing) {
		t.Fatalf("expected missing lock, got %v", err)
	}
}

func TestVoiceRoomLockRepositoryMissing(t *testing.T) {
	repo := NewVoiceRoomLockRepository()
	_, err := repo.GetVoiceRoomLock(context.Background(), "guild-1", "missing")
	if !errors.Is(err, ports.ErrVoiceRoomLockMissing) {
		t.Fatalf("expected missing lock error, got %v", err)
	}
}

func TestVoiceRoomStateRepositorySaveGetDelete(t *testing.T) {
	repo := NewVoiceRoomStateRepository()
	state := domain.VoiceRoomState{GuildID: " guild-1 ", ChannelID: " voice-1 "}
	if err := repo.SaveVoiceRoomState(context.Background(), state); err != nil {
		t.Fatalf("save state: %v", err)
	}
	got, err := repo.GetVoiceRoomState(context.Background(), "guild-1", "voice-1")
	if err != nil {
		t.Fatalf("get state: %v", err)
	}
	if got.GuildID != "guild-1" || got.ChannelID != "voice-1" || len(repo.Saved) != 1 {
		t.Fatalf("state=%#v saved=%#v", got, repo.Saved)
	}
	if err := repo.SaveVoiceRoomState(context.Background(), state); err != nil {
		t.Fatalf("save duplicate state: %v", err)
	}
	if len(repo.Saved) != 1 {
		t.Fatalf("duplicate state should not append, saved=%#v", repo.Saved)
	}
	if err := repo.DeleteVoiceRoomState(context.Background(), "guild-1", "voice-1"); err != nil {
		t.Fatalf("delete state: %v", err)
	}
	if _, err := repo.GetVoiceRoomState(context.Background(), "guild-1", "voice-1"); !errors.Is(err, ports.ErrVoiceRoomStateMissing) {
		t.Fatalf("expected missing state, got %v", err)
	}
}
