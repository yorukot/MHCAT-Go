package moderation

import (
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/commands"
)

func TestWarningHistoryDefinitionMatchesLegacyShape(t *testing.T) {
	definition := WarningHistoryDefinition()
	if definition.Name != "警告紀錄" || definition.Description != "收尋一位使用者的警告" {
		t.Fatalf("definition = %#v", definition)
	}
	if len(definition.Options) != 1 {
		t.Fatalf("options = %#v", definition.Options)
	}
	option := definition.Options[0]
	if option.Type != commands.OptionTypeUser || option.Name != "使用者" || option.Description != "要收尋的使用者!" || !option.Required {
		t.Fatalf("option = %#v", option)
	}
	if !commands.IsManagedForScope(definition, commands.Scope{Kind: commands.ScopeGuild, GuildID: "guild-1"}) {
		t.Fatal("warning history command should be managed for guild staging")
	}
	if err := commands.ValidateRegistry(commands.NewRegistry(commands.Scope{Kind: commands.ScopeGuild, GuildID: "guild-1"}, Definitions())); err != nil {
		t.Fatalf("validate registry: %v", err)
	}
}
