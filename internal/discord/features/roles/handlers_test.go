package roles

import (
	"context"
	"strings"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/customid"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/events"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/interactions"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakediscord"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakemongo"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakeusage"
)

func TestReactionSetHandlerStoresConfigAndAddsReaction(t *testing.T) {
	repo := fakemongo.NewRoleSelectionRepository()
	discord := fakediscord.NewSideEffects()
	discord.AssignableRoles["guild-1/role-1"] = true
	usage := &fakeusage.Tracker{}
	module := NewModule(repo, discord, discord, discord, discord, discord, usage)
	interaction := fakediscord.SlashInteractionWithOptions(RoleReactionSetCommandName, "", map[string]string{
		"訊息url": "https://discord.com/channels/guild-1/channel-1/message-1",
		"身分組":   "role-1",
		"表情符號":  "✅",
	})
	interaction.Actor.PermissionBits = permissionManageMessages
	responder := fakediscord.NewResponder()

	if err := module.ReactionSetHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Edits) != 1 || responder.Edits[0].Embeds[0].Title != roleSelectionDoneEmoji+" | 表情符號選取身分組成功設定" {
		t.Fatalf("edits = %#v", responder.Edits)
	}
	if _, ok := repo.Reactions["guild-1/message-1/✅"]; !ok {
		t.Fatalf("reaction config not saved: %#v", repo.Reactions)
	}
	if len(discord.Reactions) != 1 || discord.Reactions[0].ChannelID != "channel-1" {
		t.Fatalf("reactions = %#v", discord.Reactions)
	}
	if len(usage.Events) != 1 || usage.Events[0].CommandName != RoleReactionSetCommandName {
		t.Fatalf("usage = %#v", usage.Events)
	}
}

func TestReactionDeleteHandlerDeletesConfig(t *testing.T) {
	repo := fakemongo.NewRoleSelectionRepository()
	repo.Reactions["guild-1/message-1/✅"] = domain.RoleReactionConfig{GuildID: "guild-1", MessageID: "message-1", React: "✅", RoleID: "role-1"}
	discord := fakediscord.NewSideEffects()
	module := NewModule(repo, discord, discord, discord, discord, discord, nil)
	interaction := fakediscord.SlashInteractionWithOptions(RoleReactionDeleteCommandName, "", map[string]string{
		"訊息url": "https://discord.com/channels/guild-1/channel-1/message-1",
		"表情符號":  "✅",
	})
	interaction.Actor.PermissionBits = permissionManageMessages
	responder := fakediscord.NewResponder()

	if err := module.ReactionDeleteHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if _, ok := repo.Reactions["guild-1/message-1/✅"]; ok {
		t.Fatalf("reaction config should be deleted")
	}
	if got := responder.Edits[0].Embeds[0].Title; got != "表情符號選取身分組成功刪除" {
		t.Fatalf("title = %q", got)
	}
}

func TestButtonSetupShowsLegacyModalAndStoresButtonConfigs(t *testing.T) {
	repo := fakemongo.NewRoleSelectionRepository()
	discord := fakediscord.NewSideEffects()
	discord.AssignableRoles["guild-1/role-1"] = true
	module := NewModuleWithIDGenerator(repo, discord, discord, discord, discord, discord, nil, func() string { return "2026070901011234" })
	interaction := fakediscord.SlashInteractionWithOptions(RoleButtonCommandName, "", map[string]string{"身分組": "role-1"})
	interaction.Actor.PermissionBits = permissionManageMessages
	responder := fakediscord.NewResponder()

	if err := module.ButtonSetupHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Modals) != 1 || responder.Modals[0].CustomID != "nal" || responder.Modals[0].Title != "領取身分系統!" {
		t.Fatalf("modals = %#v", responder.Modals)
	}
	if got := responder.Modals[0].Rows[0].Inputs[0].CustomID; got != "roleaddcontent2026070901011234" {
		t.Fatalf("input custom id = %q", got)
	}
	if _, ok := repo.Buttons["guild-1/2026070901011234add"]; !ok {
		t.Fatalf("add button config missing: %#v", repo.Buttons)
	}
	if _, ok := repo.Buttons["guild-1/2026070901011234delete"]; !ok {
		t.Fatalf("delete button config missing: %#v", repo.Buttons)
	}
}

func TestButtonModalSendsLegacyPanel(t *testing.T) {
	repo := fakemongo.NewRoleSelectionRepository()
	discord := fakediscord.NewSideEffects()
	module := NewModule(repo, discord, discord, discord, discord, discord, nil)
	interaction := fakediscord.ModalInteraction(interactions.ModalKey{})
	interaction.ChannelID = "channel-1"
	interaction.ModalFields = []customid.ModalField{{CustomID: "roleaddcontent2026070901011234", Value: "點按鈕領身分"}}
	responder := fakediscord.NewResponder()

	if err := module.ButtonModalHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(discord.Sent) != 1 {
		t.Fatalf("sent = %#v", discord.Sent)
	}
	panel := discord.Sent[0].Message
	if panel.Embeds[0].Title != "選取身分組" || panel.Embeds[0].Description != "點按鈕領身分" {
		t.Fatalf("panel = %#v", panel)
	}
	if panel.Components[0].Components[0].CustomID != "2026070901011234add" || panel.Components[0].Components[1].CustomID != "2026070901011234delete" {
		t.Fatalf("components = %#v", panel.Components)
	}
}

func TestButtonApplyAddsAndRemovesRole(t *testing.T) {
	repo := fakemongo.NewRoleSelectionRepository()
	repo.Buttons["guild-1/button-add"] = domain.RoleButtonConfig{GuildID: "guild-1", Number: "button-add", RoleID: "role-1"}
	repo.Buttons["guild-1/button-delete"] = domain.RoleButtonConfig{GuildID: "guild-1", Number: "button-delete", RoleID: "role-1"}
	discord := fakediscord.NewSideEffects()
	discord.AssignableRoles["guild-1/role-1"] = true
	module := NewModule(repo, discord, discord, discord, discord, discord, nil)

	add := fakediscord.ComponentInteractionFromID("button-add")
	add.Actor.PermissionBits = 0
	responder := fakediscord.NewResponder()
	if err := module.ButtonApplyHandler(false)(context.Background(), add, responder); err != nil {
		t.Fatalf("add: %v", err)
	}
	if len(discord.AddedRoles) != 1 || discord.AddedRoles[0].RoleID != "role-1" {
		t.Fatalf("added roles = %#v", discord.AddedRoles)
	}
	remove := fakediscord.ComponentInteractionFromID("button-delete")
	remove.Actor.RoleIDs = []string{"role-1"}
	responder = fakediscord.NewResponder()
	if err := module.ButtonApplyHandler(true)(context.Background(), remove, responder); err != nil {
		t.Fatalf("remove: %v", err)
	}
	if len(discord.RemovedRoles) != 1 || discord.RemovedRoles[0].RoleID != "role-1" {
		t.Fatalf("removed roles = %#v", discord.RemovedRoles)
	}
}

func TestReactionEventsApplyRolesAndIgnoreMissingConfig(t *testing.T) {
	repo := fakemongo.NewRoleSelectionRepository()
	repo.Reactions["guild-1/message-1/emoji-1"] = domain.RoleReactionConfig{GuildID: "guild-1", MessageID: "message-1", React: "emoji-1", RoleID: "role-1"}
	discord := fakediscord.NewSideEffects()
	module := NewModule(repo, discord, nil, discord, discord, discord, nil)

	err := module.ReactionEventHandler(false)(context.Background(), events.Event{
		Type:      events.TypeReactionAdd,
		GuildID:   "guild-1",
		MessageID: "message-1",
		UserID:    "user-1",
		Reaction:  &events.Reaction{EmojiID: "emoji-1"},
	})
	if err != nil {
		t.Fatalf("add event: %v", err)
	}
	err = module.ReactionEventHandler(true)(context.Background(), events.Event{
		Type:      events.TypeReactionRemove,
		GuildID:   "guild-1",
		MessageID: "missing",
		UserID:    "user-1",
		Reaction:  &events.Reaction{EmojiName: "✅"},
	})
	if err != nil {
		t.Fatalf("missing config should be ignored: %v", err)
	}
	if len(discord.AddedRoles) != 1 || discord.AddedRoles[0].RoleID != "role-1" {
		t.Fatalf("added roles = %#v", discord.AddedRoles)
	}
}

func TestButtonApplyAlreadyAssignedUsesLegacyError(t *testing.T) {
	repo := fakemongo.NewRoleSelectionRepository()
	repo.Buttons["guild-1/button-add"] = domain.RoleButtonConfig{GuildID: "guild-1", Number: "button-add", RoleID: "role-1"}
	discord := fakediscord.NewSideEffects()
	discord.AssignableRoles["guild-1/role-1"] = true
	module := NewModule(repo, discord, discord, discord, discord, discord, nil)
	interaction := fakediscord.ComponentInteractionFromID("button-add")
	interaction.Actor.RoleIDs = []string{"role-1"}
	responder := fakediscord.NewResponder()

	if err := module.ButtonApplyHandler(false)(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Edits) != 1 || !strings.Contains(responder.Edits[0].Embeds[0].Title, "你已經擁有身分組了") {
		t.Fatalf("edits = %#v", responder.Edits)
	}
}
