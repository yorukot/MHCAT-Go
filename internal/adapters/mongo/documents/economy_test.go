package documents

import (
	"math"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestCoinDocumentLegacyBSONDecodesMixedNumericTypes(t *testing.T) {
	raw, err := bson.Marshal(bson.D{
		{Key: "guild", Value: "guild-1"},
		{Key: "member", Value: "user-1"},
		{Key: "coin", Value: "1234"},
		{Key: "today", Value: int32(7)},
	})
	if err != nil {
		t.Fatalf("marshal fixture: %v", err)
	}
	var document CoinDocument
	if err := bson.Unmarshal(raw, &document); err != nil {
		t.Fatalf("decode document: %v", err)
	}
	balance := document.ToDomain()
	if balance.GuildID != "guild-1" || balance.UserID != "user-1" || balance.Coins != 1234 || balance.CoinsText != "1234" || balance.Today != 7 || balance.TodayText != "7" {
		t.Fatalf("balance = %#v", balance)
	}
}

func TestCoinDocumentMissingNumericFieldsDecodeSafe(t *testing.T) {
	raw, err := bson.Marshal(bson.D{
		{Key: "guild", Value: "guild-1"},
		{Key: "member", Value: "user-1"},
	})
	if err != nil {
		t.Fatalf("marshal fixture: %v", err)
	}
	var document CoinDocument
	if err := bson.Unmarshal(raw, &document); err != nil {
		t.Fatalf("decode document: %v", err)
	}
	balance := document.ToDomain()
	if balance.Coins != 0 || balance.CoinsText != "undefined" || balance.Today != 0 || balance.TodayText != "undefined" {
		t.Fatalf("balance = %#v", balance)
	}
}

func TestCoinDocumentPreservesMongooseCoinDisplay(t *testing.T) {
	tests := []struct {
		name  string
		value any
		want  string
	}{
		{name: "null", value: nil, want: "null"},
		{name: "decimal", value: 125.5, want: "125.5"},
		{name: "positive infinity", value: math.Inf(1), want: "Infinity"},
		{name: "negative infinity", value: math.Inf(-1), want: "-Infinity"},
		{name: "malformed", value: bson.A{1}, want: "undefined"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			raw, err := bson.Marshal(bson.D{{Key: "guild", Value: "guild-1"}, {Key: "member", Value: "user-1"}, {Key: "coin", Value: test.value}, {Key: "today", Value: test.value}})
			if err != nil {
				t.Fatalf("marshal fixture: %v", err)
			}
			var document CoinDocument
			if err := bson.Unmarshal(raw, &document); err != nil {
				t.Fatalf("decode document: %v", err)
			}
			balance := document.ToDomain()
			if balance.CoinsText != test.want || balance.TodayText != test.want {
				t.Fatalf("balance = %#v, want display %q", balance, test.want)
			}
		})
	}
}

func TestGiftChangeDocumentLegacyBSONDecodesDefaults(t *testing.T) {
	raw, err := bson.Marshal(bson.D{
		{Key: "guild", Value: "guild-1"},
		{Key: "coin_number", Value: int64(700)},
		{Key: "sign_coin", Value: "20"},
		{Key: "channel", Value: "channel-1"},
		{Key: "xp_multiple", Value: float64(2.5)},
		{Key: "time", Value: nil},
	})
	if err != nil {
		t.Fatalf("marshal fixture: %v", err)
	}
	var document GiftChangeDocument
	if err := bson.Unmarshal(raw, &document); err != nil {
		t.Fatalf("decode document: %v", err)
	}
	config := document.ToDomain()
	if config.GachaCost != 700 || config.GachaCostText != "700" || config.SignCoins != 20 || config.SignCoinsText != "20" || config.ChannelID != "channel-1" || config.XPMultiple != 2.5 || config.XPMultipleText != "2.5" || config.ResetMarker != 0 || config.ResetMarkerText != "null" {
		t.Fatalf("config = %#v", config)
	}
}

func TestGiftChangeDocumentPreservesMongooseCoinNumberDisplay(t *testing.T) {
	tests := []struct {
		name  string
		value any
		want  string
	}{
		{name: "missing", value: bson.Undefined{}, want: "undefined"},
		{name: "null", value: nil, want: "null"},
		{name: "decimal", value: 700.5, want: "700.5"},
		{name: "infinity", value: math.Inf(1), want: "Infinity"},
		{name: "malformed", value: bson.D{{Key: "bad", Value: true}}, want: "undefined"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			raw, err := bson.Marshal(bson.D{
				{Key: "guild", Value: "guild-1"},
				{Key: "coin_number", Value: test.value},
				{Key: "sign_coin", Value: test.value},
				{Key: "xp_multiple", Value: test.value},
				{Key: "time", Value: test.value},
			})
			if err != nil {
				t.Fatalf("marshal fixture: %v", err)
			}
			var document GiftChangeDocument
			if err := bson.Unmarshal(raw, &document); err != nil {
				t.Fatalf("decode document: %v", err)
			}
			config := document.ToDomain()
			if config.GachaCostText != test.want || config.SignCoinsText != test.want || config.XPMultipleText != test.want || config.ResetMarkerText != test.want {
				t.Fatalf("config = %#v, want display %q", config, test.want)
			}
		})
	}
}

func TestGiftChangeUpdateFromDomainUsesLegacyFieldNames(t *testing.T) {
	update := GiftChangeUpdateFromDomain(documentTestEconomyConfig())
	raw, err := bson.Marshal(update)
	if err != nil {
		t.Fatalf("marshal update: %v", err)
	}
	var decoded map[string]any
	if err := bson.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("decode update: %v", err)
	}
	if decoded["coin_number"] != int64(700) || decoded["sign_coin"] != int64(30) || decoded["channel"] != "channel-1" || decoded["time"] != int64(3600) {
		t.Fatalf("decoded update = %#v", decoded)
	}
	if decoded["xp_multiple"] != 2.5 {
		t.Fatalf("xp_multiple = %#v", decoded["xp_multiple"])
	}
}

func TestGiftChangeUpdateUsesPreservedCooldownScalar(t *testing.T) {
	update := GiftChangeUpdateFromDomain(domain.EconomyConfig{ResetMarker: math.MaxInt64, ResetMarkerText: "9223372036854778000"})
	raw, err := bson.Marshal(update)
	if err != nil {
		t.Fatalf("marshal update: %v", err)
	}
	var decoded map[string]any
	if err := bson.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("decode update: %v", err)
	}
	if got, ok := decoded["time"].(float64); !ok || got != 9.223372036854778e18 {
		t.Fatalf("time = %#v", decoded["time"])
	}
}

func TestShopItemDocumentPreservesMongoosePriceDisplay(t *testing.T) {
	for _, test := range []struct {
		name  string
		value any
		want  string
	}{
		{name: "null", value: nil, want: "null"},
		{name: "decimal", value: 20.5, want: "20.5"},
		{name: "infinity", value: math.Inf(1), want: "Infinity"},
		{name: "malformed", value: bson.D{{Key: "bad", Value: true}}, want: "undefined"},
	} {
		t.Run(test.name, func(t *testing.T) {
			raw, err := bson.Marshal(bson.D{{Key: "need_coin", Value: test.value}, {Key: "commodity_count", Value: test.value}})
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			var document ShopItemDocument
			if err := bson.Unmarshal(raw, &document); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			item := document.ToDomain()
			if item.NeedCoinsText != test.want || item.CountText != test.want {
				t.Fatalf("item = %#v, want scalar text %q", item, test.want)
			}
		})
	}
}

func documentTestEconomyConfig() domain.EconomyConfig {
	return domain.EconomyConfig{
		GuildID:     "guild-1",
		GachaCost:   700,
		SignCoins:   30,
		ChannelID:   "channel-1",
		XPMultiple:  2.5,
		ResetMarker: 3600,
	}
}

func TestEconomyConfigDefaultGachaCost(t *testing.T) {
	var document GiftChangeDocument
	if got := document.ToDomain().EffectiveGachaCost(); got != 500 {
		t.Fatalf("default gacha cost = %d", got)
	}
}

func TestShopItemWriteDocumentPreservesLegacyTextWhitespace(t *testing.T) {
	write := ShopItemWriteDocumentFromDomain(domain.ShopItem{
		GuildID:     " guild-1 ",
		CommodityID: 1001,
		Name:        " VIP ",
		NeedCoins:   50,
		Description: " role reward ",
		Code:        " CODE ",
		RoleID:      " role-1 ",
		Count:       1,
	})
	raw, err := bson.Marshal(write)
	if err != nil {
		t.Fatalf("marshal shop item: %v", err)
	}
	var decoded map[string]any
	if err := bson.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("decode shop item: %v", err)
	}
	if decoded["guild"] != "guild-1" || decoded["role"] != "role-1" {
		t.Fatalf("identifiers were not normalized: %#v", decoded)
	}
	if decoded["name"] != " VIP " || decoded["commodity_description"] != " role reward " || decoded["code"] != " CODE " {
		t.Fatalf("shop text was normalized: %#v", decoded)
	}
}

func TestSignListDocumentDecodesNestedLegacyDate(t *testing.T) {
	raw, err := bson.Marshal(bson.D{
		{Key: "guild", Value: "guild-1"},
		{Key: "member", Value: "user-1"},
		{Key: "date", Value: bson.D{
			{Key: "2026", Value: bson.D{
				{Key: "07", Value: bson.A{"4", "5"}},
			}},
		}},
	})
	if err != nil {
		t.Fatalf("marshal fixture: %v", err)
	}
	var document SignListDocument
	if err := bson.Unmarshal(raw, &document); err != nil {
		t.Fatalf("decode sign list: %v", err)
	}
	calendar := document.ToDomain()
	if !calendar.HasDay("2026", "07", "4") || !calendar.HasDay("2026", "07", "5") {
		t.Fatalf("calendar = %#v", calendar)
	}
}
