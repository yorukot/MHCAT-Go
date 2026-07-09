package poll

import "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/commands"

const manageMessagesPermission = "8192"

func Definitions() []commands.Definition {
	return []commands.Definition{CreateDefinition()}
}

func CreateDefinition() commands.Definition {
	return commands.Definition{
		Type:                     commands.CommandTypeChatInput,
		Name:                     "投票創建",
		Description:              "創建一個萬能的投票",
		DefaultMemberPermissions: stringPtr(manageMessagesPermission),
		Ownership:                commands.ManagedOwnership("poll-wave-a", commands.ScopeGuild),
		Options: []commands.Option{
			{
				Type:        commands.OptionTypeString,
				Name:        "問題",
				Description: "輸入你要問的問題!ex:我要買甚麼?",
				Required:    true,
			},
			{
				Type:        commands.OptionTypeString,
				Name:        "選項",
				Description: "輸入回答的選項，請用^將各個選項分開 ex:電腦^手機^兩個都要^!",
				Required:    true,
			},
		},
	}
}

func stringPtr(value string) *string {
	return &value
}
