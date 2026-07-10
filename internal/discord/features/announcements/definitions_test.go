package announcements

import (
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/commands"
)

func TestConfigDefinitionMatchesLegacyShape(t *testing.T) {
	definition := ConfigDefinition()
	if definition.Name != ConfigCommandName || definition.Description != "設定公告在哪發送" {
		t.Fatalf("definition = %#v", definition)
	}
	if definition.DefaultMemberPermissions != nil {
		t.Fatalf("legacy config command must remain publicly discoverable: %#v", definition.DefaultMemberPermissions)
	}
	if definition.DocsURL != "https://docsmhcat.yorukot.meocs/ann_set" {
		t.Fatalf("docs URL changed: %q", definition.DocsURL)
	}
	if len(definition.Options) != 3 {
		t.Fatalf("options = %#v", definition.Options)
	}
	names := []string{subcommandOnce, subcommandBound, subcommandDeleteBound}
	for i, name := range names {
		if definition.Options[i].Name != name || definition.Options[i].Type != commands.OptionTypeSubCommand {
			t.Fatalf("subcommand %d = %#v", i, definition.Options[i])
		}
	}
	if got := definition.Options[0].Options[0].ChannelTypes; len(got) != 2 || got[0] != 0 || got[1] != 5 {
		t.Fatalf("channel types = %#v", got)
	}
	registry := commands.NewRegistry(commands.Scope{Kind: commands.ScopeGuild, GuildID: "guild"}, Definitions())
	if err := commands.ValidateRegistry(registry); err != nil {
		t.Fatalf("validate registry: %v", err)
	}
}

func TestSendDefinitionMatchesLegacyShape(t *testing.T) {
	definition := SendDefinition()
	if definition.Name != SendCommandName || definition.Description != "發送公告訊息" {
		t.Fatalf("definition = %#v", definition)
	}
	if definition.DocsURL != "https://docsmhcat.yorukot.me/docs/ann" {
		t.Fatalf("docs URL changed: %q", definition.DocsURL)
	}
	if definition.DefaultMemberPermissions != nil {
		t.Fatalf("legacy send command should not set default member permissions: %#v", definition.DefaultMemberPermissions)
	}
	if len(definition.Options) != 0 {
		t.Fatalf("send command must not add options: %#v", definition.Options)
	}
	registry := commands.NewRegistry(commands.Scope{Kind: commands.ScopeGuild, GuildID: "guild"}, SendDefinitions())
	if err := commands.ValidateRegistry(registry); err != nil {
		t.Fatalf("validate registry: %v", err)
	}
}
