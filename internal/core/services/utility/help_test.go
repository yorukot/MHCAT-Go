package utility_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/services/utility"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/commands"
)

func TestHelpOverviewIncludesImplementedCommands(t *testing.T) {
	service := utility.NewHelpService(commands.NewRegistry(commands.Scope{Kind: commands.ScopeGlobal}, []commands.Definition{
		{Name: "ticket", Description: "not implemented", Hidden: true},
		{Name: "help", Description: "使用我開始使用"},
		{Name: "ping", Description: "查看我的ping"},
	}))
	got := service.Overview()
	for _, want := range []string{"/help", "/ping"} {
		if !strings.Contains(got, want) {
			t.Fatalf("overview missing %q:\n%s", want, got)
		}
	}
	if strings.Contains(got, "/ticket") {
		t.Fatalf("overview listed hidden/unimplemented command:\n%s", got)
	}
}

func TestHelpDetailImplementedCommand(t *testing.T) {
	service := utility.NewHelpService(commands.NewRegistry(commands.Scope{Kind: commands.ScopeGlobal}, []commands.Definition{
		{Name: "ping", Description: "查看我的ping", DocsURL: "https://docsmhcat.yorukot.me/docs/ping"},
	}))
	got, err := service.Detail("ping")
	if err != nil {
		t.Fatalf("detail: %v", err)
	}
	if !strings.Contains(got, "Name: /ping") || !strings.Contains(got, "查看我的ping") {
		t.Fatalf("unexpected detail:\n%s", got)
	}
}

func TestHelpDetailUnknownCommand(t *testing.T) {
	service := utility.NewHelpService(commands.EmptyRegistry(commands.Scope{Kind: commands.ScopeGlobal}))
	_, err := service.Detail("missing")
	if !errors.Is(err, utility.ErrHelpCommandNotFound) {
		t.Fatalf("expected ErrHelpCommandNotFound, got %v", err)
	}
}

func TestHelpOutputDeterministic(t *testing.T) {
	registry := commands.NewRegistry(commands.Scope{Kind: commands.ScopeGlobal}, []commands.Definition{
		{Name: "ping", Description: "查看我的ping"},
		{Name: "help", Description: "使用我開始使用"},
	})
	service := utility.NewHelpService(registry)
	if first, second := service.Overview(), service.Overview(); first != second {
		t.Fatalf("overview not deterministic:\n%s\n%s", first, second)
	}
}
