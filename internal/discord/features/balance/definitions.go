package balance

import "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/commands"

const CommandName = "查看餘額"

func Definitions() []commands.Definition {
	return []commands.Definition{Definition()}
}

func Definition() commands.Definition {
	return commands.Definition{
		Type:        commands.CommandTypeChatInput,
		Name:        CommandName,
		Description: "查看剩餘餘額",
		Ownership:   commands.ManagedOwnership("balance-query", commands.ScopeGuild),
	}
}
