package commands_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/commands"
)

func TestValidateValidChatInputCommand(t *testing.T) {
	registry := commands.NewRegistry(commands.Scope{Kind: commands.ScopeGlobal}, []commands.Definition{
		{Type: commands.CommandTypeChatInput, Name: "ping", Description: "Check bot latency"},
	})
	if err := commands.ValidateRegistry(registry); err != nil {
		t.Fatalf("valid chat input command failed validation: %v", err)
	}
}

func TestValidateValidUserCommand(t *testing.T) {
	registry := commands.NewRegistry(commands.Scope{Kind: commands.ScopeGlobal}, []commands.Definition{
		{Type: commands.CommandTypeUser, Name: "View Profile"},
	})
	if err := commands.ValidateRegistry(registry); err != nil {
		t.Fatalf("valid user command failed validation: %v", err)
	}
}

func TestValidateDuplicateCommandNamesFail(t *testing.T) {
	registry := commands.NewRegistry(commands.Scope{Kind: commands.ScopeGlobal}, []commands.Definition{
		{Type: commands.CommandTypeChatInput, Name: "ping", Description: "one"},
		{Type: commands.CommandTypeChatInput, Name: "ping", Description: "two"},
	})
	if err := commands.ValidateRegistry(registry); !errors.Is(err, commands.ErrInvalidRegistry) {
		t.Fatalf("expected ErrInvalidRegistry, got %v", err)
	}
}

func TestValidateInvalidChatInputNameFails(t *testing.T) {
	registry := commands.NewRegistry(commands.Scope{Kind: commands.ScopeGlobal}, []commands.Definition{
		{Type: commands.CommandTypeChatInput, Name: "Ping", Description: "invalid uppercase"},
	})
	if err := commands.ValidateRegistry(registry); !errors.Is(err, commands.ErrInvalidRegistry) {
		t.Fatalf("expected ErrInvalidRegistry, got %v", err)
	}
}

func TestValidateUnicodeLegacyOptionName(t *testing.T) {
	registry := commands.NewRegistry(commands.Scope{Kind: commands.ScopeGlobal}, []commands.Definition{
		{
			Type:        commands.CommandTypeChatInput,
			Name:        "help",
			Description: "使用我開始使用",
			Options: []commands.Option{
				{Type: commands.OptionTypeString, Name: "指令名稱", Description: "輸入指令名稱(可不輸入)!"},
			},
		},
	})
	if err := commands.ValidateRegistry(registry); err != nil {
		t.Fatalf("unicode legacy option failed validation: %v", err)
	}
}

func TestValidateUnicodeLegacyCommandNameCountsRunes(t *testing.T) {
	registry := commands.NewRegistry(commands.Scope{Kind: commands.ScopeGlobal}, []commands.Definition{
		{
			Type:        commands.CommandTypeChatInput,
			Name:        "選取身分組刪除-表情符號",
			Description: "選取身分組刪除-表情符號版(進行刪除)",
		},
	})
	if err := commands.ValidateRegistry(registry); err != nil {
		t.Fatalf("unicode legacy command name failed validation: %v", err)
	}
}

func TestValidateMissingDescriptionFailsForChatInput(t *testing.T) {
	registry := commands.NewRegistry(commands.Scope{Kind: commands.ScopeGlobal}, []commands.Definition{
		{Type: commands.CommandTypeChatInput, Name: "ping"},
	})
	if err := commands.ValidateRegistry(registry); !errors.Is(err, commands.ErrInvalidRegistry) {
		t.Fatalf("expected ErrInvalidRegistry, got %v", err)
	}
}

func TestValidateOptionsOver25Fail(t *testing.T) {
	options := make([]commands.Option, 26)
	for index := range options {
		options[index] = commands.Option{
			Type:        commands.OptionTypeString,
			Name:        fmt.Sprintf("option_%02d", index),
			Description: "option",
		}
	}
	registry := commands.NewRegistry(commands.Scope{Kind: commands.ScopeGlobal}, []commands.Definition{
		{Type: commands.CommandTypeChatInput, Name: "many", Description: "too many", Options: options},
	})
	if err := commands.ValidateRegistry(registry); !errors.Is(err, commands.ErrInvalidRegistry) {
		t.Fatalf("expected ErrInvalidRegistry, got %v", err)
	}
}

func TestValidateDuplicateOptionNamesFail(t *testing.T) {
	registry := commands.NewRegistry(commands.Scope{Kind: commands.ScopeGlobal}, []commands.Definition{
		{
			Type:        commands.CommandTypeChatInput,
			Name:        "echo",
			Description: "Echo",
			Options: []commands.Option{
				{Type: commands.OptionTypeString, Name: "value", Description: "value"},
				{Type: commands.OptionTypeString, Name: "value", Description: "duplicate"},
			},
		},
	})
	if err := commands.ValidateRegistry(registry); !errors.Is(err, commands.ErrInvalidRegistry) {
		t.Fatalf("expected ErrInvalidRegistry, got %v", err)
	}
}

func TestValidateRequiredOptionAfterOptionalFails(t *testing.T) {
	registry := commands.NewRegistry(commands.Scope{Kind: commands.ScopeGlobal}, []commands.Definition{
		{
			Type:        commands.CommandTypeChatInput,
			Name:        "echo",
			Description: "Echo",
			Options: []commands.Option{
				{Type: commands.OptionTypeString, Name: "optional", Description: "optional"},
				{Type: commands.OptionTypeString, Name: "required", Description: "required", Required: true},
			},
		},
	})
	if err := commands.ValidateRegistry(registry); !errors.Is(err, commands.ErrInvalidRegistry) {
		t.Fatalf("expected ErrInvalidRegistry, got %v", err)
	}
}

func TestValidateChoicesAndChannelTypes(t *testing.T) {
	registry := commands.NewRegistry(commands.Scope{Kind: commands.ScopeGlobal}, []commands.Definition{
		{
			Type:        commands.CommandTypeChatInput,
			Name:        "translate",
			Description: "Translate",
			Options: []commands.Option{
				{
					Type:        commands.OptionTypeString,
					Name:        "target",
					Description: "language",
					Required:    true,
					Choices: []commands.Choice{
						{Name: "繁體中文", Value: "zh-TW"},
						{Name: "English", Value: "en"},
					},
				},
				{
					Type:         commands.OptionTypeChannel,
					Name:         "channel",
					Description:  "channel",
					ChannelTypes: []int{0, 5},
				},
			},
		},
	})
	if err := commands.ValidateRegistry(registry); err != nil {
		t.Fatalf("valid choices/channel types failed validation: %v", err)
	}
}

func TestValidateChoiceOnUnsupportedTypeFails(t *testing.T) {
	registry := commands.NewRegistry(commands.Scope{Kind: commands.ScopeGlobal}, []commands.Definition{{
		Type:        commands.CommandTypeChatInput,
		Name:        "bad",
		Description: "bad",
		Options: []commands.Option{{
			Type:        commands.OptionTypeBoolean,
			Name:        "flag",
			Description: "flag",
			Choices:     []commands.Choice{{Name: "yes", Value: true}},
		}},
	}})
	if err := commands.ValidateRegistry(registry); !errors.Is(err, commands.ErrInvalidRegistry) {
		t.Fatalf("expected ErrInvalidRegistry, got %v", err)
	}
}

func TestValidateChannelTypesOnNonChannelFails(t *testing.T) {
	registry := commands.NewRegistry(commands.Scope{Kind: commands.ScopeGlobal}, []commands.Definition{{
		Type:        commands.CommandTypeChatInput,
		Name:        "bad",
		Description: "bad",
		Options: []commands.Option{{
			Type:         commands.OptionTypeString,
			Name:         "value",
			Description:  "value",
			ChannelTypes: []int{0},
		}},
	}})
	if err := commands.ValidateRegistry(registry); !errors.Is(err, commands.ErrInvalidRegistry) {
		t.Fatalf("expected ErrInvalidRegistry, got %v", err)
	}
}
