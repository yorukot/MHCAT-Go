package voice

import "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/commands"

const (
	VoiceRoomSetCommandName    = "語音包廂設置"
	VoiceRoomDeleteCommandName = "語音包廂刪除"
	VoiceRoomLockCommandName   = "上鎖頻道"
	manageMessagesPermission   = "8192"

	optionTriggerChannel = "語音頻道"
	optionRoomName       = "設定頻道名稱"
	optionOwnerLock      = "是否予許房主上鎖"
	optionUserLimit      = "設定人數上限"
	optionChannelOrGroup = "頻道或類別"
	optionLockPassword   = "密碼"
)

func Definitions() []commands.Definition {
	return []commands.Definition{SetDefinition(), DeleteDefinition()}
}

func LockDefinitions() []commands.Definition {
	return []commands.Definition{LockDefinition()}
}

func SetDefinition() commands.Definition {
	return commands.Definition{
		Type:                     commands.CommandTypeChatInput,
		Name:                     VoiceRoomSetCommandName,
		Description:              "設定語音包廂",
		DefaultMemberPermissions: stringPtr(manageMessagesPermission),
		Ownership:                commands.ManagedOwnership("voice-room-config", commands.ScopeGuild),
		Options: []commands.Option{
			{
				Type:         commands.OptionTypeChannel,
				Name:         optionTriggerChannel,
				Description:  "設定哪個頻道加入後會開啟語音包廂",
				Required:     true,
				ChannelTypes: []int{2, 13},
			},
			{
				Type:        commands.OptionTypeString,
				Name:        optionRoomName,
				Description: "設定開啟的語音包廂要叫做甚麼 輸入{name}及代表使用者名稱",
				Required:    true,
			},
			{
				Type:        commands.OptionTypeBoolean,
				Name:        optionOwnerLock,
				Description: "設定是否予許房主將活動語音頻道上鎖(房主打指令即可上鎖)",
				Required:    true,
			},
			{
				Type:        commands.OptionTypeInteger,
				Name:        optionUserLimit,
				Description: "設定頻道人數上限(如果不填，即為無上限)",
			},
		},
	}
}

func DeleteDefinition() commands.Definition {
	return commands.Definition{
		Type:                     commands.CommandTypeChatInput,
		Name:                     VoiceRoomDeleteCommandName,
		Description:              "刪除語音包廂設置",
		DefaultMemberPermissions: stringPtr(manageMessagesPermission),
		Ownership:                commands.ManagedOwnership("voice-room-config", commands.ScopeGuild),
		Options: []commands.Option{{
			Type:        commands.OptionTypeChannel,
			Name:        optionChannelOrGroup,
			Description: "刪除加入某個頻道後會創建新頻道的那個`某個頻道`或是類別裡的所有設定",
			Required:    true,
		}},
	}
}

func LockDefinition() commands.Definition {
	return commands.Definition{
		Type:        commands.CommandTypeChatInput,
		Name:        VoiceRoomLockCommandName,
		Description: "設定語音包廂密碼",
		Ownership:   commands.ManagedOwnership("voice-room-lock", commands.ScopeGuild),
		Options: []commands.Option{{
			Type:        commands.OptionTypeString,
			Name:        optionLockPassword,
			Description: "設定該包廂密碼，如想不設定密碼，可直接忽略此選項",
		}},
	}
}

func stringPtr(value string) *string {
	return &value
}
