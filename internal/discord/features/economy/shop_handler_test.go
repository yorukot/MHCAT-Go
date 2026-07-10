package economy

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/interactions"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakediscord"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakemongo"
)

func TestShopAddCreatesLegacyItem(t *testing.T) {
	repo := fakemongo.NewEconomyRepository()
	sideEffects := fakediscord.NewSideEffects()
	sideEffects.AssignableRoles["guild-1/role-1"] = true
	module := NewShopModule(repo, nil, sideEffects, sideEffects, sideEffects, nil, shopFixedClock{now: time.UnixMilli(1_710_000_000_000)})

	interaction := shopAddInteraction()
	interaction.Actor.PermissionBits = shopManageMessagesBit
	responder := fakediscord.NewResponder()
	if err := module.ShopHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("shop add: %v", err)
	}
	item, err := repo.GetShopItem(context.Background(), "guild-1", 1_710_000_000_000)
	if err != nil {
		t.Fatalf("get created item: %v", err)
	}
	if item.Name != "VIP" || item.NeedCoins != 50 || item.Description != "role reward" || item.Code != "CODE" || item.RoleID != "role-1" || item.Count != 2 || !item.AutoDelete {
		t.Fatalf("created item = %#v", item)
	}
	if len(responder.Edits) != 1 || !strings.Contains(responder.Edits[0].Embeds[0].Fields[0].Value, "商品id:`1710000000000`") {
		t.Fatalf("response = %#v", responder.Edits)
	}
}

func TestShopAddRequiresManageMessages(t *testing.T) {
	module := NewShopModule(fakemongo.NewEconomyRepository(), nil, nil, nil, nil, nil, shopFixedClock{now: time.UnixMilli(1_710_000_000_000)})

	responder := fakediscord.NewResponder()
	if err := module.ShopHandler()(context.Background(), shopAddInteraction(), responder); err != nil {
		t.Fatalf("shop add: %v", err)
	}
	if len(responder.Edits) != 1 || !strings.Contains(responder.Edits[0].Embeds[0].Title, "你需要有`查詢跟購買大家都可用") {
		t.Fatalf("permission response = %#v", responder.Edits)
	}
}

func TestShopAddUsesLegacyUTF16NameLength(t *testing.T) {
	repo := fakemongo.NewEconomyRepository()
	module := NewShopModule(repo, nil, nil, nil, nil, nil, shopFixedClock{now: time.UnixMilli(1_710_000_000_000)})
	interaction := shopAddInteraction()
	interaction.Actor.PermissionBits = shopManageMessagesBit
	interaction.Options[shopOptionName] = strings.Repeat("\U0001F600", 8)
	responder := fakediscord.NewResponder()

	if err := module.ShopHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("shop add: %v", err)
	}
	if len(responder.Edits) != 1 || !strings.Contains(responder.Edits[0].Embeds[0].Title, "商品名請少於15字") {
		t.Fatalf("response = %#v", responder.Edits)
	}
	if _, err := repo.GetShopItem(context.Background(), "guild-1", 1_710_000_000_000); !errors.Is(err, ports.ErrShopItemMissing) {
		t.Fatalf("overlong item was created: %v", err)
	}
}

func TestShopListRendersLegacyFieldsAndButtons(t *testing.T) {
	repo := fakemongo.NewEconomyRepository()
	repo.PutShopItem(domain.ShopItem{GuildID: "guild-1", CommodityID: 1001, Name: "VIP", NeedCoins: 50, Description: "role reward", Count: 1})
	module := NewShopModule(repo, nil, nil, nil, nil, nil, nil)
	open := fakediscord.SlashInteractionWithOptions(ShopCommandName, shopSubcommandList, nil)
	open.ID = "interaction-1"

	responder := fakediscord.NewResponder()
	if err := module.ShopHandler()(context.Background(), open, responder); err != nil {
		t.Fatalf("shop list: %v", err)
	}
	if len(responder.Edits) != 1 || len(responder.Edits[0].Components) != 1 {
		t.Fatalf("list response = %#v", responder.Edits)
	}
	embed := responder.Edits[0].Embeds[0]
	if !strings.Contains(embed.Title, "以下是") || !strings.Contains(embed.Fields[0].Value, "**商品id:**`1001`") {
		t.Fatalf("list embed = %#v", embed)
	}
	button := responder.Edits[0].Components[0].Components[0]
	if button.CustomID != "1001" || button.Label != "VIP" {
		t.Fatalf("button = %#v", button)
	}

	detail := fakediscord.ComponentInteractionFromID("1001")
	detail.MessageID = "message-1"
	detail.OriginalInteractionID = "interaction-1"
	detailResponder := fakediscord.NewResponder()
	if err := module.ShopItemHandler()(context.Background(), detail, detailResponder); err != nil {
		t.Fatalf("show item detail: %v", err)
	}
	if len(detailResponder.Updates) != 1 || !strings.Contains(detailResponder.Updates[0].Embeds[0].Title, "VIP") {
		t.Fatalf("detail response = %#v", detailResponder.Updates)
	}

	secondResponder := fakediscord.NewResponder()
	if err := module.ShopItemHandler()(context.Background(), detail, secondResponder); err != nil {
		t.Fatalf("repeat item click: %v", err)
	}
	if len(secondResponder.Replies) != 0 || len(secondResponder.Updates) != 0 {
		t.Fatalf("max-one collector responded twice: replies %#v updates %#v", secondResponder.Replies, secondResponder.Updates)
	}
}

func TestShopDetailAndQuantityUseLegacyRandomColors(t *testing.T) {
	repo := fakemongo.NewEconomyRepository()
	repo.PutShopItem(domain.ShopItem{GuildID: "guild-1", CommodityID: 1001, Name: "VIP", NeedCoins: 50, Description: "role reward", Count: 2})
	module := NewShopModule(repo, nil, nil, nil, nil, nil, nil)
	colors := []int{0x123456, 0x654321, 0xABCDEF}
	colorIndex := 0
	module.color = func() int {
		color := colors[colorIndex]
		colorIndex++
		return color
	}
	seedShopBrowse(&module, "interaction-1")

	detail := fakediscord.ComponentInteractionFromID("1001")
	detail.MessageID = "message-1"
	detail.OriginalInteractionID = "interaction-1"
	detailResponder := fakediscord.NewResponder()
	if err := module.ShopItemHandler()(context.Background(), detail, detailResponder); err != nil {
		t.Fatalf("show detail: %v", err)
	}
	if got := detailResponder.Updates[0].Embeds[0].Color; got != colors[0] {
		t.Fatalf("detail color = %#x, want %#x", got, colors[0])
	}

	start := fakediscord.ComponentInteractionFromID("1001ghp")
	start.MessageID = "message-1"
	startResponder := fakediscord.NewResponder()
	if err := module.ShopItemHandler()(context.Background(), start, startResponder); err != nil {
		t.Fatalf("start purchase: %v", err)
	}
	if got := startResponder.Updates[0].Embeds[0].Color; got != colors[1] {
		t.Fatalf("initial quantity color = %#x, want %#x", got, colors[1])
	}

	digit := fakediscord.ComponentInteractionFromID("1ghp_number")
	digit.MessageID = "message-1"
	digitResponder := fakediscord.NewResponder()
	if err := module.ShopQuantityHandler()(context.Background(), digit, digitResponder); err != nil {
		t.Fatalf("enter quantity: %v", err)
	}
	if got := digitResponder.Updates[0].Embeds[0].Color; got != colors[2] {
		t.Fatalf("updated quantity color = %#x, want %#x", got, colors[2])
	}
}

func TestShopComponentsRejectDifferentRequester(t *testing.T) {
	module := NewShopModule(fakemongo.NewEconomyRepository(), nil, nil, nil, nil, nil, nil)
	tests := []struct {
		name     string
		customID string
		handler  interactions.Handler
	}{
		{name: "item", customID: "1001", handler: module.ShopItemHandler()},
		{name: "purchase", customID: "1001ghp", handler: module.ShopItemHandler()},
		{name: "quantity", customID: "1ghp_number", handler: module.ShopQuantityHandler()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			interaction := fakediscord.ComponentInteractionFromID(tt.customID)
			interaction.Actor.UserID = "user-2"
			interaction.OriginalInteractionUserID = "user-1"
			responder := fakediscord.NewResponder()

			if err := tt.handler(context.Background(), interaction, responder); err != nil {
				t.Fatalf("handle component: %v", err)
			}
			if len(responder.Replies) != 1 || !responder.Replies[0].Ephemeral || len(responder.Updates) != 0 {
				t.Fatalf("response = replies %#v updates %#v", responder.Replies, responder.Updates)
			}
			embed := responder.Replies[0].Embeds[0]
			if embed.Title != "<a:error:980086028113182730> | 你不是查詢者無法使用!" || embed.Color != shopUnauthorizedColor {
				t.Fatalf("unauthorized embed = %#v", embed)
			}
		})
	}
}

func TestShopListCollectorExpiresAtLegacyDeadline(t *testing.T) {
	now := time.Unix(1_000, 0)
	repo := fakemongo.NewEconomyRepository()
	repo.PutShopItem(domain.ShopItem{GuildID: "guild-1", CommodityID: 1001, Name: "VIP", NeedCoins: 50, Count: 1})
	clock := &shopMutableClock{now: now}
	module := NewShopModule(repo, nil, nil, nil, nil, nil, clock)
	open := fakediscord.SlashInteractionWithOptions(ShopCommandName, shopSubcommandList, nil)
	open.ID = "interaction-1"
	if err := module.ShopHandler()(context.Background(), open, fakediscord.NewResponder()); err != nil {
		t.Fatalf("open shop: %v", err)
	}
	clock.now = now.Add(shopCollectorTTL)
	interaction := fakediscord.ComponentInteractionFromID("1001")
	interaction.MessageID = "message-1"
	interaction.OriginalInteractionID = "interaction-1"
	responder := fakediscord.NewResponder()

	if err := module.ShopItemHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("expired item click: %v", err)
	}
	if len(responder.Replies) != 0 || len(responder.Updates) != 0 {
		t.Fatalf("expired collector responded: replies %#v updates %#v", responder.Replies, responder.Updates)
	}
}

func TestShopDetailCollectorDoesNotExtendWhenPurchaseStarts(t *testing.T) {
	now := time.Unix(1_000, 0)
	repo := fakemongo.NewEconomyRepository()
	repo.PutShopItem(domain.ShopItem{GuildID: "guild-1", CommodityID: 1001, Name: "VIP", NeedCoins: 50, Count: 2})
	clock := &shopMutableClock{now: now}
	module := NewShopModule(repo, nil, nil, nil, nil, nil, clock)
	seedShopBrowse(&module, "interaction-1")

	detail := fakediscord.ComponentInteractionFromID("1001")
	detail.MessageID = "message-1"
	detail.OriginalInteractionID = "interaction-1"
	if err := module.ShopItemHandler()(context.Background(), detail, fakediscord.NewResponder()); err != nil {
		t.Fatalf("show detail: %v", err)
	}

	clock.now = now.Add(shopCollectorTTL - time.Second)
	purchase := fakediscord.ComponentInteractionFromID("1001ghp")
	purchase.MessageID = "message-1"
	if err := module.ShopItemHandler()(context.Background(), purchase, fakediscord.NewResponder()); err != nil {
		t.Fatalf("start purchase: %v", err)
	}

	clock.now = now.Add(shopCollectorTTL)
	digit := fakediscord.ComponentInteractionFromID("1ghp_number")
	digit.MessageID = "message-1"
	responder := fakediscord.NewResponder()
	if err := module.ShopQuantityHandler()(context.Background(), digit, responder); err != nil {
		t.Fatalf("expired quantity click: %v", err)
	}
	if len(responder.Replies) != 0 || len(responder.Updates) != 0 {
		t.Fatalf("expired collector responded: replies %#v updates %#v", responder.Replies, responder.Updates)
	}
}

func TestShopCollectorsDoNotSurviveProcessRestart(t *testing.T) {
	now := time.Unix(1_000, 0)
	repo := fakemongo.NewEconomyRepository()
	repo.PutShopItem(domain.ShopItem{GuildID: "guild-1", CommodityID: 1001, Name: "VIP", NeedCoins: 50, Count: 1})
	first := NewShopModule(repo, nil, nil, nil, nil, nil, shopFixedClock{now: now})
	seedShopBrowse(&first, "interaction-1")
	detail := fakediscord.ComponentInteractionFromID("1001")
	detail.MessageID = "message-1"
	detail.OriginalInteractionID = "interaction-1"
	if err := first.ShopItemHandler()(context.Background(), detail, fakediscord.NewResponder()); err != nil {
		t.Fatalf("show detail before restart: %v", err)
	}
	module := NewShopModule(repo, nil, nil, nil, nil, nil, shopFixedClock{now: now})
	interaction := fakediscord.ComponentInteractionFromID("1001ghp")
	interaction.MessageID = "message-1"
	responder := fakediscord.NewResponder()

	if err := module.ShopItemHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("orphaned purchase click: %v", err)
	}
	if len(responder.Replies) != 0 || len(responder.Updates) != 0 {
		t.Fatalf("orphaned collector responded: replies %#v updates %#v", responder.Replies, responder.Updates)
	}
}

func TestShopPurchaseFlowMutatesCoinsInventoryRoleAndCode(t *testing.T) {
	repo := fakemongo.NewEconomyRepository()
	repo.PutShopItem(domain.ShopItem{GuildID: "guild-1", CommodityID: 1001, Name: "VIP", NeedCoins: 20, Description: "role reward", Code: "CODE", AutoDelete: true, RoleID: "role-1", Count: 1})
	repo.PutBalance(domain.CoinBalance{GuildID: "guild-1", UserID: "user-1", Coins: 50})
	sideEffects := fakediscord.NewSideEffects()
	module := NewShopModule(repo, nil, sideEffects, sideEffects, sideEffects, nil, nil)
	showShopDetail(t, &module, 1001, "message-1")

	start := fakediscord.ComponentInteractionFromID("1001ghp")
	start.MessageID = "message-1"
	responder := fakediscord.NewResponder()
	if err := module.ShopItemHandler()(context.Background(), start, responder); err != nil {
		t.Fatalf("start purchase: %v", err)
	}
	if len(responder.Updates) != 1 || len(responder.Updates[0].Components) != 4 {
		t.Fatalf("quantity prompt = %#v", responder.Updates)
	}

	digit := fakediscord.ComponentInteractionFromID("1ghp_number")
	digit.MessageID = "message-1"
	responder = fakediscord.NewResponder()
	if err := module.ShopQuantityHandler()(context.Background(), digit, responder); err != nil {
		t.Fatalf("digit: %v", err)
	}
	if len(responder.Updates) != 1 || !strings.Contains(responder.Updates[0].Embeds[0].Description, "目前選擇數量:`1`") {
		t.Fatalf("digit update = %#v", responder.Updates)
	}

	confirm := fakediscord.ComponentInteractionFromID("confirmghp_number1001")
	confirm.MessageID = "message-1"
	responder = fakediscord.NewResponder()
	if err := module.ShopQuantityHandler()(context.Background(), confirm, responder); err != nil {
		t.Fatalf("confirm: %v", err)
	}
	balance, err := repo.GetCoinBalance(context.Background(), "guild-1", "user-1")
	if err != nil || balance.Coins != 30 {
		t.Fatalf("balance=%#v err=%v", balance, err)
	}
	if _, err := repo.GetShopItem(context.Background(), "guild-1", 1001); !errors.Is(err, ports.ErrShopItemMissing) {
		t.Fatalf("expected item deletion, got %v", err)
	}
	if len(sideEffects.AddedRoles) != 1 || sideEffects.AddedRoles[0].RoleID != "role-1" {
		t.Fatalf("added roles = %#v", sideEffects.AddedRoles)
	}
	if len(sideEffects.DirectMessages) != 1 || !strings.Contains(sideEffects.DirectMessages[0].Message.Embeds[0].Description, "CODE") {
		t.Fatalf("direct messages = %#v", sideEffects.DirectMessages)
	}
	if len(responder.Updates) != 1 || !strings.Contains(responder.Updates[0].Embeds[0].Description, "您已成功購買:VIP") {
		t.Fatalf("success update = %#v", responder.Updates)
	}
}

func TestShopPurchaseInsufficientCoinsUsesLegacyError(t *testing.T) {
	repo := fakemongo.NewEconomyRepository()
	repo.PutShopItem(domain.ShopItem{GuildID: "guild-1", CommodityID: 1001, Name: "VIP", NeedCoins: 20, Description: "role reward", AutoDelete: true, Count: 1})
	repo.PutBalance(domain.CoinBalance{GuildID: "guild-1", UserID: "user-1", Coins: 5})
	module := NewShopModule(repo, nil, nil, nil, nil, nil, nil)
	showShopDetail(t, &module, 1001, "message-1")

	start := fakediscord.ComponentInteractionFromID("1001ghp")
	start.MessageID = "message-1"
	if err := module.ShopItemHandler()(context.Background(), start, fakediscord.NewResponder()); err != nil {
		t.Fatalf("start purchase: %v", err)
	}
	digit := fakediscord.ComponentInteractionFromID("1ghp_number")
	digit.MessageID = "message-1"
	if err := module.ShopQuantityHandler()(context.Background(), digit, fakediscord.NewResponder()); err != nil {
		t.Fatalf("digit: %v", err)
	}
	confirm := fakediscord.ComponentInteractionFromID("confirmghp_number1001")
	confirm.MessageID = "message-1"
	responder := fakediscord.NewResponder()
	if err := module.ShopQuantityHandler()(context.Background(), confirm, responder); err != nil {
		t.Fatalf("confirm: %v", err)
	}
	if len(responder.Updates) != 1 || !strings.Contains(responder.Updates[0].Embeds[0].Title, "你的代幣不夠") {
		t.Fatalf("insufficient response = %#v", responder.Updates)
	}
	balance, _ := repo.GetCoinBalance(context.Background(), "guild-1", "user-1")
	if balance.Coins != 5 {
		t.Fatalf("balance mutated = %#v", balance)
	}
}

func shopAddInteraction() interactions.Interaction {
	return fakediscord.SlashInteractionWithOptions(ShopCommandName, shopSubcommandAdd, map[string]string{
		shopOptionName:        "VIP",
		shopOptionPrice:       "50",
		shopOptionDescription: "role reward",
		shopOptionAutoDelete:  "true",
		shopOptionCode:        "CODE",
		shopOptionRole:        "role-1",
		shopOptionCount:       "2",
	})
}

func seedShopBrowse(module *Module, interactionID string) {
	now := module.shopNow()
	module.shopSessions.PutBrowse(shopBrowseSession{
		GuildID:       "guild-1",
		UserID:        "user-1",
		InteractionID: interactionID,
		ExpiresAt:     now.Add(shopCollectorTTL),
	}, now)
}

func showShopDetail(t *testing.T, module *Module, commodityID int64, messageID string) {
	t.Helper()
	interactionID := messageID + "-interaction"
	seedShopBrowse(module, interactionID)
	interaction := fakediscord.ComponentInteractionFromID(strconv.FormatInt(commodityID, 10))
	interaction.MessageID = messageID
	interaction.OriginalInteractionID = interactionID
	if err := module.ShopItemHandler()(context.Background(), interaction, fakediscord.NewResponder()); err != nil {
		t.Fatalf("show shop detail: %v", err)
	}
}

type shopFixedClock struct {
	now time.Time
}

type shopMutableClock struct {
	now time.Time
}

func (c *shopMutableClock) Now() time.Time {
	return c.now
}

func (c shopFixedClock) Now() time.Time {
	return c.now
}
