package safety

import (
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/commands"
)

func TestAntiScamDefinitionMatchesLegacyMetadata(t *testing.T) {
	definition := AntiScamDefinition()
	if definition.Name != AntiScamCommandName {
		t.Fatalf("name = %q", definition.Name)
	}
	if definition.Description != "設定是否開啟防詐騙網址功能(輸入這個指令就會更改)" {
		t.Fatalf("description = %q", definition.Description)
	}
	if len(definition.Options) != 0 {
		t.Fatalf("options = %#v", definition.Options)
	}
	if definition.DefaultMemberPermissions != nil {
		t.Fatalf("legacy toggle command should not have default permissions: %#v", definition.DefaultMemberPermissions)
	}
	if definition.Ownership == nil || !definition.Ownership.Managed || definition.Ownership.SinceWave != "anti-scam-config" {
		t.Fatalf("ownership = %#v", definition.Ownership)
	}
}

func TestScamReportDefinitionMatchesLegacyMetadata(t *testing.T) {
	definition := ScamReportDefinition()
	if definition.Name != ScamReportCommandName {
		t.Fatalf("name = %q", definition.Name)
	}
	if definition.Description != "回報詐騙網站" {
		t.Fatalf("description = %q", definition.Description)
	}
	if definition.DefaultMemberPermissions != nil {
		t.Fatalf("default permissions = %#v", definition.DefaultMemberPermissions)
	}
	if definition.Ownership == nil || !definition.Ownership.Managed || definition.Ownership.SinceWave != "anti-scam-report" {
		t.Fatalf("ownership = %#v", definition.Ownership)
	}
	if len(definition.Options) != 1 {
		t.Fatalf("options = %#v", definition.Options)
	}
	option := definition.Options[0]
	if option.Name != "網址" || option.Description != "回報網址" || option.Type != commands.OptionTypeString || !option.Required {
		t.Fatalf("option = %#v", option)
	}
}

func TestDefinitionsValidate(t *testing.T) {
	if err := commands.ValidateRegistry(commands.NewRegistry(commands.Scope{Kind: commands.ScopeGuild, GuildID: "guild-1"}, Definitions())); err != nil {
		t.Fatalf("validate definitions: %v", err)
	}
}

func TestSafetyDefinitionGroupsValidateSeparately(t *testing.T) {
	for name, definitions := range map[string][]commands.Definition{
		"config": ConfigDefinitions(),
		"report": ReportDefinitions(),
	} {
		if err := commands.ValidateRegistry(commands.NewRegistry(commands.Scope{Kind: commands.ScopeGuild, GuildID: "guild-1"}, definitions)); err != nil {
			t.Fatalf("validate %s definitions: %v", name, err)
		}
	}
}
