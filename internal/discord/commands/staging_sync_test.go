package commands_test

import (
	"errors"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/commands"
)

func TestValidateStagingSyncAcceptsBuiltinRegistry(t *testing.T) {
	scope := commands.Scope{Kind: commands.ScopeGuild, GuildID: "guild-1"}
	if err := commands.ValidateStagingSync(commands.BuiltinRegistry(scope), commands.StagingSyncOptions{
		Scope:            scope,
		ExpectedCommands: []string{"help", "ping", "info"},
	}); err != nil {
		t.Fatalf("validate staging sync: %v", err)
	}
}

func TestValidateStagingSyncRejectsGlobalScope(t *testing.T) {
	err := commands.ValidateStagingSync(commands.BuiltinRegistry(commands.Scope{Kind: commands.ScopeGlobal}), commands.StagingSyncOptions{
		Scope:            commands.Scope{Kind: commands.ScopeGlobal},
		ExpectedCommands: []string{"help", "ping", "info"},
	})
	if !errors.Is(err, commands.ErrUnsafeOperation) {
		t.Fatalf("expected unsafe operation, got %v", err)
	}
}

func TestValidateStagingSyncRejectsDeleteAndBulkOverwrite(t *testing.T) {
	scope := commands.Scope{Kind: commands.ScopeGuild, GuildID: "guild-1"}
	for _, opts := range []commands.StagingSyncOptions{
		{Scope: scope, ExpectedCommands: []string{"help", "ping", "info"}, AllowDelete: true},
		{Scope: scope, ExpectedCommands: []string{"help", "ping", "info"}, AllowBulkOverwrite: true},
	} {
		if err := commands.ValidateStagingSync(commands.BuiltinRegistry(scope), opts); !errors.Is(err, commands.ErrUnsafeOperation) {
			t.Fatalf("expected unsafe operation for %#v, got %v", opts, err)
		}
	}
}

func TestValidateStagingSyncRejectsUnexpectedCommand(t *testing.T) {
	scope := commands.Scope{Kind: commands.ScopeGuild, GuildID: "guild-1"}
	registry := commands.NewRegistry(scope, append(commands.BuiltinDefinitions(), commands.Definition{
		Type:        commands.CommandTypeChatInput,
		Name:        "ticket",
		Description: "Ticket",
		Ownership:   commands.ManagedOwnership("5.1", commands.ScopeGuild),
	}))
	err := commands.ValidateStagingSync(registry, commands.StagingSyncOptions{
		Scope:            scope,
		ExpectedCommands: []string{"help", "ping", "info"},
	})
	if !errors.Is(err, commands.ErrUnsafeOperation) {
		t.Fatalf("expected unsafe operation, got %v", err)
	}
}

func TestValidateStagingSyncRejectsUnmanagedCommand(t *testing.T) {
	scope := commands.Scope{Kind: commands.ScopeGuild, GuildID: "guild-1"}
	registry := commands.NewRegistry(scope, []commands.Definition{
		{Type: commands.CommandTypeChatInput, Name: "help", Description: "Help"},
		{Type: commands.CommandTypeChatInput, Name: "ping", Description: "Ping", Ownership: commands.ManagedOwnership("5.1", commands.ScopeGuild)},
		{Type: commands.CommandTypeChatInput, Name: "info", Description: "Info", Ownership: commands.ManagedOwnership("5.1", commands.ScopeGuild)},
	})
	err := commands.ValidateStagingSync(registry, commands.StagingSyncOptions{
		Scope:            scope,
		ExpectedCommands: []string{"help", "ping", "info"},
	})
	if !errors.Is(err, commands.ErrUnsafeOperation) {
		t.Fatalf("expected unsafe operation, got %v", err)
	}
}

func TestStagingDiffSkipsUnknownRemoteCommands(t *testing.T) {
	scope := commands.Scope{Kind: commands.ScopeGuild, GuildID: "guild-1"}
	plan, err := commands.Diff(commands.BuiltinRegistry(scope), []commands.RemoteCommand{
		{ID: "remote-ticket", Definition: commands.Definition{Type: commands.CommandTypeChatInput, Name: "ticket", Description: "Ticket"}},
	}, commands.DiffOptions{Scope: scope})
	if err != nil {
		t.Fatalf("diff: %v", err)
	}
	assertOperation(t, plan, commands.OperationSkipped, "ticket")
}
