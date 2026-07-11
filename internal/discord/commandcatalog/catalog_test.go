package commandcatalog

import (
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/commands"
)

func TestAllDefinitionsIsCompleteValidAndUnique(t *testing.T) {
	definitions := AllDefinitions()
	if len(definitions) != 74 {
		t.Fatalf("definition count=%d, want 74", len(definitions))
	}
	registry := commands.NewRegistry(commands.Scope{Kind: commands.ScopeGuild, GuildID: "catalog-test"}, definitions)
	if err := commands.ValidateRegistry(registry); err != nil {
		t.Fatalf("validate command catalog: %v", err)
	}
	seen := make(map[string]struct{}, len(definitions))
	for _, definition := range definitions {
		if _, exists := seen[definition.Name]; exists {
			t.Fatalf("duplicate command name %q", definition.Name)
		}
		seen[definition.Name] = struct{}{}
	}
}
