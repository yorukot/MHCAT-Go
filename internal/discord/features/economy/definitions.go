package economy

import (
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/commands"
)

const (
	CoinQueryCommandName         = "代幣查詢"
	CoinAdminCommandName         = "代幣增加"
	CoinRankCommandName          = "代幣排行榜"
	ProfileCommandName           = "my-profile"
	RockPaperScissorsCommandName = "剪刀石頭布"
	SignInCommandName            = "簽到"
	SignInListCommandName        = "簽到列表"
	EconomySettingsCommandName   = "coin-related-settings"
	manageMessagesPermission     = "8192"
	textChannelType              = 0
	announcementChannelType      = 5
)

func Definitions() []commands.Definition {
	return []commands.Definition{CoinQueryDefinition()}
}

func SignInDefinitions() []commands.Definition {
	return []commands.Definition{SignInDefinition(), SignInListDefinition()}
}

func SettingsDefinitions() []commands.Definition {
	return []commands.Definition{SettingsDefinition()}
}

func CoinAdminDefinitions() []commands.Definition {
	return []commands.Definition{CoinAdminDefinition()}
}

func CoinRankDefinitions() []commands.Definition {
	return []commands.Definition{CoinRankDefinition()}
}

func RockPaperScissorsDefinitions() []commands.Definition {
	return []commands.Definition{RockPaperScissorsDefinition()}
}

func ProfileDefinitions() []commands.Definition {
	return []commands.Definition{ProfileDefinition()}
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

func CoinRankDefinition() commands.Definition {
	return commands.Definition{
		Type:        commands.CommandTypeChatInput,
		Name:        CoinRankCommandName,
		Description: "查詢代幣的排行榜",
		Ownership:   commands.ManagedOwnership("economy-coin-rank", commands.ScopeGuild),
	}
}

func RockPaperScissorsDefinition() commands.Definition {
	return commands.Definition{
		Type:        commands.CommandTypeChatInput,
		Name:        RockPaperScissorsCommandName,
		Description: "跟電腦剪刀時候布來獲得代幣(有賺有賠)",
		DocsURL:     "https://docsmhcat.yorukot.me/docs/required_coins",
		Ownership:   commands.ManagedOwnership("economy-rps", commands.ScopeGuild),
		Options: []commands.Option{
			{
				Type:        commands.OptionTypeInteger,
				Name:        "使用多少代幣來進行",
				Description: "要用多少代幣進行賭注(贏的話會多這些，輸的話這些代幣會全被拿走，平手會被扣這些的一半)",
				Required:    true,
			},
			{
				Type:        commands.OptionTypeString,
				Name:        "剪刀石頭或布",
				Description: "選擇要剪刀石頭還是布",
				Required:    true,
				Choices: []commands.Choice{
					{Name: string(domain.RockPaperScissorsChoiceScissors), Value: string(domain.RockPaperScissorsChoiceScissors)},
					{Name: string(domain.RockPaperScissorsChoiceRock), Value: string(domain.RockPaperScissorsChoiceRock)},
					{Name: string(domain.RockPaperScissorsChoicePaper), Value: string(domain.RockPaperScissorsChoicePaper)},
				},
			},
		},
	}
}

func ProfileDefinition() commands.Definition {
	return commands.Definition{
		Type:        commands.CommandTypeChatInput,
		Name:        ProfileCommandName,
		Description: "Check about data in specific server!!",
		NameLocalizations: map[string]string{
			"zh-TW": "我的檔案",
			"zh-CN": "我的档案",
			"en-US": "my-profile",
			"en-GB": "my-profile",
		},
		DescriptionLocalizations: map[string]string{
			"zh-TW": "查看自己在伺服器內的所有資料!",
			"zh-CN": "查看自己在伺服器内的所有资料!",
			"en-US": "Check about data in specific server!",
			"en-GB": "Check about data in specific server!",
		},
		Options: []commands.Option{{
			Type:        commands.OptionTypeUser,
			Name:        "user",
			Description: "查詢某位使用者的資料",
			NameLocalizations: map[string]string{
				"zh-TW": "使用者",
				"zh-CN": "使用者",
				"en-US": "user",
				"en-GB": "user",
			},
			DescriptionLocalizations: map[string]string{
				"zh-TW": "查詢某位使用者的資料!",
				"zh-CN": "查询某位使用者的资料!",
				"en-US": "Check a users data!",
				"en-GB": "Check a users data!",
			},
		}},
		DocsURL:   "https://docsmhcat.yorukot.me/docs/snig",
		Ownership: commands.ManagedOwnership("economy-profile", commands.ScopeGuild),
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

func SignInListDefinition() commands.Definition {
	return commands.Definition{
		Type:        commands.CommandTypeChatInput,
		Name:        SignInListCommandName,
		Description: "查看今天有誰簽到了",
		DocsURL:     "https://docsmhcat.yorukot.me/docs/snig",
		Ownership:   commands.ManagedOwnership("economy-signin", commands.ScopeGuild),
	}
}

func CoinAdminDefinition() commands.Definition {
	return commands.Definition{
		Type:        commands.CommandTypeChatInput,
		Name:        CoinAdminCommandName,
		Description: "改變扭蛋數量",
		DocsURL:     "https://docsmhcat.yorukot.meocs/coin_increase",
		Ownership:   commands.ManagedOwnership("economy-coin-admin", commands.ScopeGuild),
		Options: []commands.Option{
			{
				Type:        commands.OptionTypeUser,
				Name:        "使用者",
				Description: "要改變的人",
				Required:    true,
			},
			{
				Type:        commands.OptionTypeString,
				Name:        "增加或減少",
				Description: "輸入這個獎品叫甚麼，以及簡單概述",
				DescriptionLocalizations: map[string]string{
					"en-US": "test",
					"zh-TW": "tetse",
				},
				Required: true,
				Choices: []commands.Choice{
					{Name: "增加", Value: string(domain.CoinAdminOperationAdd)},
					{Name: "減少", Value: string(domain.CoinAdminOperationReduce)},
				},
			},
			{
				Type:        commands.OptionTypeInteger,
				Name:        "數量",
				Description: "增加或減少的數量",
				Required:    true,
			},
		},
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
