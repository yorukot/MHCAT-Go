package stats

import "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/commands"

const (
	StatsQueryCommandName    = "統計系統查詢"
	StatsCreateCommandName   = "統計系統創建"
	StatsRoleCommandName     = "統計身分組人數"
	StatsDeleteCommandName   = "統計系統刪除"
	manageMessagesPermission = "8192"

	statsOptionChannelType = "統計頻道類型"
	statsOptionStat        = "統計選項"
	statsOptionRole        = "身分組"
)

func Definitions() []commands.Definition {
	return []commands.Definition{QueryDefinition(), CreateDefinition(), RoleDefinition(), DeleteDefinition()}
}

func QueryDefinitions() []commands.Definition {
	return []commands.Definition{QueryDefinition()}
}

func DeleteDefinitions() []commands.Definition {
	return []commands.Definition{DeleteDefinition()}
}

func CreateDefinitions() []commands.Definition {
	return []commands.Definition{CreateDefinition()}
}

func RoleDefinitions() []commands.Definition {
	return []commands.Definition{RoleDefinition()}
}

func QueryDefinition() commands.Definition {
	return commands.Definition{
		Type:        commands.CommandTypeChatInput,
		Name:        StatsQueryCommandName,
		Description: "查詢統計消息",
		Ownership:   commands.ManagedOwnership("stats-query", commands.ScopeGuild),
	}
}

func CreateDefinition() commands.Definition {
	return commands.Definition{
		Type:                     commands.CommandTypeChatInput,
		Name:                     StatsCreateCommandName,
		Description:              "創建統計消息",
		DefaultMemberPermissions: stringPtr(manageMessagesPermission),
		Ownership:                commands.ManagedOwnership("stats-create", commands.ScopeGuild),
		Options: []commands.Option{
			{
				Type:        commands.OptionTypeString,
				Name:        statsOptionChannelType,
				Description: "輸入統計頻道要是文字頻道還是語音頻道",
				Required:    true,
				Choices: []commands.Choice{
					{Name: "文字頻道", Value: "文字頻道"},
					{Name: "語音頻道", Value: "語音頻道"},
				},
			},
			{
				Type:        commands.OptionTypeString,
				Name:        statsOptionStat,
				Description: "輸入統計要統計甚麼",
				Choices: []commands.Choice{
					{Name: "頻道數量", Value: "頻道數量"},
					{Name: "文字頻道數量", Value: "文字頻道數量"},
					{Name: "語音頻道數量", Value: "語音頻道數量"},
				},
			},
		},
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

func RoleDefinition() commands.Definition {
	return commands.Definition{
		Type:                     commands.CommandTypeChatInput,
		Name:                     StatsRoleCommandName,
		Description:              "統計某個特定的身分組的人數",
		DefaultMemberPermissions: stringPtr(manageMessagesPermission),
		Ownership:                commands.ManagedOwnership("stats-role-count", commands.ScopeGuild),
		Options: []commands.Option{
			{
				Type:        commands.OptionTypeString,
				Name:        statsOptionChannelType,
				Description: "輸入統計頻道要是文字頻道還是語音頻道",
				Required:    true,
				Choices: []commands.Choice{
					{Name: "文字頻道", Value: "文字頻道"},
					{Name: "語音頻道", Value: "語音頻道"},
				},
			},
			{
				Type:        commands.OptionTypeRole,
				Name:        statsOptionRole,
				Description: "輸入要統計哪個身分組",
				Required:    true,
			},
		},
	}
}

func stringPtr(value string) *string {
	return &value
}
