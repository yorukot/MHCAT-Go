package commands_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/commands"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakediscord"
)

func TestPlanSyncDryRunPerformsNoWrites(t *testing.T) {
	registry := registryWith(definition("ping", "Ping"))
	client := &fakediscord.CommandSyncClient{}
	plan, err := commands.PlanSync(context.Background(), client, registry, commands.SyncOptions{Scope: registry.Scope, DryRun: true})
	if err != nil {
		t.Fatalf("plan sync: %v", err)
	}
	result, err := commands.ExecutePlan(context.Background(), client, registry, plan, commands.SyncOptions{Scope: registry.Scope, DryRun: true})
	if err != nil {
		t.Fatalf("execute dry-run: %v", err)
	}
	if result.Applied || result.Writes != 0 {
		t.Fatalf("dry-run result = %#v", result)
	}
	if len(client.Created) != 0 || len(client.Updated) != 0 || len(client.Deleted) != 0 {
		t.Fatalf("dry-run performed writes: %#v", client)
	}
}

func TestExecutePlanApplyCreatesAndUpdates(t *testing.T) {
	registry := registryWith(definition("hello", "Hello"), definition("ping", "New"))
	client := &fakediscord.CommandSyncClient{
		Remote: []commands.RemoteCommand{{ID: "remote-ping", Definition: definition("ping", "Old"), Owned: true}},
	}
	plan, err := commands.PlanSync(context.Background(), client, registry, commands.SyncOptions{Scope: registry.Scope})
	if err != nil {
		t.Fatalf("plan sync: %v", err)
	}
	result, err := commands.ExecutePlan(context.Background(), client, registry, plan, commands.SyncOptions{Scope: registry.Scope, DryRun: false})
	if err != nil {
		t.Fatalf("execute apply: %v", err)
	}
	if !result.Applied || result.Writes != 2 {
		t.Fatalf("apply result = %#v", result)
	}
	if len(client.Created) != 1 || client.Created[0].Name != "hello" {
		t.Fatalf("created = %#v", client.Created)
	}
	if len(client.Updated) != 1 || client.Updated[0].RemoteID != "remote-ping" {
		t.Fatalf("updated = %#v", client.Updated)
	}
}

func TestExecutePlanDoesNotDeleteWithoutAllowDelete(t *testing.T) {
	registry := registryWith()
	client := &fakediscord.CommandSyncClient{
		Remote: []commands.RemoteCommand{{ID: "remote-old", Definition: definition("old", "Old"), Owned: true}},
	}
	plan, err := commands.PlanSync(context.Background(), client, registry, commands.SyncOptions{Scope: registry.Scope, AllowDelete: false})
	if err != nil {
		t.Fatalf("plan sync: %v", err)
	}
	result, err := commands.ExecutePlan(context.Background(), client, registry, plan, commands.SyncOptions{Scope: registry.Scope, DryRun: false, AllowDelete: false})
	if err != nil {
		t.Fatalf("execute plan: %v", err)
	}
	if result.Writes != 0 {
		t.Fatalf("expected zero writes, got %#v", result)
	}
	if len(client.Deleted) != 0 {
		t.Fatalf("delete happened without allow-delete: %#v", client.Deleted)
	}
}

func TestFormatPlanTextDoesNotIncludeTokenLikeValues(t *testing.T) {
	plan := commands.Plan{Operations: []commands.PlannedOperation{{
		Operation:   commands.OperationCreate,
		Scope:       commands.ScopeGuild,
		GuildID:     "guild-1",
		CommandType: commands.CommandTypeChatInput,
		CommandName: "ping",
		Reason:      "desired command is missing remotely",
		Risk:        commands.RiskLow,
	}}}
	var output bytes.Buffer
	if err := commands.FormatPlan(&output, plan, "text"); err != nil {
		t.Fatalf("format plan: %v", err)
	}
	if strings.Contains(output.String(), "Bot ") || strings.Contains(output.String(), "token") {
		t.Fatalf("plan output contained token-like content: %q", output.String())
	}
}
