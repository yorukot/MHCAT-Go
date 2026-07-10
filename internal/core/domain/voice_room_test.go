package domain_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
)

func TestVoiceRoomConfigValidate(t *testing.T) {
	valid := domain.VoiceRoomConfig{
		GuildID:          "guild-1",
		TriggerChannelID: "voice-1",
		Name:             "{name} 的包廂",
		Limit:            99,
		Lock:             true,
	}
	if err := valid.Validate(); err != nil {
		t.Fatalf("validate valid config: %v", err)
	}

	invalid := valid
	invalid.Limit = 100
	if err := invalid.Validate(); !errors.Is(err, domain.ErrInvalidVoiceRoomConfig) {
		t.Fatalf("expected invalid limit error, got %v", err)
	}

	invalid = valid
	invalid.Name = ""
	if err := invalid.Validate(); !errors.Is(err, domain.ErrInvalidVoiceRoomConfig) {
		t.Fatalf("expected invalid name error, got %v", err)
	}

	whitespaceName := valid
	whitespaceName.Name = "  "
	if err := whitespaceName.Validate(); err != nil {
		t.Fatalf("legacy accepts whitespace-only room names: %v", err)
	}
}

func TestVoiceRoomLockNormalizeAndValidate(t *testing.T) {
	lock := domain.VoiceRoomLock{
		GuildID:        " guild-1 ",
		ChannelID:      " voice-1 ",
		Password:       " secret ",
		OwnerID:        " owner-1 ",
		TextChannelID:  " text-1 ",
		AllowedUserIDs: []string{" user-1 ", " ", "user-2"},
	}
	normalized := lock.Normalize()
	if normalized.GuildID != "guild-1" ||
		normalized.ChannelID != "voice-1" ||
		normalized.Password != "secret" ||
		normalized.OwnerID != "owner-1" ||
		normalized.TextChannelID != "text-1" ||
		!reflect.DeepEqual(normalized.AllowedUserIDs, []string{"user-1", "user-2"}) {
		t.Fatalf("normalized lock = %#v", normalized)
	}
	if err := normalized.Validate(); err != nil {
		t.Fatalf("validate normalized lock: %v", err)
	}

	withoutPassword := normalized
	withoutPassword.Password = ""
	if err := withoutPassword.Validate(); err != nil {
		t.Fatalf("password is optional: %v", err)
	}

	lockSeed := normalized
	lockSeed.Password = ""
	lockSeed.TextChannelID = ""
	if err := lockSeed.Validate(); err != nil {
		t.Fatalf("dynamic lock seed text channel is optional: %v", err)
	}

	for _, invalid := range []domain.VoiceRoomLock{
		{ChannelID: "voice-1", OwnerID: "owner-1", TextChannelID: "text-1"},
		{GuildID: "guild-1", OwnerID: "owner-1", TextChannelID: "text-1"},
		{GuildID: "guild-1", ChannelID: "voice-1", TextChannelID: "text-1"},
	} {
		if err := invalid.Validate(); !errors.Is(err, domain.ErrInvalidVoiceRoomLock) {
			t.Fatalf("expected invalid lock error for %#v, got %v", invalid, err)
		}
	}
}
