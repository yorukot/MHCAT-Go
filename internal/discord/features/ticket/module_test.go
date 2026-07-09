package ticket

import (
	"context"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/commands"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/customid"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/interactions"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakediscord"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakemongo"
)

func TestModuleCommandsValidate(t *testing.T) {
	module := NewModule(fakemongo.NewTicketConfigRepository(), nil)
	registry := commands.NewRegistry(commands.Scope{Kind: commands.ScopeGuild, GuildID: "guild-1"}, module.Commands())
	if err := commands.ValidateRegistry(registry); err != nil {
		t.Fatalf("ticket command registry validation failed: %v", err)
	}

	got := map[string]commands.Definition{}
	for _, definition := range registry.Commands {
		got[definition.Name] = definition
	}
	setup := got["私人頻道設置"]
	if setup.Name == "" || setup.DefaultMemberPermissions == nil || *setup.DefaultMemberPermissions != manageMessagesPermission {
		t.Fatalf("setup definition permissions = %#v", setup.DefaultMemberPermissions)
	}
	if len(setup.Options) != 2 || setup.Options[0].Name != "類別" || setup.Options[1].Name != "管理員身分組" {
		t.Fatalf("setup options = %#v", setup.Options)
	}
	if setup.Ownership == nil || !setup.Ownership.Managed {
		t.Fatalf("setup ownership = %#v", setup.Ownership)
	}

	deleteCommand := got["私人頻道刪除"]
	if deleteCommand.Name == "" || deleteCommand.DefaultMemberPermissions == nil || *deleteCommand.DefaultMemberPermissions != manageMessagesPermission {
		t.Fatalf("delete definition permissions = %#v", deleteCommand.DefaultMemberPermissions)
	}
}

func TestModuleRegistersRoutes(t *testing.T) {
	module := NewModule(fakemongo.NewTicketConfigRepository(), nil)
	router := interactions.NewRouter()
	router.SetCustomIDParser(interactions.DefaultCustomIDParser{})
	if err := module.RegisterRoutes(router); err != nil {
		t.Fatalf("register routes: %v", err)
	}

	responder := fakediscord.NewResponder()
	interaction := fakediscord.SlashInteractionWithOptions("私人頻道設置", "", map[string]string{
		"類別":     "123456789012345678",
		"管理員身分組": "987654321098765432",
	})
	interaction.Actor.PermissionBits = permissionManageMessages
	if err := router.Handle(context.Background(), interaction, responder); err != nil {
		t.Fatalf("route setup: %v", err)
	}
	if len(responder.Modals) != 1 {
		t.Fatalf("modals = %#v", responder.Modals)
	}
}

func TestModuleRegistersLegacyTicketPanelModalRoute(t *testing.T) {
	sideEffects := fakediscord.NewSideEffects()
	module := NewModuleWithSideEffects(fakemongo.NewTicketConfigRepository(), nil, nil, sideEffects, "")
	router := interactions.NewRouter()
	router.SetCustomIDParser(interactions.DefaultCustomIDParser{})
	if err := module.RegisterRoutes(router); err != nil {
		t.Fatalf("register routes: %v", err)
	}

	responder := fakediscord.NewResponder()
	interaction := interactions.Interaction{
		Type:      interactions.TypeModal,
		CustomID:  "nal",
		ChannelID: "panel-channel",
		Actor:     interactions.Actor{GuildID: testGuildID, UserID: "user-1"},
		ModalFields: []customid.ModalField{
			{CustomID: "ticketcolor", Value: "#00ff00"},
			{CustomID: "tickettitle", Value: "建立私人頻道"},
			{CustomID: "ticketcontent", Value: "按下按鈕創建客服頻道"},
		},
	}
	if err := router.Handle(context.Background(), interaction, responder); err != nil {
		t.Fatalf("route legacy ticket modal: %v", err)
	}
	if len(sideEffects.Sent) != 1 || sideEffects.Sent[0].Message.Components[0].Components[0].CustomID != "tic" {
		t.Fatalf("sent messages = %#v", sideEffects.Sent)
	}
}
