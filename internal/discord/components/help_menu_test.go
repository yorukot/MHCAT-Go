package components_test

import (
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/commands"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/components"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/customid"
)

func TestBuildHelpMenuUsesVersionedCustomID(t *testing.T) {
	menu, err := components.BuildHelpMenu(commands.BuiltinRegistry(commands.Scope{Kind: commands.ScopeGlobal}))
	if err != nil {
		t.Fatalf("build help menu: %v", err)
	}
	parsed, err := customid.ParseComponent(menu.CustomID)
	if err != nil {
		t.Fatalf("parse custom id: %v", err)
	}
	if parsed.Feature != "help" || parsed.Action != "category" || parsed.Version != customid.VersionV1 {
		t.Fatalf("parsed = %#v", parsed)
	}
	if len(menu.CustomID) > customid.MaxCustomIDLength {
		t.Fatalf("custom id length = %d", len(menu.CustomID))
	}
}

func TestBuildHelpMenuExcludesHiddenCommands(t *testing.T) {
	registry := commands.NewRegistry(commands.Scope{Kind: commands.ScopeGlobal}, []commands.Definition{
		{Name: "help", Description: "help"},
		{Name: "ticket", Description: "hidden", Hidden: true},
	})
	menu, err := components.BuildHelpMenu(registry)
	if err != nil {
		t.Fatalf("build help menu: %v", err)
	}
	if len(menu.Options) != 1 || menu.Options[0].Value != "help" {
		t.Fatalf("options = %#v", menu.Options)
	}
}
