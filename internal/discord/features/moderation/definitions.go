package moderation

import "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/commands"

const WarningHistoryCommandName = "警告紀錄"

func Definitions() []commands.Definition {
	return []commands.Definition{WarningHistoryDefinition()}
}

func WarningHistoryDefinition() commands.Definition {
	return commands.Definition{
		Type:        commands.CommandTypeChatInput,
		Name:        WarningHistoryCommandName,
		Description: "收尋一位使用者的警告",
		Ownership:   commands.ManagedOwnership("warnings", commands.ScopeGuild),
		Options: []commands.Option{{
			Type:        commands.OptionTypeUser,
			Name:        "使用者",
			Description: "要收尋的使用者!",
			Required:    true,
		}},
	}
}
