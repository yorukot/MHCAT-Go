package documents

import (
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestBalanceDocumentToDomainPreservesLegacyPrice(t *testing.T) {
	priceType, priceValue, err := bson.MarshalValue("12.5")
	if err != nil {
		t.Fatalf("marshal price: %v", err)
	}
	document := BalanceDocument{
		Guild: "guild-1",
		Price: bson.RawValue{Type: priceType, Value: priceValue},
	}
	balance := document.ToDomain()
	if balance.GuildID != "guild-1" || balance.Amount != "12.5" {
		t.Fatalf("balance = %#v", balance)
	}
}

func TestBalanceDocumentToDomainFormatsNumericPrice(t *testing.T) {
	priceType, priceValue, err := bson.MarshalValue(42.25)
	if err != nil {
		t.Fatalf("marshal price: %v", err)
	}
	balance := (BalanceDocument{
		Guild: "guild-1",
		Price: bson.RawValue{Type: priceType, Value: priceValue},
	}).ToDomain()
	if balance.Amount != "42.25" {
		t.Fatalf("balance = %#v", balance)
	}
}
