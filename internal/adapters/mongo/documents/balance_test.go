package documents

import (
	"testing"
	"time"

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

func TestBalanceDocumentToDomainPreservesMongooseNumberCategories(t *testing.T) {
	decimal, err := bson.ParseDecimal128("-3.5")
	if err != nil {
		t.Fatalf("parse decimal: %v", err)
	}
	for _, test := range []struct {
		name  string
		value any
		want  string
	}{
		{name: "undefined", value: bson.Undefined{}, want: "undefined"},
		{name: "null", value: nil, want: "0"},
		{name: "true", value: true, want: "1"},
		{name: "false", value: false, want: "0"},
		{name: "date", value: time.UnixMilli(-5), want: "-5"},
		{name: "decimal", value: decimal, want: "-3.5"},
		{name: "document", value: bson.D{{Key: "invalid", Value: 1}}, want: "undefined"},
	} {
		t.Run(test.name, func(t *testing.T) {
			if test.value == nil {
				if got := legacyPriceString(bson.RawValue{Type: bson.TypeNull}); got != test.want {
					t.Fatalf("legacyPriceString() = %q, want %q", got, test.want)
				}
				return
			}
			valueType, value, err := bson.MarshalValue(test.value)
			if err != nil {
				t.Fatalf("marshal value: %v", err)
			}
			got := legacyPriceString(bson.RawValue{Type: valueType, Value: value})
			if got != test.want {
				t.Fatalf("legacyPriceString() = %q, want %q", got, test.want)
			}
		})
	}

	if got := legacyPriceString(bson.RawValue{}); got != "undefined" {
		t.Fatalf("missing price = %q, want undefined", got)
	}
}
