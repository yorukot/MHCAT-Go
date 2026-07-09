package balance

import (
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/commands"
)

func TestDefinitionMatchesLegacyCommand(t *testing.T) {
	definition := Definition()
	registry := commands.NewRegistry(commands.Scope{Kind: commands.ScopeGuild, GuildID: "guild"}, []commands.Definition{definition})
	if err := commands.ValidateRegistry(registry); err != nil {
		t.Fatalf("validate registry: %v", err)
	}
	if definition.Name != CommandName || definition.Description != "查看剩餘餘額" {
		t.Fatalf("definition = %#v", definition)
	}
	if !commands.IsManagedForScope(definition, commands.Scope{Kind: commands.ScopeGuild, GuildID: "guild"}) {
		t.Fatalf("definition should be managed for guild scope: %#v", definition)
	}
}
