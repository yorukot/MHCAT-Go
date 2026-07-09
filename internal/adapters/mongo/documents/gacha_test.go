package documents

import (
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
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

func TestGiftWriteDocumentFromDomainPreservesLegacyBSONFields(t *testing.T) {
	document := GiftWriteDocumentFromDomain(domain.GachaPrizeConfig{
		GuildID:    "guild-1",
		Name:       "大獎",
		Code:       "code-1",
		Chance:     12.5,
		AutoDelete: false,
		Count:      3,
		GiveCoin:   7,
	})
	if document.Guild != "guild-1" || document.GiftName != "大獎" || document.GiftCode == nil || *document.GiftCode != "code-1" {
		t.Fatalf("document identity = %#v", document)
	}
	if document.GiftChance != 12.5 || document.AutoDelete || document.GiftCount != 3 || document.GiveCoin != 7 {
		t.Fatalf("document values = %#v", document)
	}
}

func TestGiftWriteDocumentFromDomainStoresMissingCodeAsNil(t *testing.T) {
	document := GiftWriteDocumentFromDomain(domain.GachaPrizeConfig{GuildID: "guild-1", Name: "大獎", Count: 1})
	if document.GiftCode != nil {
		t.Fatalf("expected nil gift code, got %#v", document.GiftCode)
	}
}
