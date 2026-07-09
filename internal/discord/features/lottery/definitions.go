package lottery

import "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/commands"

const (
	LotteryCreateCommandName = "抽獎設置"
	manageMessagesPermission = "8192"
)

func Definitions() []commands.Definition {
	return []commands.Definition{CreateDefinition()}
}

func CreateDefinition() commands.Definition {
	return commands.Definition{
		Type:                     commands.CommandTypeChatInput,
		Name:                     LotteryCreateCommandName,
		Description:              "設置抽獎訊息",
		DefaultMemberPermissions: stringPtr(manageMessagesPermission),
		DocsURL:                  "https://docsmhcat.yorukot.meocs/lotter",
		Ownership:                commands.ManagedOwnership("lottery-disabled-command", commands.ScopeGuild),
		Options: []commands.Option{
			{
				Type:        commands.OptionTypeString,
				Name:        "截止日期",
				Description: "輸入多久後截止ex:01d02h10m(1天2小時10分鐘後截止)也可以輸入單一ex:01d or 03h10m",
				Required:    true,
			},
			{
				Type:        commands.OptionTypeInteger,
				Name:        "抽出幾位中獎者",
				Description: "請輸入要抽出幾位中獎者",
				Required:    true,
			},
			{
				Type:        commands.OptionTypeString,
				Name:        "獎品",
				Description: "輸入獎品要甚麼",
				Required:    true,
			},
			{
				Type:        commands.OptionTypeString,
				Name:        "內文",
				Description: "輸入抽獎訊息內文",
				Required:    true,
			},
			{
				Type:        commands.OptionTypeRole,
				Name:        "可以抽的身分組",
				Description: "輸入哪個身分組可以抽(要有這個身分組才能抽)(選填)",
			},
			{
				Type:        commands.OptionTypeRole,
				Name:        "不能抽的身分組",
				Description: "輸入哪個身分組不能抽(有這個身分組就不能抽)(選填)",
			},
			{
				Type:        commands.OptionTypeInteger,
				Name:        "最高抽獎人數",
				Description: "設定最多只能有幾位參加",
			},
		},
	}
}

func stringPtr(value string) *string {
	return &value
}
