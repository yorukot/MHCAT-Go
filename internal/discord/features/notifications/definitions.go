package notifications

import "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/commands"

const (
	AutoNotificationListCommandName   = "自動通知列表"
	AutoNotificationDeleteCommandName = "自動通知刪除"
	manageMessagesPermission          = "8192"
	optionID                          = "id"
)

func Definitions() []commands.Definition {
	return []commands.Definition{ListDefinition(), DeleteDefinition()}
}

func ListDefinition() commands.Definition {
	return commands.Definition{
		Type:                     commands.CommandTypeChatInput,
		Name:                     AutoNotificationListCommandName,
		Description:              "查看所有的自動通知列表",
		DefaultMemberPermissions: stringPtr(manageMessagesPermission),
		DocsURL:                  "https://youtu.be/D43zPrZU5Fw",
		Ownership:                commands.ManagedOwnership("auto-notification-config", commands.ScopeGuild),
	}
}

func DeleteDefinition() commands.Definition {
	return commands.Definition{
		Type:                     commands.CommandTypeChatInput,
		Name:                     AutoNotificationDeleteCommandName,
		Description:              "刪除之前設定的自動通知",
		DefaultMemberPermissions: stringPtr(manageMessagesPermission),
		DocsURL:                  "https://youtu.be/D43zPrZU5Fw",
		Ownership:                commands.ManagedOwnership("auto-notification-config", commands.ScopeGuild),
		Options: []commands.Option{{
			Type:        commands.OptionTypeString,
			Name:        optionID,
			Description: "輸入要刪除的自動通知id!",
			Required:    true,
		}},
	}
}

func stringPtr(value string) *string {
	return &value
}
