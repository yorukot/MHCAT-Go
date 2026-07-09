package xp

import "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/commands"

const (
	TextXPProfileCommandName     = "聊天經驗"
	TextXPSetCommandName         = "聊天經驗設定"
	TextXPDeleteCommandName      = "聊天經驗刪除"
	VoiceXPProfileCommandName    = "語音經驗"
	VoiceXPSetCommandName        = "語音經驗設定"
	VoiceXPDeleteCommandName     = "語音經驗刪除"
	TextXPRewardRoleCommandName  = "聊天經驗身分組設定"
	VoiceXPRewardRoleCommandName = "語音經驗身分組設定"
	XPAdminCommandName           = "經驗值改變"
	manageMessagesPermission     = "8192"
	kickMembersPermission        = "2"
)

func Definitions() []commands.Definition {
	return TextDefinitions()
}

func TextDefinitions() []commands.Definition {
	return []commands.Definition{TextXPSetDefinition(), TextXPDeleteDefinition()}
}

func VoiceDefinitions() []commands.Definition {
	return []commands.Definition{VoiceXPSetDefinition(), VoiceXPDeleteDefinition()}
}

func DisabledProfileDefinitions() []commands.Definition {
	return []commands.Definition{TextXPProfileDefinition(), VoiceXPProfileDefinition()}
}

func RewardRoleDefinitions() []commands.Definition {
	return []commands.Definition{TextXPRewardRoleDefinition(), VoiceXPRewardRoleDefinition()}
}

func AdminDefinitions() []commands.Definition {
	return []commands.Definition{XPAdminDefinition()}
}

func TextXPProfileDefinition() commands.Definition {
	return xpProfileDefinition(TextXPProfileCommandName, "查詢聊天經驗")
}

func VoiceXPProfileDefinition() commands.Definition {
	return xpProfileDefinition(VoiceXPProfileCommandName, "查詢語音經驗")
}

func TextXPSetDefinition() commands.Definition {
	return commands.Definition{
		Type:                     commands.CommandTypeChatInput,
		Name:                     TextXPSetCommandName,
		Description:              "設定聊天經驗通知要在哪發送",
		DefaultMemberPermissions: stringPtr(manageMessagesPermission),
		DocsURL:                  "https://docsmhcat.yorukot.me/docs/chat_xp_set",
		Ownership:                commands.ManagedOwnership("text-xp-config", commands.ScopeGuild),
		Options: []commands.Option{
			{
				Type:         commands.OptionTypeChannel,
				Name:         "頻道",
				Description:  "輸入頻道!",
				Required:     true,
				ChannelTypes: []int{0, 5},
			},
			{
				Type:        commands.OptionTypeString,
				Name:        "訊息",
				Description: "當有人升等的訊息，輸入:{level}為等級，{user}為tag使用者",
			},
			{
				Type:        commands.OptionTypeString,
				Name:        "顏色",
				Description: "輸入玩家查詢的主題要甚麼顏色(默認為白色)!",
			},
		},
	}
}

func VoiceXPSetDefinition() commands.Definition {
	return commands.Definition{
		Type:                     commands.CommandTypeChatInput,
		Name:                     VoiceXPSetCommandName,
		Description:              "設定語音經驗通知要在哪發送",
		DefaultMemberPermissions: stringPtr(manageMessagesPermission),
		DocsURL:                  "https://docsmhcat.yorukot.me/docs/voice_xp_set",
		Ownership:                commands.ManagedOwnership("voice-xp-config", commands.ScopeGuild),
		Options: []commands.Option{
			{
				Type:         commands.OptionTypeChannel,
				Name:         "頻道",
				Description:  "輸入頻道!",
				Required:     true,
				ChannelTypes: []int{0, 5},
			},
			{
				Type:        commands.OptionTypeString,
				Name:        "訊息",
				Description: "當有人升等的訊息，輸入:{level}為等級，{user}為tag使用者",
			},
			{
				Type:        commands.OptionTypeString,
				Name:        "顏色",
				Description: "輸入玩家查詢的主題要甚麼顏色(默認為白色)!",
			},
			{
				Type:        commands.OptionTypeString,
				Name:        "背景",
				Description: "輸入玩家查詢的背景(默認為discord色)支援png和jpg(可使用discord的複製連結)最佳大小為931*231",
			},
		},
	}
}

func TextXPDeleteDefinition() commands.Definition {
	return commands.Definition{
		Type:                     commands.CommandTypeChatInput,
		Name:                     TextXPDeleteCommandName,
		Description:              "刪除聊天經驗發送訊息設置",
		DefaultMemberPermissions: stringPtr(manageMessagesPermission),
		DocsURL:                  "https://docsmhcat.yorukot.me/docs/chat_xp_delete",
		Ownership:                commands.ManagedOwnership("text-xp-config", commands.ScopeGuild),
	}
}

func VoiceXPDeleteDefinition() commands.Definition {
	return commands.Definition{
		Type:                     commands.CommandTypeChatInput,
		Name:                     VoiceXPDeleteCommandName,
		Description:              "刪除語音發送訊息設置",
		DefaultMemberPermissions: stringPtr(manageMessagesPermission),
		DocsURL:                  "https://docsmhcat.yorukot.me/docs/voice_xp_delete",
		Ownership:                commands.ManagedOwnership("voice-xp-config", commands.ScopeGuild),
	}
}

func TextXPRewardRoleDefinition() commands.Definition {
	return xpRewardRoleDefinition(TextXPRewardRoleCommandName, "設定聊天經驗通知要在哪發送", "text-xp-role-config", "https://docsmhcat.yorukot.me.xyz.xyz/docs/chat_xp_set", "輸入之前設定的身分組")
}

func VoiceXPRewardRoleDefinition() commands.Definition {
	return xpRewardRoleDefinition(VoiceXPRewardRoleCommandName, "設定語音經驗通知要在哪發送(兼增加、刪除、設定查詢)", "voice-xp-role-config", "https://docsmhcat.yorukot.me/docs/chat_xp_set", "當到達設定的等級時時，要給甚麼身份組")
}

func XPAdminDefinition() commands.Definition {
	return commands.Definition{
		Type:                     commands.CommandTypeChatInput,
		Name:                     XPAdminCommandName,
		Description:              "增加某人的經驗值",
		DefaultMemberPermissions: stringPtr(kickMembersPermission),
		Ownership:                commands.ManagedOwnership("xp-admin", commands.ScopeGuild),
		Options: []commands.Option{
			{
				Type:        commands.OptionTypeSubCommand,
				Name:        "聊天經驗改變",
				Description: "增加聊天經驗",
				Options: []commands.Option{
					{Type: commands.OptionTypeUser, Name: "使用者", Description: "要增加的對象!", Required: true},
					{Type: commands.OptionTypeInteger, Name: "經驗值", Description: "要增加多少經驗值!", Required: true},
				},
			},
			{
				Type:        commands.OptionTypeSubCommand,
				Name:        "語音經驗改變",
				Description: "增加語音經驗",
				Options: []commands.Option{
					{Type: commands.OptionTypeUser, Name: "使用者", Description: "要增加的對象!", Required: true},
					{Type: commands.OptionTypeInteger, Name: "經驗值", Description: "要增加多少經驗值!", Required: true},
				},
			},
		},
	}
}

func xpProfileDefinition(name string, description string) commands.Definition {
	return commands.Definition{
		Type:        commands.CommandTypeChatInput,
		Name:        name,
		Description: description,
		Ownership:   commands.ManagedOwnership("xp-profile-disabled", commands.ScopeGuild),
		Options: []commands.Option{
			{
				Type:        commands.OptionTypeUser,
				Name:        "玩家",
				Description: "輸入玩家!",
			},
		},
	}
}

func xpRewardRoleDefinition(name string, description string, owner string, docsURL string, deleteRoleDescription string) commands.Definition {
	return commands.Definition{
		Type:                     commands.CommandTypeChatInput,
		Name:                     name,
		Description:              description,
		DefaultMemberPermissions: stringPtr(manageMessagesPermission),
		DocsURL:                  docsURL,
		Ownership:                commands.ManagedOwnership(owner, commands.ScopeGuild),
		Options: []commands.Option{
			{
				Type:        commands.OptionTypeSubCommand,
				Name:        "增加",
				Description: "當有人的等級達到後要給予身分組",
				Options: []commands.Option{
					{Type: commands.OptionTypeInteger, Name: "等級", Description: "輸入要在幾等時給予身分組!", Required: true},
					{Type: commands.OptionTypeRole, Name: "身分組", Description: "當到達設定的等級時時，要給甚麼身份組", Required: true},
					{Type: commands.OptionTypeBoolean, Name: "是否自動刪除", Description: "當使用者等級不再是所設定的等級時自動將此身分組刪除(默認為否)"},
				},
			},
			{
				Type:        commands.OptionTypeSubCommand,
				Name:        "刪除",
				Description: "刪除之前的設定",
				Options: []commands.Option{
					{Type: commands.OptionTypeInteger, Name: "等級", Description: "輸入之前設定的身分組!", Required: true},
					{Type: commands.OptionTypeRole, Name: "身分組", Description: deleteRoleDescription, Required: true},
				},
			},
			{
				Type:        commands.OptionTypeSubCommand,
				Name:        "設定查詢",
				Description: "查看之前的設定",
			},
		},
	}
}

func stringPtr(value string) *string {
	return &value
}
