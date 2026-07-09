package gacha

import "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/commands"

const (
	GachaPrizeListCommandName   = "扭蛋獎池查詢"
	GachaPrizeDeleteCommandName = "扭蛋獎池刪除"

	gachaManageMessagesPermission = "8192"
	gachaPrizeNameOption          = "獎品名稱"
)

func Definitions() []commands.Definition {
	return PrizeListDefinitions()
}

func AllDefinitions() []commands.Definition {
	return append(PrizeListDefinitions(), PrizeDeleteDefinitions()...)
}

func PrizeListDefinitions() []commands.Definition {
	return []commands.Definition{PrizeListDefinition()}
}

func PrizeDeleteDefinitions() []commands.Definition {
	return []commands.Definition{PrizeDeleteDefinition()}
}

func PrizeListDefinition() commands.Definition {
	return commands.Definition{
		Type:        commands.CommandTypeChatInput,
		Name:        GachaPrizeListCommandName,
		Description: "增加扭蛋的獎池",
		DocsURL:     "https://docsmhcat.yorukot.me/docs/prize_search",
		Ownership:   commands.ManagedOwnership("gacha-prize-list", commands.ScopeGuild),
	}
}

func PrizeDeleteDefinition() commands.Definition {
	return commands.Definition{
		Type:                     commands.CommandTypeChatInput,
		Name:                     GachaPrizeDeleteCommandName,
		Description:              "刪除扭蛋的獎池",
		DocsURL:                  "https://docsmhcat.yorukot.me.xyz/docs/prize_removal",
		DefaultMemberPermissions: stringPtr(gachaManageMessagesPermission),
		Ownership:                commands.ManagedOwnership("gacha-prize-delete", commands.ScopeGuild),
		Options: []commands.Option{{
			Type:        commands.OptionTypeString,
			Name:        gachaPrizeNameOption,
			Description: "輸入這個獎品叫甚麼，以及簡單概述",
			Required:    true,
		}},
	}
}

func stringPtr(value string) *string {
	return &value
}
