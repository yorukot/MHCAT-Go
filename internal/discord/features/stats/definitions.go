package stats

import "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/commands"

const (
	StatsQueryCommandName    = "統計系統查詢"
	StatsDeleteCommandName   = "統計系統刪除"
	manageMessagesPermission = "8192"
)

func Definitions() []commands.Definition {
	return []commands.Definition{QueryDefinition(), DeleteDefinition()}
}

func QueryDefinitions() []commands.Definition {
	return []commands.Definition{QueryDefinition()}
}

func DeleteDefinitions() []commands.Definition {
	return []commands.Definition{DeleteDefinition()}
}

func QueryDefinition() commands.Definition {
	return commands.Definition{
		Type:        commands.CommandTypeChatInput,
		Name:        StatsQueryCommandName,
		Description: "查詢統計消息",
		Ownership:   commands.ManagedOwnership("stats-query", commands.ScopeGuild),
	}
}

func DeleteDefinition() commands.Definition {
	return commands.Definition{
		Type:                     commands.CommandTypeChatInput,
		Name:                     StatsDeleteCommandName,
		Description:              "刪除統計消息",
		DefaultMemberPermissions: stringPtr(manageMessagesPermission),
		Ownership:                commands.ManagedOwnership("stats-delete", commands.ScopeGuild),
	}
}

func stringPtr(value string) *string {
	return &value
}
