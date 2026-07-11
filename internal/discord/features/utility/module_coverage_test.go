package utility

import (
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/commands"
)

func TestUtilityModuleName(t *testing.T) {
	module := NewModule(commands.Registry{}, nil, nil, nil)
	if name := module.Name(); name != "utility" {
		t.Fatalf("utility module name = %q", name)
	}
	if len(module.Commands()) == 0 {
		t.Fatal("utility module commands must not be empty")
	}
}
