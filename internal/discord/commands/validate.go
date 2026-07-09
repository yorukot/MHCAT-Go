package commands

import (
	"errors"
	"fmt"
	"unicode"
	"unicode/utf8"
)

var (
	ErrInvalidRegistry = errors.New("invalid command registry")
)

func ValidateRegistry(registry Registry) error {
	if registry.Scope.Kind == "" {
		registry.Scope.Kind = ScopeGlobal
	}
	if registry.Scope.Kind != ScopeGlobal && registry.Scope.Kind != ScopeGuild {
		return fmt.Errorf("%w: scope must be global or guild", ErrInvalidRegistry)
	}
	if registry.Scope.Kind == ScopeGuild && registry.Scope.GuildID == "" {
		return fmt.Errorf("%w: guild scope requires guild id", ErrInvalidRegistry)
	}

	seen := map[string]struct{}{}
	for _, definition := range registry.Commands {
		if err := validateDefinition(definition); err != nil {
			return err
		}
		key := fmt.Sprintf("%s:%s:%d:%s", registry.Scope.Kind, registry.Scope.GuildID, definition.Type, definition.Name)
		if _, ok := seen[key]; ok {
			return fmt.Errorf("%w: duplicate command %q type %d", ErrInvalidRegistry, definition.Name, definition.Type)
		}
		seen[key] = struct{}{}
	}
	return nil
}

func validateDefinition(definition Definition) error {
	if definition.Type == 0 {
		definition.Type = CommandTypeChatInput
	}
	if definition.Name == "" {
		return fmt.Errorf("%w: command name is required", ErrInvalidRegistry)
	}
	if utf8.RuneCountInString(definition.Name) > 32 {
		return fmt.Errorf("%w: command name %q exceeds 32 characters", ErrInvalidRegistry, definition.Name)
	}
	switch definition.Type {
	case CommandTypeChatInput:
		if !validCommandToken(definition.Name) {
			return fmt.Errorf("%w: chat input command name %q is invalid", ErrInvalidRegistry, definition.Name)
		}
		if definition.Description == "" {
			return fmt.Errorf("%w: chat input command %q requires description", ErrInvalidRegistry, definition.Name)
		}
		if err := validateOptions(definition.Options); err != nil {
			return fmt.Errorf("command %q: %w", definition.Name, err)
		}
	case CommandTypeUser, CommandTypeMessage:
		if len(definition.Options) > 0 {
			return fmt.Errorf("%w: context command %q must not have options", ErrInvalidRegistry, definition.Name)
		}
	default:
		return fmt.Errorf("%w: command %q has unsupported type %d", ErrInvalidRegistry, definition.Name, definition.Type)
	}
	return nil
}

func validateOptions(options []Option) error {
	if len(options) > 25 {
		return fmt.Errorf("%w: option count exceeds 25", ErrInvalidRegistry)
	}
	seen := map[string]struct{}{}
	optionalSeen := false
	for _, option := range options {
		if option.Name == "" {
			return fmt.Errorf("%w: option name is required", ErrInvalidRegistry)
		}
		if !validCommandToken(option.Name) {
			return fmt.Errorf("%w: option name %q is invalid", ErrInvalidRegistry, option.Name)
		}
		if _, ok := seen[option.Name]; ok {
			return fmt.Errorf("%w: duplicate option %q", ErrInvalidRegistry, option.Name)
		}
		seen[option.Name] = struct{}{}
		if optionalSeen && option.Required {
			return fmt.Errorf("%w: required option %q appears after optional option", ErrInvalidRegistry, option.Name)
		}
		if !option.Required {
			optionalSeen = true
		}
		switch option.Type {
		case OptionTypeSubCommand, OptionTypeSubCommandGroup:
			if len(option.Choices) > 0 || len(option.ChannelTypes) > 0 {
				return fmt.Errorf("%w: subcommand option %q must not have choices or channel types", ErrInvalidRegistry, option.Name)
			}
			if err := validateOptions(option.Options); err != nil {
				return err
			}
		case OptionTypeString, OptionTypeInteger, OptionTypeBoolean, OptionTypeUser, OptionTypeChannel, OptionTypeRole, OptionTypeMentionable, OptionTypeNumber, OptionTypeAttachment:
			if len(option.Options) > 0 {
				return fmt.Errorf("%w: non-subcommand option %q must not have child options", ErrInvalidRegistry, option.Name)
			}
			if len(option.Choices) > 0 {
				if option.Type != OptionTypeString && option.Type != OptionTypeInteger && option.Type != OptionTypeNumber {
					return fmt.Errorf("%w: option %q has choices for unsupported type", ErrInvalidRegistry, option.Name)
				}
				if err := validateChoices(option.Choices); err != nil {
					return fmt.Errorf("option %q: %w", option.Name, err)
				}
			}
			if len(option.ChannelTypes) > 0 && option.Type != OptionTypeChannel {
				return fmt.Errorf("%w: option %q has channel types but is not a channel option", ErrInvalidRegistry, option.Name)
			}
		default:
			return fmt.Errorf("%w: option %q has unsupported type %d", ErrInvalidRegistry, option.Name, option.Type)
		}
	}
	return nil
}

func validateChoices(choices []Choice) error {
	if len(choices) > 25 {
		return fmt.Errorf("%w: choice count exceeds 25", ErrInvalidRegistry)
	}
	seen := map[string]struct{}{}
	for _, choice := range choices {
		if choice.Name == "" {
			return fmt.Errorf("%w: choice name is required", ErrInvalidRegistry)
		}
		if _, ok := seen[choice.Name]; ok {
			return fmt.Errorf("%w: duplicate choice %q", ErrInvalidRegistry, choice.Name)
		}
		seen[choice.Name] = struct{}{}
		if choice.Value == nil {
			return fmt.Errorf("%w: choice %q requires value", ErrInvalidRegistry, choice.Name)
		}
	}
	return nil
}

func validCommandToken(value string) bool {
	if value == "" || utf8.RuneCountInString(value) > 32 {
		return false
	}
	for _, r := range value {
		switch {
		case r >= 'a' && r <= 'z':
			continue
		case r >= '0' && r <= '9':
			continue
		case r == '_' || r == '-':
			continue
		case unicode.IsLetter(r) || unicode.IsNumber(r):
			if unicode.IsUpper(r) {
				return false
			}
			continue
		default:
			return false
		}
	}
	return true
}
