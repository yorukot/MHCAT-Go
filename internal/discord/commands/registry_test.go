package commands_test

import (
	"reflect"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/commands"
)

func TestEmptyRegistryValid(t *testing.T) {
	registry := commands.EmptyRegistry(commands.Scope{Kind: commands.ScopeGlobal})
	if err := commands.ValidateRegistry(registry); err != nil {
		t.Fatalf("empty registry should be valid: %v", err)
	}
}

func TestRegistryDeterministicOrdering(t *testing.T) {
	registry := commands.NewRegistry(commands.Scope{Kind: commands.ScopeGlobal}, []commands.Definition{
		{Type: commands.CommandTypeUser, Name: "View Profile"},
		{Type: commands.CommandTypeChatInput, Name: "zeta", Description: "last"},
		{Type: commands.CommandTypeChatInput, Name: "alpha", Description: "first"},
	})
	got := []string{registry.Commands[0].Name, registry.Commands[1].Name, registry.Commands[2].Name}
	want := []string{"alpha", "zeta", "View Profile"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("sorted names = %#v, want %#v", got, want)
	}
}

func TestEnabledDefinitionsFiltersLocalOnlyCommands(t *testing.T) {
	registry := commands.NewRegistry(commands.Scope{Kind: commands.ScopeGlobal}, []commands.Definition{
		{Type: commands.CommandTypeChatInput, Name: "visible", Description: "visible command"},
		{Type: commands.CommandTypeChatInput, Name: "hidden", Description: "hidden command", Hidden: true},
		{Type: commands.CommandTypeChatInput, Name: "internal", Description: "internal command", Internal: true},
	})
	enabled := commands.EnabledDefinitions(registry)
	if len(enabled) != 1 || enabled[0].Name != "visible" {
		t.Fatalf("enabled definitions = %#v", enabled)
	}
}
