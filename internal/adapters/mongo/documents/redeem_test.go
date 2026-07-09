package documents

import (
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
