package safety

import "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/commands"

const (
	AntiScamCommandName      = "防詐騙網址"
	ScamReportCommandName    = "詐騙網址回報"
	manageMessagesPermission = "8192"
)

func Definitions() []commands.Definition {
	return append(ConfigDefinitions(), ReportDefinitions()...)
}

func ConfigDefinitions() []commands.Definition {
	return []commands.Definition{AntiScamDefinition()}
}

func ReportDefinitions() []commands.Definition {
	return []commands.Definition{ScamReportDefinition()}
}

func AntiScamDefinition() commands.Definition {
	return commands.Definition{
		Type:                     commands.CommandTypeChatInput,
		Name:                     AntiScamCommandName,
		Description:              "設定是否開啟防詐騙網址功能(輸入這個指令就會更改)",
		DefaultMemberPermissions: stringPtr(manageMessagesPermission),
		Ownership:                commands.ManagedOwnership("anti-scam-config", commands.ScopeGuild),
	}
}

func ScamReportDefinition() commands.Definition {
	return commands.Definition{
		Type:        commands.CommandTypeChatInput,
		Name:        ScamReportCommandName,
		Description: "回報詐騙網站",
		Ownership:   commands.ManagedOwnership("anti-scam-report", commands.ScopeGuild),
		Options: []commands.Option{{
			Type:        commands.OptionTypeString,
			Name:        "網址",
			Description: "回報網址",
			Required:    true,
		}},
	}
}

func stringPtr(value string) *string {
	return &value
}
