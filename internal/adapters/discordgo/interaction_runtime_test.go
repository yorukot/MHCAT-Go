package discordgo

import (
	"errors"
	"testing"

	dgo "github.com/bwmarrin/discordgo"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/customid"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/interactions"
	discordruntime "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/runtime"
)

func TestRuntimeInteractionSlashCommand(t *testing.T) {
	session := testSession()
	event := slashEvent("help", nil)
	event.Token = "private-token"
	interaction, responder, err := session.RuntimeInteraction(&dgo.InteractionCreate{Interaction: event})
	if err != nil {
		t.Fatalf("runtime interaction: %v", err)
	}
	if interaction.CommandName != "help" || interaction.Type != interactions.TypeSlash {
		t.Fatalf("interaction = %#v", interaction)
	}
	if interaction.ID == "private-token" || interaction.CustomID == "private-token" {
		t.Fatalf("interaction leaked token: %#v", interaction)
	}
	if responder == nil {
		t.Fatal("responder is nil")
	}
}

func TestRuntimeInteractionSubcommand(t *testing.T) {
	session := testSession()
	event := slashEvent("info", []*dgo.ApplicationCommandInteractionDataOption{
		{Name: "bot", Type: dgo.ApplicationCommandOptionSubCommand},
	})
	interaction, _, err := session.RuntimeInteraction(&dgo.InteractionCreate{Interaction: event})
	if err != nil {
		t.Fatalf("runtime interaction: %v", err)
	}
	if interaction.Subcommand != "bot" {
		t.Fatalf("subcommand = %q", interaction.Subcommand)
	}
}

func TestRuntimeInteractionUserOption(t *testing.T) {
	session := testSession()
	event := slashEvent("info", []*dgo.ApplicationCommandInteractionDataOption{
		{
			Name: "user",
			Type: dgo.ApplicationCommandOptionSubCommand,
			Options: []*dgo.ApplicationCommandInteractionDataOption{
				{Name: "user", Type: dgo.ApplicationCommandOptionUser, Value: "123456789012345678"},
			},
		},
	})
	interaction, _, err := session.RuntimeInteraction(&dgo.InteractionCreate{Interaction: event})
	if err != nil {
		t.Fatalf("runtime interaction: %v", err)
	}
	if interaction.Subcommand != "user" || interaction.Options["user"] != "123456789012345678" {
		t.Fatalf("interaction = %#v", interaction)
	}
}

func TestRuntimeInteractionComponent(t *testing.T) {
	session := testSession()
	event := &dgo.Interaction{
		ID:        "interaction-1",
		Type:      dgo.InteractionMessageComponent,
		GuildID:   "guild-1",
		ChannelID: "channel-1",
		Data: dgo.MessageComponentInteractionData{
			CustomID:      "mhcat:v1:help:category:overview",
			ComponentType: dgo.SelectMenuComponent,
			Values:        []string{"overview"},
		},
		Message: &dgo.Message{ID: "panel-message"},
		Member:  &dgo.Member{User: &dgo.User{ID: "user-1"}, Permissions: 8192},
	}
	interaction, _, err := session.RuntimeInteraction(&dgo.InteractionCreate{Interaction: event})
	if err != nil {
		t.Fatalf("runtime interaction: %v", err)
	}
	if interaction.CustomID != "mhcat:v1:help:category:overview" || len(interaction.Values) != 1 || interaction.MessageID != "panel-message" || interaction.Actor.PermissionBits != 8192 {
		t.Fatalf("interaction = %#v", interaction)
	}
	parsed, err := customid.ParseComponent(interaction.CustomID)
	if err != nil {
		t.Fatalf("parse component: %v", err)
	}
	if parsed.Feature != "help" {
		t.Fatalf("parsed = %#v", parsed)
	}
}

func TestRuntimeInteractionModal(t *testing.T) {
	session := testSession()
	event := &dgo.Interaction{
		ID:   "interaction-1",
		Type: dgo.InteractionModalSubmit,
		Data: dgo.ModalSubmitInteractionData{
			CustomID: "mhcat:v1:ticket:rename:state=abc123",
			Components: []dgo.MessageComponent{
				&dgo.ActionsRow{Components: []dgo.MessageComponent{
					&dgo.TextInput{CustomID: "title", Value: "new title"},
				}},
			},
		},
		User: &dgo.User{ID: "user-1"},
	}
	interaction, _, err := session.RuntimeInteraction(&dgo.InteractionCreate{Interaction: event})
	if err != nil {
		t.Fatalf("runtime interaction: %v", err)
	}
	if interaction.CustomID == "" || len(interaction.ModalFields) != 1 {
		t.Fatalf("interaction = %#v", interaction)
	}
}

func TestRuntimeInteractionUnsupportedType(t *testing.T) {
	session := testSession()
	_, _, err := session.RuntimeInteraction(&dgo.InteractionCreate{Interaction: &dgo.Interaction{Type: dgo.InteractionType(99)}})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRuntimeInteractionNilEvent(t *testing.T) {
	session := testSession()
	_, _, err := session.RuntimeInteraction(nil)
	if !errors.Is(err, discordruntime.ErrInvalidRuntimeEvent) {
		t.Fatalf("expected ErrInvalidRuntimeEvent, got %v", err)
	}
}

func testSession() *Session {
	return &Session{session: &dgo.Session{}, ready: make(chan struct{})}
}

func slashEvent(name string, options []*dgo.ApplicationCommandInteractionDataOption) *dgo.Interaction {
	return &dgo.Interaction{
		ID:        "interaction-1",
		Type:      dgo.InteractionApplicationCommand,
		GuildID:   "guild-1",
		ChannelID: "channel-1",
		Data: dgo.ApplicationCommandInteractionData{
			Name:    name,
			Options: options,
		},
		Member: &dgo.Member{User: &dgo.User{ID: "user-1"}, Roles: []string{"role-1"}},
	}
}
