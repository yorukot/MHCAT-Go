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

func TestPrizeDeleteDefinitionMatchesLegacy(t *testing.T) {
	definition := PrizeDeleteDefinition()
	if definition.Name != GachaPrizeDeleteCommandName || definition.Description != "刪除扭蛋的獎池" {
		t.Fatalf("definition = %#v", definition)
	}
	if definition.DefaultMemberPermissions == nil || *definition.DefaultMemberPermissions != gachaManageMessagesPermission {
		t.Fatalf("permissions = %#v", definition.DefaultMemberPermissions)
	}
	if len(definition.Options) != 1 {
		t.Fatalf("expected one option, got %#v", definition.Options)
	}
	option := definition.Options[0]
	if option.Type != commands.OptionTypeString || option.Name != gachaPrizeNameOption || option.Description != "輸入這個獎品叫甚麼，以及簡單概述" || !option.Required {
		t.Fatalf("option = %#v", option)
	}
	if definition.Ownership == nil || definition.Ownership.Owner != commands.OwnerMHCATRefactor || definition.Ownership.SinceWave != "gacha-prize-delete" || !definition.Ownership.Managed {
		t.Fatalf("ownership = %#v", definition.Ownership)
	}
	if err := commands.ValidateRegistry(commands.NewRegistry(commands.Scope{Kind: commands.ScopeGuild, GuildID: "guild-1"}, AllDefinitions())); err != nil {
		t.Fatalf("validate: %v", err)
	}
}
