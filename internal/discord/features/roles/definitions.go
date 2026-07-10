package roles

import "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/commands"

const (
	RoleButtonCommandName         = "選取身分組-按鈕"
	RoleReactionSetCommandName    = "選取身分組-表情符號"
	RoleReactionDeleteCommandName = "選取身分組刪除-表情符號"
)

func Definitions() []commands.Definition {
	return []commands.Definition{
		RoleButtonDefinition(),
		RoleReactionSetDefinition(),
		RoleReactionDeleteDefinition(),
	}
}

func RoleButtonDefinition() commands.Definition {
	return commands.Definition{
		Type:        commands.CommandTypeChatInput,
		Name:        RoleButtonCommandName,
		Description: "設定領取身分組的消息(點按鈕自動增加身分組)",
		Ownership:   commands.ManagedOwnership("role-selection", commands.ScopeGuild),
		Options: []commands.Option{{
			Type:        commands.OptionTypeRole,
			Name:        "身分組",
			Description: "輸入身分組!",
			Required:    true,
		}},
	}
}

func RoleReactionSetDefinition() commands.Definition {
	return commands.Definition{
		Type:        commands.CommandTypeChatInput,
		Name:        RoleReactionSetCommandName,
		Description: "設定領取身分組的消息-點按鈕自動增加身分組(如要更改某個表情符號所給予的身分組，請一樣打這個指令)",
		Ownership:   commands.ManagedOwnership("role-selection", commands.ScopeGuild),
		Options: []commands.Option{
			{
				Type:        commands.OptionTypeString,
				Name:        "訊息url",
				Description: "輸入訊息的url(對訊息點右鍵按複製訊息連結)!",
				Required:    true,
			},
			{
				Type:        commands.OptionTypeRole,
				Name:        "身分組",
				Description: "輸入要給的身分組!",
				Required:    true,
			},
			{
				Type:        commands.OptionTypeString,
				Name:        "表情符號",
				Description: "請輸入要放在訊息下面的表情符號",
				Required:    true,
			},
		},
	}
}

func RoleReactionDeleteDefinition() commands.Definition {
	return commands.Definition{
		Type:        commands.CommandTypeChatInput,
		Name:        RoleReactionDeleteCommandName,
		Description: "選取身分組刪除-表情符號版(進行刪除)",
		Ownership:   commands.ManagedOwnership("role-selection", commands.ScopeGuild),
		Options: []commands.Option{
			{
				Type:        commands.OptionTypeString,
				Name:        "訊息url",
				Description: "輸入訊息的url(對訊息點右鍵按複製訊息連結)!",
				Required:    true,
			},
			{
				Type:        commands.OptionTypeString,
				Name:        "表情符號",
				Description: "請輸入要放在訊息下面的表情符號",
				Required:    true,
			},
		},
	}
}
