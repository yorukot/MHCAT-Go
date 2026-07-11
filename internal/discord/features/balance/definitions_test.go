package balance

import (
	"reflect"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/commands"
)

func TestDefinitionMatchesLegacyCommand(t *testing.T) {
	definition := Definition()
	registry := commands.NewRegistry(commands.Scope{Kind: commands.ScopeGuild, GuildID: "guild"}, []commands.Definition{definition})
	if err := commands.ValidateRegistry(registry); err != nil {
		t.Fatalf("validate registry: %v", err)
	}
	want := commands.Definition{
		Type:        commands.CommandTypeChatInput,
		Name:        "查看餘額",
		Description: "查看剩餘餘額",
		Ownership:   commands.ManagedOwnership("balance-query", commands.ScopeGuild),
	}
	if !reflect.DeepEqual(definition, want) {
		t.Fatalf("definition = %#v, want %#v", definition, want)
	}
	if !commands.IsManagedForScope(definition, commands.Scope{Kind: commands.ScopeGuild, GuildID: "guild"}) {
		t.Fatalf("definition should be managed for guild scope: %#v", definition)
	}
}
