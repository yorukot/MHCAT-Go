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

func TestConfigServiceSaveNormalizesIDsAndPreservesVoiceRoomName(t *testing.T) {
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
	if saved.GuildID != "guild-1" || saved.TriggerChannelID != "voice-1" || saved.ParentID != "category-1" || saved.Name != " {name} 的包廂 " || saved.Limit != 12 || !saved.Lock {
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

func TestRoomServiceTriggerConfig(t *testing.T) {
	configs := fakemongo.NewVoiceRoomConfigRepository()
	configs.Configs["guild-1\x00trigger-1"] = domain.VoiceRoomConfig{
		GuildID:          "guild-1",
		TriggerChannelID: "trigger-1",
		ParentID:         "parent-1",
		Name:             " {name} room ",
		Limit:            8,
		Lock:             true,
	}
	service := coreservice.NewRoomService(configs, fakemongo.NewVoiceRoomStateRepository(), fakemongo.NewVoiceRoomLockRepository())
	config, ok, err := service.TriggerConfig(context.Background(), " guild-1 ", " trigger-1 ")
	if err != nil {
		t.Fatalf("trigger config: %v", err)
	}
	if !ok || config.Name != " {name} room " || config.ParentID != "parent-1" || !config.Lock {
		t.Fatalf("config=%#v ok=%t", config, ok)
	}
	if _, ok, err := service.TriggerConfig(context.Background(), "guild-1", "missing"); err != nil || ok {
		t.Fatalf("missing trigger should no-op: ok=%t err=%v", ok, err)
	}
}

func TestRoomServiceTrackAndDeleteDynamicRoom(t *testing.T) {
	states := fakemongo.NewVoiceRoomStateRepository()
	locks := fakemongo.NewVoiceRoomLockRepository()
	service := coreservice.NewRoomService(fakemongo.NewVoiceRoomConfigRepository(), states, locks)
	if err := service.TrackDynamicRoom(context.Background(), " guild-1 ", " voice-1 ", " owner-1 ", true); err != nil {
		t.Fatalf("track dynamic room: %v", err)
	}
	if _, ok := states.States["guild-1\x00voice-1"]; !ok {
		t.Fatalf("state not saved: %#v", states.States)
	}
	lock := locks.Locks["guild-1\x00voice-1"]
	if lock.OwnerID != "owner-1" || lock.Password != "" || lock.TextChannelID != "" {
		t.Fatalf("seed lock = %#v", lock)
	}
	tracked, err := service.IsDynamicRoom(context.Background(), "guild-1", "voice-1")
	if err != nil || !tracked {
		t.Fatalf("tracked=%t err=%v", tracked, err)
	}
	if err := service.DeleteDynamicRoomLock(context.Background(), "guild-1", "voice-1"); err != nil {
		t.Fatalf("delete lock: %v", err)
	}
	if err := service.DeleteDynamicRoomState(context.Background(), "guild-1", "voice-1"); err != nil {
		t.Fatalf("delete state: %v", err)
	}
	if tracked, err := service.IsDynamicRoom(context.Background(), "guild-1", "voice-1"); err != nil || tracked {
		t.Fatalf("deleted tracked=%t err=%v", tracked, err)
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
	if saved.GuildID != "guild-1" || saved.ChannelID != "voice-1" || saved.OwnerID != "owner-1" || saved.TextChannelID != "text-1" || saved.Password != " secret " || !saved.PasswordPresent {
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

func TestLockServiceUsesExactLegacyOwnerAndAllowedIDs(t *testing.T) {
	repo := fakemongo.NewVoiceRoomLockRepository()
	repo.Locks["guild-1\x00voice-1"] = domain.VoiceRoomLock{
		GuildID:         "guild-1",
		ChannelID:       "voice-1",
		Password:        "secret",
		PasswordPresent: true,
		OwnerID:         " owner-1 ",
		TextChannelID:   "text-1",
		AllowedUserIDs:  []string{" user-1 "},
	}
	service := coreservice.NewLockService(repo)
	if err := service.SetPassword(context.Background(), "guild-1", "voice-1", "owner-1", "text-1", "new"); !errors.Is(err, ports.ErrVoiceRoomLockNotOwner) {
		t.Fatalf("expected exact owner mismatch, got %v", err)
	}
	_, prompt, err := service.LockedJoinPrompt(context.Background(), "guild-1", "voice-1", "user-1")
	if err != nil || !prompt {
		t.Fatalf("spaced allowed id should not match: prompt=%t err=%v", prompt, err)
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

func TestLockServiceAnswerPasswordAllowsUser(t *testing.T) {
	repo := fakemongo.NewVoiceRoomLockRepository()
	repo.Locks["guild-1\x00voice-1"] = domain.VoiceRoomLock{
		GuildID:         "guild-1",
		ChannelID:       "voice-1",
		Password:        " secret ",
		PasswordPresent: true,
		OwnerID:         "owner-1",
		TextChannelID:   "text-1",
		AllowedUserIDs:  []string{"user-2"},
	}
	service := coreservice.NewLockService(repo)
	if err := service.AnswerPassword(context.Background(), " guild-1 ", " voice-1 ", " user-1 ", " secret "); err != nil {
		t.Fatalf("answer password: %v", err)
	}
	lock := repo.Locks["guild-1\x00voice-1"]
	if got := lock.AllowedUserIDs; len(got) != 2 || got[0] != "user-2" || got[1] != "user-1" {
		t.Fatalf("allowed users = %#v", got)
	}
	if err := service.AnswerPassword(context.Background(), "guild-1", "voice-1", "user-1", " secret "); err != nil {
		t.Fatalf("answer password duplicate: %v", err)
	}
	if got := repo.Locks["guild-1\x00voice-1"].AllowedUserIDs; len(got) != 2 {
		t.Fatalf("duplicate user should not be appended: %#v", got)
	}
}

func TestLockServiceAnswerPasswordRejectsWrongPassword(t *testing.T) {
	repo := fakemongo.NewVoiceRoomLockRepository()
	repo.Locks["guild-1\x00voice-1"] = domain.VoiceRoomLock{
		GuildID:       "guild-1",
		ChannelID:     "voice-1",
		Password:      "secret",
		OwnerID:       "owner-1",
		TextChannelID: "text-1",
	}
	service := coreservice.NewLockService(repo)
	err := service.AnswerPassword(context.Background(), "guild-1", "voice-1", "user-1", "wrong")
	if !errors.Is(err, coreservice.ErrVoiceRoomLockPasswordMismatch) {
		t.Fatalf("expected password mismatch, got %v", err)
	}
	if len(repo.Locks["guild-1\x00voice-1"].AllowedUserIDs) != 0 {
		t.Fatalf("wrong password should not allow user: %#v", repo.Locks["guild-1\x00voice-1"])
	}
}

func TestLockServiceAnswerPasswordDoesNotTrimLegacyInput(t *testing.T) {
	repo := fakemongo.NewVoiceRoomLockRepository()
	repo.Locks["guild-1\x00voice-1"] = domain.VoiceRoomLock{
		GuildID:         "guild-1",
		ChannelID:       "voice-1",
		Password:        " secret ",
		PasswordPresent: true,
		OwnerID:         "owner-1",
		TextChannelID:   "text-1",
	}
	service := coreservice.NewLockService(repo)
	err := service.AnswerPassword(context.Background(), "guild-1", "voice-1", "user-1", "secret")
	if !errors.Is(err, coreservice.ErrVoiceRoomLockPasswordMismatch) {
		t.Fatalf("expected exact password mismatch, got %v", err)
	}
}

func TestLockServiceLockedJoinPrompt(t *testing.T) {
	repo := fakemongo.NewVoiceRoomLockRepository()
	repo.Locks["guild-1\x00voice-1"] = domain.VoiceRoomLock{
		GuildID:        "guild-1",
		ChannelID:      "voice-1",
		Password:       "secret",
		OwnerID:        "owner-1",
		TextChannelID:  "text-1",
		AllowedUserIDs: []string{"user-2"},
	}
	service := coreservice.NewLockService(repo)
	lock, prompt, err := service.LockedJoinPrompt(context.Background(), " guild-1 ", " voice-1 ", " user-1 ")
	if err != nil {
		t.Fatalf("locked join prompt: %v", err)
	}
	if !prompt || lock.ChannelID != "voice-1" || lock.TextChannelID != "text-1" {
		t.Fatalf("prompt=%t lock=%#v", prompt, lock)
	}
	if _, prompt, err := service.LockedJoinPrompt(context.Background(), "guild-1", "voice-1", "user-2"); err != nil || prompt {
		t.Fatalf("allowed user should not prompt: prompt=%t err=%v", prompt, err)
	}
	if _, prompt, err := service.LockedJoinPrompt(context.Background(), "guild-1", "missing", "user-1"); err != nil || prompt {
		t.Fatalf("missing lock should no-op: prompt=%t err=%v", prompt, err)
	}
}

func TestLockServiceTreatsExplicitEmptyBSONPasswordAsLocked(t *testing.T) {
	repo := fakemongo.NewVoiceRoomLockRepository()
	repo.Locks["guild-1\x00voice-1"] = domain.VoiceRoomLock{
		GuildID:         "guild-1",
		ChannelID:       "voice-1",
		PasswordPresent: true,
		OwnerID:         "owner-1",
		TextChannelID:   "text-1",
	}
	service := coreservice.NewLockService(repo)
	lock, prompt, err := service.LockedJoinPrompt(context.Background(), "guild-1", "voice-1", "user-1")
	if err != nil || !prompt || !lock.HasPassword() {
		t.Fatalf("prompt=%t lock=%#v err=%v", prompt, lock, err)
	}
	if err := service.AnswerPassword(context.Background(), "guild-1", "voice-1", "user-1", ""); err != nil {
		t.Fatalf("answer explicit empty password: %v", err)
	}
}
