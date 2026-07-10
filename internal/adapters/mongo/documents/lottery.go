package documents

import (
	"strconv"
	"strings"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type LotteryDocument struct {
	Guild           string        `bson:"guild" json:"guild"`
	Date            bson.RawValue `bson:"date" json:"date"`
	Gift            string        `bson:"gift" json:"gift"`
	HowManyWinner   bson.RawValue `bson:"howmanywinner" json:"howmanywinner"`
	ID              string        `bson:"id" json:"id"`
	Member          bson.RawValue `bson:"member" json:"member"`
	End             bson.RawValue `bson:"end" json:"end"`
	MessageChannel  string        `bson:"message_channel" json:"message_channel"`
	RequiredRole    bson.RawValue `bson:"yesrole" json:"yesrole"`
	ForbiddenRole   bson.RawValue `bson:"norole" json:"norole"`
	MaxParticipants bson.RawValue `bson:"maxNumber" json:"maxNumber"`
	Owner           bson.RawValue `bson:"owner" json:"owner"`
}

func (d LotteryDocument) ToDomain() domain.Lottery {
	return domain.Lottery{
		GuildID:         d.Guild,
		ID:              d.ID,
		EndsAtUnix:      legacyInt64(d.Date),
		Gift:            d.Gift,
		WinnerCount:     int(legacyInt64(d.HowManyWinner)),
		Participants:    lotteryParticipants(d.Member),
		Ended:           lotteryLegacyBool(d.End),
		ChannelID:       d.MessageChannel,
		RequiredRoleID:  lotteryLegacyString(d.RequiredRole),
		ForbiddenRoleID: lotteryLegacyString(d.ForbiddenRole),
		MaxParticipants: int(legacyInt64(d.MaxParticipants)),
		OwnerID:         lotteryLegacyString(d.Owner),
	}.Normalized()
}

func lotteryParticipants(value bson.RawValue) []domain.LotteryParticipant {
	array, ok := value.ArrayOK()
	if !ok {
		return nil
	}
	values, err := array.Values()
	if err != nil {
		return nil
	}
	participants := make([]domain.LotteryParticipant, 0, len(values))
	for _, value := range values {
		document, ok := value.DocumentOK()
		if !ok {
			continue
		}
		userID := strings.TrimSpace(lotteryLegacyString(document.Lookup("id")))
		if userID == "" {
			continue
		}
		joinedAt := document.Lookup("time")
		participants = append(participants, domain.LotteryParticipant{
			UserID:         userID,
			JoinedAtMillis: legacyInt64(joinedAt),
			JoinedAtRaw:    lotteryLegacyString(joinedAt),
		})
	}
	return participants
}

func lotteryLegacyString(value bson.RawValue) string {
	if value.Type == 0 || value.Type == bson.TypeNull || value.Type == bson.TypeUndefined {
		return ""
	}
	if text, ok := value.StringValueOK(); ok {
		return strings.TrimSpace(text)
	}
	if parsed, ok := value.AsInt64OK(); ok {
		return strconv.FormatInt(parsed, 10)
	}
	if parsed, ok := value.DoubleOK(); ok {
		return strconv.FormatFloat(parsed, 'f', -1, 64)
	}
	return ""
}

func lotteryLegacyBool(value bson.RawValue) bool {
	if parsed, ok := value.BooleanOK(); ok {
		return parsed
	}
	parsed, err := strconv.ParseBool(lotteryLegacyString(value))
	return err == nil && parsed
}
