package voice

import (
	"reflect"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/commands"
)

func TestDefinitionsMatchLegacyVoiceRoomCommands(t *testing.T) {
	definitions := Definitions()
	if len(definitions) != 2 {
		t.Fatalf("definitions len = %d", len(definitions))
	}
	set := definitions[0]
	if set.Name != VoiceRoomSetCommandName || set.Description != "設定語音包廂" {
		t.Fatalf("set definition = %#v", set)
	}
	if set.DefaultMemberPermissions == nil || *set.DefaultMemberPermissions != manageMessagesPermission {
		t.Fatalf("set default permissions = %#v", set.DefaultMemberPermissions)
	}
	if len(set.Options) != 4 {
		t.Fatalf("set options = %#v", set.Options)
	}
	if set.Options[0].Type != commands.OptionTypeChannel || set.Options[0].Name != optionTriggerChannel || !set.Options[0].Required || !reflect.DeepEqual(set.Options[0].ChannelTypes, []int{2, 13}) {
		t.Fatalf("trigger option = %#v", set.Options[0])
	}
	if set.Options[1].Type != commands.OptionTypeString || set.Options[1].Name != optionRoomName || !set.Options[1].Required {
		t.Fatalf("room-name option = %#v", set.Options[1])
	}
	if set.Options[2].Type != commands.OptionTypeBoolean || set.Options[2].Name != optionOwnerLock || !set.Options[2].Required {
		t.Fatalf("owner-lock option = %#v", set.Options[2])
	}
	if set.Options[3].Type != commands.OptionTypeInteger || set.Options[3].Name != optionUserLimit || set.Options[3].Required {
		t.Fatalf("limit option = %#v", set.Options[3])
	}

	deleteDefinition := definitions[1]
	if deleteDefinition.Name != VoiceRoomDeleteCommandName || deleteDefinition.Description != "刪除語音包廂設置" {
		t.Fatalf("delete definition = %#v", deleteDefinition)
	}
	if deleteDefinition.DefaultMemberPermissions == nil || *deleteDefinition.DefaultMemberPermissions != manageMessagesPermission {
		t.Fatalf("delete default permissions = %#v", deleteDefinition.DefaultMemberPermissions)
	}
	if len(deleteDefinition.Options) != 1 || deleteDefinition.Options[0].Name != optionChannelOrGroup || !deleteDefinition.Options[0].Required {
		t.Fatalf("delete options = %#v", deleteDefinition.Options)
	}
}
