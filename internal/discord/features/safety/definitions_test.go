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
	if definition.DefaultMemberPermissions == nil || *definition.DefaultMemberPermissions != manageMessagesPermission {
		t.Fatalf("default permissions = %#v", definition.DefaultMemberPermissions)
	}
	if definition.Ownership == nil || !definition.Ownership.Managed || definition.Ownership.SinceWave != "anti-scam-config" {
		t.Fatalf("ownership = %#v", definition.Ownership)
	}
}

func TestDefinitionsValidate(t *testing.T) {
	if err := commands.ValidateRegistry(commands.NewRegistry(commands.Scope{Kind: commands.ScopeGuild, GuildID: "guild-1"}, Definitions())); err != nil {
		t.Fatalf("validate definitions: %v", err)
	}
}
