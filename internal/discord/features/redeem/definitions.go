package redeem

import "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/commands"

const (
	CommandName = "兌換"
	optionCode  = "代碼"
)

func Definitions() []commands.Definition {
	return []commands.Definition{Definition()}
}

func Definition() commands.Definition {
	return commands.Definition{
		Type:        commands.CommandTypeChatInput,
		Name:        CommandName,
		Description: "兌換代碼",
		Ownership:   commands.ManagedOwnership("redeem", commands.ScopeGuild),
		Options: []commands.Option{{
			Type:        commands.OptionTypeString,
			Name:        optionCode,
			Description: "輸入您的代碼",
			Required:    true,
		}},
	}
}
