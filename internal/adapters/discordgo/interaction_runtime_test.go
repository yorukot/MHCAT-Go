package discordgo

import (
	"errors"
	"strings"
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
	if interaction.ApplicationID != "application-1" {
		t.Fatalf("application id = %q", interaction.ApplicationID)
	}
	if interaction.ID == "private-token" || interaction.CustomID == "private-token" {
		t.Fatalf("interaction leaked token: %#v", interaction)
	}
	if responder == nil {
		t.Fatal("responder is nil")
	}
}

func TestRuntimeInteractionPreservesUserAndGuildDisplayAvatars(t *testing.T) {
	session := testSession()
	event := slashEvent("coin-related-settings", nil)
	event.Member.User.Avatar = "user-avatar"
	event.Member.Avatar = "guild-avatar"
	interaction, _, err := session.RuntimeInteraction(&dgo.InteractionCreate{Interaction: event})
	if err != nil {
		t.Fatalf("runtime interaction: %v", err)
	}
	if !strings.Contains(interaction.Actor.AvatarURL, "/avatars/user-1/user-avatar") {
		t.Fatalf("user avatar = %q", interaction.Actor.AvatarURL)
	}
	if !strings.Contains(interaction.Actor.GuildAvatarURL, "/guilds/guild-1/users/user-1/avatars/guild-avatar") {
		t.Fatalf("guild avatar = %q", interaction.Actor.GuildAvatarURL)
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
	data := event.ApplicationCommandData()
	data.Resolved = &dgo.ApplicationCommandInteractionDataResolved{Users: map[string]*dgo.User{
		"123456789012345678": {ID: "123456789012345678", Username: "Yoru"},
	}}
	event.Data = data
	interaction, _, err := session.RuntimeInteraction(&dgo.InteractionCreate{Interaction: event})
	if err != nil {
		t.Fatalf("runtime interaction: %v", err)
	}
	if interaction.Subcommand != "user" || interaction.Options["user"] != "123456789012345678" || interaction.CommandOptions["user"].UserName != "Yoru" {
		t.Fatalf("interaction = %#v", interaction)
	}
}

func TestRuntimeInteractionChannelOptionMetadata(t *testing.T) {
	session := testSession()
	event := slashEvent("語音包廂設置", []*dgo.ApplicationCommandInteractionDataOption{
		{Name: "語音頻道", Type: dgo.ApplicationCommandOptionChannel, Value: "voice-1"},
	})
	event.Data = dgo.ApplicationCommandInteractionData{
		Name: event.ApplicationCommandData().Name,
		Options: []*dgo.ApplicationCommandInteractionDataOption{
			{Name: "語音頻道", Type: dgo.ApplicationCommandOptionChannel, Value: "voice-1"},
		},
		Resolved: &dgo.ApplicationCommandInteractionDataResolved{
			Channels: map[string]*dgo.Channel{
				"voice-1": {
					ID:       "voice-1",
					Name:     "Create room",
					Type:     dgo.ChannelTypeGuildVoice,
					ParentID: "category-1",
				},
			},
		},
	}
	interaction, _, err := session.RuntimeInteraction(&dgo.InteractionCreate{Interaction: event})
	if err != nil {
		t.Fatalf("runtime interaction: %v", err)
	}
	value := interaction.CommandOptions["語音頻道"]
	if value.String != "voice-1" || value.ChannelName != "Create room" || value.ChannelType != int(dgo.ChannelTypeGuildVoice) || value.ChannelParentID != "category-1" {
		t.Fatalf("channel option = %#v", value)
	}
}

func TestRuntimeInteractionIncludesCachedBotAvatar(t *testing.T) {
	session := testSession()
	session.session.State = dgo.NewState()
	session.session.State.User = &dgo.User{ID: "bot-1", Avatar: "bot-avatar"}
	if err := session.session.State.GuildAdd(&dgo.Guild{
		ID: "guild-1",
		Roles: []*dgo.Role{
			{ID: "guild-1", Position: 0},
			{ID: "bot-color", Position: 1, Color: 0x123456},
		},
		Members: []*dgo.Member{{User: session.session.State.User, Roles: []string{"bot-color"}}},
	}); err != nil {
		t.Fatalf("seed bot display color: %v", err)
	}

	interaction, _, err := session.RuntimeInteraction(&dgo.InteractionCreate{Interaction: slashEvent("ping", nil)})
	if err != nil {
		t.Fatalf("runtime interaction: %v", err)
	}
	if interaction.BotAvatarURL == "" || !strings.Contains(interaction.BotAvatarURL, "bot-avatar") {
		t.Fatalf("bot avatar url = %q", interaction.BotAvatarURL)
	}
	if interaction.BotDisplayColor != 0x123456 {
		t.Fatalf("bot display color = %#x", interaction.BotDisplayColor)
	}
}

func TestRuntimeInteractionPopulatesActorVoiceState(t *testing.T) {
	session := testSession()
	session.session.State = dgo.NewState()
	if err := session.session.State.GuildAdd(&dgo.Guild{
		ID: "guild-1",
		VoiceStates: []*dgo.VoiceState{{
			GuildID:   "guild-1",
			UserID:    "user-1",
			ChannelID: "voice-1",
		}},
	}); err != nil {
		t.Fatalf("seed voice state: %v", err)
	}
	interaction, _, err := session.RuntimeInteraction(&dgo.InteractionCreate{Interaction: slashEvent("上鎖頻道", nil)})
	if err != nil {
		t.Fatalf("runtime interaction: %v", err)
	}
	if interaction.Actor.VoiceChannelID != "voice-1" {
		t.Fatalf("voice channel id = %q", interaction.Actor.VoiceChannelID)
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
		Message: &dgo.Message{
			ID: "panel-message",
			InteractionMetadata: &dgo.MessageInteractionMetadata{
				ID:   "request-interaction-1",
				User: &dgo.User{ID: "requester-1"},
			},
		},
		Member: &dgo.Member{User: &dgo.User{ID: "user-1"}, Permissions: 8192},
	}
	interaction, _, err := session.RuntimeInteraction(&dgo.InteractionCreate{Interaction: event})
	if err != nil {
		t.Fatalf("runtime interaction: %v", err)
	}
	if interaction.CustomID != "mhcat:v1:help:category:overview" || len(interaction.Values) != 1 || interaction.MessageID != "panel-message" || interaction.OriginalInteractionID != "request-interaction-1" || interaction.OriginalInteractionUserID != "requester-1" || interaction.Actor.PermissionBits != 8192 {
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
		AppID:     "application-1",
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
