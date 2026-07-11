package documents

import (
	"math"
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
		{name: "null", value: nil, want: "null"},
		{name: "empty string", value: "", want: "null"},
		{name: "whitespace string", value: "   ", want: "0"},
		{name: "numeric string", value: "12.5", want: "12.5"},
		{name: "hex string", value: "0x10", want: "16"},
		{name: "exponent string", value: "1e3", want: "1000"},
		{name: "malformed string", value: "abc", want: "undefined"},
		{name: "true", value: true, want: "1"},
		{name: "false", value: false, want: "0"},
		{name: "date", value: time.UnixMilli(-5), want: "-5"},
		{name: "decimal", value: decimal, want: "-3.5"},
		{name: "safe int64", value: int64(9_007_199_254_740_991), want: "9007199254740991"},
		{name: "rounded int64", value: int64(9_007_199_254_740_993), want: "9007199254740992"},
		{name: "small scientific", value: 1e-7, want: "1e-7"},
		{name: "decimal threshold", value: 1e20, want: "100000000000000000000"},
		{name: "scientific threshold", value: 1e21, want: "1e+21"},
		{name: "positive infinity", value: math.Inf(1), want: "Infinity"},
		{name: "negative infinity", value: math.Inf(-1), want: "-Infinity"},
		{name: "nan", value: math.NaN(), want: "undefined"},
		{name: "numeric binary", value: bson.Binary{Subtype: 0, Data: []byte("5")}, want: "5"},
		{name: "malformed binary", value: bson.Binary{Subtype: 0, Data: []byte("abc")}, want: "undefined"},
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
