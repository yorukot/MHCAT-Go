package utility_test

import (
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/commands"
	featureutility "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/features/utility"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/interactions"
)

func TestModuleCommandsValid(t *testing.T) {
	module := featureutility.NewModule(commands.BuiltinRegistry(commands.Scope{Kind: commands.ScopeGlobal}), nil, nil, nil)
	registry := commands.NewRegistry(commands.Scope{Kind: commands.ScopeGlobal}, module.Commands())
	if err := commands.ValidateRegistry(registry); err != nil {
		t.Fatalf("validate module commands: %v", err)
	}
}

func TestModuleRegistersRoutes(t *testing.T) {
	module := featureutility.NewModule(commands.BuiltinRegistry(commands.Scope{Kind: commands.ScopeGlobal}), nil, nil, nil)
	router := interactions.NewRouter()
	if err := module.RegisterRoutes(router); err != nil {
		t.Fatalf("register routes: %v", err)
	}
}
