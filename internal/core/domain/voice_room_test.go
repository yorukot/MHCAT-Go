package domain_test

import (
	"errors"
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
	invalid.Name = " "
	if err := invalid.Validate(); !errors.Is(err, domain.ErrInvalidVoiceRoomConfig) {
		t.Fatalf("expected invalid name error, got %v", err)
	}
}
