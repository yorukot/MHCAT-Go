package documents

import (
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestGiftDocumentToDomainPreservesLegacyFields(t *testing.T) {
	chanceType, chanceValue, err := bson.MarshalValue("12.5")
	if err != nil {
		t.Fatalf("marshal chance: %v", err)
	}
	countType, countValue, err := bson.MarshalValue(int32(3))
	if err != nil {
		t.Fatalf("marshal count: %v", err)
	}
	document := GiftDocument{
		Guild:      "guild-1",
		GiftName:   "大獎",
		GiftChance: bson.RawValue{Type: chanceType, Value: chanceValue},
		GiftCount:  bson.RawValue{Type: countType, Value: countValue},
	}
	prize := document.ToDomain()
	if prize.GuildID != "guild-1" || prize.Name != "大獎" || prize.Chance != 12.5 || prize.Count != 3 {
		t.Fatalf("prize = %#v", prize)
	}
}
