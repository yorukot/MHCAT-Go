package runtime_test

import (
	"errors"
	"testing"

	discordruntime "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/runtime"
)

func TestRuntimeErrorsAreTyped(t *testing.T) {
	if !errors.Is(discordruntime.ErrInvalidRuntimeEvent, discordruntime.ErrInvalidRuntimeEvent) {
		t.Fatal("runtime errors must be detectable")
	}
}
