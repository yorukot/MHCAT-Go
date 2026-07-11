package domain

import (
	"math"
	"testing"
)

func TestShopItemPurchaseCostUsesJavaScriptNumberArithmetic(t *testing.T) {
	item := ShopItem{NeedCoins: math.MaxInt64}

	if cost, ok := item.PurchaseCost(1); !ok || cost != float64(math.MaxInt64) {
		t.Fatalf("single item cost = %v, ok = %t", cost, ok)
	}
	if cost, ok := item.PurchaseCost(2); !ok || cost != float64(math.MaxInt64)*2 {
		t.Fatalf("large cost = %v, ok = %t", cost, ok)
	}
}

func TestShopItemPurchaseCostPreservesLegacyScalars(t *testing.T) {
	for _, test := range []struct {
		text string
		want float64
	}{
		{text: "20.5", want: 41},
		{text: "null", want: 0},
		{text: "-2", want: -4},
		{text: "Infinity", want: math.Inf(1)},
	} {
		if got, ok := (ShopItem{NeedCoinsText: test.text}).PurchaseCost(2); !ok || got != test.want {
			t.Fatalf("cost for %q = %v, ok=%t", test.text, got, ok)
		}
	}
	if _, ok := (ShopItem{NeedCoinsText: "undefined"}).PurchaseCost(2); ok {
		t.Fatal("malformed price should fail")
	}
}

func TestShopItemStockNumberPreservesLegacyScalars(t *testing.T) {
	for _, test := range []struct {
		text string
		want float64
	}{
		{text: "1.5", want: 1.5},
		{text: "null", want: 0},
		{text: "Infinity", want: math.Inf(1)},
	} {
		if got, ok := (ShopItem{CountText: test.text}).StockNumber(); !ok || got != test.want {
			t.Fatalf("stock for %q = %v, ok=%t", test.text, got, ok)
		}
	}
	if _, ok := (ShopItem{CountText: "undefined"}).StockNumber(); ok {
		t.Fatal("malformed stock should fail")
	}
}
