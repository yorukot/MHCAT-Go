package documents

import (
	"strconv"
	"strings"

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
	if value.Type == 0 || value.Type == bson.TypeNull || value.Type == bson.TypeUndefined {
		return "0"
	}
	if text, ok := value.StringValueOK(); ok {
		text = strings.TrimSpace(text)
		if text != "" {
			return text
		}
		return "0"
	}
	if parsed, ok := value.DoubleOK(); ok {
		return strconv.FormatFloat(parsed, 'f', -1, 64)
	}
	if parsed, ok := value.AsInt64OK(); ok {
		return strconv.FormatInt(parsed, 10)
	}
	return "0"
}
