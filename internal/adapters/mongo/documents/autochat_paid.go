package documents

import (
	"math"
	"strconv"
	"strings"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type AutoChatPaidDocument struct {
	ID      any           `bson:"_id" json:"_id"`
	Guild   string        `bson:"guild" json:"guild"`
	ResidC  bson.RawValue `bson:"resid_c" json:"resid_c"`
	ResidP  bson.RawValue `bson:"resid_p" json:"resid_p"`
	Reply   bson.RawValue `bson:"reply" json:"reply"`
	Message bson.RawValue `bson:"message" json:"message"`
	Time    bson.RawValue `bson:"time" json:"time"`
}

func (d AutoChatPaidDocument) ToResponse() (domain.AutoChatPaidResponse, bool) {
	timeMilli, ok := LegacyExactInt64(d.Time)
	if !ok {
		return domain.AutoChatPaidResponse{}, false
	}
	message, ok := legacyMongooseString(d.Message)
	if !ok {
		return domain.AutoChatPaidResponse{}, false
	}
	return domain.AutoChatPaidResponse{
		GuildID:          d.Guild,
		Content:          message,
		RequestTimeMilli: timeMilli,
		Reply:            legacyMongooseBoolean(d.Reply),
	}, true
}

func legacyMongooseString(value bson.RawValue) (string, bool) {
	if text, ok := value.StringValueOK(); ok {
		return text, true
	}
	if text, ok := value.SymbolOK(); ok {
		return text, true
	}
	if text, ok := value.JavaScriptOK(); ok {
		return text, true
	}
	if parsed, ok := value.BooleanOK(); ok {
		return strconv.FormatBool(parsed), true
	}
	if parsed, ok := value.DoubleOK(); ok {
		return legacyJavaScriptNumberString(parsed), true
	}
	if parsed, ok := value.Int32OK(); ok {
		return strconv.FormatInt(int64(parsed), 10), true
	}
	if parsed, ok := value.Int64OK(); ok {
		return strconv.FormatInt(parsed, 10), true
	}
	if parsed, ok := value.Decimal128OK(); ok {
		return parsed.String(), true
	}
	if parsed, ok := value.ObjectIDOK(); ok {
		return parsed.Hex(), true
	}
	return "", false
}

func legacyMongooseBoolean(value bson.RawValue) bool {
	if parsed, ok := value.BooleanOK(); ok {
		return parsed
	}
	if parsed, ok := value.StringValueOK(); ok {
		switch parsed {
		case "true", "1", "yes":
			return true
		case "false", "0", "no":
			return false
		}
	}
	if parsed, ok := value.AsFloat64OK(); ok {
		return parsed == 1
	}
	return false
}

func legacyJavaScriptNumberString(value float64) string {
	switch {
	case math.IsNaN(value):
		return "NaN"
	case math.IsInf(value, 1):
		return "Infinity"
	case math.IsInf(value, -1):
		return "-Infinity"
	case value == 0:
		return "0"
	}
	absolute := math.Abs(value)
	if absolute >= 1e-6 && absolute < 1e21 {
		return strconv.FormatFloat(value, 'f', -1, 64)
	}
	mantissa, exponent, _ := strings.Cut(strconv.FormatFloat(value, 'e', -1, 64), "e")
	parsedExponent, _ := strconv.Atoi(exponent)
	if parsedExponent >= 0 {
		return mantissa + "e+" + strconv.Itoa(parsedExponent)
	}
	return mantissa + "e" + strconv.Itoa(parsedExponent)
}

func LegacyExactInt64(value bson.RawValue) (int64, bool) {
	switch value.Type {
	case bson.TypeInt32:
		return int64(value.Int32()), true
	case bson.TypeInt64:
		return value.Int64(), true
	case bson.TypeDouble:
		parsed := value.Double()
		const int64Limit = float64(uint64(1) << 63)
		if math.IsNaN(parsed) || math.IsInf(parsed, 0) || math.Trunc(parsed) != parsed || parsed < -int64Limit || parsed >= int64Limit {
			return 0, false
		}
		return int64(parsed), true
	default:
		return 0, false
	}
}
