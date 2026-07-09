package moderation

import "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/commands"

const WarningHistoryCommandName = "警告紀錄"
const WarningSettingsCommandName = "警告設定"
const WarningRemoveCommandName = "警告清除"
const WarningRemoveAllCommandName = "警告全部清除"
const WarningIssueCommandName = "警告"
const CleanupCommandName = "刪除訊息"

const (
	warningSettingsOptionAction    = "執行的動作"
	warningSettingsOptionThreshold = "幾次警告後執行動作"
	warningOptionUser              = "使用者"
	warningIssueOptionReason       = "原因"
	warningRemoveOptionIndex       = "第幾項"
	cleanupOptionCount             = "刪除數量"
	cleanupOptionUser              = "使用者"
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

func IssueDefinitions() []commands.Definition {
	return []commands.Definition{WarningIssueDefinition()}
}

func CleanupDefinitions() []commands.Definition {
	return []commands.Definition{CleanupDefinition()}
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

func WarningIssueDefinition() commands.Definition {
	return commands.Definition{
		Type:        commands.CommandTypeChatInput,
		Name:        WarningIssueCommandName,
		Description: "警告一個使用者",
		Ownership:   commands.ManagedOwnership("warning-issue", commands.ScopeGuild),
		Options: []commands.Option{
			{
				Type:        commands.OptionTypeUser,
				Name:        warningOptionUser,
				Description: "要警告的使用者!",
				Required:    true,
			},
			{
				Type:        commands.OptionTypeString,
				Name:        warningIssueOptionReason,
				Description: "警告他的原因",
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

func CleanupDefinition() commands.Definition {
	return commands.Definition{
		Type:        commands.CommandTypeChatInput,
		Name:        CleanupCommandName,
		Description: "刪除大量訊息",
		Ownership:   commands.ManagedOwnership("message-cleanup", commands.ScopeGuild),
		Options: []commands.Option{
			{
				Type:        commands.OptionTypeInteger,
				Name:        cleanupOptionCount,
				Description: "設定要刪除幾個訊息(最高1000超過200需要管理者權限)(只能刪除14天內的消息)",
				Required:    true,
			},
			{
				Type:        commands.OptionTypeUser,
				Name:        cleanupOptionUser,
				Description: "選擇是否要刪除某個特定的使用者的訊息(如填選這項，第一項代表的將是檢測訊息數量)",
				Required:    false,
			},
		},
	}
}
