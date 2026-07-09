package autochat

import (
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/commands"
)

func TestDefinitionsMatchLegacyCommandShape(t *testing.T) {
	definitions := Definitions()
	if len(definitions) != 2 {
		t.Fatalf("definitions = %#v", definitions)
	}
	set := definitions[0]
	if set.Name != AutoChatSetCommandName || set.Description != "設定自動聊天頻道要在哪裡發送" {
		t.Fatalf("set definition = %#v", set)
	}
	if set.DefaultMemberPermissions == nil || *set.DefaultMemberPermissions != manageMessagesPermission {
		t.Fatalf("permissions = %#v", set.DefaultMemberPermissions)
	}
	if len(set.Options) != 1 || set.Options[0].Name != optionChannel || set.Options[0].Type != commands.OptionTypeChannel || !set.Options[0].Required {
		t.Fatalf("channel option = %#v", set.Options)
	}
	if len(set.Options[0].ChannelTypes) != 2 || set.Options[0].ChannelTypes[0] != 0 || set.Options[0].ChannelTypes[1] != 5 {
		t.Fatalf("channel types = %#v", set.Options[0].ChannelTypes)
	}
	deleteCommand := definitions[1]
	if deleteCommand.Name != AutoChatDeleteCommandName || deleteCommand.Description != "刪除自動聊天頻道要在哪裡發送" {
		t.Fatalf("delete definition = %#v", deleteCommand)
	}
	if deleteCommand.DefaultMemberPermissions == nil || *deleteCommand.DefaultMemberPermissions != manageMessagesPermission {
		t.Fatalf("delete permissions = %#v", deleteCommand.DefaultMemberPermissions)
	}
}
