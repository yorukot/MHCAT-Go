package utility

import (
	"fmt"
	"strings"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/interactions"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/responses"
)

const (
	legacyHelpMenuCustomID = "helphelphelphelpmenu"
	legacyHelpColor        = 0x5865F2
)

type legacyHelpCategory struct {
	Name        string
	Description string
	Emoji       string
	Commands    []legacyHelpCommand
}

type legacyHelpCommand struct {
	Name        string
	Description string
	Emoji       string
	DocsURL     string
	UserPerms   string
}

var legacyHelpCategories = []legacyHelpCategory{
	{Name: "代幣系統", Description: "類似於各位常見的經濟系統", Emoji: "<:money:997374193026994236>", Commands: []legacyHelpCommand{
		{Name: "代幣增加", Description: "改變扭蛋數量", Emoji: "<:income:997374186794258452>", DocsURL: "https://docsmhcat.yorukot.meocs/coin_increase"},
		{Name: "代幣查詢", Description: "查詢你有多少代幣", Emoji: "<:money:997374193026994236>", DocsURL: "https://docsmhcat.yorukot.meocs/coin"},
		{Name: "coin-related-settings", Description: "Various settings related to tokens", Emoji: "<:coins:997374177944281190>", DocsURL: "https://docsmhcat.yorukot.meocs/required_coins", UserPerms: "訊息管理"},
		{Name: "代幣排行榜", Description: "查詢代幣的排行榜", Emoji: "<:levelup:990254382845157406>"},
		{Name: "代幣重製", Description: "重製所有人的代幣，或者是進行代幣改變幣值", Emoji: "<:money:997374193026994236>", DocsURL: "https://docs.mhcat.xyz/docs/coin"},
		{Name: "代幣遊戲", Description: "遊玩有關代幣的小遊戲", Emoji: "<:blackjack1:1005469910689923142>"},
		{Name: "代幣商店", Description: "使用你所賺到的代幣買一些特別的東西吧!", Emoji: "<:store:1001118704651743372>"},
		{Name: "剪刀石頭布", Description: "跟電腦剪刀時候布來獲得代幣(有賺有賠)", Emoji: "<:coins:997374177944281190>", DocsURL: "https://docsmhcat.yorukot.me/docs/required_coins"},
		{Name: "簽到", Description: "簽到來獲得代幣", Emoji: "<:sign:997374180632825896>", DocsURL: "https://docsmhcat.yorukot.me/docs/snig"},
		{Name: "簽到列表", Description: "查看今天有誰簽到了", Emoji: "<:sign:997374180632825896>", DocsURL: "https://docsmhcat.yorukot.me/docs/snig"},
		{Name: "my-profile", Description: "Check about data in specific server!!", Emoji: "<:sign:997374180632825896>", DocsURL: "https://docsmhcat.yorukot.me/docs/snig"},
	}},
	{Name: "公告系統", Description: "讓你的公告不再是普通的文字", Emoji: "<:megaphone:985943890148327454>", Commands: []legacyHelpCommand{
		{Name: "公告發送", Description: "發送公告訊息", Emoji: "<:megaphone:985943890148327454>", DocsURL: "https://docsmhcat.yorukot.me/docs/ann"},
		{Name: "公告頻道設置", Description: "設定公告在哪發送", Emoji: "<:configuration:984010500608249886>"},
	}},
	{Name: "加入設置", Description: "當玩加入或離開時，要做一些動作", Emoji: "🪂", Commands: []legacyHelpCommand{
		{Name: "加入訊息設置", Description: "設定玩家加入時發送甚麼訊息", Emoji: "<:comments:985944111725019246>"},
		{Name: "加入身份組設置", Description: "設定玩家加入時要給甚麼身份組", Emoji: "<:roleplaying:985945121264635964>"},
		{Name: "加入身份組刪除", Description: "刪除之前設定的加入身份組", Emoji: "<:delete:985944877663678505>", DocsURL: "https://docsmhcat.yorukot.me/docs/join_role_delete"},
		{Name: "退出訊息設置", Description: "設定玩家退出時發送甚麼訊息", Emoji: "<:comments:985944111725019246>"},
		{Name: "驗證", Description: "確保你不是機器人", Emoji: "<:tickmark:985949769224556614>"},
		{Name: "驗證設置", Description: "設置驗證完成後要給甚麼身份組", Emoji: "<:configuration:984010500608249886>"},
	}},
	{Name: "實用工具", Description: "查看一些實用的功能", Emoji: "<:bestpractice:986070549115596950>", Commands: []legacyHelpCommand{
		{Name: "自動聊天頻道", Description: "設定自動聊天頻道要在哪裡發送", Emoji: "<:configuration:984010500608249886>", DocsURL: "https://docsmhcat.yorukot.me/docs/chat_xp_set"},
		{Name: "自動聊天頻道刪除", Description: "刪除自動聊天頻道要在哪裡發送", Emoji: "<:delete:985944877663678505>", DocsURL: "https://docsmhcat.yorukot.me/docs/chat_xp_set"},
		{Name: "help", Description: "使用我開始使用", Emoji: "<:help:985948179709186058>", DocsURL: "https://docsmhcat.yorukot.me/docs/help"},
		{Name: "info", Description: "Check all informations.", Emoji: "<:info:985946738403737620>"},
		{Name: "ping", Description: "查看我的ping", Emoji: "<:icons_goodping:1084881470075703367>"},
		{Name: "翻譯", Description: "翻譯成各種語言", Emoji: "<:help:985948179709186058>", DocsURL: "https://docsmhcat.yorukot.me/docs/translate"},
	}},
	{Name: "打工系統", Description: "可以在閒暇之餘做些工作賺些代幣", Emoji: "<:working:1048617967799242772>", Commands: []legacyHelpCommand{
		{Name: "打工系統", Description: "用自己的心血來獲得一些獎勵吧!", Emoji: "<:working:1048617967799242772>"},
	}},
	{Name: "扭蛋系統", Description: "使用簽到以及聊天獲得的代幣進行扭蛋", Emoji: "<:vendingmachine:997374191651274823>", Commands: []legacyHelpCommand{
		{Name: "扭蛋", Description: "進行扭蛋，有機會抽中各種大獎喔!!!!", Emoji: "<:gashapon:997374176526610472>"},
		{Name: "扭蛋獎池刪除", Description: "刪除扭蛋的獎池", Emoji: "<:trashbin:995991389043163257>"},
		{Name: "扭蛋獎品編輯", Description: "增加扭蛋的獎池裡的獎品的數量", Emoji: "<:add1:981722904251215872>", DocsURL: "https://docsmhcat.yorukot.me/docs/prize_add"},
		{Name: "扭蛋獎池增加", Description: "增加扭蛋的獎池", Emoji: "<:add1:981722904251215872>", DocsURL: "https://docsmhcat.yorukot.me/docs/prize_add"},
		{Name: "扭蛋獎池查詢", Description: "增加扭蛋的獎池", Emoji: "<:list:992002476360343602>", DocsURL: "https://docsmhcat.yorukot.me/docs/prize_search"},
	}},
	{Name: "抽獎系統", Description: "一起來進行抽獎，來獲得非常棒的獎品吧", Emoji: "<:lottery:985946439253381200>", Commands: []legacyHelpCommand{
		{Name: "抽獎設置", Description: "設置抽獎訊息", Emoji: "<:lottery:985946439253381200>"},
	}},
	{Name: "生日系統", Description: "當有人生日的時候給他一些祝福吧", Emoji: "<:cake:1065654305983570041>", Commands: []legacyHelpCommand{
		{Name: "生日系統", Description: "讓你的伺服器可以為生日的人慶生!", Emoji: "<:working:1048617967799242772>"},
	}},
	{Name: "私人頻道", Description: "一個簡單的客服頻道系統", Emoji: "<:ticket:985945491093205073>", Commands: []legacyHelpCommand{
		{Name: "私人頻道設置", Description: "設置私人頻道", Emoji: "<:ticket:985945491093205073>"},
		{Name: "私人頻道刪除", Description: "刪除之前設置的私人頻道", Emoji: "<:delete:985944877663678505>"},
	}},
	{Name: "管理系統", Description: "管理員一定要看這個，非常適合管理伺服器", Emoji: "<:manager:986069915129769994>", Commands: []legacyHelpCommand{
		{Name: "查看餘額", Description: "查看剩餘餘額", Emoji: "<:logfile:985948561625710663>"},
		{Name: "刪除訊息", Description: "刪除大量訊息", Emoji: "<:delete:985944877663678505>"},
		{Name: "set-log-channel", Description: "Set where log messages should send", Emoji: "<:logfile:985948561625710663>"},
		{Name: "刪除資料", Description: "刪除之前設置過的資料", Emoji: "<:logfile:985948561625710663>"},
		{Name: "兌換", Description: "兌換代碼", Emoji: "<:logfile:985948561625710663>"},
		{Name: "投票創建", Description: "創建一個萬能的投票", Emoji: "<:logfile:985948561625710663>"},
		{Name: "選取身分組-按鈕", Description: "設定領取身分組的消息(點按鈕自動增加身分組)", Emoji: "<:add:985948803469279303>"},
		{Name: "選取身分組-表情符號", Description: "設定領取身分組的消息-點按鈕自動增加身分組(如要更改某個表情符號所給予的身分組，請一樣打這個指令)", Emoji: "<:add:985948803469279303>"},
		{Name: "選取身分組刪除-表情符號", Description: "選取身分組刪除-表情符號版(進行刪除)", Emoji: "<:add:985948803469279303>"},
	}},
	{Name: "統計系統", Description: "頻道變身成一個伺服器資料報告", Emoji: "<:statistics:986108146747600928>", Commands: []legacyHelpCommand{
		{Name: "統計系統查詢", Description: "查詢統計消息", Emoji: "<:searching:986107902777491497>"},
		{Name: "統計系統創建", Description: "創建統計消息", Emoji: "<:statistics:986108146747600928>"},
		{Name: "統計系統刪除", Description: "刪除統計消息", Emoji: "<:delete1:986068526387314690>"},
		{Name: "統計身分組人數", Description: "統計某個特定的身分組的人數", Emoji: "<:statistics:986108146747600928>"},
	}},
	{Name: "經驗系統", Description: "各種經驗系統，例如語音經驗及聊天經驗", Emoji: "<:level1:985947371957547088>", Commands: []legacyHelpCommand{
		{Name: "經驗值重製", Description: "重製整個伺服器的經驗", Emoji: "<:onehour:1000310711941087293>"},
		{Name: "聊天經驗身分組設定", Description: "設定聊天經驗通知要在哪發送", Emoji: "<:configuration:984010500608249886>"},
		{Name: "聊天排行榜", Description: "查詢聊天經驗的排行榜", Emoji: "<:level1:985947371957547088>"},
		{Name: "聊天經驗設定", Description: "設定聊天經驗通知要在哪發送", Emoji: "<:configuration:984010500608249886>", DocsURL: "https://docsmhcat.yorukot.me/docs/chat_xp_set"},
		{Name: "聊天經驗刪除", Description: "刪除聊天經驗發送訊息設置", Emoji: "<:delete:985944877663678505>", DocsURL: "https://docsmhcat.yorukot.me/docs/chat_xp_delete"},
		{Name: "聊天經驗", Description: "查詢聊天經驗", Emoji: "<:level1:985947371957547088>"},
		{Name: "語音經驗身分組設定", Description: "設定語音經驗通知要在哪發送(兼增加、刪除、設定查詢)", Emoji: "<:configuration:984010500608249886>", DocsURL: "https://docsmhcat.yorukot.me/docs/chat_xp_set"},
		{Name: "語音排行榜", Description: "查詢語音經驗的排行榜", Emoji: "<:level1:985947371957547088>"},
		{Name: "語音經驗設定", Description: "設定語音經驗通知要在哪發送", Emoji: "<:configuration:984010500608249886>", DocsURL: "https://docsmhcat.yorukot.me/docs/voice_xp_set"},
		{Name: "語音經驗刪除", Description: "刪除語音發送訊息設置", Emoji: "<:delete:985944877663678505>", DocsURL: "https://docsmhcat.yorukot.me/docs/voice_xp_delete"},
		{Name: "語音經驗", Description: "查詢語音經驗", Emoji: "<:level1:985947371957547088>"},
		{Name: "經驗值改變", Description: "增加某人的經驗值", Emoji: "<:onehour:1000310711941087293>"},
	}},
	{Name: "群組防護", Description: "防止群組被炸，被各種爆破", Emoji: "<:shield:1000309930043133992>", Commands: []legacyHelpCommand{
		{Name: "帳號需創建時數", Description: "設定用戶要創建多久才能加入這個伺服器", Emoji: "<:onehour:1000310711941087293>"},
		{Name: "防詐騙網址", Description: "設定是否開啟防詐騙網址功能(輸入這個指令就會更改)", Emoji: "<:fraudalert:1000408260777611355>"},
		{Name: "詐騙網址回報", Description: "回報詐騙網站", Emoji: "<:fraudalert:1000408260777611355>"},
	}},
	{Name: "自動通知", Description: "在一個特定的時間發送通知", Emoji: "<:alarmclock:997415306530131980>", Commands: []legacyHelpCommand{
		{Name: "自動通知刪除", Description: "刪除之前設定的自動通知", Emoji: "<:trashbin:995991389043163257>", DocsURL: "https://docsmhcat.yorukot.me/docs/chat_xp_set"},
		{Name: "自動通知列表", Description: "查看所有的自動通知列表", Emoji: "<:list:992002476360343602>", DocsURL: "https://docsmhcat.yorukot.me/docs/chat_xp_set"},
		{Name: "automatic-notification", Description: "Set where automatic notification should be send", Emoji: "<:configuration:984010500608249886>", DocsURL: "https://docsmhcat.yorukot.me/docs/chat_xp_set"},
	}},
	{Name: "語音包廂", Description: "創建一個可活動的語音頻道", Emoji: "🔉", Commands: []legacyHelpCommand{
		{Name: "上鎖頻道", Description: "設定語音包廂密碼", Emoji: "<:mapsandflags:985949507114131587>"},
		{Name: "語音包廂設置", Description: "設定語音包廂", Emoji: "<:mapsandflags:985949507114131587>"},
		{Name: "語音包廂刪除", Description: "刪除語音包廂設置", Emoji: "<:delete:985944877663678505>"},
	}},
	{Name: "警告系統", Description: "讓你的警告可以進行紀錄，還可自動停權", Emoji: "<:warning:985590881698590730>", Commands: []legacyHelpCommand{
		{Name: "警告設定", Description: "警告的各種設定", Emoji: "<:configuration:984010500608249886>"},
		{Name: "警告全部清除", Description: "清除一個使用者的全部警告", Emoji: "<:trashbin:986308183674990592>"},
		{Name: "警告清除", Description: "清除一個使用者的某個警告", Emoji: "<:delete1:986068526387314690>"},
		{Name: "警告", Description: "警告一個使用者", Emoji: "<:warning:985590881698590730>"},
		{Name: "警告紀錄", Description: "收尋一位使用者的警告", Emoji: "<:searching:986107902777491497>"},
	}},
}

func legacyHelpOverview(interaction interactions.Interaction) responses.Message {
	fields := make([]responses.EmbedField, 0, len(legacyHelpCategories))
	options := make([]responses.SelectOption, 0, len(legacyHelpCategories))
	for _, category := range legacyHelpCategories {
		fields = append(fields, responses.EmbedField{
			Name:   fmt.Sprintf("%s - %s", category.Emoji, category.Name),
			Value:  fmt.Sprintf("`%s`", category.Description),
			Inline: true,
		})
		options = append(options, responses.SelectOption{
			Label:       strings.ToUpper(category.Name),
			Value:       strings.ToLower(category.Name),
			Description: category.Description,
			Emoji:       category.Emoji,
		})
	}
	return responses.Message{
		Embeds: []responses.Embed{{
			Author: &responses.EmbedAuthor{
				Name:    "MHCAT",
				IconURL: "https://i.imgur.com/AQAodBA.png",
				URL:     "https://discord.com/api/oauth2/authorize?client_id=964185876559196181&permissions=8&scope=bot%20applications.commands",
			},
			Description: "**<a:cool:984263702897360897> 嗨嗨，你發現了酷東西\n使用我來讓你的discord更棒!!\n想要了解某個類別請使用下方的選單\n如要查看特定的指令請使用`/help 指令名稱`\n\n<:9605discordslashcommand:982559784429563925> 指令一律使用斜線命令，只需打`/指令名稱`即可使用**\n\n<a:buycoffeeforme:986560638304256051> [幫我買杯咖啡!](https://www.buymeacoffee.com/mhcat)\n\n[隱私權聲明](https://docsmhcat.yorukot.me/terms/privacy_policy) [服務條款](https://docsmhcat.yorukot.me/terms/Terms_of_Service)",
			Color:       legacyHelpColor,
			Footer:      legacyFooter(interaction),
			Fields:      fields,
		}},
		Components: []responses.ComponentRow{
			{Components: []responses.Component{{
				Type:        responses.ComponentTypeSelect,
				CustomID:    legacyHelpMenuCustomID,
				Placeholder: "📜 選擇命令類別",
				Options:     options,
			}}},
			{Components: []responses.Component{
				{Type: responses.ComponentTypeButton, Style: responses.ButtonStyleLink, URL: "https://dsc.gg/mhcat", Label: "邀請我", Emoji: "<a:catjump:984807173529931837>"},
				{Type: responses.ComponentTypeButton, Style: responses.ButtonStyleLink, URL: "https://discord.gg/7g7VE2Sqna", Label: "支援伺服器", Emoji: "<:customerservice:986268421144592415>"},
				{Type: responses.ComponentTypeButton, Style: responses.ButtonStyleLink, URL: "https://mhcat.yorukot.me", Label: "官方網站", Emoji: "<:worldwideweb:986268131284627507>"},
			}},
		},
	}
}

func legacyHelpCategoryMessage(interaction interactions.Interaction, selected string) (responses.Message, bool) {
	selected = strings.ToLower(strings.TrimSpace(selected))
	for _, category := range legacyHelpCategories {
		if strings.ToLower(category.Name) != selected {
			continue
		}
		fields := make([]responses.EmbedField, 0, len(category.Commands))
		for _, command := range category.Commands {
			name := fmt.Sprintf("%s`%s`", command.Emoji, command.Name)
			value := command.Description
			if value == "" {
				value = "沒有說明"
			}
			fields = append(fields, responses.EmbedField{
				Name:   name,
				Value:  fmt.Sprintf("```fix\n%s```", value),
				Inline: true,
			})
		}
		return responses.Message{
			Embeds: []responses.Embed{{
				Title:       fmt.Sprintf("__%s %s 指令!__", category.Emoji, category.Name),
				Description: "> 使用 `/help 指令名稱:` 以獲取有關指令的更多信息!\n> 例: `/help 指令名稱:公告發送`\n\n",
				Color:       legacyHelpColor,
				Footer:      legacyFooter(interaction),
				Fields:      fields,
			}},
			Ephemeral: true,
		}, true
	}
	return responses.Message{}, false
}

func legacyHelpCommandDetail(interaction interactions.Interaction, query string) (responses.Message, bool) {
	query = strings.ToLower(strings.TrimSpace(query))
	for _, category := range legacyHelpCategories {
		for _, command := range category.Commands {
			if strings.ToLower(command.Name) != query {
				continue
			}
			perms := command.UserPerms
			if perms == "" {
				perms = "這個指令大家都可以用喔"
			}
			docs := "此指令目前沒有教學"
			if command.DocsURL != "" {
				docs = fmt.Sprintf("[__**點我立刻前往教學頁面**__](%s)", command.DocsURL)
			}
			return responses.Message{
				Embeds: []responses.Embed{{
					Title:  "**<:9605discordslashcommand:982559784429563925> 指令資料**",
					Color:  legacyHelpColor,
					Footer: legacyFooter(interaction),
					Fields: []responses.EmbedField{
						{Name: "<:id:985950321975128094>**指令名稱:**", Value: fmt.Sprintf("```%s```", command.Name)},
						{Name: "<:editinfo:985950967566569503>**指令說明:**", Value: fmt.Sprintf("```%s```", command.Description)},
						{Name: "<:key:986059580821868544>**指令權限需求(用戶需要有甚麼權限才能使用):**", Value: fmt.Sprintf("```%s```", perms)},
						{Name: "<:creativeteaching:986060052949524600>**指令文檔教學:**", Value: docs},
					},
				}},
			}, true
		}
	}
	return responses.Message{}, false
}

func legacyHelpInvalidCommand() responses.Message {
	return responses.Message{
		Embeds: []responses.Embed{{
			Title: "無效的指令! 使用 `/help` 查看所有指令!",
			Color: 0xEA0000,
		}},
	}
}

func legacyFooter(interaction interactions.Interaction) *responses.EmbedFooter {
	text := "使用者的查詢"
	if interaction.Actor.UserTag != "" {
		text = interaction.Actor.UserTag + "的查詢"
	}
	return &responses.EmbedFooter{Text: text, IconURL: interaction.Actor.AvatarURL}
}
