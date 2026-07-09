package discordgo_test

import (
	"errors"
	"testing"

	dgo "github.com/bwmarrin/discordgo"
	adapter "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/adapters/discordgo"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/customid"
)

func TestComponentInputFromDiscordButton(t *testing.T) {
	input, err := adapter.ComponentInputFromDiscord(dgo.MessageComponentInteractionData{
		CustomID:      "mhcat:v1:ticket:close:",
		ComponentType: dgo.ButtonComponent,
	})
	if err != nil {
		t.Fatalf("component input: %v", err)
	}
	if input.CustomID != "mhcat:v1:ticket:close:" {
		t.Fatalf("custom id = %q", input.CustomID)
	}
}

func TestComponentInputFromDiscordSelectValues(t *testing.T) {
	input, err := adapter.ComponentInputFromDiscord(dgo.MessageComponentInteractionData{
		CustomID:      "mhcat:v1:poll:vote:",
		ComponentType: dgo.SelectMenuComponent,
		Values:        []string{"opt_1", "opt_2"},
	})
	if err != nil {
		t.Fatalf("component input: %v", err)
	}
	if len(input.Values) != 2 || input.Values[0] != "opt_1" || input.Values[1] != "opt_2" {
		t.Fatalf("values = %#v", input.Values)
	}
	input.Values[0] = "changed"
	again, err := adapter.ComponentInputFromDiscord(dgo.MessageComponentInteractionData{
		CustomID:      "mhcat:v1:poll:vote:",
		ComponentType: dgo.SelectMenuComponent,
		Values:        []string{"opt_1"},
	})
	if err != nil {
		t.Fatalf("component input again: %v", err)
	}
	if again.Values[0] != "opt_1" {
		t.Fatalf("values were not copied safely")
	}
}

func TestComponentInputFromDiscordUnsupportedType(t *testing.T) {
	_, err := adapter.ComponentInputFromDiscord(dgo.MessageComponentInteractionData{
		CustomID:      "mhcat:v1:ticket:close:",
		ComponentType: dgo.ComponentType(99),
	})
	if !errors.Is(err, customid.ErrUnsupportedComponent) {
		t.Fatalf("expected ErrUnsupportedComponent, got %v", err)
	}
}

func TestModalInputFromDiscord(t *testing.T) {
	input, err := adapter.ModalInputFromDiscord(dgo.ModalSubmitInteractionData{
		CustomID: "mhcat:v1:ticket:rename:state=abc123",
		Components: []dgo.MessageComponent{
			&dgo.ActionsRow{Components: []dgo.MessageComponent{
				&dgo.TextInput{CustomID: "title", Value: "new title"},
			}},
		},
	})
	if err != nil {
		t.Fatalf("modal input: %v", err)
	}
	if input.CustomID != "mhcat:v1:ticket:rename:state=abc123" {
		t.Fatalf("custom id = %q", input.CustomID)
	}
	if len(input.Fields) != 1 || input.Fields[0].CustomID != "title" || input.Fields[0].Value != "new title" {
		t.Fatalf("fields = %#v", input.Fields)
	}
}

func TestModalInputFromDiscordUnsupportedComponent(t *testing.T) {
	_, err := adapter.ModalInputFromDiscord(dgo.ModalSubmitInteractionData{
		CustomID: "mhcat:v1:ticket:rename:",
		Components: []dgo.MessageComponent{
			&dgo.TextInput{CustomID: "title", Value: "wrong level"},
		},
	})
	if !errors.Is(err, customid.ErrUnsupportedComponent) {
		t.Fatalf("expected ErrUnsupportedComponent, got %v", err)
	}
}
