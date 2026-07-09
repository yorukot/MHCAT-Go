package stats

import (
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/commands"
)

func TestQueryDefinitionMatchesLegacyMetadata(t *testing.T) {
	definition := QueryDefinition()
	if definition.Name != StatsQueryCommandName {
		t.Fatalf("name = %q", definition.Name)
	}
	if definition.Description != "查詢統計消息" {
		t.Fatalf("description = %q", definition.Description)
	}
	if len(definition.Options) != 0 {
		t.Fatalf("options = %#v", definition.Options)
	}
	if definition.DefaultMemberPermissions != nil {
		t.Fatalf("permissions = %#v", definition.DefaultMemberPermissions)
	}
}

func TestDeleteDefinitionMatchesLegacyMetadata(t *testing.T) {
	definition := DeleteDefinition()
	if definition.Name != StatsDeleteCommandName {
		t.Fatalf("name = %q", definition.Name)
	}
	if definition.Description != "刪除統計消息" {
		t.Fatalf("description = %q", definition.Description)
	}
	if len(definition.Options) != 0 {
		t.Fatalf("options = %#v", definition.Options)
	}
	if definition.DefaultMemberPermissions == nil || *definition.DefaultMemberPermissions != manageMessagesPermission {
		t.Fatalf("permissions = %#v", definition.DefaultMemberPermissions)
	}
}

func TestCreateDefinitionMatchesLegacyMetadata(t *testing.T) {
	definition := CreateDefinition()
	if definition.Name != StatsCreateCommandName {
		t.Fatalf("name = %q", definition.Name)
	}
	if definition.Description != "創建統計消息" {
		t.Fatalf("description = %q", definition.Description)
	}
	if definition.DefaultMemberPermissions == nil || *definition.DefaultMemberPermissions != manageMessagesPermission {
		t.Fatalf("permissions = %#v", definition.DefaultMemberPermissions)
	}
	if len(definition.Options) != 2 {
		t.Fatalf("options = %#v", definition.Options)
	}
	channelType := definition.Options[0]
	if channelType.Type != commands.OptionTypeString || channelType.Name != statsOptionChannelType || !channelType.Required {
		t.Fatalf("channel type option = %#v", channelType)
	}
	if len(channelType.Choices) != 2 || channelType.Choices[0].Name != "文字頻道" || channelType.Choices[1].Name != "語音頻道" {
		t.Fatalf("channel choices = %#v", channelType.Choices)
	}
	stat := definition.Options[1]
	if stat.Type != commands.OptionTypeString || stat.Name != statsOptionStat || stat.Required {
		t.Fatalf("stat option = %#v", stat)
	}
	if len(stat.Choices) != 3 || stat.Choices[0].Name != "頻道數量" || stat.Choices[2].Name != "語音頻道數量" {
		t.Fatalf("stat choices = %#v", stat.Choices)
	}
}

func TestDefinitionsValidate(t *testing.T) {
	if err := commands.ValidateRegistry(commands.NewRegistry(commands.Scope{Kind: commands.ScopeGuild, GuildID: "guild-1"}, Definitions())); err != nil {
		t.Fatalf("validate definitions: %v", err)
	}
}
