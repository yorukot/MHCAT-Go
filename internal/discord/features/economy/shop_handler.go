package economy

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	coreeconomy "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/services/economy"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/interactions"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/responses"
)

const (
	shopSubcommandAdd       = "商品增加"
	shopSubcommandDelete    = "商品刪除"
	shopSubcommandList      = "商品查詢"
	shopOptionName          = "商品名"
	shopOptionPrice         = "商品所需代幣"
	shopOptionDescription   = "商品描述"
	shopOptionAutoDelete    = "是否要在購買後自動刪除"
	shopOptionCode          = "序號"
	shopOptionRole          = "商品是否為身分組"
	shopOptionCount         = "商品數量"
	shopOptionID            = "商品id"
	shopErrorColor          = 0xED4245
	shopUnauthorizedColor   = 0xEA0000
	shopSuccessColor        = 0x53FF53
	shopManageMessagesBit   = int64(8192)
	shopErrorEmoji          = "<a:Discord_AnimatedNo:1015989839809757295>"
	shopStoreEmoji          = "<:store:1001118704651743372>"
	shopDoneEmoji           = "<a:green_tick:994529015652163614>"
	shopDeleteEmoji         = "<:delete:985944877663678505>"
	shopAddDocsPath         = "allcommands/%E4%BB%A3%E5%B9%A3%E7%B3%BB%E7%B5%B1/ghp_shop#%E5%A2%9E%E5%8A%A0%E5%95%86%E5%93%81"
	shopDeleteDocsPath      = "allcommands/%E4%BB%A3%E5%B9%A3%E7%B3%BB%E7%B5%B1/ghp_shop#%E5%88%AA%E9%99%A4%E5%95%86%E5%93%81"
	shopPurchaseDocsPath    = "allcommands/%E4%BB%A3%E5%B9%A3%E7%B3%BB%E7%B5%B1/ghp_shop#%E8%B3%BC%E8%B2%B7%E5%95%86%E5%93%81"
	shopRoleHighDocsPath    = "MHCAT/bug#%E8%BA%AB%E5%88%86%E7%B5%84%E4%BD%8D%E9%9A%8E%E8%AA%BF%E9%AB%98%E6%98%AF%E7%94%9A%E9%BA%BC%E6%84%8F%E6%80%9D"
	shopFallbackGuildName   = "這個伺服器"
	shopGenericErrorContent = "很抱歉，出現了未知的錯誤，請重試!"
)

func (m Module) ShopHandler() interactions.Handler {
	return func(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
		if err := responder.Defer(ctx, responses.DeferOptions{}); err != nil {
			return err
		}
		switch interaction.Subcommand {
		case shopSubcommandAdd:
			return m.handleShopAdd(ctx, interaction, responder)
		case shopSubcommandDelete:
			return m.handleShopDelete(ctx, interaction, responder)
		case shopSubcommandList:
			return m.handleShopList(ctx, interaction, responder)
		default:
			return responder.EditOriginal(ctx, shopErrorMessage(shopGenericErrorContent, shopPurchaseDocsPath))
		}
	}
}

func (m Module) handleShopAdd(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
	if !interaction.Actor.HasPermission(shopManageMessagesBit) {
		return responder.EditOriginal(ctx, shopErrorMessage("你需要有`查詢跟購買大家都可用，剩下皆須訊息管理`才能使用此指令", shopAddDocsPath))
	}
	name := strings.TrimSpace(interaction.Options[shopOptionName])
	if len([]rune(name)) > coreeconomy.MaxLegacyShopNameRunes {
		return responder.EditOriginal(ctx, shopErrorMessage("商品名請少於15字!", shopAddDocsPath))
	}
	price, ok := integerOption(interaction, shopOptionPrice)
	if !ok || price <= 0 {
		return responder.EditOriginal(ctx, shopErrorMessage("`商品所需代幣`不可為負數或0!!!", shopAddDocsPath))
	}
	count := int64(1)
	if value, ok := integerOption(interaction, shopOptionCount); ok {
		count = value
	}
	if count <= 0 {
		return responder.EditOriginal(ctx, shopErrorMessage("`商品所需代幣`不可為負數或0!!!", shopAddDocsPath))
	}
	roleID := strings.TrimSpace(interaction.Options[shopOptionRole])
	if roleID != "" {
		assignable, err := m.canAssignShopRole(ctx, interaction.Actor.GuildID, roleID)
		if err != nil || !assignable {
			return responder.EditOriginal(ctx, shopErrorMessage("我沒有足夠的權限，請將我的身分組位階調高是!", shopRoleHighDocsPath))
		}
	}
	autoDelete, ok := boolOption(interaction, shopOptionAutoDelete)
	if !ok {
		return responder.EditOriginal(ctx, shopErrorMessage(shopGenericErrorContent, shopAddDocsPath))
	}
	item := domain.ShopItem{
		GuildID:     interaction.Actor.GuildID,
		CommodityID: m.nextShopCommodityID(),
		Name:        name,
		NeedCoins:   price,
		Description: strings.TrimSpace(interaction.Options[shopOptionDescription]),
		Code:        strings.TrimSpace(interaction.Options[shopOptionCode]),
		AutoDelete:  autoDelete,
		RoleID:      roleID,
		Count:       count,
	}
	created, err := m.shop.Create(ctx, item)
	if err != nil {
		switch {
		case errors.Is(err, ports.ErrShopItemLimit):
			return responder.EditOriginal(ctx, shopErrorMessage("很抱歉，商品數量已達到上限!請刪除商品!", shopAddDocsPath))
		case errors.Is(err, ports.ErrShopItemExists):
			return responder.EditOriginal(ctx, shopErrorMessage("很抱歉，資料重複，請等待幾秒後重試!", shopAddDocsPath))
		case errors.Is(err, domain.ErrInvalidShopItem):
			return responder.EditOriginal(ctx, shopErrorMessage(shopGenericErrorContent, shopAddDocsPath))
		default:
			return responder.EditOriginal(ctx, shopErrorMessage(shopGenericErrorContent, shopAddDocsPath))
		}
	}
	if err := responder.EditOriginal(ctx, shopAddSuccessMessage(created)); err != nil {
		return err
	}
	return m.trackCommand(ctx, interaction, ShopCommandName)
}

func (m Module) handleShopDelete(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
	if !interaction.Actor.HasPermission(shopManageMessagesBit) {
		return responder.EditOriginal(ctx, shopErrorMessage("你需要有`查詢跟購買大家都可用，剩下皆須訊息管理`才能使用此指令", shopDeleteDocsPath))
	}
	commodityID, ok := integerOption(interaction, shopOptionID)
	if !ok {
		return responder.EditOriginal(ctx, shopErrorMessage(shopGenericErrorContent, shopDeleteDocsPath))
	}
	if _, err := m.shop.Delete(ctx, interaction.Actor.GuildID, commodityID); err != nil {
		if errors.Is(err, ports.ErrShopItemMissing) {
			return responder.EditOriginal(ctx, shopErrorMessage("很抱歉，找不到這個商品，是不是打錯了?!", shopDeleteDocsPath))
		}
		return responder.EditOriginal(ctx, shopErrorMessage(shopGenericErrorContent, shopDeleteDocsPath))
	}
	if err := responder.EditOriginal(ctx, shopDeleteSuccessMessage()); err != nil {
		return err
	}
	return m.trackCommand(ctx, interaction, ShopCommandName)
}

func (m Module) handleShopList(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
	items, err := m.shop.List(ctx, interaction.Actor.GuildID)
	if err != nil {
		if errors.Is(err, ports.ErrShopItemMissing) {
			if editErr := responder.EditOriginal(ctx, shopErrorMessage("目前商店沒有任何商品喔", shopPurchaseDocsPath)); editErr != nil {
				return editErr
			}
			return m.trackCommand(ctx, interaction, ShopCommandName)
		}
		return responder.EditOriginal(ctx, shopErrorMessage(shopGenericErrorContent, shopPurchaseDocsPath))
	}
	if err := responder.EditOriginal(ctx, shopListMessage(items, m.shopGuildName(ctx, interaction.Actor.GuildID), interaction.Actor.UserTag, interaction.Actor.AvatarURL, m.color())); err != nil {
		return err
	}
	return m.trackCommand(ctx, interaction, ShopCommandName)
}

func (m Module) ShopItemHandler() interactions.Handler {
	return func(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
		if !shopInteractionOwnedByActor(interaction) {
			return responder.Reply(ctx, shopUnauthorizedMessage())
		}
		raw := strings.TrimSpace(interaction.CustomID)
		if strings.HasSuffix(raw, "ghp") {
			commodityID, ok := shopCommodityIDFromCustomID(strings.TrimSuffix(raw, "ghp"))
			if !ok {
				return responder.UpdateMessage(ctx, shopUpdateErrorMessage(shopGenericErrorContent, shopPurchaseDocsPath))
			}
			item, err := m.shop.Detail(ctx, interaction.Actor.GuildID, commodityID)
			if err != nil {
				return responder.UpdateMessage(ctx, shopDetailError(err))
			}
			m.shopSessions.Put(shopSession{
				GuildID:     interaction.Actor.GuildID,
				UserID:      interaction.Actor.UserID,
				MessageID:   interaction.MessageID,
				CommodityID: commodityID,
			})
			return responder.UpdateMessage(ctx, shopQuantityMessage(item, "", m.color()))
		}
		commodityID, ok := shopCommodityIDFromCustomID(raw)
		if !ok {
			return responder.UpdateMessage(ctx, shopUpdateErrorMessage(shopGenericErrorContent, shopPurchaseDocsPath))
		}
		item, err := m.shop.Detail(ctx, interaction.Actor.GuildID, commodityID)
		if err != nil {
			return responder.UpdateMessage(ctx, shopDetailError(err))
		}
		return responder.UpdateMessage(ctx, shopDetailMessage(item, m.color()))
	}
}

func (m Module) ShopQuantityHandler() interactions.Handler {
	return func(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
		if !shopInteractionOwnedByActor(interaction) {
			return responder.Reply(ctx, shopUnauthorizedMessage())
		}
		raw := strings.TrimSpace(interaction.CustomID)
		if strings.HasPrefix(raw, "confirmghp_number") {
			commodityID, ok := shopCommodityIDFromCustomID(strings.TrimPrefix(raw, "confirmghp_number"))
			if !ok {
				return responder.Reply(ctx, shopEphemeralContentError(":x: | 很抱歉，找不到這個商品，請重試!"))
			}
			return m.confirmShopQuantity(ctx, interaction, responder, commodityID)
		}
		session, ok := m.shopSessions.GetByMessage(interaction.Actor.GuildID, interaction.Actor.UserID, interaction.MessageID)
		if !ok {
			return responder.UpdateMessage(ctx, shopUpdateErrorMessage(shopGenericErrorContent, shopPurchaseDocsPath))
		}
		item, err := m.shop.Detail(ctx, interaction.Actor.GuildID, session.CommodityID)
		if err != nil {
			return responder.UpdateMessage(ctx, shopDetailError(err))
		}
		switch {
		case raw == "backghp_number":
			if session.Quantity != "" {
				session.Quantity = session.Quantity[:len(session.Quantity)-1]
			}
			m.shopSessions.Update(session)
			return responder.UpdateMessage(ctx, shopQuantityMessage(item, session.Quantity, m.color()))
		case strings.HasSuffix(raw, "ghp_number") && len(raw) > len("ghp_number"):
			digit := strings.TrimSuffix(raw, "ghp_number")
			if len(digit) != 1 || digit[0] < '0' || digit[0] > '9' {
				return responder.UpdateMessage(ctx, shopUpdateErrorMessage(shopGenericErrorContent, shopPurchaseDocsPath))
			}
			next := session.Quantity + digit
			quantity, _ := strconv.ParseInt(next, 10, 64)
			if quantity > item.Count {
				return responder.Reply(ctx, shopEphemeralContentError(":x: | 你輸入的數量不可超過商品數量!"))
			}
			if item.RoleID != "" && quantity > 1 {
				return responder.Reply(ctx, shopEphemeralContentError(":x: | 此商品為身分組商品，只能夠買一樣!"))
			}
			session.Quantity = next
			m.shopSessions.Update(session)
			return responder.UpdateMessage(ctx, shopQuantityMessage(item, session.Quantity, m.color()))
		default:
			return responder.UpdateMessage(ctx, shopUpdateErrorMessage(shopGenericErrorContent, shopPurchaseDocsPath))
		}
	}
}

func (m Module) confirmShopQuantity(ctx context.Context, interaction interactions.Interaction, responder responses.Responder, commodityID int64) error {
	session, ok := m.shopSessions.Get(interaction.Actor.GuildID, interaction.Actor.UserID, interaction.MessageID, commodityID)
	if !ok {
		return responder.UpdateMessage(ctx, shopUpdateErrorMessage(shopGenericErrorContent, shopPurchaseDocsPath))
	}
	quantity, err := strconv.ParseInt(strings.TrimSpace(session.Quantity), 10, 64)
	if err != nil || quantity <= 0 {
		return responder.Reply(ctx, shopEphemeralContentError(":x: | 購買數量必須大於0!"))
	}
	result, err := m.shop.Purchase(ctx, domain.ShopPurchaseCommand{
		GuildID:     interaction.Actor.GuildID,
		UserID:      interaction.Actor.UserID,
		CommodityID: commodityID,
		Quantity:    quantity,
	})
	if err != nil {
		switch {
		case errors.Is(err, ports.ErrShopItemMissing):
			return responder.Reply(ctx, shopEphemeralContentError(":x: | 很抱歉，找不到這個商品，請重試!"))
		case errors.Is(err, ports.ErrShopInsufficientCoin):
			return responder.UpdateMessage(ctx, shopUpdateErrorMessage("你的代幣不夠欸!使用/簽到或是多講話，都可以獲得代幣喔!", shopPurchaseDocsPath))
		case errors.Is(err, ports.ErrShopQuantityInvalid):
			return responder.Reply(ctx, shopEphemeralContentError(":x: | 你輸入的數量不可超過商品數量!"))
		default:
			return responder.UpdateMessage(ctx, shopUpdateErrorMessage(shopGenericErrorContent, shopPurchaseDocsPath))
		}
	}
	m.sendShopPurchaseSideEffects(ctx, interaction, result)
	m.shopSessions.Delete(session)
	return responder.UpdateMessage(ctx, shopPurchaseSuccessMessage(result))
}

func (m Module) canAssignShopRole(ctx context.Context, guildID string, roleID string) (bool, error) {
	if m.roleInspector == nil {
		return false, nil
	}
	return m.roleInspector.CanAssignRole(ctx, guildID, roleID)
}

func (m Module) nextShopCommodityID() int64 {
	clock := m.clock
	if clock == nil {
		clock = ports.SystemClock{}
	}
	return clock.Now().UnixMilli()
}

func (m Module) shopGuildName(ctx context.Context, guildID string) string {
	if m.discord == nil || strings.TrimSpace(guildID) == "" {
		return shopFallbackGuildName
	}
	info, err := m.discord.GuildInfo(ctx, guildID)
	if err != nil || strings.TrimSpace(info.Name) == "" {
		return shopFallbackGuildName
	}
	return info.Name
}

func (m Module) sendShopPurchaseSideEffects(ctx context.Context, interaction interactions.Interaction, result domain.ShopPurchaseResult) {
	if result.Item.RoleID != "" && m.roles != nil {
		_ = m.roles.AddRole(ctx, interaction.Actor.GuildID, interaction.Actor.UserID, result.Item.RoleID)
	}
	if result.Item.Code != "" && m.direct != nil {
		_, _ = m.direct.SendDirectMessage(ctx, interaction.Actor.UserID, ports.OutboundMessage{
			Embeds: []ports.OutboundEmbed{{
				Title:       fmt.Sprintf("%s您已成功購買`%s`", shopDoneEmoji, result.Item.Name),
				Description: fmt.Sprintf("<:security:997374179257102396> 您的獎品代碼:\n`%s`", result.Item.Code),
				Color:       shopSuccessColor,
			}},
			AllowedMentions: ports.AllowedMentions{},
		})
	}
}

func shopAddSuccessMessage(item domain.ShopItem) responses.Message {
	return responses.Message{
		Embeds: []responses.Embed{{
			Title:       shopStoreEmoji + " 代幣商店系統",
			Description: "已為您添加該商品!",
			Fields: []responses.EmbedField{{
				Name:  fmt.Sprintf("<:id:985950321975128094> 商品名稱: %s", item.Name),
				Value: fmt.Sprintf("商品id:`%d`\n需要代幣數: `%d`\n商品描述:`%s`\n商品是否自動刪除:`%t`\n身分組:`%s\n商品數量:%d`", item.CommodityID, item.NeedCoins, item.Description, item.AutoDelete, shopRoleDisplay(item.RoleID), item.Count),
			}},
			Color: shopSuccessColor,
		}},
		AllowedMentions: &responses.AllowedMentions{},
	}
}

func shopDeleteSuccessMessage() responses.Message {
	return responses.Message{
		Embeds: []responses.Embed{{
			Title:       shopStoreEmoji + " 代幣商店系統",
			Description: shopDeleteEmoji + "已為您刪除該商品!",
			Color:       shopSuccessColor,
		}},
		AllowedMentions: &responses.AllowedMentions{},
	}
}

func shopListMessage(items []domain.ShopItem, guildName string, userTag string, avatarURL string, color int) responses.Message {
	fields := make([]responses.EmbedField, 0, len(items))
	for _, item := range items {
		fields = append(fields, responses.EmbedField{
			Name:   fmt.Sprintf("<:id:985950321975128094> **商品名 :** `%s`", item.Name),
			Value:  fmt.Sprintf("💰 **商品價錢 :**`%d`\n<:productdescription:1001163044560314398> **商品說明 :**`%s`\n<:id:985950321975128094> **商品id:**`%d`", item.NeedCoins, item.Description, item.CommodityID),
			Inline: true,
		})
	}
	rows := make([]responses.ComponentRow, 0, 5)
	for start := 0; start < len(items) && start < coreeconomy.MaxLegacyShopItems; start += 5 {
		end := start + 5
		if end > len(items) {
			end = len(items)
		}
		row := responses.ComponentRow{}
		for _, item := range items[start:end] {
			row.Components = append(row.Components, responses.Component{
				Type:     responses.ComponentTypeButton,
				CustomID: strconv.FormatInt(item.CommodityID, 10),
				Label:    item.Name,
				Style:    responses.ButtonStylePrimary,
			})
		}
		rows = append(rows, row)
	}
	return responses.Message{
		Embeds: []responses.Embed{{
			Title:       fmt.Sprintf("<:list:992002476360343602> 以下是%s的商店", guildName),
			Description: "<a:arrow_pink:996242460294512690> **點擊下方的按扭進行購買及查詢詳情!**",
			Color:       color,
			Fields:      fields,
			Footer: &responses.EmbedFooter{
				Text:    fmt.Sprintf("%s的查詢", userTag),
				IconURL: avatarURL,
			},
		}},
		Components:      rows,
		AllowedMentions: &responses.AllowedMentions{},
	}
}

func shopDetailMessage(item domain.ShopItem, color int) responses.Message {
	return responses.Message{
		Embeds: []responses.Embed{{
			Title:       fmt.Sprintf("<:creativeteaching:986060052949524600> 以下是%s的詳細資料", item.Name),
			Description: shopItemDetailDescription(item),
			Color:       color,
		}},
		Components: []responses.ComponentRow{{
			Components: []responses.Component{{
				Type:     responses.ComponentTypeButton,
				CustomID: fmt.Sprintf("%dghp", item.CommodityID),
				Label:    "購買該商品",
				Emoji:    "<:addtocart:1010884094088978474>",
				Style:    responses.ButtonStyleSuccess,
			}},
		}},
		AllowedMentions: &responses.AllowedMentions{},
	}
}

func shopQuantityMessage(item domain.ShopItem, quantity string, color int) responses.Message {
	description := shopItemDetailDescription(item) + "目前選擇數量:"
	if quantity != "" {
		description += fmt.Sprintf("`%s`", quantity)
	}
	return responses.Message{
		Embeds: []responses.Embed{{
			Title:       "<:choose:1007244640958808088> 請選擇購買數量!",
			Description: description,
			Color:       color,
		}},
		Components:      shopQuantityRows(item.CommodityID),
		AllowedMentions: &responses.AllowedMentions{},
	}
}

func shopQuantityRows(commodityID int64) []responses.ComponentRow {
	return []responses.ComponentRow{
		{Components: []responses.Component{
			shopQuantityButton("1ghp_number", "<:numberone:1005471516407906324>", responses.ButtonStyleSecondary),
			shopQuantityButton("2ghp_number", "<:number2:1005471518018510950>", responses.ButtonStyleSecondary),
			shopQuantityButton("3ghp_number", "<:number3:1005471519574597672>", responses.ButtonStyleSecondary),
		}},
		{Components: []responses.Component{
			shopQuantityButton("4ghp_number", "<:numberfour:1005471521147473950>", responses.ButtonStyleSecondary),
			shopQuantityButton("5ghp_number", "<:number5:1005471522649022517>", responses.ButtonStyleSecondary),
			shopQuantityButton("6ghp_number", "<:six:1005471524721020948>", responses.ButtonStyleSecondary),
		}},
		{Components: []responses.Component{
			shopQuantityButton("7ghp_number", "<:seven:1005471526222581760>", responses.ButtonStyleSecondary),
			shopQuantityButton("8ghp_number", "<:number8:1005471527891898398>", responses.ButtonStyleSecondary),
			shopQuantityButton("9ghp_number", "<:number9:1005471529699655780>", responses.ButtonStyleSecondary),
		}},
		{Components: []responses.Component{
			shopQuantityButton("backghp_number", "<:previous:1010916328045035560>", responses.ButtonStyleDanger),
			shopQuantityButton("0ghp_number", "<:zero1:1010925602066399273>", responses.ButtonStyleSecondary),
			shopQuantityButton(fmt.Sprintf("confirmghp_number%d", commodityID), "<:confirm:1010916326405054515>", responses.ButtonStyleSuccess),
		}},
	}
}

func shopQuantityButton(customID string, emoji string, style responses.ButtonStyle) responses.Component {
	return responses.Component{
		Type:     responses.ComponentTypeButton,
		CustomID: customID,
		Emoji:    emoji,
		Style:    style,
	}
}

func shopPurchaseSuccessMessage(result domain.ShopPurchaseResult) responses.Message {
	return responses.Message{
		Embeds: []responses.Embed{{
			Title:       shopStoreEmoji + " 代幣商店系統",
			Description: fmt.Sprintf("%s您已成功購買:%s\n數量:%d!", shopDoneEmoji, result.Item.Name, result.Quantity),
			Color:       shopSuccessColor,
		}},
		AllowedMentions: &responses.AllowedMentions{},
	}
}

func shopItemDetailDescription(item domain.ShopItem) string {
	return fmt.Sprintf("<:id:1010884394791207003> 商品id:\n```%d```<:pricetag:1010884565822349392> 商品價格:\n```%d 個代幣```<:sign:997374180632825896> 商品描述:\n```%s```<:trashbin:995991389043163257> 是否自動刪除:\n```%t```<:roleplaying:985945121264635964> 是否會附加身分組:\n%s\n <:counter:994585977207140423> 商品數量:\n```%d```", item.CommodityID, item.NeedCoins, item.Description, item.AutoDelete, shopRoleDetailDisplay(item.RoleID), item.Count)
}

func shopErrorMessage(content string, docsPath string) responses.Message {
	return responses.Message{
		Embeds: []responses.Embed{{
			Title:       fmt.Sprintf("%s | %s", shopErrorEmoji, content),
			Description: shopDocsDescription(docsPath),
			Color:       shopErrorColor,
		}},
		AllowedMentions: &responses.AllowedMentions{},
	}
}

func shopUpdateErrorMessage(content string, docsPath string) responses.Message {
	message := shopErrorMessage(content, docsPath)
	message.Components = []responses.ComponentRow{}
	return message
}

func shopDetailError(err error) responses.Message {
	if errors.Is(err, ports.ErrShopItemMissing) {
		return shopUpdateErrorMessage("很抱歉，找不到這個商品，請於幾秒鐘後重試!", shopPurchaseDocsPath)
	}
	return shopUpdateErrorMessage(shopGenericErrorContent, shopPurchaseDocsPath)
}

func shopEphemeralContentError(content string) responses.Message {
	return responses.Message{
		Content:         content,
		Ephemeral:       true,
		AllowedMentions: &responses.AllowedMentions{},
	}
}

func shopInteractionOwnedByActor(interaction interactions.Interaction) bool {
	ownerID := strings.TrimSpace(interaction.OriginalInteractionUserID)
	return ownerID == "" || ownerID == strings.TrimSpace(interaction.Actor.UserID)
}

func shopUnauthorizedMessage() responses.Message {
	return responses.Message{
		Ephemeral: true,
		Embeds: []responses.Embed{{
			Title: "<a:error:980086028113182730> | 你不是查詢者無法使用!",
			Color: shopUnauthorizedColor,
		}},
		AllowedMentions: &responses.AllowedMentions{},
	}
}

func shopDocsDescription(docsPath string) string {
	docsPath = strings.TrimSpace(docsPath)
	if docsPath == "" {
		return ""
	}
	return fmt.Sprintf("<:MHCATdarkdsadsadsadsadsadas1:1079853990541529208> [立即前往官方文檔查看問題](https://docsmhcat.yorukot.me/%s)", docsPath)
}

func shopRoleDisplay(roleID string) string {
	if strings.TrimSpace(roleID) == "" {
		return "null"
	}
	return fmt.Sprintf("<@&%s>", roleID)
}

func shopRoleDetailDisplay(roleID string) string {
	if strings.TrimSpace(roleID) == "" {
		return "不會"
	}
	return fmt.Sprintf("<@&%s>", roleID)
}

func shopCommodityIDFromCustomID(value string) (int64, bool) {
	commodityID, err := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
	return commodityID, err == nil && commodityID > 0
}

func boolOption(interaction interactions.Interaction, name string) (bool, bool) {
	if value, ok := interaction.CommandOptions[name]; ok && value.Type == interactions.CommandOptionBoolean {
		return value.Bool, true
	}
	parsed, err := strconv.ParseBool(strings.TrimSpace(interaction.Options[name]))
	return parsed, err == nil
}
