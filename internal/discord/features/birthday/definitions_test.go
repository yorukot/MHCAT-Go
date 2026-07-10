package birthday

import (
	"fmt"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/commands"
)

func TestDefinitionMatchesLegacyCommandShape(t *testing.T) {
	definition := Definition()
	if definition.Name != BirthdayCommandName || definition.Description != "讓你的伺服器可以為生日的人慶生!" {
		t.Fatalf("definition = %#v", definition)
	}
	if definition.DefaultMemberPermissions != nil {
		t.Fatalf("legacy birthday command should not set default member permissions: %#v", definition.DefaultMemberPermissions)
	}
	if len(definition.Options) != 5 {
		t.Fatalf("options = %#v", definition.Options)
	}
	config := definition.Options[0]
	if config.Type != commands.OptionTypeSubCommand || config.Name != subcommandConfig || config.Description != "設定祝福語" || len(config.Options) != 5 {
		t.Fatalf("config subcommand = %#v", config)
	}
	if config.Options[0].Name != optionMessage || config.Options[0].Type != commands.OptionTypeString || config.Options[0].Description != "設定祝福語，{user}代表tag這個使用者 {name}代表這個使用者的名稱 {age}代表使用者年紀" || !config.Options[0].Required {
		t.Fatalf("message option = %#v", config.Options[0])
	}
	if config.Options[1].Name != optionChannel || config.Options[1].Type != commands.OptionTypeChannel || config.Options[1].Description != "要發通知的頻道" || !config.Options[1].Required || len(config.Options[1].ChannelTypes) != 0 {
		t.Fatalf("channel option = %#v", config.Options[1])
	}
	if config.Options[2].Name != optionEveryoneCanSet || config.Options[2].Type != commands.OptionTypeBoolean || config.Options[2].Description != "使用者是否可以自己設定自己的生日日期跟通知時間" || !config.Options[2].Required {
		t.Fatalf("can-set option = %#v", config.Options[2])
	}
	if config.Options[3].Name != optionUTC || config.Options[3].Type != commands.OptionTypeString || config.Options[3].Description != "設定屬於你伺服器的時區(台灣、香港、新加坡、馬來西亞、中國是UTC+8，日本是UTC+9) " || !config.Options[3].Required || len(config.Options[3].Choices) != 24 || config.Options[3].Choices[0].Name != "UTC+0" || config.Options[3].Choices[8].Value != "+08:00" || config.Options[3].Choices[23].Name != "UTC+23" {
		t.Fatalf("utc option = %#v", config.Options[3])
	}
	if config.Options[4].Name != optionRole || config.Options[4].Type != commands.OptionTypeRole || config.Options[4].Description != "當使用者那天生日時自動給予身分組，在當天結束後自動刪除" || config.Options[4].Required {
		t.Fatalf("role option = %#v", config.Options[4])
	}
	for hour, choice := range config.Options[3].Choices {
		if choice.Name != fmt.Sprintf("UTC+%d", hour) || choice.Value != fmt.Sprintf("+%02d:00", hour) {
			t.Fatalf("utc choice %d = %#v", hour, choice)
		}
	}

	add := definition.Options[1]
	if add.Type != commands.OptionTypeSubCommand || add.Name != subcommandAdd || add.Description != "新增使用者的生日資料" || len(add.Options) != 4 {
		t.Fatalf("add subcommand = %#v", add)
	}
	if add.Options[0].Name != optionBirthdayMonth || add.Options[0].Type != commands.OptionTypeInteger || add.Options[0].Description != "格式為mm ex: 1(1月出生) ex:12(12月出生)" || !add.Options[0].Required ||
		add.Options[1].Name != optionBirthdayDay || add.Options[1].Type != commands.OptionTypeInteger || add.Options[1].Description != "格式為dd ex: 1(1日出生) 31(31日出生)" || !add.Options[1].Required ||
		add.Options[2].Name != optionUser || add.Options[2].Type != commands.OptionTypeUser || add.Options[2].Description != "要設定的使用者(需有管理員權限)" || add.Options[2].Required ||
		add.Options[3].Name != optionBirthdayYear || add.Options[3].Type != commands.OptionTypeInteger || add.Options[3].Description != "格式為yyyy ex: 2023(2023年出生)" || add.Options[3].Required {
		t.Fatalf("add options = %#v", add.Options)
	}

	deleteOption := definition.Options[2]
	if deleteOption.Type != commands.OptionTypeSubCommand || deleteOption.Name != subcommandDelete || deleteOption.Description != "刪除某使用者的資料" || len(deleteOption.Options) != 1 || deleteOption.Options[0].Name != optionUser || deleteOption.Options[0].Type != commands.OptionTypeUser || deleteOption.Options[0].Description != "輸入你要刪除的使用者!" || !deleteOption.Options[0].Required {
		t.Fatalf("delete subcommand = %#v", deleteOption)
	}

	allowAdmin := definition.Options[3]
	if allowAdmin.Type != commands.OptionTypeSubCommand || allowAdmin.Name != subcommandAllowAdmin || allowAdmin.Description != "是否允許管理員設定你的生日 防止打擾到你(預設為true)" || len(allowAdmin.Options) != 1 || allowAdmin.Options[0].Name != optionAllowAdmin || allowAdmin.Options[0].Type != commands.OptionTypeBoolean || allowAdmin.Options[0].Description != "是否允許!true為允許 false為不允許" || !allowAdmin.Options[0].Required {
		t.Fatalf("allow-admin subcommand = %#v", allowAdmin)
	}

	list := definition.Options[4]
	if list.Type != commands.OptionTypeSubCommand || list.Name != subcommandList || list.Description != "這個伺服器內的使用者生日列表" || len(list.Options) != 0 {
		t.Fatalf("list subcommand = %#v", list)
	}
}
