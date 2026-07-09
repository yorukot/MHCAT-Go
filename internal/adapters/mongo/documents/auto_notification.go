package documents

import (
	"strconv"
	"strings"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type AutoNotificationScheduleDocument struct {
	Guild   string        `bson:"guild" json:"guild"`
	Channel string        `bson:"channel" json:"channel"`
	ID      string        `bson:"id" json:"id"`
	Cron    bson.RawValue `bson:"cron" json:"cron"`
}

func (d AutoNotificationScheduleDocument) ToDomain() domain.AutoNotificationSchedule {
	cron, pending := legacyNullableString(d.Cron)
	return domain.AutoNotificationSchedule{
		GuildID:   d.Guild,
		ID:        d.ID,
		Cron:      cron,
		ChannelID: d.Channel,
		Pending:   pending,
	}
}

func legacyNullableString(value bson.RawValue) (string, bool) {
	if value.Type == 0 || value.Type == bson.TypeNull || value.Type == bson.TypeUndefined {
		return "", true
	}
	if text, ok := value.StringValueOK(); ok {
		return text, false
	}
	if parsed, ok := value.AsInt64OK(); ok {
		return strconv.FormatInt(parsed, 10), false
	}
	if parsed, ok := value.DoubleOK(); ok {
		return strings.TrimRight(strings.TrimRight(strconv.FormatFloat(parsed, 'f', 6, 64), "0"), "."), false
	}
	return "", false
}
