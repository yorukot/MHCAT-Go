package documents

import (
	"math"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type AutoChatPaidDocument struct {
	ID      any           `bson:"_id" json:"_id"`
	Guild   string        `bson:"guild" json:"guild"`
	ResidC  bson.RawValue `bson:"resid_c" json:"resid_c"`
	ResidP  bson.RawValue `bson:"resid_p" json:"resid_p"`
	Reply   bool          `bson:"reply" json:"reply"`
	Message string        `bson:"message" json:"message"`
	Time    bson.RawValue `bson:"time" json:"time"`
}

func (d AutoChatPaidDocument) ToResponse() (domain.AutoChatPaidResponse, bool) {
	timeMilli, ok := LegacyExactInt64(d.Time)
	if !ok {
		return domain.AutoChatPaidResponse{}, false
	}
	return domain.AutoChatPaidResponse{
		GuildID:          d.Guild,
		Content:          d.Message,
		RequestTimeMilli: timeMilli,
		Reply:            d.Reply,
	}, true
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
