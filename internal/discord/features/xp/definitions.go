package xp

import "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/commands"

const (
	TextXPProfileCommandName  = "聊天經驗"
	TextXPSetCommandName      = "聊天經驗設定"
	TextXPDeleteCommandName   = "聊天經驗刪除"
	VoiceXPProfileCommandName = "語音經驗"
	VoiceXPSetCommandName     = "語音經驗設定"
	VoiceXPDeleteCommandName  = "語音經驗刪除"
	manageMessagesPermission  = "8192"
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

func stringPtr(value string) *string {
	return &value
}
