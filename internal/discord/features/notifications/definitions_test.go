package notifications

import (
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/commands"
)

func TestDefinitionsMatchLegacy(t *testing.T) {
	setup := SetupDefinition()
	if setup.Name != AutoNotificationSetupCommandName || setup.Description != "Set where automatic notification should be send" {
		t.Fatalf("setup definition = %#v", setup)
	}
	if setup.DefaultMemberPermissions != nil {
		t.Fatalf("legacy setup command should not have default permissions: %#v", setup.DefaultMemberPermissions)
	}
	if len(setup.Options) != 1 || setup.Options[0].Name != optionChannel || setup.Options[0].Type != commands.OptionTypeChannel || !setup.Options[0].Required {
		t.Fatalf("setup options = %#v", setup.Options)
	}
	if len(setup.Options[0].ChannelTypes) != 2 || setup.Options[0].ChannelTypes[0] != 0 || setup.Options[0].ChannelTypes[1] != 5 {
		t.Fatalf("setup channel types = %#v", setup.Options[0].ChannelTypes)
	}
	list := ListDefinition()
	if list.Name != AutoNotificationListCommandName || list.Description != "查看所有的自動通知列表" {
		t.Fatalf("list definition = %#v", list)
	}
	if list.DefaultMemberPermissions != nil {
		t.Fatalf("legacy list command should not have default permissions: %#v", list.DefaultMemberPermissions)
	}
	deleteDefinition := DeleteDefinition()
	if deleteDefinition.Name != AutoNotificationDeleteCommandName || deleteDefinition.Description != "刪除之前設定的自動通知" {
		t.Fatalf("delete definition = %#v", deleteDefinition)
	}
	if deleteDefinition.DefaultMemberPermissions != nil {
		t.Fatalf("legacy delete command should not have default permissions: %#v", deleteDefinition.DefaultMemberPermissions)
	}
	if len(deleteDefinition.Options) != 1 || deleteDefinition.Options[0].Name != optionID || deleteDefinition.Options[0].Type != commands.OptionTypeString || !deleteDefinition.Options[0].Required {
		t.Fatalf("delete options = %#v", deleteDefinition.Options)
	}
	if err := commands.ValidateRegistry(commands.NewRegistry(commands.Scope{Kind: commands.ScopeGuild, GuildID: "guild-1"}, Definitions())); err != nil {
		t.Fatalf("validate: %v", err)
	}
}
