package redeem

import (
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/commands"
)

func TestDefinitionMatchesLegacyMetadata(t *testing.T) {
	definition := Definition()
	if definition.Name != CommandName || definition.Description != "兌換代碼" {
		t.Fatalf("definition = %#v", definition)
	}
	if len(definition.Options) != 1 {
		t.Fatalf("options = %#v", definition.Options)
	}
	option := definition.Options[0]
	if option.Name != optionCode || option.Description != "輸入您的代碼" || option.Type != commands.OptionTypeString || !option.Required {
		t.Fatalf("option = %#v", option)
	}
	if definition.DefaultMemberPermissions != nil {
		t.Fatalf("permissions = %#v", definition.DefaultMemberPermissions)
	}
}

func TestDefinitionsValidate(t *testing.T) {
	registry := commands.NewRegistry(commands.Scope{Kind: commands.ScopeGuild, GuildID: "guild"}, Definitions())
	if err := commands.ValidateRegistry(registry); err != nil {
		t.Fatalf("validate registry: %v", err)
	}
}
