package documents

import (
	"math"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type RedeemCodeDocument struct {
	ID    any           `bson:"_id" json:"_id"`
	Code  string        `bson:"code" json:"code"`
	Price bson.RawValue `bson:"price" json:"price"`
	Time  bson.RawValue `bson:"time" json:"time"`
}

func (d RedeemCodeDocument) ToDomain() domain.RedeemCode {
	return domain.RedeemCode{
		Identity:        d.ID,
		Code:            d.Code,
		Price:           legacyRedeemNumber(d.Price),
		CreatedAtMillis: legacyRedeemNumber(d.Time),
	}
}

func LegacyBalancePriceFloat(value bson.RawValue) float64 {
	return legacyRedeemNumber(value)
}

func legacyRedeemNumber(value bson.RawValue) float64 {
	parsed, ok := LegacyMongooseNumber(value)
	if !ok {
		return math.NaN()
	}
	return parsed
}
