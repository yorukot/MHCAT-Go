package domain

import (
	"errors"
	"testing"
)

func TestAntiScamConfigValidateRequiresGuildID(t *testing.T) {
	config := AntiScamConfig{GuildID: " "}
	if err := config.Validate(); !errors.Is(err, ErrInvalidAntiScamConfig) {
		t.Fatalf("expected invalid anti-scam config, got %v", err)
	}
}

func TestAntiScamConfigValidateAllowsOpenFalse(t *testing.T) {
	config := AntiScamConfig{GuildID: "guild-1", Open: false}
	if err := config.Validate(); err != nil {
		t.Fatalf("validate anti-scam config: %v", err)
	}
}
