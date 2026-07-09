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

func TestWarningSettingsDefinitionMatchesLegacyShape(t *testing.T) {
	definition := WarningSettingsDefinition()
	if definition.Name != "警告設定" || definition.Description != "警告的各種設定" {
		t.Fatalf("definition = %#v", definition)
	}
	if len(definition.Options) != 2 {
		t.Fatalf("options = %#v", definition.Options)
	}
	action := definition.Options[0]
	if action.Type != commands.OptionTypeString || action.Name != warningSettingsOptionAction || action.Description != "警告他的原因" || !action.Required {
		t.Fatalf("action option = %#v", action)
	}
	if len(action.Choices) != 2 || action.Choices[0].Name != "停權" || action.Choices[0].Value != "停權" || action.Choices[1].Name != "踢出" || action.Choices[1].Value != "踢出" {
		t.Fatalf("choices = %#v", action.Choices)
	}
	threshold := definition.Options[1]
	if threshold.Type != commands.OptionTypeInteger || threshold.Name != warningSettingsOptionThreshold || threshold.Description != "被警告幾次後要執行這個動作!" || !threshold.Required {
		t.Fatalf("threshold option = %#v", threshold)
	}
	if !commands.IsManagedForScope(definition, commands.Scope{Kind: commands.ScopeGuild, GuildID: "guild-1"}) {
		t.Fatal("warning settings command should be managed for guild staging")
	}
	if err := commands.ValidateRegistry(commands.NewRegistry(commands.Scope{Kind: commands.ScopeGuild, GuildID: "guild-1"}, SettingsDefinitions())); err != nil {
		t.Fatalf("validate registry: %v", err)
	}
}
