package moderation

import "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/commands"

const WarningHistoryCommandName = "警告紀錄"
const WarningSettingsCommandName = "警告設定"

const (
	warningSettingsOptionAction    = "執行的動作"
	warningSettingsOptionThreshold = "幾次警告後執行動作"
)

func Definitions() []commands.Definition {
	return []commands.Definition{WarningHistoryDefinition()}
}

func SettingsDefinitions() []commands.Definition {
	return []commands.Definition{WarningSettingsDefinition()}
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
