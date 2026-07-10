package logging

import "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/commands"

const LoggingConfigCommandName = "set-log-channel"

func Definitions() []commands.Definition {
	return []commands.Definition{LoggingConfigDefinition()}
}

func LoggingConfigDefinition() commands.Definition {
	return commands.Definition{
		Type:        commands.CommandTypeChatInput,
		Name:        LoggingConfigCommandName,
		Description: "Set where log messages should send",
		NameLocalizations: map[string]string{
			"zh-TW": "設置日誌",
			"zh-CN": "设置日志",
			"en-US": "set-log-channel",
			"en-GB": "set-log-channel",
		},
		DescriptionLocalizations: map[string]string{
			"en-US": "Set where log messages should send",
			"en-GB": "Set where log messages should send",
			"zh-TW": "設置日誌訊息要在哪發送",
			"zh-CN": "设置日志讯息要在哪发送",
		},
		Ownership: commands.ManagedOwnership("logging-config", commands.ScopeGuild),
		Options: []commands.Option{{
			Type:        commands.OptionTypeChannel,
			Name:        "channel",
			Description: "Enter log channel!",
			NameLocalizations: map[string]string{
				"en-US": "channel",
				"en-GB": "channel",
				"zh-TW": "頻道",
				"zh-CN": "頻道",
			},
			DescriptionLocalizations: map[string]string{
				"en-US": "Enter log channel!",
				"en-GB": "Enter log channel!",
				"zh-TW": "輸入日誌頻道!",
				"zh-CN": "输入日志频道",
			},
			Required:     true,
			ChannelTypes: []int{0, 5},
		}},
	}
}
