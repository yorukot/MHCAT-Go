package onboarding

import (
	"context"
	"strings"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/interactions"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakediscord"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakemongo"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakeusage"
)

func TestVerificationSetHandlerSavesConfigWithLegacySuccess(t *testing.T) {
	repo := fakemongo.NewVerificationConfigRepository()
	roles := fakediscord.NewSideEffects()
	roles.AssignableRoles["guild-1/role-1"] = true
	usage := &fakeusage.Tracker{}
	module := NewVerificationModule(repo, roles, usage)
	interaction := fakediscord.SlashInteractionWithOptions(VerificationSetCommandName, "", map[string]string{
		"身分組": "role-1",
		"改名":  "{name} | MHCAT",
	})
	interaction.Actor.PermissionBits = permissionManageMessages
	responder := fakediscord.NewResponder()

	if err := module.VerificationSetHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Defers) != 1 {
		t.Fatalf("defers = %#v", responder.Defers)
	}
	if len(responder.Edits) != 1 || len(responder.Edits[0].Embeds) != 1 {
		t.Fatalf("edits = %#v", responder.Edits)
	}
	embed := responder.Edits[0].Embeds[0]
	if embed.Title != "<a:green_tick:994529015652163614> 設置成功!" {
		t.Fatalf("embed title = %#v", embed)
	}
	if embed.Description != "<:roleplaying:985945121264635964>身分組: <@&role-1>!\n <:id:985950321975128094>改名為:{name} | MHCAT" {
		t.Fatalf("description = %q", embed.Description)
	}
	if responder.Edits[0].AllowedMentions == nil || responder.Edits[0].AllowedMentions.ParseRoles {
		t.Fatalf("mentions should be suppressed: %#v", responder.Edits[0].AllowedMentions)
	}
	saved := repo.Configs["guild-1"]
	if saved.RoleID != "role-1" || saved.RenameTemplate != "{name} | MHCAT" {
		t.Fatalf("saved = %#v", saved)
	}
	if len(usage.Events) != 1 || usage.Events[0].CommandName != VerificationSetCommandName || usage.Events[0].Feature != "verification-config" {
		t.Fatalf("usage = %#v", usage.Events)
	}
}

func TestVerificationSetHandlerDisplaysNullRenameWhenMissing(t *testing.T) {
	repo := fakemongo.NewVerificationConfigRepository()
	roles := fakediscord.NewSideEffects()
	roles.AssignableRoles["guild-1/role-1"] = true
	module := NewVerificationModule(repo, roles, nil)
	interaction := fakediscord.SlashInteractionWithOptions(VerificationSetCommandName, "", map[string]string{"身分組": "role-1"})
	interaction.Actor.PermissionBits = permissionManageMessages
	responder := fakediscord.NewResponder()

	if err := module.VerificationSetHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if !strings.Contains(responder.Edits[0].Embeds[0].Description, "改名為:null") {
		t.Fatalf("description = %q", responder.Edits[0].Embeds[0].Description)
	}
	if repo.Configs["guild-1"].RenameTemplate != "" {
		t.Fatalf("saved = %#v", repo.Configs["guild-1"])
	}
}

func TestVerificationSetHandlerRejectsPermissionAndUnassignableRole(t *testing.T) {
	repo := fakemongo.NewVerificationConfigRepository()
	roles := fakediscord.NewSideEffects()
	module := NewVerificationModule(repo, roles, nil)
	interaction := fakediscord.SlashInteractionWithOptions(VerificationSetCommandName, "", map[string]string{"身分組": "role-1"})
	responder := fakediscord.NewResponder()

	if err := module.VerificationSetHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("permission handler: %v", err)
	}
	if !strings.Contains(responder.Edits[0].Embeds[0].Title, "你需要有`訊息管理`才能使用此指令") {
		t.Fatalf("permission response = %#v", responder.Edits)
	}

	interaction.Actor.PermissionBits = permissionManageMessages
	responder = fakediscord.NewResponder()
	if err := module.VerificationSetHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("unassignable handler: %v", err)
	}
	if !strings.Contains(responder.Edits[0].Embeds[0].Title, "我沒有權限為大家增加這個身分組") {
		t.Fatalf("unassignable response = %#v", responder.Edits)
	}
}

func TestVerificationModuleRegistersOnlyVerificationRoute(t *testing.T) {
	repo := fakemongo.NewVerificationConfigRepository()
	roles := fakediscord.NewSideEffects()
	roles.AssignableRoles["guild-1/role-1"] = true
	module := NewVerificationModule(repo, roles, nil)
	if module.Name() != "verification-config" {
		t.Fatalf("module name = %q", module.Name())
	}
	router := interactions.NewRouter()
	if err := module.RegisterRoutes(router); err != nil {
		t.Fatalf("register: %v", err)
	}
	responder := fakediscord.NewResponder()
	interaction := fakediscord.SlashInteractionWithOptions(VerificationSetCommandName, "", map[string]string{"身分組": "role-1"})
	interaction.Actor.PermissionBits = permissionManageMessages
	if err := router.Handle(context.Background(), interaction, responder); err != nil {
		t.Fatalf("dispatch: %v", err)
	}
	if len(responder.Edits) != 1 {
		t.Fatalf("edits = %#v", responder.Edits)
	}
}
