package gacha

import "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/commands"

const GachaPrizeListCommandName = "扭蛋獎池查詢"

func Definitions() []commands.Definition {
	return []commands.Definition{PrizeListDefinition()}
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
