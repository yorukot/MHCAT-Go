package birthday

import (
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/commands"
)

func TestDefinitionMatchesLegacyCommandShape(t *testing.T) {
	definition := Definition()
	if definition.Name != BirthdayCommandName || definition.Description != "讓你的伺服器可以為生日的人慶生!" {
		t.Fatalf("definition = %#v", definition)
	}
	if definition.DefaultMemberPermissions != nil {
		t.Fatalf("legacy birthday command should not set default member permissions: %#v", definition.DefaultMemberPermissions)
	}
	if len(definition.Options) != 5 {
		t.Fatalf("options = %#v", definition.Options)
	}
	config := definition.Options[0]
	if config.Type != commands.OptionTypeSubCommand || config.Name != subcommandConfig || len(config.Options) != 5 {
		t.Fatalf("config subcommand = %#v", config)
	}
	if config.Options[0].Name != optionMessage || config.Options[0].Type != commands.OptionTypeString || !config.Options[0].Required {
		t.Fatalf("message option = %#v", config.Options[0])
	}
	if config.Options[1].Name != optionChannel || config.Options[1].Type != commands.OptionTypeChannel || !config.Options[1].Required || len(config.Options[1].ChannelTypes) != 0 {
		t.Fatalf("channel option = %#v", config.Options[1])
	}
	if config.Options[2].Name != optionEveryoneCanSet || config.Options[2].Type != commands.OptionTypeBoolean || !config.Options[2].Required {
		t.Fatalf("can-set option = %#v", config.Options[2])
	}
	if config.Options[3].Name != optionUTC || len(config.Options[3].Choices) != 24 || config.Options[3].Choices[8].Value != "+08:00" || config.Options[3].Choices[23].Name != "UTC+23" {
		t.Fatalf("utc option = %#v", config.Options[3])
	}
	if config.Options[4].Name != optionRole || config.Options[4].Type != commands.OptionTypeRole || config.Options[4].Required {
		t.Fatalf("role option = %#v", config.Options[4])
	}
}
