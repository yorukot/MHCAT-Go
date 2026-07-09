package work

import (
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/commands"
)

func TestWorkDefinitionMatchesLegacyShape(t *testing.T) {
	def := Definition()
	if def.Name != "打工系統" || def.Description != "用自己的心血來獲得一些獎勵吧!" {
		t.Fatalf("unexpected command metadata: %#v", def)
	}
	if len(def.Options) != 6 {
		t.Fatalf("expected 6 legacy subcommands, got %d", len(def.Options))
	}
	names := []string{"打工系統設定", "新增打工事項", "打工事項刪除", "打工介面", "增加個人精力", "增加全體精力"}
	for i, want := range names {
		if def.Options[i].Name != want || def.Options[i].Type != commands.OptionTypeSubCommand {
			t.Fatalf("option %d = %#v, want %q subcommand", i, def.Options[i], want)
		}
	}
	if def.Ownership == nil || !commands.IsManagedForScope(def, commands.Scope{Kind: commands.ScopeGuild, GuildID: "guild-1"}) {
		t.Fatalf("work command must be managed for guild staging")
	}
	if err := commands.ValidateRegistry(commands.NewRegistry(commands.Scope{Kind: commands.ScopeGuild, GuildID: "guild-1"}, Definitions())); err != nil {
		t.Fatalf("work registry validation failed: %v", err)
	}
}

func TestWorkAddSubcommandOptions(t *testing.T) {
	def := Definition()
	var add commands.Option
	for _, option := range def.Options {
		if option.Name == "新增打工事項" {
			add = option
			break
		}
	}
	if add.Name == "" {
		t.Fatal("missing add work subcommand")
	}
	if len(add.Options) != 5 {
		t.Fatalf("expected 5 add-work options, got %d", len(add.Options))
	}
	if add.Options[1].Type != commands.OptionTypeNumber || add.Options[4].Type != commands.OptionTypeRole || add.Options[4].Required {
		t.Fatalf("unexpected add-work options: %#v", add.Options)
	}
}
