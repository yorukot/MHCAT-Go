package interactions_test

import (
	"errors"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/interactions"
)

func TestParseCommandOptionsStringIntegerBoolean(t *testing.T) {
	parsed, err := interactions.ParseCommandOptions([]interactions.CommandOption{
		{Name: "name", Type: interactions.CommandOptionString, Value: "ping"},
		{Name: "count", Type: interactions.CommandOptionInteger, Value: int64(2)},
		{Name: "verbose", Type: interactions.CommandOptionBoolean, Value: true},
	})
	if err != nil {
		t.Fatalf("parse options: %v", err)
	}
	if parsed.Options["name"] != "ping" || parsed.Values["count"].Int != 2 || !parsed.Values["verbose"].Bool {
		t.Fatalf("parsed = %#v", parsed)
	}
}

func TestParseCommandOptionsUser(t *testing.T) {
	parsed, err := interactions.ParseCommandOptions([]interactions.CommandOption{
		{Name: "user", Type: interactions.CommandOptionUser, Value: "123456789012345678"},
	})
	if err != nil {
		t.Fatalf("parse options: %v", err)
	}
	if parsed.Options["user"] != "123456789012345678" || parsed.Values["user"].String != "123456789012345678" {
		t.Fatalf("parsed = %#v", parsed)
	}
}

func TestParseCommandOptionsDiscordIDsAndNumber(t *testing.T) {
	parsed, err := interactions.ParseCommandOptions([]interactions.CommandOption{
		{Name: "channel", Type: interactions.CommandOptionChannel, Value: "channel-1"},
		{Name: "role", Type: interactions.CommandOptionRole, Value: "role-1"},
		{Name: "mentionable", Type: interactions.CommandOptionMentionable, Value: "user-or-role-1"},
		{Name: "amount", Type: interactions.CommandOptionNumber, Value: 12.5},
	})
	if err != nil {
		t.Fatalf("parse options: %v", err)
	}
	if parsed.Options["channel"] != "channel-1" || parsed.Options["role"] != "role-1" || parsed.Options["mentionable"] != "user-or-role-1" {
		t.Fatalf("parsed ids = %#v", parsed.Options)
	}
	if parsed.Values["amount"].Float != 12.5 || parsed.Options["amount"] != "12.5" {
		t.Fatalf("parsed number = %#v", parsed.Values["amount"])
	}
}

func TestParseCommandOptionsSubcommand(t *testing.T) {
	parsed, err := interactions.ParseCommandOptions([]interactions.CommandOption{
		{
			Name: "bot",
			Type: interactions.CommandOptionSubcommand,
			Options: []interactions.CommandOption{
				{Name: "detail", Type: interactions.CommandOptionString, Value: "basic"},
			},
		},
	})
	if err != nil {
		t.Fatalf("parse options: %v", err)
	}
	if parsed.Subcommand != "bot" || parsed.Options["detail"] != "basic" {
		t.Fatalf("parsed = %#v", parsed)
	}
}

func TestParseCommandOptionsSubcommandGroup(t *testing.T) {
	parsed, err := interactions.ParseCommandOptions([]interactions.CommandOption{
		{
			Name: "admin",
			Type: interactions.CommandOptionSubcommandGroup,
			Options: []interactions.CommandOption{
				{Name: "status", Type: interactions.CommandOptionSubcommand},
			},
		},
	})
	if err != nil {
		t.Fatalf("parse options: %v", err)
	}
	if parsed.SubcommandGroup != "admin" || parsed.Subcommand != "status" {
		t.Fatalf("parsed = %#v", parsed)
	}
}

func TestParseCommandOptionsMissingOptional(t *testing.T) {
	parsed, err := interactions.ParseCommandOptions(nil)
	if err != nil {
		t.Fatalf("parse options: %v", err)
	}
	if parsed.Subcommand != "" || len(parsed.Options) != 0 {
		t.Fatalf("parsed = %#v", parsed)
	}
}

func TestParseCommandOptionsDuplicateOptionFails(t *testing.T) {
	_, err := interactions.ParseCommandOptions([]interactions.CommandOption{
		{Name: "name", Type: interactions.CommandOptionString, Value: "one"},
		{Name: "name", Type: interactions.CommandOptionString, Value: "two"},
	})
	if !errors.Is(err, interactions.ErrInvalidCommandOption) {
		t.Fatalf("expected ErrInvalidCommandOption, got %v", err)
	}
}

func TestParseCommandOptionsMalformedDoesNotPanic(t *testing.T) {
	_, err := interactions.ParseCommandOptions([]interactions.CommandOption{
		{Name: "count", Type: interactions.CommandOptionInteger, Value: "not-int"},
	})
	if !errors.Is(err, interactions.ErrInvalidCommandOption) {
		t.Fatalf("expected ErrInvalidCommandOption, got %v", err)
	}
}
