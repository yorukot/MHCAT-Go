package notifications

import "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/commands"

const (
	AutoNotificationSetupCommandName  = "automatic-notification"
	AutoNotificationListCommandName   = "自動通知列表"
	AutoNotificationDeleteCommandName = "自動通知刪除"
	manageMessagesPermission          = "8192"
	optionChannel                     = "channel"
	optionID                          = "id"
)

func Definitions() []commands.Definition {
	return []commands.Definition{SetupDefinition(), ListDefinition(), DeleteDefinition()}
}

func SetupDefinition() commands.Definition {
	return commands.Definition{
		Type:        commands.CommandTypeChatInput,
		Name:        AutoNotificationSetupCommandName,
		Description: "Set where automatic notification should be send",
		NameLocalizations: map[string]string{
			"en-US": "automatic-notification",
			"zh-TW": "自動通知",
			"zh-CN": "自动通知",
		},
		DescriptionLocalizations: map[string]string{
			"en-US": "Set where automatic notification should be send",
			"zh-TW": "设置自动聊天频道要在哪发送",
		},
		DefaultMemberPermissions: stringPtr(manageMessagesPermission),
		DocsURL:                  "https://youtu.be/D43zPrZU5Fw",
		Ownership:                commands.ManagedOwnership("auto-notification-config", commands.ScopeGuild),
		Options: []commands.Option{{
			Type:        commands.OptionTypeChannel,
			Name:        optionChannel,
			Description: "Enter channel to send!",
			NameLocalizations: map[string]string{
				"en-US": "channel",
				"zh-TW": "頻道",
				"zh-CN": "頻道",
			},
			DescriptionLocalizations: map[string]string{
				"en-US": "Enter channel to send!",
				"zh-TW": "輸入要發送的頻道!",
				"zh-CN": "输入要发送的频道",
			},
			Required:     true,
			ChannelTypes: []int{0, 5},
		}},
	}
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
