package documents

import (
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
	if balance.GuildID != "guild-1" || balance.UserID != "user-1" || balance.Coins != 1234 || balance.Today != 7 {
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
	if balance.Coins != 0 || balance.Today != 0 {
		t.Fatalf("balance = %#v", balance)
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
	if config.GachaCost != 700 || config.SignCoins != 20 || config.ChannelID != "channel-1" || config.XPMultiple != 2.5 || config.ResetMarker != 0 {
		t.Fatalf("config = %#v", config)
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
