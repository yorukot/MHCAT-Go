package economy

import (
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/commands"
)

func TestCoinQueryDefinitionMatchesLegacyCommand(t *testing.T) {
	definition := CoinQueryDefinition()
	registry := commands.NewRegistry(commands.Scope{Kind: commands.ScopeGuild, GuildID: "guild"}, []commands.Definition{definition})
	if err := commands.ValidateRegistry(registry); err != nil {
		t.Fatalf("validate registry: %v", err)
	}
	if definition.Name != "代幣查詢" || definition.Description != "查詢你有多少代幣" {
		t.Fatalf("unexpected command definition: %#v", definition)
	}
	if len(definition.Options) != 1 {
		t.Fatalf("expected one option, got %#v", definition.Options)
	}
	option := definition.Options[0]
	if option.Type != commands.OptionTypeUser || option.Name != "使用者" || option.Description != "要查詢的使用者" || option.Required {
		t.Fatalf("unexpected user option: %#v", option)
	}
	if !commands.IsManagedForScope(definition, commands.Scope{Kind: commands.ScopeGuild, GuildID: "guild"}) {
		t.Fatal("coin query command should be marked managed for guild-scoped staging sync")
	}
}

func TestSignInDefinitionsMatchLegacyCommands(t *testing.T) {
	definitions := SignInDefinitions()
	registry := commands.NewRegistry(commands.Scope{Kind: commands.ScopeGuild, GuildID: "guild"}, definitions)
	if err := commands.ValidateRegistry(registry); err != nil {
		t.Fatalf("validate registry: %v", err)
	}
	if len(definitions) != 2 {
		t.Fatalf("definitions = %#v", definitions)
	}
	if definitions[0].Name != SignInCommandName || definitions[0].Description != "簽到來獲得代幣" {
		t.Fatalf("sign-in definition = %#v", definitions[0])
	}
	if definitions[1].Name != SignInListCommandName || definitions[1].Description != "查看今天有誰簽到了" {
		t.Fatalf("sign-in list definition = %#v", definitions[1])
	}
	for _, definition := range definitions {
		if !commands.IsManagedForScope(definition, commands.Scope{Kind: commands.ScopeGuild, GuildID: "guild"}) {
			t.Fatalf("definition should be managed for guild scope: %#v", definition)
		}
	}
}

func TestSettingsDefinitionMatchesLegacyCommand(t *testing.T) {
	definition := SettingsDefinition()
	registry := commands.NewRegistry(commands.Scope{Kind: commands.ScopeGuild, GuildID: "guild"}, []commands.Definition{definition})
	if err := commands.ValidateRegistry(registry); err != nil {
		t.Fatalf("validate registry: %v", err)
	}
	if definition.Name != EconomySettingsCommandName || definition.Description != "Various settings related to tokens" {
		t.Fatalf("unexpected settings definition: %#v", definition)
	}
	if definition.NameLocalizations["zh-TW"] != "代幣相關設定" || definition.NameLocalizations["zh-CN"] != "代币相关设定" {
		t.Fatalf("localizations = %#v", definition.NameLocalizations)
	}
	if definition.DefaultMemberPermissions == nil || *definition.DefaultMemberPermissions != manageMessagesPermission {
		t.Fatalf("default member permissions = %#v", definition.DefaultMemberPermissions)
	}
	if len(definition.Options) != 5 {
		t.Fatalf("expected five options, got %#v", definition.Options)
	}
	wantNames := []string{"coin-raffle-takes", "check-in-cooldown-time", "check-in-give-coins", "notification-channel", "level-up-multiply-amount"}
	for i, want := range wantNames {
		if definition.Options[i].Name != want || !definition.Options[i].Required {
			t.Fatalf("option %d = %#v", i, definition.Options[i])
		}
	}
	if definition.Options[3].Type != commands.OptionTypeChannel || len(definition.Options[3].ChannelTypes) != 2 || definition.Options[3].ChannelTypes[0] != textChannelType || definition.Options[3].ChannelTypes[1] != announcementChannelType {
		t.Fatalf("notification channel option = %#v", definition.Options[3])
	}
	if definition.Options[4].Type != commands.OptionTypeNumber {
		t.Fatalf("xp multiplier option = %#v", definition.Options[4])
	}
	if !commands.IsManagedForScope(definition, commands.Scope{Kind: commands.ScopeGuild, GuildID: "guild"}) {
		t.Fatal("settings command should be marked managed for guild-scoped staging sync")
	}
}
