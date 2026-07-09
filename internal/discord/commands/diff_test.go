package commands_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/commands"
)

func TestDiffCreatesNewCommand(t *testing.T) {
	registry := registryWith(definition("ping", "Ping"))
	plan, err := commands.Diff(registry, nil, commands.DiffOptions{Scope: registry.Scope})
	if err != nil {
		t.Fatalf("diff: %v", err)
	}
	assertOperation(t, plan, commands.OperationCreate, "ping")
}

func TestDiffUnchangedCommand(t *testing.T) {
	registry := registryWith(definition("ping", "Ping"))
	remote := []commands.RemoteCommand{{ID: "remote-1", Definition: definition("ping", "Ping"), Owned: true}}
	plan, err := commands.Diff(registry, remote, commands.DiffOptions{Scope: registry.Scope})
	if err != nil {
		t.Fatalf("diff: %v", err)
	}
	assertOperation(t, plan, commands.OperationUnchanged, "ping")
}

func TestDiffUpdatesChangedCommand(t *testing.T) {
	registry := registryWith(definition("ping", "New description"))
	remote := []commands.RemoteCommand{{ID: "remote-1", Definition: definition("ping", "Old description"), Owned: true}}
	plan, err := commands.Diff(registry, remote, commands.DiffOptions{Scope: registry.Scope})
	if err != nil {
		t.Fatalf("diff: %v", err)
	}
	assertOperation(t, plan, commands.OperationUpdate, "ping")
}

func TestDiffDeleteRemoteCommandSkippedByDefault(t *testing.T) {
	registry := registryWith()
	remote := []commands.RemoteCommand{{ID: "remote-1", Definition: definition("old", "Old"), Owned: true}}
	plan, err := commands.Diff(registry, remote, commands.DiffOptions{Scope: registry.Scope})
	if err != nil {
		t.Fatalf("diff: %v", err)
	}
	assertOperation(t, plan, commands.OperationSkipped, "old")
}

func TestDiffDeleteRemoteCommandOnlyWhenAllowed(t *testing.T) {
	registry := registryWith()
	remote := []commands.RemoteCommand{{ID: "remote-1", Definition: definition("old", "Old"), Owned: true}}
	plan, err := commands.Diff(registry, remote, commands.DiffOptions{Scope: registry.Scope, AllowDelete: true})
	if err != nil {
		t.Fatalf("diff: %v", err)
	}
	assertOperation(t, plan, commands.OperationDelete, "old")
}

func TestDiffUnknownRemoteCommandNotDeletedByDefault(t *testing.T) {
	registry := registryWith()
	remote := []commands.RemoteCommand{{ID: "remote-1", Definition: definition("unknown", "Unknown"), Owned: false}}
	plan, err := commands.Diff(registry, remote, commands.DiffOptions{Scope: registry.Scope, AllowDelete: true})
	if err != nil {
		t.Fatalf("diff: %v", err)
	}
	assertOperation(t, plan, commands.OperationSkipped, "unknown")
}

func TestStableHashIgnoresLocalOnlyFields(t *testing.T) {
	plain := definition("ping", "Ping")
	local := plain
	local.Hidden = true
	local.Internal = true
	local.DocsURL = "https://example.invalid/docs"
	if commands.StableHash(plain) != commands.StableHash(local) {
		t.Fatal("stable hash should ignore local-only fields")
	}
}

func TestStableHashIgnoresRemoteOnlyFields(t *testing.T) {
	left := commands.RemoteCommand{
		ID:            "one",
		ApplicationID: "app-one",
		GuildID:       "guild-one",
		Version:       "version-one",
		Definition:    definition("ping", "Ping"),
	}
	right := commands.RemoteCommand{
		ID:            "two",
		ApplicationID: "app-two",
		GuildID:       "guild-two",
		Version:       "version-two",
		Definition:    definition("ping", "Ping"),
	}
	if commands.StableHash(left.Definition) != commands.StableHash(right.Definition) {
		t.Fatal("stable hash should ignore remote-only fields")
	}
}

func TestFormatPlanDeterministic(t *testing.T) {
	registry := registryWith(definition("zeta", "Zeta"), definition("alpha", "Alpha"))
	plan, err := commands.Diff(registry, nil, commands.DiffOptions{Scope: registry.Scope})
	if err != nil {
		t.Fatalf("diff: %v", err)
	}
	var first bytes.Buffer
	var second bytes.Buffer
	if err := commands.FormatPlan(&first, plan, "json"); err != nil {
		t.Fatalf("format first: %v", err)
	}
	if err := commands.FormatPlan(&second, plan, "json"); err != nil {
		t.Fatalf("format second: %v", err)
	}
	if first.String() != second.String() {
		t.Fatalf("plan output was not deterministic:\n%s\n---\n%s", first.String(), second.String())
	}
	if !strings.Contains(first.String(), `"command_name": "alpha"`) {
		t.Fatalf("expected formatted plan to include alpha: %s", first.String())
	}
}

func TestBulkOverwritePlanMarkedDangerous(t *testing.T) {
	registry := registryWith(definition("ping", "Ping"))
	plan, err := commands.BulkOverwritePlan(registry, commands.DiffOptions{Scope: registry.Scope, AllowBulkOverwrite: true})
	if err != nil {
		t.Fatalf("bulk overwrite plan: %v", err)
	}
	if len(plan.Operations) != 1 || plan.Operations[0].Operation != commands.OperationDangerous || plan.Operations[0].Risk != commands.RiskHigh {
		t.Fatalf("unexpected bulk overwrite plan: %#v", plan)
	}
}

func registryWith(definitions ...commands.Definition) commands.Registry {
	return commands.NewRegistry(commands.Scope{Kind: commands.ScopeGuild, GuildID: "guild-1"}, definitions)
}

func definition(name, description string) commands.Definition {
	return commands.Definition{Type: commands.CommandTypeChatInput, Name: name, Description: description}
}

func assertOperation(t *testing.T, plan commands.Plan, operation commands.Operation, name string) {
	t.Helper()
	for _, planned := range plan.Operations {
		if planned.Operation == operation && planned.CommandName == name {
			return
		}
	}
	t.Fatalf("operation %s for %s not found in %#v", operation, name, plan.Operations)
}
