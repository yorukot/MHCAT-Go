package stats

import (
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/commands"
)

func TestQueryDefinitionMatchesLegacyMetadata(t *testing.T) {
	definition := QueryDefinition()
	if definition.Name != StatsQueryCommandName {
		t.Fatalf("name = %q", definition.Name)
	}
	if definition.Description != "查詢統計消息" {
		t.Fatalf("description = %q", definition.Description)
	}
	if len(definition.Options) != 0 {
		t.Fatalf("options = %#v", definition.Options)
	}
	if definition.DefaultMemberPermissions != nil {
		t.Fatalf("permissions = %#v", definition.DefaultMemberPermissions)
	}
}

func TestDefinitionsValidate(t *testing.T) {
	if err := commands.ValidateRegistry(commands.NewRegistry(commands.Scope{Kind: commands.ScopeGuild, GuildID: "guild-1"}, Definitions())); err != nil {
		t.Fatalf("validate definitions: %v", err)
	}
}
