package announcements

import "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/commands"

const (
	ConfigCommandName = "公告頻道設置"
	SendCommandName   = "公告發送"

	subcommandOnce        = "一次性公告頻道"
	subcommandBound       = "綁定公告頻道"
	subcommandDeleteBound = "綁定公告頻道刪除"

	optionChannel = "頻道"
	optionTag     = "標註"
	optionColor   = "顏色"
	optionTitle   = "標題"
)

func Definitions() []commands.Definition {
	return ConfigDefinitions()
}

func ConfigDefinitions() []commands.Definition {
	return []commands.Definition{ConfigDefinition()}
}

func SendDefinitions() []commands.Definition {
	return []commands.Definition{SendDefinition()}
}

func ConfigDefinition() commands.Definition {
	return commands.Definition{
		Type:        commands.CommandTypeChatInput,
		Name:        ConfigCommandName,
		Description: "設定公告在哪發送",
		DocsURL:     "https://docsmhcat.yorukot.meocs/ann_set",
		Ownership:   commands.ManagedOwnership("announcement-config", commands.ScopeGuild),
		Options: []commands.Option{
			{
				Type:        commands.OptionTypeSubCommand,
				Name:        subcommandOnce,
				Description: "設定一次性公告頻道要在哪發送",
				Options: []commands.Option{
					channelOption(),
				},
			},
			{
				Type:        commands.OptionTypeSubCommand,
				Name:        subcommandBound,
				Description: "設定綁定型公告要在哪發送以及發送時的格式",
				Options: []commands.Option{
					channelOption(),
					{
						Type:        commands.OptionTypeString,
						Name:        optionTag,
						Description: "輸入要標註哪個身分組!",
						Required:    true,
					},
					{
						Type:        commands.OptionTypeString,
						Name:        optionColor,
						Description: "輸入這個綁定公告頻道的設定!(隨機顏色請輸入Random)",
						Required:    true,
					},
					{
						Type:        commands.OptionTypeString,
						Name:        optionTitle,
						Description: "輸入公告發送的頻道!",
						Required:    true,
					},
				},
			},
			{
				Type:        commands.OptionTypeSubCommand,
				Name:        subcommandDeleteBound,
				Description: "刪除之前的設定",
				Options: []commands.Option{
					channelOption(),
				},
			},
		},
	}
}

func SendDefinition() commands.Definition {
	return commands.Definition{
		Type:        commands.CommandTypeChatInput,
		Name:        SendCommandName,
		Description: "發送公告訊息",
		DocsURL:     "https://docsmhcat.yorukot.me/docs/ann",
		Ownership:   commands.ManagedOwnership("announcement-send", commands.ScopeGuild),
	}
}

func channelOption() commands.Option {
	return commands.Option{
		Type:         commands.OptionTypeChannel,
		Name:         optionChannel,
		Description:  "輸入公告發送的頻道!",
		Required:     true,
		ChannelTypes: []int{0, 5},
	}
}
