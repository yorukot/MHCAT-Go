package documents

import (
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type GiftDocument struct {
	Guild      string        `bson:"guild" json:"guild"`
	GiftName   string        `bson:"gift_name" json:"gift_name"`
	GiftCode   string        `bson:"gift_code" json:"gift_code"`
	GiftChance bson.RawValue `bson:"gift_chence" json:"gift_chence"`
	AutoDelete bool          `bson:"auto_delete" json:"auto_delete"`
	GiftCount  bson.RawValue `bson:"gift_count" json:"gift_count"`
	GiveCoin   bson.RawValue `bson:"give_coin" json:"give_coin"`
}

type GiftWriteDocument struct {
	Guild      string  `bson:"guild" json:"guild"`
	GiftName   string  `bson:"gift_name" json:"gift_name"`
	GiftCode   *string `bson:"gift_code" json:"gift_code"`
	GiftChance float64 `bson:"gift_chence" json:"gift_chence"`
	AutoDelete bool    `bson:"auto_delete" json:"auto_delete"`
	GiftCount  int64   `bson:"gift_count" json:"gift_count"`
	GiveCoin   int64   `bson:"give_coin" json:"give_coin"`
}

func GiftWriteDocumentFromDomain(prize domain.GachaPrizeConfig) GiftWriteDocument {
	document := GiftWriteDocument{
		Guild:      prize.GuildID,
		GiftName:   prize.Name,
		GiftChance: prize.Chance,
		AutoDelete: prize.AutoDelete,
		GiftCount:  prize.Count,
		GiveCoin:   prize.GiveCoin,
	}
	if prize.Code != "" {
		code := prize.Code
		document.GiftCode = &code
	}
	return document
}

func (d GiftDocument) ToDomain() domain.GachaPrize {
	return domain.GachaPrize{
		GuildID: d.Guild,
		Name:    d.GiftName,
		Chance:  legacyFloat64(d.GiftChance),
		Count:   legacyInt64(d.GiftCount),
	}
}
