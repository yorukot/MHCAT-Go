package interactions

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
)

var ErrInvalidCommandOption = errors.New("invalid command option")

type CommandOptionType string

const (
	CommandOptionString          CommandOptionType = "string"
	CommandOptionInteger         CommandOptionType = "integer"
	CommandOptionNumber          CommandOptionType = "number"
	CommandOptionBoolean         CommandOptionType = "boolean"
	CommandOptionUser            CommandOptionType = "user"
	CommandOptionChannel         CommandOptionType = "channel"
	CommandOptionRole            CommandOptionType = "role"
	CommandOptionMentionable     CommandOptionType = "mentionable"
	CommandOptionSubcommand      CommandOptionType = "subcommand"
	CommandOptionSubcommandGroup CommandOptionType = "subcommand_group"
)

type CommandOption struct {
	Name            string
	Type            CommandOptionType
	Value           any
	Options         []CommandOption
	ChannelName     string
	ChannelType     int
	ChannelParentID string
	UserName        string
}

type CommandOptionValue struct {
	Type            CommandOptionType
	String          string
	Int             int64
	Float           float64
	Bool            bool
	ChannelName     string
	ChannelType     int
	ChannelParentID string
	UserName        string
}

type ParsedCommandOptions struct {
	SubcommandGroup string
	Subcommand      string
	Options         map[string]string
	Values          map[string]CommandOptionValue
}

func ParseCommandOptions(options []CommandOption) (ParsedCommandOptions, error) {
	parsed := ParsedCommandOptions{
		Options: map[string]string{},
		Values:  map[string]CommandOptionValue{},
	}
	for _, option := range options {
		if option.Name == "" {
			return parsed, fmt.Errorf("%w: option name is required", ErrInvalidCommandOption)
		}
		switch option.Type {
		case CommandOptionSubcommand:
			if parsed.Subcommand != "" {
				return parsed, fmt.Errorf("%w: duplicate subcommand", ErrInvalidCommandOption)
			}
			parsed.Subcommand = option.Name
			if err := parseLeafOptions(option.Options, parsed); err != nil {
				return parsed, err
			}
		case CommandOptionSubcommandGroup:
			if parsed.SubcommandGroup != "" {
				return parsed, fmt.Errorf("%w: duplicate subcommand group", ErrInvalidCommandOption)
			}
			parsed.SubcommandGroup = option.Name
			for _, child := range option.Options {
				if child.Type != CommandOptionSubcommand {
					return parsed, fmt.Errorf("%w: subcommand group child must be subcommand", ErrInvalidCommandOption)
				}
				if parsed.Subcommand != "" {
					return parsed, fmt.Errorf("%w: duplicate subcommand", ErrInvalidCommandOption)
				}
				parsed.Subcommand = child.Name
				if err := parseLeafOptions(child.Options, parsed); err != nil {
					return parsed, err
				}
			}
		default:
			if err := parseLeafOption(option, parsed); err != nil {
				return parsed, err
			}
		}
	}
	return parsed, nil
}

func parseLeafOptions(options []CommandOption, parsed ParsedCommandOptions) error {
	for _, option := range options {
		if err := parseLeafOption(option, parsed); err != nil {
			return err
		}
	}
	return nil
}

func parseLeafOption(option CommandOption, parsed ParsedCommandOptions) error {
	if option.Name == "" {
		return fmt.Errorf("%w: option name is required", ErrInvalidCommandOption)
	}
	if _, exists := parsed.Values[option.Name]; exists {
		return fmt.Errorf("%w: duplicate option %q", ErrInvalidCommandOption, option.Name)
	}
	value := CommandOptionValue{Type: option.Type}
	switch option.Type {
	case CommandOptionString:
		value.String = fmt.Sprint(option.Value)
		parsed.Options[option.Name] = value.String
	case CommandOptionInteger:
		integer, err := integerValue(option.Value)
		if err != nil {
			return fmt.Errorf("%w: %s", ErrInvalidCommandOption, option.Name)
		}
		value.Int = integer
		value.String = strconv.FormatInt(integer, 10)
		parsed.Options[option.Name] = value.String
	case CommandOptionBoolean:
		boolean, ok := option.Value.(bool)
		if !ok {
			return fmt.Errorf("%w: %s", ErrInvalidCommandOption, option.Name)
		}
		value.Bool = boolean
		value.String = strconv.FormatBool(boolean)
		parsed.Options[option.Name] = value.String
	case CommandOptionNumber:
		number, err := numberValue(option.Value)
		if err != nil {
			return fmt.Errorf("%w: %s", ErrInvalidCommandOption, option.Name)
		}
		value.Float = number
		value.String = strconv.FormatFloat(number, 'f', -1, 64)
		parsed.Options[option.Name] = value.String
	case CommandOptionUser, CommandOptionChannel, CommandOptionRole, CommandOptionMentionable:
		value.String = fmt.Sprint(option.Value)
		if value.String == "" {
			return fmt.Errorf("%w: %s", ErrInvalidCommandOption, option.Name)
		}
		if option.Type == CommandOptionChannel {
			value.ChannelName = option.ChannelName
			value.ChannelType = option.ChannelType
			value.ChannelParentID = option.ChannelParentID
		}
		if option.Type == CommandOptionUser {
			value.UserName = option.UserName
		}
		parsed.Options[option.Name] = value.String
	default:
		return fmt.Errorf("%w: unsupported option type %q", ErrInvalidCommandOption, option.Type)
	}
	parsed.Values[option.Name] = value
	return nil
}

func integerValue(value any) (int64, error) {
	switch typed := value.(type) {
	case int:
		return int64(typed), nil
	case int64:
		return typed, nil
	case int32:
		return int64(typed), nil
	case float64:
		return int64(typed), nil
	case json.Number:
		return typed.Int64()
	default:
		return 0, fmt.Errorf("unsupported integer value")
	}
}

func numberValue(value any) (float64, error) {
	switch typed := value.(type) {
	case int:
		return float64(typed), nil
	case int64:
		return float64(typed), nil
	case int32:
		return float64(typed), nil
	case float64:
		return typed, nil
	case float32:
		return float64(typed), nil
	case json.Number:
		return typed.Float64()
	default:
		return 0, fmt.Errorf("unsupported number value")
	}
}
