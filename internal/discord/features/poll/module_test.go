package poll

import (
	"context"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/commands"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/interactions"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakediscord"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakemongo"
)

func TestModuleCommandsValidate(t *testing.T) {
	module := NewModule(fakemongo.NewPollRepository(), nil)
	registry := commands.NewRegistry(commands.Scope{Kind: commands.ScopeGuild, GuildID: "guild-1"}, module.Commands())
	if err := commands.ValidateRegistry(registry); err != nil {
		t.Fatalf("poll command registry validation failed: %v", err)
	}
	if len(registry.Commands) != 1 {
		t.Fatalf("commands = %#v", registry.Commands)
	}
	definition := registry.Commands[0]
	if definition.Name != "投票創建" || definition.Description != "創建一個萬能的投票" {
		t.Fatalf("definition = %#v", definition)
	}
	if definition.DefaultMemberPermissions == nil || *definition.DefaultMemberPermissions != manageMessagesPermission {
		t.Fatalf("permissions = %#v", definition.DefaultMemberPermissions)
	}
	if len(definition.Options) != 2 || definition.Options[0].Name != "問題" || definition.Options[1].Name != "選項" {
		t.Fatalf("options = %#v", definition.Options)
	}
}

func TestModuleRegistersPollRoutes(t *testing.T) {
	repo := fakemongo.NewPollRepository()
	sideEffects := fakediscord.NewSideEffects()
	module := NewModuleWithSideEffects(repo, nil, sideEffects, sideEffects, nil)
	router := interactions.NewRouter()
	router.SetCustomIDParser(interactions.DefaultCustomIDParser{})
	if err := module.RegisterRoutes(router); err != nil {
		t.Fatalf("register routes: %v", err)
	}
	responder := fakediscord.NewResponder()
	interaction := pollCreateInteraction("問題", "A^B")
	if err := router.Handle(context.Background(), interaction, responder); err != nil {
		t.Fatalf("route create: %v", err)
	}
	if len(sideEffects.Sent) != 1 {
		t.Fatalf("sent = %#v", sideEffects.Sent)
	}
}
