package gacha

import (
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/commands"
)

func TestPrizeListDefinitionMatchesLegacy(t *testing.T) {
	definition := PrizeListDefinition()
	if definition.Name != GachaPrizeListCommandName || definition.Description != "增加扭蛋的獎池" {
		t.Fatalf("definition = %#v", definition)
	}
	if len(definition.Options) != 0 {
		t.Fatalf("expected no options, got %#v", definition.Options)
	}
	if definition.Ownership == nil || definition.Ownership.Owner != commands.OwnerMHCATRefactor || definition.Ownership.SinceWave != "gacha-prize-list" || !definition.Ownership.Managed {
		t.Fatalf("ownership = %#v", definition.Ownership)
	}
	if err := commands.ValidateRegistry(commands.NewRegistry(commands.Scope{Kind: commands.ScopeGuild, GuildID: "guild-1"}, Definitions())); err != nil {
		t.Fatalf("validate: %v", err)
	}
}
