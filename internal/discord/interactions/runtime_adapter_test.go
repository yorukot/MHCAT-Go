package interactions_test

import (
	"errors"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/interactions"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakediscord"
)

func TestValidateRuntimeInteractionSlash(t *testing.T) {
	if err := interactions.ValidateRuntimeInteraction(fakediscord.SlashInteraction("help")); err != nil {
		t.Fatalf("validate runtime interaction: %v", err)
	}
}

func TestValidateRuntimeInteractionInvalid(t *testing.T) {
	err := interactions.ValidateRuntimeInteraction(interactions.Interaction{Type: interactions.TypeSlash})
	if !errors.Is(err, interactions.ErrInvalidCommandOption) {
		t.Fatalf("expected ErrInvalidCommandOption, got %v", err)
	}
}
