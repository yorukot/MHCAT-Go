package domain

import (
	"math"
	"testing"
)

func TestShopItemPurchaseCostRejectsOverflow(t *testing.T) {
	item := ShopItem{NeedCoins: math.MaxInt64}

	if cost, ok := item.PurchaseCost(1); !ok || cost != math.MaxInt64 {
		t.Fatalf("single item cost = %d, ok = %t", cost, ok)
	}
	if cost, ok := item.PurchaseCost(2); ok || cost != 0 {
		t.Fatalf("overflow cost = %d, ok = %t", cost, ok)
	}
}
