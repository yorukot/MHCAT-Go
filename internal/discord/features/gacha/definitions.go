package gacha

import "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/commands"

const (
	GachaPrizeCreateCommandName = "扭蛋獎池增加"
	GachaPrizeListCommandName   = "扭蛋獎池查詢"
	GachaPrizeDeleteCommandName = "扭蛋獎池刪除"

	gachaManageMessagesPermission = "8192"
	gachaPrizeNameOption          = "獎品名稱"
	gachaPrizeChanceOption        = "機率"
	gachaPrizeCodeOption          = "獎品代碼"
	gachaPrizeAutoDeleteOption    = "自動刪除"
	gachaPrizeCountOption         = "獎品數量"
	gachaPrizeGiveCoinOption      = "給予硬幣"
)

func Definitions() []commands.Definition {
	return PrizeListDefinitions()
}

func AllDefinitions() []commands.Definition {
	definitions := PrizeListDefinitions()
	definitions = append(definitions, PrizeCreateDefinitions()...)
	definitions = append(definitions, PrizeDeleteDefinitions()...)
	return definitions
}

func PrizeListDefinitions() []commands.Definition {
	return []commands.Definition{PrizeListDefinition()}
}

func PrizeDeleteDefinitions() []commands.Definition {
	return []commands.Definition{PrizeDeleteDefinition()}
}

func PrizeCreateDefinitions() []commands.Definition {
	return []commands.Definition{PrizeCreateDefinition()}
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

func PrizeCreateDefinition() commands.Definition {
	return commands.Definition{
		Type:                     commands.CommandTypeChatInput,
		Name:                     GachaPrizeCreateCommandName,
		Description:              "增加扭蛋的獎池",
		DocsURL:                  "https://docsmhcat.yorukot.me/docs/prize_add",
		DefaultMemberPermissions: stringPtr(gachaManageMessagesPermission),
		Ownership:                commands.ManagedOwnership("gacha-prize-create", commands.ScopeGuild),
		Options: []commands.Option{
			{
				Type:        commands.OptionTypeString,
				Name:        gachaPrizeNameOption,
				Description: "輸入這個獎品叫甚麼，以及簡單概述",
				Required:    true,
			},
			{
				Type:        commands.OptionTypeNumber,
				Name:        gachaPrizeChanceOption,
				Description: "輸入中獎機率(百分比)ex:10 代表10% 0.1代表0.1%",
				Required:    true,
			},
			{
				Type:        commands.OptionTypeString,
				Name:        gachaPrizeCodeOption,
				Description: "填上獎品的代碼ex:stram序號nitro連結等",
			},
			{
				Type:        commands.OptionTypeBoolean,
				Name:        gachaPrizeAutoDeleteOption,
				Description: "抽中後是否自動刪除(預設為true，如果填否的話連獎品數量都不會變)",
			},
			{
				Type:        commands.OptionTypeInteger,
				Name:        gachaPrizeCountOption,
				Description: "該獎品的數量",
			},
			{
				Type:        commands.OptionTypeInteger,
				Name:        gachaPrizeGiveCoinOption,
				Description: "當抽中後是否要給予代幣",
			},
		},
	}
}

func stringPtr(value string) *string {
	return &value
}
