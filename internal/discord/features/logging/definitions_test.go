package logging

import (
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/commands"
)

func TestLoggingConfigDefinitionMatchesLegacyShape(t *testing.T) {
	definition := LoggingConfigDefinition()
	if definition.Name != "set-log-channel" || definition.NameLocalizations["zh-TW"] != "設置日誌" || definition.NameLocalizations["zh-CN"] != "设置日志" {
		t.Fatalf("definition name/localizations = %#v", definition)
	}
	if definition.Description != "Set where log messages should send" || definition.DescriptionLocalizations["zh-TW"] != "設置日誌訊息要在哪發送" {
		t.Fatalf("definition description/localizations = %#v", definition)
	}
	if definition.DefaultMemberPermissions == nil || *definition.DefaultMemberPermissions != manageMessagesPermission {
		t.Fatalf("default permissions = %#v", definition.DefaultMemberPermissions)
	}
	if definition.Ownership == nil || !definition.Ownership.Managed {
		t.Fatalf("ownership = %#v", definition.Ownership)
	}
	if len(definition.Options) != 1 {
		t.Fatalf("options = %#v", definition.Options)
	}
	option := definition.Options[0]
	if option.Type != commands.OptionTypeChannel || option.Name != "channel" || !option.Required {
		t.Fatalf("channel option = %#v", option)
	}
	if len(option.ChannelTypes) != 2 || option.ChannelTypes[0] != 0 || option.ChannelTypes[1] != 5 {
		t.Fatalf("channel types = %#v", option.ChannelTypes)
	}
}
