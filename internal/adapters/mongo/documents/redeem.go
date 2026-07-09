package documents

import (
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type RedeemCodeDocument struct {
	Code  string        `bson:"code" json:"code"`
	Price bson.RawValue `bson:"price" json:"price"`
	Time  bson.RawValue `bson:"time" json:"time"`
}

func (d RedeemCodeDocument) ToDomain() domain.RedeemCode {
	return domain.RedeemCode{
		Code:            d.Code,
		Price:           legacyFloat64(d.Price),
		CreatedAtMillis: legacyInt64(d.Time),
	}
}

func LegacyBalancePriceFloat(value bson.RawValue) float64 {
	return legacyFloat64(value)
}
