package discordgo

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	dgo "github.com/bwmarrin/discordgo"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/commands"
)

func TestCommandSyncClientCRUDUsesScopedRESTEndpoints(t *testing.T) {
	client, err := NewCommandSyncClient("test-token", "app-1")
	if err != nil {
		t.Fatalf("new command sync client: %v", err)
	}
	type recordedRequest struct {
		method string
		path   string
		body   string
	}
	var requests []recordedRequest
	client.session.Client.Transport = responderRoundTripFunc(func(request *http.Request) (*http.Response, error) {
		body, err := io.ReadAll(request.Body)
		if err != nil {
			t.Fatalf("read request body: %v", err)
		}
		requests = append(requests, recordedRequest{method: request.Method, path: request.URL.Path, body: string(body)})
		var responseBody string
		switch request.Method {
		case http.MethodGet:
			responseBody = `[{"id":"remote-z","application_id":"app-1","guild_id":"guild-1","version":"1","type":1,"name":"zeta","description":"Z"},{"id":"remote-a","application_id":"app-1","guild_id":"guild-1","version":"1","type":1,"name":"alpha","description":"A"}]`
		case http.MethodPost:
			responseBody = `{"id":"created","application_id":"app-1","guild_id":"guild-1","version":"2","type":1,"name":"localized","description":"Localized"}`
		case http.MethodPatch:
			responseBody = `{"id":"remote-1","application_id":"app-1","guild_id":"guild-1","version":"3","type":1,"name":"localized","description":"Localized"}`
		case http.MethodDelete:
			return responderHTTPResponse(request, http.StatusNoContent, ""), nil
		case http.MethodPut:
			responseBody = `[{"id":"bulk-z","application_id":"app-1","guild_id":"guild-1","version":"4","type":1,"name":"zeta","description":"Z"},{"id":"bulk-a","application_id":"app-1","guild_id":"guild-1","version":"4","type":1,"name":"alpha","description":"A"}]`
		default:
			t.Fatalf("unexpected command sync method %s", request.Method)
		}
		return responderHTTPResponse(request, http.StatusOK, responseBody), nil
	})

	scope := commands.Scope{Kind: commands.ScopeGuild, GuildID: "guild-1"}
	remote, err := client.ListCommands(context.Background(), scope)
	if err != nil {
		t.Fatalf("list commands: %v", err)
	}
	if len(remote) != 2 || remote[0].Definition.Name != "alpha" || remote[1].Definition.Name != "zeta" {
		t.Fatalf("listed commands = %#v", remote)
	}
	definition := localizedCommandDefinition()
	created, err := client.CreateCommand(context.Background(), scope, definition)
	if err != nil || created.ID != "created" {
		t.Fatalf("create command: remote=%#v err=%v", created, err)
	}
	updated, err := client.UpdateCommand(context.Background(), scope, "remote-1", definition)
	if err != nil || updated.Version != "3" {
		t.Fatalf("update command: remote=%#v err=%v", updated, err)
	}
	if err := client.DeleteCommand(context.Background(), scope, "remote-1"); err != nil {
		t.Fatalf("delete command: %v", err)
	}
	bulk, err := client.BulkOverwriteCommands(context.Background(), scope, []commands.Definition{
		{Type: commands.CommandTypeChatInput, Name: "zeta", Description: "Z"},
		{Type: commands.CommandTypeChatInput, Name: "alpha", Description: "A"},
	})
	if err != nil {
		t.Fatalf("bulk overwrite commands: %v", err)
	}
	if len(bulk) != 2 || bulk[0].Definition.Name != "alpha" || bulk[1].Definition.Name != "zeta" {
		t.Fatalf("bulk commands = %#v", bulk)
	}

	want := []recordedRequest{
		{method: http.MethodGet, path: "/api/v9/applications/app-1/guilds/guild-1/commands"},
		{method: http.MethodPost, path: "/api/v9/applications/app-1/guilds/guild-1/commands"},
		{method: http.MethodPatch, path: "/api/v9/applications/app-1/guilds/guild-1/commands/remote-1"},
		{method: http.MethodDelete, path: "/api/v9/applications/app-1/guilds/guild-1/commands/remote-1"},
		{method: http.MethodPut, path: "/api/v9/applications/app-1/guilds/guild-1/commands"},
	}
	if len(requests) != len(want) {
		t.Fatalf("requests = %#v", requests)
	}
	for index := range want {
		if requests[index].method != want[index].method || requests[index].path != want[index].path {
			t.Fatalf("request[%d] = %#v, want %#v", index, requests[index], want[index])
		}
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(requests[1].body), &payload); err != nil {
		t.Fatalf("decode create payload: %v", err)
	}
	if !strings.Contains(requests[1].body, `"name_localizations":{"zh-TW":"在地化"}`) ||
		!strings.Contains(requests[1].body, `"description_localizations":{"zh-TW":"在地化說明"}`) ||
		payload["default_member_permissions"] != "8" || payload["nsfw"] != true {
		t.Fatalf("create payload = %s", requests[1].body)
	}
}

func TestCommandSyncClientPropagatesCancellationToEveryRequest(t *testing.T) {
	tests := []struct {
		name string
		call func(context.Context, *CommandSyncClient) error
	}{
		{name: "list", call: func(ctx context.Context, client *CommandSyncClient) error {
			_, err := client.ListCommands(ctx, commands.Scope{})
			return err
		}},
		{name: "create", call: func(ctx context.Context, client *CommandSyncClient) error {
			_, err := client.CreateCommand(ctx, commands.Scope{}, commands.Definition{Name: "test", Description: "Test"})
			return err
		}},
		{name: "update", call: func(ctx context.Context, client *CommandSyncClient) error {
			_, err := client.UpdateCommand(ctx, commands.Scope{}, "remote-1", commands.Definition{Name: "test", Description: "Test"})
			return err
		}},
		{name: "delete", call: func(ctx context.Context, client *CommandSyncClient) error {
			return client.DeleteCommand(ctx, commands.Scope{}, "remote-1")
		}},
		{name: "bulk overwrite", call: func(ctx context.Context, client *CommandSyncClient) error {
			_, err := client.BulkOverwriteCommands(ctx, commands.Scope{}, []commands.Definition{{Name: "test", Description: "Test"}})
			return err
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client, err := NewCommandSyncClient("test-token", "app-1")
			if err != nil {
				t.Fatalf("new command sync client: %v", err)
			}
			started := make(chan struct{})
			client.session.Client.Transport = responderRoundTripFunc(func(request *http.Request) (*http.Response, error) {
				close(started)
				<-request.Context().Done()
				return nil, request.Context().Err()
			})
			ctx, cancel := context.WithCancel(context.Background())
			done := make(chan error, 1)
			go func() { done <- test.call(ctx, client) }()
			select {
			case <-started:
			case <-time.After(time.Second):
				t.Fatal("request did not reach transport")
			}
			cancel()
			select {
			case err := <-done:
				if !errors.Is(err, context.Canceled) {
					t.Fatalf("request error = %v", err)
				}
			case <-time.After(time.Second):
				t.Fatal("request did not stop after context cancellation")
			}
		})
	}
}

func TestCommandSyncClientRejectsInvalidConfigurationAndEmptyResponses(t *testing.T) {
	if _, err := NewCommandSyncClient("", "app-1"); !errors.Is(err, errCommandSyncClientNotConfigured) {
		t.Fatalf("empty token error = %v", err)
	}
	if _, err := NewCommandSyncClient("token", " "); !errors.Is(err, errCommandSyncClientNotConfigured) {
		t.Fatalf("empty application error = %v", err)
	}
	client, err := NewCommandSyncClient(" token ", " app-1 ")
	if err != nil {
		t.Fatalf("new trimmed command sync client: %v", err)
	}
	if client.session.Token != "Bot token" || client.applicationID != "app-1" {
		t.Fatalf("trimmed client token=%q application=%q", client.session.Token, client.applicationID)
	}
	var nilClient *CommandSyncClient
	if _, err := nilClient.ListCommands(context.Background(), commands.Scope{}); !errors.Is(err, errCommandSyncClientNotConfigured) {
		t.Fatalf("nil client error = %v", err)
	}
	if err := (&CommandSyncClient{}).DeleteCommand(context.Background(), commands.Scope{}, "remote-1"); !errors.Is(err, errCommandSyncClientNotConfigured) {
		t.Fatalf("zero-value client error = %v", err)
	}
	if _, err := client.ListCommands(nil, commands.Scope{}); !errors.Is(err, errCommandSyncClientNotConfigured) {
		t.Fatalf("nil context error = %v", err)
	}

	client.session.Client.Transport = responderRoundTripFunc(func(request *http.Request) (*http.Response, error) {
		body := "null"
		if request.Method == http.MethodGet || request.Method == http.MethodPut {
			body = `[{"id":"valid","type":1,"name":"valid","description":"Valid"},null]`
		}
		return responderHTTPResponse(request, http.StatusOK, body), nil
	})
	listed, err := client.ListCommands(context.Background(), commands.Scope{})
	if err != nil || len(listed) != 1 || listed[0].ID != "valid" {
		t.Fatalf("list with null command: remote=%#v err=%v", listed, err)
	}
	if _, err := client.CreateCommand(context.Background(), commands.Scope{}, commands.Definition{}); err == nil || !strings.Contains(err.Error(), "empty response") {
		t.Fatalf("empty create response error = %v", err)
	}
	if _, err := client.UpdateCommand(context.Background(), commands.Scope{}, "remote-1", commands.Definition{}); err == nil || !strings.Contains(err.Error(), "empty response") {
		t.Fatalf("empty update response error = %v", err)
	}
	bulk, err := client.BulkOverwriteCommands(context.Background(), commands.Scope{}, nil)
	if err != nil || len(bulk) != 1 || bulk[0].ID != "valid" {
		t.Fatalf("bulk with null command: remote=%#v err=%v", bulk, err)
	}
}

func TestCommandSyncConversionsSkipEmptyDiscordOptionsAndChoices(t *testing.T) {
	remote := fromDiscordCommand(&dgo.ApplicationCommand{
		Type:        dgo.ChatApplicationCommand,
		Name:        "test",
		Description: "Test",
		Options: []*dgo.ApplicationCommandOption{
			nil,
			{
				Type:        dgo.ApplicationCommandOptionString,
				Name:        "choice",
				Description: "Choice",
				Choices: []*dgo.ApplicationCommandOptionChoice{
					nil,
					{Name: "Yes", Value: "yes"},
				},
			},
		},
	})
	if len(remote.Definition.Options) != 1 || len(remote.Definition.Options[0].Choices) != 1 || remote.Definition.Options[0].Choices[0].Value != "yes" {
		t.Fatalf("converted definition = %#v", remote.Definition)
	}
}

func localizedCommandDefinition() commands.Definition {
	permissions := "8"
	return commands.Definition{
		Type:                     commands.CommandTypeChatInput,
		Name:                     "localized",
		Description:              "Localized",
		NameLocalizations:        map[string]string{"zh-TW": "在地化"},
		DescriptionLocalizations: map[string]string{"zh-TW": "在地化說明"},
		DefaultMemberPermissions: &permissions,
		Contexts:                 []int{0},
		IntegrationTypes:         []int{0},
		NSFW:                     true,
		Options: []commands.Option{{
			Type:                     commands.OptionTypeString,
			Name:                     "choice",
			Description:              "Choice",
			NameLocalizations:        map[string]string{"zh-TW": "選項"},
			DescriptionLocalizations: map[string]string{"zh-TW": "選項說明"},
			Choices: []commands.Choice{{
				Name:              "Yes",
				NameLocalizations: map[string]string{"zh-TW": "是"},
				Value:             "yes",
			}},
		}},
	}
}

func TestCommandOptionChoicesAndChannelTypesRoundTrip(t *testing.T) {
	definition := commands.Definition{
		Type:        commands.CommandTypeChatInput,
		Name:        "translate",
		Description: "Translate",
		Options: []commands.Option{
			{
				Type:        commands.OptionTypeString,
				Name:        "target",
				Description: "language",
				Choices: []commands.Choice{
					{Name: "繁體中文", Value: "zh-TW"},
					{Name: "English", Value: "en"},
				},
			},
			{
				Type:         commands.OptionTypeChannel,
				Name:         "channel",
				Description:  "channel",
				ChannelTypes: []int{int(dgo.ChannelTypeGuildText), int(dgo.ChannelTypeGuildNews)},
			},
		},
	}
	discordCommand := toDiscordCommand(definition)
	if len(discordCommand.Options) != 2 || len(discordCommand.Options[0].Choices) != 2 {
		t.Fatalf("discord options = %#v", discordCommand.Options)
	}
	if got := discordCommand.Options[1].ChannelTypes; len(got) != 2 || got[0] != dgo.ChannelTypeGuildText || got[1] != dgo.ChannelTypeGuildNews {
		t.Fatalf("channel types = %#v", got)
	}
	roundTrip := fromDiscordCommand(&dgo.ApplicationCommand{
		Type:        dgo.ChatApplicationCommand,
		Name:        discordCommand.Name,
		Description: discordCommand.Description,
		Options:     discordCommand.Options,
	}).Definition
	if len(roundTrip.Options[0].Choices) != 2 || roundTrip.Options[0].Choices[0].Value != "zh-TW" {
		t.Fatalf("round-trip choices = %#v", roundTrip.Options[0].Choices)
	}
	if len(roundTrip.Options[1].ChannelTypes) != 2 || roundTrip.Options[1].ChannelTypes[1] != int(dgo.ChannelTypeGuildNews) {
		t.Fatalf("round-trip channel types = %#v", roundTrip.Options[1].ChannelTypes)
	}
}
