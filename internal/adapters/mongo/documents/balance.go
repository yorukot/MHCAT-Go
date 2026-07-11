package documents

import (
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type BalanceDocument struct {
	Guild string        `bson:"guild" json:"guild"`
	Price bson.RawValue `bson:"price" json:"price"`
}

func (d BalanceDocument) ToDomain() domain.Balance {
	return domain.Balance{
		GuildID: d.Guild,
		Amount:  legacyPriceString(d.Price),
	}
}

func legacyPriceString(value bson.RawValue) string {
	if value.Type == 0 || value.Type == bson.TypeUndefined {
		return "undefined"
	}
	if value.Type == bson.TypeNull {
		return "null"
	}
	if text, ok := value.StringValueOK(); ok {
		if text == "" {
			return "null"
		}
	}
	if _, data, ok := value.BinaryOK(); ok {
		parsed, valid := legacyJavaScriptNumericString(string(data))
		if !valid {
			return "undefined"
		}
		return legacyJavaScriptNumberString(parsed)
	}
	parsed, ok := LegacyMongooseNumber(value)
	if !ok {
		return "undefined"
	}
	return legacyJavaScriptNumberString(parsed)
}
