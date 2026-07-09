package birthday

import (
	"fmt"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/commands"
)

const (
	BirthdayCommandName = "生日系統"

	subcommandConfig     = "祝福語設定"
	subcommandAdd        = "增加"
	subcommandDelete     = "刪除"
	subcommandAllowAdmin = "是否允許管理員設定"
	subcommandList       = "生日列表"

	optionMessage        = "祝福語"
	optionChannel        = "頻道"
	optionEveryoneCanSet = "是否可以自行設定生日"
	optionUTC            = "時區"
	optionRole           = "給予身分組"
	optionBirthdayMonth  = "生日月份"
	optionBirthdayDay    = "生日日期"
	optionUser           = "使用者"
	optionBirthdayYear   = "生日年份"
	optionAllowAdmin     = "是否"
)

func Definitions() []commands.Definition {
	return []commands.Definition{Definition()}
}

func Definition() commands.Definition {
	return commands.Definition{
		Type:        commands.CommandTypeChatInput,
		Name:        BirthdayCommandName,
		Description: "讓你的伺服器可以為生日的人慶生!",
		Ownership:   commands.ManagedOwnership("birthday-config", commands.ScopeGuild),
		Options: []commands.Option{
			{
				Type:        commands.OptionTypeSubCommand,
				Name:        subcommandConfig,
				Description: "設定祝福語",
				Options: []commands.Option{
					{
						Type:        commands.OptionTypeString,
						Name:        optionMessage,
						Description: "設定祝福語，{user}代表tag這個使用者 {name}代表這個使用者的名稱 {age}代表使用者年紀",
						Required:    true,
					},
					{
						Type:        commands.OptionTypeChannel,
						Name:        optionChannel,
						Description: "要發通知的頻道",
						Required:    true,
					},
					{
						Type:        commands.OptionTypeBoolean,
						Name:        optionEveryoneCanSet,
						Description: "使用者是否可以自己設定自己的生日日期跟通知時間",
						Required:    true,
					},
					{
						Type:        commands.OptionTypeString,
						Name:        optionUTC,
						Description: "設定屬於你伺服器的時區(台灣、香港、新加坡、馬來西亞、中國是UTC+8，日本是UTC+9) ",
						Required:    true,
						Choices:     legacyUTCChoices(),
					},
					{
						Type:        commands.OptionTypeRole,
						Name:        optionRole,
						Description: "當使用者那天生日時自動給予身分組，在當天結束後自動刪除",
					},
				},
			},
			{
				Type:        commands.OptionTypeSubCommand,
				Name:        subcommandAdd,
				Description: "新增使用者的生日資料",
				Options: []commands.Option{
					{Type: commands.OptionTypeInteger, Name: optionBirthdayMonth, Description: "格式為mm ex: 1(1月出生) ex:12(12月出生)", Required: true},
					{Type: commands.OptionTypeInteger, Name: optionBirthdayDay, Description: "格式為dd ex: 1(1日出生) 31(31日出生)", Required: true},
					{Type: commands.OptionTypeUser, Name: optionUser, Description: "要設定的使用者(需有管理員權限)"},
					{Type: commands.OptionTypeInteger, Name: optionBirthdayYear, Description: "格式為yyyy ex: 2023(2023年出生)"},
				},
			},
			{
				Type:        commands.OptionTypeSubCommand,
				Name:        subcommandDelete,
				Description: "刪除某使用者的資料",
				Options: []commands.Option{{
					Type:        commands.OptionTypeUser,
					Name:        optionUser,
					Description: "輸入你要刪除的使用者!",
					Required:    true,
				}},
			},
			{
				Type:        commands.OptionTypeSubCommand,
				Name:        subcommandAllowAdmin,
				Description: "是否允許管理員設定你的生日 防止打擾到你(預設為true)",
				Options: []commands.Option{{
					Type:        commands.OptionTypeBoolean,
					Name:        optionAllowAdmin,
					Description: "是否允許!true為允許 false為不允許",
					Required:    true,
				}},
			},
			{
				Type:        commands.OptionTypeSubCommand,
				Name:        subcommandList,
				Description: "這個伺服器內的使用者生日列表",
			},
		},
	}
}

func legacyUTCChoices() []commands.Choice {
	choices := make([]commands.Choice, 0, 24)
	for hour := 0; hour < 24; hour++ {
		choices = append(choices, commands.Choice{
			Name:  fmt.Sprintf("UTC+%d", hour),
			Value: fmt.Sprintf("+%02d:00", hour),
		})
	}
	return choices
}
