package onboarding

import (
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/commands"
)

const (
	JoinRoleSetCommandName     = "加入身份組設置"
	JoinRoleDeleteCommandName  = "加入身份組刪除"
	JoinMessageSetCommandName  = "加入訊息設置"
	LeaveMessageSetCommandName = "退出訊息設置"
	VerificationCommandName    = "驗證"
	VerificationSetCommandName = "驗證設置"
	AccountAgeCommandName      = "帳號需創建時數"
	manageMessagesPermission   = "8192"
)

func Definitions() []commands.Definition {
	return JoinRoleDefinitions()
}

func JoinRoleDefinitions() []commands.Definition {
	return []commands.Definition{JoinRoleSetDefinition(), JoinRoleDeleteDefinition()}
}

func MessageDefinitions() []commands.Definition {
	return []commands.Definition{JoinMessageDefinition(), LeaveMessageDefinition()}
}

func VerificationDefinitions() []commands.Definition {
	return []commands.Definition{VerificationSetDefinition()}
}

func VerificationFlowDefinitions() []commands.Definition {
	return []commands.Definition{VerificationDefinition()}
}

func AccountAgeDefinitions() []commands.Definition {
	return []commands.Definition{AccountAgeDefinition()}
}

func JoinRoleSetDefinition() commands.Definition {
	return commands.Definition{
		Type:        commands.CommandTypeChatInput,
		Name:        JoinRoleSetCommandName,
		Description: "設定玩家加入時要給甚麼身份組",
		DocsURL:     "https://docsmhcat.yorukot.meocs/join_role",
		Ownership:   commands.ManagedOwnership("join-role-config", commands.ScopeGuild),
		Options: []commands.Option{
			{
				Type:        commands.OptionTypeRole,
				Name:        "身分組",
				Description: "輸入身分組!",
				Required:    true,
			},
			{
				Type:        commands.OptionTypeString,
				Name:        "給人還是給機器人",
				Description: "請選擇(預設為給所有人)!",
				Choices: []commands.Choice{
					{Name: "給全部人", Value: domain.JoinRoleGiveAllUsers},
					{Name: "機器人", Value: domain.JoinRoleGiveBots},
					{Name: "成員", Value: domain.JoinRoleGiveMembers},
				},
			},
		},
	}
}

func JoinRoleDeleteDefinition() commands.Definition {
	return commands.Definition{
		Type:        commands.CommandTypeChatInput,
		Name:        JoinRoleDeleteCommandName,
		Description: "刪除之前設定的加入身份組",
		DocsURL:     "https://docsmhcat.yorukot.me/docs/join_role_delete",
		Ownership:   commands.ManagedOwnership("join-role-config", commands.ScopeGuild),
		Options: []commands.Option{{
			Type:        commands.OptionTypeRole,
			Name:        "身分組",
			Description: "輸入身分組!",
			Required:    true,
		}},
	}
}

func JoinMessageDefinition() commands.Definition {
	return commands.Definition{
		Type:                     commands.CommandTypeChatInput,
		Name:                     JoinMessageSetCommandName,
		Description:              "設定玩家加入時發送甚麼訊息",
		DefaultMemberPermissions: stringPtr(manageMessagesPermission),
		DocsURL:                  "https://docsmhcat.yorukot.meocs/join_message",
		Ownership:                commands.ManagedOwnership("welcome-message-config", commands.ScopeGuild),
		Options: []commands.Option{{
			Type:         commands.OptionTypeChannel,
			Name:         "頻道",
			Description:  "輸入加入訊息要在那發送!",
			Required:     true,
			ChannelTypes: []int{0, 5},
		}},
	}
}

func LeaveMessageDefinition() commands.Definition {
	return commands.Definition{
		Type:                     commands.CommandTypeChatInput,
		Name:                     LeaveMessageSetCommandName,
		Description:              "設定玩家退出時發送甚麼訊息",
		DefaultMemberPermissions: stringPtr(manageMessagesPermission),
		Ownership:                commands.ManagedOwnership("welcome-message-config", commands.ScopeGuild),
		Options: []commands.Option{{
			Type:         commands.OptionTypeChannel,
			Name:         "頻道",
			Description:  "輸入加入訊息要在那發送!",
			Required:     true,
			ChannelTypes: []int{0, 5},
		}},
	}
}

func VerificationSetDefinition() commands.Definition {
	return commands.Definition{
		Type:        commands.CommandTypeChatInput,
		Name:        VerificationSetCommandName,
		Description: "設置驗證完成後要給甚麼身份組",
		DocsURL:     "https://docsmhcat.yorukot.me/commands/announcement.html",
		Ownership:   commands.ManagedOwnership("verification-config", commands.ScopeGuild),
		Options: []commands.Option{
			{
				Type:        commands.OptionTypeRole,
				Name:        "身分組",
				Description: "輸入身份組!",
				Required:    true,
			},
			{
				Type:        commands.OptionTypeString,
				Name:        "改名",
				Description: "輸入名稱，{name}代表原本的名稱ex:平名 | {name} 就會變成 平名 | 夜貓",
			},
		},
	}
}

func VerificationDefinition() commands.Definition {
	return commands.Definition{
		Type:        commands.CommandTypeChatInput,
		Name:        VerificationCommandName,
		Description: "確保你不是機器人",
		DocsURL:     "https://docsmhcat.yorukot.me/commands/announcement.html",
		Ownership:   commands.ManagedOwnership("verification-flow", commands.ScopeGuild),
	}
}

func AccountAgeDefinition() commands.Definition {
	return commands.Definition{
		Type:        commands.CommandTypeChatInput,
		Name:        AccountAgeCommandName,
		Description: "設定用戶要創建多久才能加入這個伺服器",
		Ownership:   commands.ManagedOwnership("account-age-config", commands.ScopeGuild),
		Options: []commands.Option{
			{
				Type:        commands.OptionTypeSubCommand,
				Name:        "小時數",
				Description: "設定用戶需要滿幾小時才能夠進入伺服器",
				Options: []commands.Option{{
					Type:        commands.OptionTypeInteger,
					Name:        "小時數",
					Description: "輸入當未滿幾個小時時要自動踢出!",
					Required:    true,
				}},
			},
			{
				Type:        commands.OptionTypeSubCommand,
				Name:        "被踢出資訊頻道",
				Description: "當有人因為未滿創建時數被踢出時要在哪裡發送",
				Options: []commands.Option{{
					Type:         commands.OptionTypeChannel,
					Name:         "頻道",
					Description:  "設定因未達創建時數而被踢出的使用者資訊!",
					Required:     true,
					ChannelTypes: []int{0, 5},
				}},
			},
			{
				Type:        commands.OptionTypeSubCommand,
				Name:        "創建時數刪除",
				Description: "刪除之前設定的小時數以及被踢出後再發的頻道",
			},
			{
				Type:        commands.OptionTypeSubCommand,
				Name:        "被踢出資訊頻道刪除",
				Description: "刪除之前的設定發送頻道",
			},
		},
	}
}

func stringPtr(value string) *string {
	return &value
}
