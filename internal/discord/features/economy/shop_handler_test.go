package economy

import (
	"context"
	"errors"
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

func TestShopListRendersLegacyFieldsAndButtons(t *testing.T) {
	repo := fakemongo.NewEconomyRepository()
	repo.PutShopItem(domain.ShopItem{GuildID: "guild-1", CommodityID: 1001, Name: "VIP", NeedCoins: 50, Description: "role reward", Count: 1})
	module := NewShopModule(repo, nil, nil, nil, nil, nil, nil)

	responder := fakediscord.NewResponder()
	if err := module.ShopHandler()(context.Background(), fakediscord.SlashInteractionWithOptions(ShopCommandName, shopSubcommandList, nil), responder); err != nil {
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
}

func TestShopPurchaseFlowMutatesCoinsInventoryRoleAndCode(t *testing.T) {
	repo := fakemongo.NewEconomyRepository()
	repo.PutShopItem(domain.ShopItem{GuildID: "guild-1", CommodityID: 1001, Name: "VIP", NeedCoins: 20, Description: "role reward", Code: "CODE", AutoDelete: true, RoleID: "role-1", Count: 1})
	repo.PutBalance(domain.CoinBalance{GuildID: "guild-1", UserID: "user-1", Coins: 50})
	sideEffects := fakediscord.NewSideEffects()
	module := NewShopModule(repo, nil, sideEffects, sideEffects, sideEffects, nil, nil)

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

type shopFixedClock struct {
	now time.Time
}

func (c shopFixedClock) Now() time.Time {
	return c.now
}
