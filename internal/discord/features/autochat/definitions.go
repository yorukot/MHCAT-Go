package autochat

import "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/commands"

const (
	AutoChatSetCommandName    = "自動聊天頻道"
	AutoChatDeleteCommandName = "自動聊天頻道刪除"
	optionChannel             = "頻道"
)

func Definitions() []commands.Definition {
	return []commands.Definition{SetDefinition(), DeleteDefinition()}
}

func SetDefinition() commands.Definition {
	return commands.Definition{
		Type:        commands.CommandTypeChatInput,
		Name:        AutoChatSetCommandName,
		Description: "設定自動聊天頻道要在哪裡發送",
		DocsURL:     "https://docsmhcat.yorukot.me/docs/chat_xp_set",
		Ownership:   commands.ManagedOwnership("autochat-config", commands.ScopeGuild),
		Options: []commands.Option{{
			Type:         commands.OptionTypeChannel,
			Name:         optionChannel,
			Description:  "輸入頻道!",
			Required:     true,
			ChannelTypes: []int{0, 5},
		}},
	}
}

func DeleteDefinition() commands.Definition {
	return commands.Definition{
		Type:        commands.CommandTypeChatInput,
		Name:        AutoChatDeleteCommandName,
		Description: "刪除自動聊天頻道要在哪裡發送",
		DocsURL:     "https://docsmhcat.yorukot.me/docs/chat_xp_set",
		Ownership:   commands.ManagedOwnership("autochat-config", commands.ScopeGuild),
	}
}
