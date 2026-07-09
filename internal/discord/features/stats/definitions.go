package stats

import "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/commands"

const StatsQueryCommandName = "統計系統查詢"

func Definitions() []commands.Definition {
	return []commands.Definition{QueryDefinition()}
}

func QueryDefinition() commands.Definition {
	return commands.Definition{
		Type:        commands.CommandTypeChatInput,
		Name:        StatsQueryCommandName,
		Description: "查詢統計消息",
		Ownership:   commands.ManagedOwnership("stats-query", commands.ScopeGuild),
	}
}
