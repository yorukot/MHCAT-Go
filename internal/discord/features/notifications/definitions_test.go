package notifications

import (
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/commands"
)

func TestDefinitionsMatchLegacy(t *testing.T) {
	list := ListDefinition()
	if list.Name != AutoNotificationListCommandName || list.Description != "查看所有的自動通知列表" {
		t.Fatalf("list definition = %#v", list)
	}
	if list.DefaultMemberPermissions == nil || *list.DefaultMemberPermissions != manageMessagesPermission {
		t.Fatalf("list permissions = %#v", list.DefaultMemberPermissions)
	}
	deleteDefinition := DeleteDefinition()
	if deleteDefinition.Name != AutoNotificationDeleteCommandName || deleteDefinition.Description != "刪除之前設定的自動通知" {
		t.Fatalf("delete definition = %#v", deleteDefinition)
	}
	if len(deleteDefinition.Options) != 1 || deleteDefinition.Options[0].Name != optionID || deleteDefinition.Options[0].Type != commands.OptionTypeString || !deleteDefinition.Options[0].Required {
		t.Fatalf("delete options = %#v", deleteDefinition.Options)
	}
	if err := commands.ValidateRegistry(commands.NewRegistry(commands.Scope{Kind: commands.ScopeGuild, GuildID: "guild-1"}, Definitions())); err != nil {
		t.Fatalf("validate: %v", err)
	}
}
