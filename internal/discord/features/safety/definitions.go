package safety

import "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/commands"

const (
	AntiScamCommandName      = "防詐騙網址"
	manageMessagesPermission = "8192"
)

func Definitions() []commands.Definition {
	return []commands.Definition{AntiScamDefinition()}
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

func stringPtr(value string) *string {
	return &value
}
