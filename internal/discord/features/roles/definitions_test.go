package roles

import (
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/commands"
)

func TestDefinitionsMatchLegacyCommands(t *testing.T) {
	definitions := Definitions()
	if len(definitions) != 3 {
		t.Fatalf("definitions = %#v", definitions)
	}
	button := definitions[0]
	if button.Name != RoleButtonCommandName || button.Description != "設定領取身分組的消息(點按鈕自動增加身分組)" {
		t.Fatalf("button definition = %#v", button)
	}
	if button.DefaultMemberPermissions != nil {
		t.Fatalf("legacy role commands should remain publicly discoverable: %#v", button.DefaultMemberPermissions)
	}
	if len(button.Options) != 1 || button.Options[0].Type != commands.OptionTypeRole || button.Options[0].Name != "身分組" || !button.Options[0].Required {
		t.Fatalf("button role option = %#v", button.Options)
	}
	set := definitions[1]
	if set.Name != RoleReactionSetCommandName || len(set.Options) != 3 {
		t.Fatalf("reaction set definition = %#v", set)
	}
	if set.DefaultMemberPermissions != nil {
		t.Fatalf("legacy reaction-set command should remain publicly discoverable: %#v", set.DefaultMemberPermissions)
	}
	if set.Options[0].Type != commands.OptionTypeString || set.Options[0].Name != "訊息url" || !set.Options[0].Required {
		t.Fatalf("set url option = %#v", set.Options[0])
	}
	if set.Options[1].Type != commands.OptionTypeRole || set.Options[1].Name != "身分組" || !set.Options[1].Required {
		t.Fatalf("set role option = %#v", set.Options[1])
	}
	if set.Options[2].Type != commands.OptionTypeString || set.Options[2].Name != "表情符號" || !set.Options[2].Required {
		t.Fatalf("set emoji option = %#v", set.Options[2])
	}
	del := definitions[2]
	if del.Name != RoleReactionDeleteCommandName || len(del.Options) != 2 {
		t.Fatalf("reaction delete definition = %#v", del)
	}
	if del.DefaultMemberPermissions != nil {
		t.Fatalf("legacy reaction-delete command should remain publicly discoverable: %#v", del.DefaultMemberPermissions)
	}
}
