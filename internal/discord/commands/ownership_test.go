package commands_test

import (
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/commands"
)

func TestManagedOwnershipScope(t *testing.T) {
	definition := commands.Definition{
		Type:        commands.CommandTypeChatInput,
		Name:        "ping",
		Description: "Ping",
		Ownership:   commands.ManagedOwnership("5.1", commands.ScopeGuild),
	}
	if !commands.IsManagedForScope(definition, commands.Scope{Kind: commands.ScopeGuild, GuildID: "guild-1"}) {
		t.Fatal("expected command to be managed for guild scope")
	}
	if commands.IsManagedForScope(definition, commands.Scope{Kind: commands.ScopeGlobal}) {
		t.Fatal("command must not be managed for global scope")
	}
}

func TestOwnershipMetadataStrippedFromEnabledDefinitionAndHash(t *testing.T) {
	plain := commands.Definition{Type: commands.CommandTypeChatInput, Name: "ping", Description: "Ping"}
	owned := plain
	owned.Ownership = commands.ManagedOwnership("5.1", commands.ScopeGuild)

	registry := commands.NewRegistry(commands.Scope{Kind: commands.ScopeGuild, GuildID: "guild-1"}, []commands.Definition{owned})
	enabled := commands.EnabledDefinitions(registry)
	if len(enabled) != 1 {
		t.Fatalf("enabled = %#v", enabled)
	}
	if enabled[0].Ownership != nil {
		t.Fatalf("ownership leaked into enabled definition: %#v", enabled[0].Ownership)
	}
	if commands.StableHash(plain) != commands.StableHash(owned) {
		t.Fatal("stable hash should ignore ownership metadata")
	}
}
