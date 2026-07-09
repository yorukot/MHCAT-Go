package lottery

import (
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/commands"
)

func TestCreateDefinitionMatchesLegacyMetadata(t *testing.T) {
	definition := CreateDefinition()
	if definition.Name != LotteryCreateCommandName {
		t.Fatalf("name = %q", definition.Name)
	}
	if definition.Description != "設置抽獎訊息" {
		t.Fatalf("description = %q", definition.Description)
	}
	if definition.DefaultMemberPermissions == nil || *definition.DefaultMemberPermissions != manageMessagesPermission {
		t.Fatalf("permissions = %#v", definition.DefaultMemberPermissions)
	}
	if definition.DocsURL != "https://docsmhcat.yorukot.meocs/lotter" {
		t.Fatalf("docs url = %q", definition.DocsURL)
	}
	if len(definition.Options) != 7 {
		t.Fatalf("options = %#v", definition.Options)
	}
	want := []struct {
		name     string
		typ      commands.OptionType
		required bool
	}{
		{"截止日期", commands.OptionTypeString, true},
		{"抽出幾位中獎者", commands.OptionTypeInteger, true},
		{"獎品", commands.OptionTypeString, true},
		{"內文", commands.OptionTypeString, true},
		{"可以抽的身分組", commands.OptionTypeRole, false},
		{"不能抽的身分組", commands.OptionTypeRole, false},
		{"最高抽獎人數", commands.OptionTypeInteger, false},
	}
	for i, wantOption := range want {
		got := definition.Options[i]
		if got.Name != wantOption.name || got.Type != wantOption.typ || got.Required != wantOption.required {
			t.Fatalf("option %d = %#v, want %#v", i, got, wantOption)
		}
	}
}

func TestDefinitionsValidate(t *testing.T) {
	if err := commands.ValidateRegistry(commands.NewRegistry(commands.Scope{Kind: commands.ScopeGuild, GuildID: "guild-1"}, Definitions())); err != nil {
		t.Fatalf("validate definitions: %v", err)
	}
}
