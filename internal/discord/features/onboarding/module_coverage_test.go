package onboarding

import (
	"errors"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
)

func TestOnboardingModuleCommandsFollowEnabledFeatures(t *testing.T) {
	module := Module{joinRoleConfigured: true, leaveConfigured: true, verificationEnabled: true, flowEnabled: true, accountAgeEnabled: true}
	if len(module.Commands()) == 0 {
		t.Fatal("enabled onboarding module commands must not be empty")
	}
}

func TestOnboardingFallbackHelpers(t *testing.T) {
	if value := stringPtr("value"); value == nil || *value != "value" {
		t.Fatalf("string pointer = %#v", value)
	}
	if color := randomEmbedColor(); color < 0 || color > 0xFFFFFF {
		t.Fatalf("embed color = %#x", color)
	}
	if usernameFromTag(" Tester#1234 ") != "Tester" || usernameFromTag("Tester") != "Tester" {
		t.Fatal("username fallback mismatch")
	}
	message := accountAgeErrorFromError(ports.ErrAccountAgeConfigMissing)
	if len(message.Embeds) != 1 || message.Embeds[0].Title == "" {
		t.Fatalf("account age error = %#v", message)
	}
	_ = accountAgeErrorFromError(errors.New("unknown"))
}
