package work

import "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/commands"

const CommandName = "打工系統"

func Definitions() []commands.Definition {
	return []commands.Definition{Definition()}
}

func Definition() commands.Definition {
	return commands.Definition{
		Type:        commands.CommandTypeChatInput,
		Name:        CommandName,
		Description: "用自己的心血來獲得一些獎勵吧!",
		DocsURL:     "https://docsmhcat.yorukot.me/docs/allcommands/%E6%89%93%E5%B7%A5%E7%B3%BB%E7%B5%B1/user_work",
		Ownership:   commands.ManagedOwnership("work", commands.ScopeGuild),
		Options: []commands.Option{
			{
				Type:        commands.OptionTypeSubCommand,
				Name:        "打工系統設定",
				Description: "設定打工系統的各項設定，伺服器第一次使用請使用這個指令!",
				Options: []commands.Option{
					{Type: commands.OptionTypeInteger, Name: "每天可獲得多少精力", Description: "每天可以獲得多少精力(每天24點發送)!", Required: true},
					{Type: commands.OptionTypeInteger, Name: "精力上限為多少", Description: "每人的精力上限最多是多少!", Required: true},
					{Type: commands.OptionTypeBoolean, Name: "是否需要驗證", Description: "防止有人使用腳本進行一些掛機行為!", Required: false},
				},
			},
			{
				Type:        commands.OptionTypeSubCommand,
				Name:        "新增打工事項",
				Description: "新增打工的事項",
				Options: []commands.Option{
					{Type: commands.OptionTypeString, Name: "打工地點名稱", Description: "打工地點名稱!(重複的話會自動刪除舊的)", Required: true},
					{Type: commands.OptionTypeNumber, Name: "耗費時間", Description: "打工一次需要耗費多少時間(小時為單位)!", Required: true},
					{Type: commands.OptionTypeInteger, Name: "耗費精力", Description: "打工一次需耗費多少的精力!", Required: true},
					{Type: commands.OptionTypeInteger, Name: "取得代幣", Description: "打工一次可取得多少代幣!", Required: true},
					{Type: commands.OptionTypeRole, Name: "身分組", Description: "允許的身分組(除了這個身分組其他所有人都不能用這個打工)!", Required: false},
				},
			},
			{
				Type:        commands.OptionTypeSubCommand,
				Name:        "打工事項刪除",
				Description: "刪除打工事項",
				Options: []commands.Option{
					{Type: commands.OptionTypeString, Name: "打工地點名稱", Description: "輸入打工地點名稱!", Required: true},
				},
			},
			{
				Type:        commands.OptionTypeSubCommand,
				Name:        "打工介面",
				Description: "在這裡一般使用者可以使用所有的東西!!",
			},
			{
				Type:        commands.OptionTypeSubCommand,
				Name:        "增加個人精力",
				Description: "增加個人的精力!!",
				Options: []commands.Option{
					{Type: commands.OptionTypeUser, Name: "使用者", Description: "輸入要給精力的使用者!", Required: true},
					{Type: commands.OptionTypeInteger, Name: "要給多少精力", Description: "輸入要給多少精力!", Required: true},
				},
			},
			{
				Type:        commands.OptionTypeSubCommand,
				Name:        "增加全體精力",
				Description: "增加全伺服器的精力!!",
				Options: []commands.Option{
					{Type: commands.OptionTypeInteger, Name: "要給多少精力", Description: "輸入要給多少精力!", Required: true},
				},
			},
		},
	}
}
