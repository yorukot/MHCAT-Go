package commands_test

import (
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/commands"
)

func TestBuiltinRegistryValid(t *testing.T) {
	registry := commands.BuiltinRegistry(commands.Scope{Kind: commands.ScopeGlobal})
	if err := commands.ValidateRegistry(registry); err != nil {
		t.Fatalf("builtin registry failed validation: %v", err)
	}
	if len(registry.Commands) != 3 {
		t.Fatalf("builtin command count = %d", len(registry.Commands))
	}
}

func TestBuiltinRegistryDeterministicOrder(t *testing.T) {
	registry := commands.BuiltinRegistry(commands.Scope{Kind: commands.ScopeGlobal})
	var names []string
	for _, definition := range registry.Commands {
		names = append(names, definition.Name)
	}
	want := []string{"help", "info", "ping"}
	for index := range want {
		if names[index] != want[index] {
			t.Fatalf("names = %#v, want %#v", names, want)
		}
	}
}

func TestBuiltinRegistryDryRunPlanCreatesSelectedDefinitions(t *testing.T) {
	scope := commands.Scope{Kind: commands.ScopeGlobal}
	plan, err := commands.Diff(commands.BuiltinRegistry(scope), nil, commands.DiffOptions{Scope: scope})
	if err != nil {
		t.Fatalf("diff: %v", err)
	}
	got := map[string]commands.Operation{}
	for _, operation := range plan.Operations {
		got[operation.CommandName] = operation.Operation
	}
	for _, name := range []string{"help", "info", "ping"} {
		if got[name] != commands.OperationCreate {
			t.Fatalf("operation for %s = %s, want create", name, got[name])
		}
	}
}

func TestInfoDefinitionIncludesLegacySubcommands(t *testing.T) {
	definition := commands.InfoDefinition()
	found := map[string]bool{}
	for _, option := range definition.Options {
		if option.Type == commands.OptionTypeSubCommand {
			found[option.Name] = true
		}
	}
	for _, name := range []string{"user", "bot", "shard", "guild"} {
		if !found[name] {
			t.Fatalf("info option %q missing from %#v", name, definition.Options)
		}
	}
}

func TestBuiltinRegistryStripsLocalOnlyMetadata(t *testing.T) {
	registry := commands.BuiltinRegistry(commands.Scope{Kind: commands.ScopeGlobal})
	enabled := commands.EnabledDefinitions(registry)
	for _, definition := range enabled {
		if definition.Hidden || definition.Internal || definition.Disabled || definition.DocsURL != "" || definition.Ownership != nil {
			t.Fatalf("local-only metadata leaked into enabled definition: %#v", definition)
		}
	}
}
