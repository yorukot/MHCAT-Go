package moderation

import "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/commands"

const WarningHistoryCommandName = "警告紀錄"
const WarningSettingsCommandName = "警告設定"
const WarningRemoveCommandName = "警告清除"
const WarningRemoveAllCommandName = "警告全部清除"

const (
	warningSettingsOptionAction    = "執行的動作"
	warningSettingsOptionThreshold = "幾次警告後執行動作"
	warningOptionUser              = "使用者"
	warningRemoveOptionIndex       = "第幾項"
)

func Definitions() []commands.Definition {
	return []commands.Definition{WarningHistoryDefinition()}
}

func SettingsDefinitions() []commands.Definition {
	return []commands.Definition{WarningSettingsDefinition()}
}

func RemovalDefinitions() []commands.Definition {
	return []commands.Definition{WarningRemoveDefinition(), WarningRemoveAllDefinition()}
}

func WarningHistoryDefinition() commands.Definition {
	return commands.Definition{
		Type:        commands.CommandTypeChatInput,
		Name:        WarningHistoryCommandName,
		Description: "收尋一位使用者的警告",
		Ownership:   commands.ManagedOwnership("warnings", commands.ScopeGuild),
		Options: []commands.Option{{
			Type:        commands.OptionTypeUser,
			Name:        warningOptionUser,
			Description: "要收尋的使用者!",
			Required:    true,
		}},
	}
}

func WarningSettingsDefinition() commands.Definition {
	return commands.Definition{
		Type:        commands.CommandTypeChatInput,
		Name:        WarningSettingsCommandName,
		Description: "警告的各種設定",
		Ownership:   commands.ManagedOwnership("warning-settings", commands.ScopeGuild),
		Options: []commands.Option{
			{
				Type:        commands.OptionTypeString,
				Name:        warningSettingsOptionAction,
				Description: "警告他的原因",
				Required:    true,
				Choices: []commands.Choice{
					{Name: "停權", Value: "停權"},
					{Name: "踢出", Value: "踢出"},
				},
			},
			{
				Type:        commands.OptionTypeInteger,
				Name:        warningSettingsOptionThreshold,
				Description: "被警告幾次後要執行這個動作!",
				Required:    true,
			},
		},
	}
}

func WarningRemoveDefinition() commands.Definition {
	return commands.Definition{
		Type:        commands.CommandTypeChatInput,
		Name:        WarningRemoveCommandName,
		Description: "清除一個使用者的某個警告",
		Ownership:   commands.ManagedOwnership("warning-removal", commands.ScopeGuild),
		Options: []commands.Option{
			{
				Type:        commands.OptionTypeUser,
				Name:        warningOptionUser,
				Description: "要清除資料的使用者!",
				Required:    true,
			},
			{
				Type:        commands.OptionTypeInteger,
				Name:        warningRemoveOptionIndex,
				Description: "要清除第幾個警告!",
				Required:    true,
			},
		},
	}
}

func WarningRemoveAllDefinition() commands.Definition {
	return commands.Definition{
		Type:        commands.CommandTypeChatInput,
		Name:        WarningRemoveAllCommandName,
		Description: "清除一個使用者的全部警告",
		Ownership:   commands.ManagedOwnership("warning-removal", commands.ScopeGuild),
		Options: []commands.Option{{
			Type:        commands.OptionTypeUser,
			Name:        warningOptionUser,
			Description: "要清除資料的使用者!",
			Required:    true,
		}},
	}
}
