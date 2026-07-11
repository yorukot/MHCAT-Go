package documents

import (
	"math"
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestRedeemCodeDocumentDecodesLegacyNumbers(t *testing.T) {
	priceType, priceValue, err := bson.MarshalValue(12.5)
	if err != nil {
		t.Fatalf("marshal price: %v", err)
	}
	timeType, timeValue, err := bson.MarshalValue(int64(1700000000000))
	if err != nil {
		t.Fatalf("marshal time: %v", err)
	}
	code := (RedeemCodeDocument{
		Code:  "abc",
		Price: bson.RawValue{Type: priceType, Value: priceValue},
		Time:  bson.RawValue{Type: timeType, Value: timeValue},
	}).ToDomain()
	if code.Code != "abc" || code.Price != 12.5 || code.CreatedAtMillis != 1700000000000 {
		t.Fatalf("code = %#v", code)
	}
}

func TestRedeemCodeDocumentPreservesMongooseNumberEdges(t *testing.T) {
	for _, test := range []struct {
		name      string
		price     any
		time      any
		wantPrice float64
		wantTime  float64
	}{
		{name: "negative price", price: -5, time: int64(1700000000000), wantPrice: -5, wantTime: 1700000000000},
		{name: "null values", price: nil, time: nil},
		{name: "numeric strings", price: "0x10", time: "1e3", wantPrice: 16, wantTime: 1000},
		{name: "missing values", price: bson.Undefined{}, time: bson.Undefined{}, wantPrice: math.NaN(), wantTime: math.NaN()},
		{name: "malformed values", price: "abc", time: bson.D{{Key: "bad", Value: 1}}, wantPrice: math.NaN(), wantTime: math.NaN()},
	} {
		t.Run(test.name, func(t *testing.T) {
			got := (RedeemCodeDocument{
				Code:  "abc",
				Price: redeemRawValue(t, test.price),
				Time:  redeemRawValue(t, test.time),
			}).ToDomain()
			if !sameRedeemFloat(got.Price, test.wantPrice) || !sameRedeemFloat(got.CreatedAtMillis, test.wantTime) {
				t.Fatalf("code = %#v, want price=%v time=%v", got, test.wantPrice, test.wantTime)
			}
		})
	}
}

func sameRedeemFloat(left float64, right float64) bool {
	return left == right || math.IsNaN(left) && math.IsNaN(right)
}

func redeemRawValue(t *testing.T, value any) bson.RawValue {
	t.Helper()
	if value == nil {
		return bson.RawValue{Type: bson.TypeNull}
	}
	valueType, encoded, err := bson.MarshalValue(value)
	if err != nil {
		t.Fatalf("marshal %T: %v", value, err)
	}
	return bson.RawValue{Type: valueType, Value: encoded}
}

func TestRedeemCodeDocumentDecodesLegacyStringNumbers(t *testing.T) {
	priceType, priceValue, err := bson.MarshalValue("12.5")
	if err != nil {
		t.Fatalf("marshal price: %v", err)
	}
	timeType, timeValue, err := bson.MarshalValue("1700000000000")
	if err != nil {
		t.Fatalf("marshal time: %v", err)
	}
	code := (RedeemCodeDocument{
		Code:  "abc",
		Price: bson.RawValue{Type: priceType, Value: priceValue},
		Time:  bson.RawValue{Type: timeType, Value: timeValue},
	}).ToDomain()
	if code.Price != 12.5 || code.CreatedAtMillis != 1700000000000 {
		t.Fatalf("code = %#v", code)
	}
}
