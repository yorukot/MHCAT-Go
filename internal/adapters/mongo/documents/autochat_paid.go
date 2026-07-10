package documents

import (
	"math"
	"strconv"
	"strings"
	"unicode"

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
	parsed, ok := LegacyMongooseNumber(value)
	const int64Limit = float64(uint64(1) << 63)
	if !ok || math.IsInf(parsed, 0) || math.Trunc(parsed) != parsed || parsed < -int64Limit || parsed >= int64Limit {
		return 0, false
	}
	return int64(parsed), true
}

func LegacyMongooseNumber(value bson.RawValue) (float64, bool) {
	switch value.Type {
	case 0, bson.TypeUndefined:
		return 0, false
	case bson.TypeNull:
		return 0, true
	}
	if text, ok := value.StringValueOK(); ok {
		return legacyJavaScriptNumericString(text)
	}
	if text, ok := value.SymbolOK(); ok {
		return legacyJavaScriptNumericString(text)
	}
	if parsed, ok := value.BooleanOK(); ok {
		if parsed {
			return 1, true
		}
		return 0, true
	}
	if parsed, ok := value.AsFloat64OK(); ok {
		if math.IsNaN(parsed) {
			return 0, false
		}
		return parsed, true
	}
	if parsed, ok := value.DateTimeOK(); ok {
		return float64(parsed), true
	}
	if parsed, ok := value.Decimal128OK(); ok {
		return legacyJavaScriptNumericString(parsed.String())
	}
	if timestamp, increment, ok := value.TimestampOK(); ok {
		return float64(uint64(timestamp)<<32 | uint64(increment)), true
	}
	if parsed, ok := value.ObjectIDOK(); ok {
		return legacyJavaScriptNumericString(parsed.Hex())
	}
	return 0, false
}

func legacyJavaScriptNumericString(value string) (float64, bool) {
	value = strings.TrimFunc(value, func(r rune) bool {
		return unicode.IsSpace(r) || r == '\ufeff'
	})
	if value == "" {
		return 0, true
	}
	switch value {
	case "Infinity", "+Infinity":
		return math.Inf(1), true
	case "-Infinity":
		return math.Inf(-1), true
	}
	if len(value) > 2 && value[0] == '0' {
		base := 0
		switch value[1] {
		case 'x', 'X':
			base = 16
		case 'b', 'B':
			base = 2
		case 'o', 'O':
			base = 8
		}
		if base != 0 {
			parsed, err := strconv.ParseUint(value[2:], base, 64)
			return float64(parsed), err == nil
		}
	}
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil || math.IsNaN(parsed) || math.IsInf(parsed, 0) {
		return 0, false
	}
	return parsed, true
}
