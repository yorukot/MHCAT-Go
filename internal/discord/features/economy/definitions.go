package economy

import "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/commands"

const (
	CoinQueryCommandName       = "代幣查詢"
	SignInCommandName          = "簽到"
	EconomySettingsCommandName = "coin-related-settings"
	manageMessagesPermission   = "8192"
	textChannelType            = 0
	announcementChannelType    = 5
)

func Definitions() []commands.Definition {
	return []commands.Definition{CoinQueryDefinition()}
}

func SignInDefinitions() []commands.Definition {
	return []commands.Definition{SignInDefinition()}
}

func SettingsDefinitions() []commands.Definition {
	return []commands.Definition{SettingsDefinition()}
}

func CoinQueryDefinition() commands.Definition {
	return commands.Definition{
		Type:        commands.CommandTypeChatInput,
		Name:        CoinQueryCommandName,
		Description: "查詢你有多少代幣",
		DocsURL:     "https://docsmhcat.yorukot.meocs/coin",
		Ownership:   commands.ManagedOwnership("economy-query", commands.ScopeGuild),
		Options: []commands.Option{
			{
				Type:        commands.OptionTypeUser,
				Name:        "使用者",
				Description: "要查詢的使用者",
				Required:    false,
			},
		},
	}
}

func SignInDefinition() commands.Definition {
	return commands.Definition{
		Type:        commands.CommandTypeChatInput,
		Name:        SignInCommandName,
		Description: "簽到來獲得代幣",
		DocsURL:     "https://docsmhcat.yorukot.me/docs/snig",
		Ownership:   commands.ManagedOwnership("economy-signin", commands.ScopeGuild),
	}
}

func SettingsDefinition() commands.Definition {
	return commands.Definition{
		Type:        commands.CommandTypeChatInput,
		Name:        EconomySettingsCommandName,
		Description: "Various settings related to tokens",
		NameLocalizations: map[string]string{
			"en-US": "coin-related-settings",
			"zh-TW": "代幣相關設定",
			"zh-CN": "代币相关设定",
		},
		DescriptionLocalizations: map[string]string{
			"en-US": "Various settings related to tokens",
			"zh-TW": "有關代幣的各項設定",
			"zh-CN": "有关代币的各项设定",
		},
		DefaultMemberPermissions: stringPtr(manageMessagesPermission),
		DocsURL:                  "https://docsmhcat.yorukot.meocs/required_coins",
		Ownership:                commands.ManagedOwnership("economy-settings", commands.ScopeGuild),
		Options: []commands.Option{
			{
				Type:        commands.OptionTypeInteger,
				Name:        "coin-raffle-takes",
				Description: "The amount of coin raffle requires",
				NameLocalizations: map[string]string{
					"en-US": "coin-raffle-takes",
					"zh-TW": "扭蛋所需代幣",
					"zh-CN": "扭蛋所需代幣",
				},
				DescriptionLocalizations: map[string]string{
					"en-US": "The amount of coin raffle requires",
					"zh-TW": "每次扭蛋所需的代幣數量",
					"zh-CN": "每次扭蛋所需的代币数量",
				},
				Required: true,
			},
			{
				Type:        commands.OptionTypeInteger,
				Name:        "check-in-cooldown-time",
				Description: "Time between check-in(Unit is hour)(If you want to set it to 0:00, type 0)",
				NameLocalizations: map[string]string{
					"en-US": "check-in-cooldown-time",
					"zh-TW": "簽到所需時間",
					"zh-CN": "簽到所需時間",
				},
				DescriptionLocalizations: map[string]string{
					"en-US": "Time between check-in(Unit is hour)(If you want to set it to 0:00, type 0)",
					"zh-TW": "每次簽到所需時間(單位為小時)(如想設置為0:00重製請打0)",
					"zh-CN": "每次签到所需时间(单位为小时)(如想设置为0:00重新制作请打0)",
				},
				Required: true,
			},
			{
				Type:        commands.OptionTypeInteger,
				Name:        "check-in-give-coins",
				Description: "How many amount of coin check-in gives",
				NameLocalizations: map[string]string{
					"en-US": "check-in-give-coins",
					"zh-TW": "簽到給予代幣數",
					"zh-CN": "簽到給予代幣數",
				},
				DescriptionLocalizations: map[string]string{
					"en-US": "How many amount of coin check-in gives",
					"zh-TW": "每次扭蛋所需的代幣數量",
					"zh-CN": "每次扭蛋所需的代币数量",
				},
				Required: true,
			},
			{
				Type:         commands.OptionTypeChannel,
				Name:         "notification-channel",
				Description:  "Channel to announcement raffle winner",
				ChannelTypes: []int{textChannelType, announcementChannelType},
				NameLocalizations: map[string]string{
					"en-US": "notification-channel",
					"zh-TW": "通知頻道",
					"zh-CN": "通知頻道",
				},
				DescriptionLocalizations: map[string]string{
					"en-US": "Channel to announcement raffle winner",
					"zh-TW": "抽中後的通知頻道",
					"zh-CN": "抽中后的通知频道",
				},
				Required: true,
			},
			{
				Type:        commands.OptionTypeNumber,
				Name:        "level-up-multiply-amount",
				Description: "How many coins you get when you level up.If your level is 9, and the multiply amount is 9,9*10=90",
				NameLocalizations: map[string]string{
					"en-US": "level-up-multiply-amount",
					"zh-TW": "等級提升給予倍數",
					"zh-CN": "等級提升給予倍數",
				},
				DescriptionLocalizations: map[string]string{
					"en-US": "How many coins you get when you level up.If your level is 9, and the multiply amount is 9,9*10=90",
					"zh-TW": "等級提升時要給等級幾倍的代幣ex:假設你提升到9等，倍數設10就會得到 9*10=90",
					"zh-CN": "等级提升时要给等级几倍的代币ex:假设你提升到9等，倍数设10就会得到9*10=90",
				},
				Required: true,
			},
		},
	}
}

func stringPtr(value string) *string {
	return &value
}
