package ticket

import "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/commands"

const manageMessagesPermission = "8192"

func Definitions() []commands.Definition {
	return []commands.Definition{
		SetupDefinition(),
		DeleteDefinition(),
	}
}

func SetupDefinition() commands.Definition {
	return commands.Definition{
		Type:                     commands.CommandTypeChatInput,
		Name:                     "私人頻道設置",
		Description:              "設置私人頻道",
		DefaultMemberPermissions: stringPtr(manageMessagesPermission),
		Ownership:                commands.ManagedOwnership("ticket-foundation", commands.ScopeGuild),
		Options: []commands.Option{
			{
				Type:         commands.OptionTypeChannel,
				Name:         "類別",
				Description:  "輸入私人頻道要在哪個類別開啟!",
				Required:     true,
				ChannelTypes: []int{4},
			},
			{
				Type:        commands.OptionTypeRole,
				Name:        "管理員身分組",
				Description: "輸入管理員身分組(有這個身分組的能夠管理私人頻道)!",
				Required:    true,
			},
		},
	}
}

func DeleteDefinition() commands.Definition {
	return commands.Definition{
		Type:                     commands.CommandTypeChatInput,
		Name:                     "私人頻道刪除",
		Description:              "刪除之前設置的私人頻道",
		DefaultMemberPermissions: stringPtr(manageMessagesPermission),
		Ownership:                commands.ManagedOwnership("ticket-foundation", commands.ScopeGuild),
	}
}

func stringPtr(value string) *string {
	return &value
}
