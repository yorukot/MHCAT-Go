package redeem

import (
	"reflect"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/commands"
)

func TestDefinitionMatchesLegacyMetadata(t *testing.T) {
	want := commands.Definition{
		Type:        commands.CommandTypeChatInput,
		Name:        "兌換",
		Description: "兌換代碼",
		Ownership:   commands.ManagedOwnership("redeem", commands.ScopeGuild),
		Options: []commands.Option{{
			Type:        commands.OptionTypeString,
			Name:        "代碼",
			Description: "輸入您的代碼",
			Required:    true,
		}},
	}
	if got := Definition(); !reflect.DeepEqual(got, want) {
		t.Fatalf("definition = %#v, want %#v", got, want)
	}
}

func TestDefinitionsValidate(t *testing.T) {
	registry := commands.NewRegistry(commands.Scope{Kind: commands.ScopeGuild, GuildID: "guild"}, Definitions())
	if err := commands.ValidateRegistry(registry); err != nil {
		t.Fatalf("validate registry: %v", err)
	}
}
