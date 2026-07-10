package documents

import (
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestLotteryDocumentDecodesLegacyMixedValues(t *testing.T) {
	raw, err := bson.Marshal(bson.D{
		{Key: "guild", Value: "guild-1"},
		{Key: "id", Value: "1700000000000999lotter"},
		{Key: "date", Value: "1700000300"},
		{Key: "gift", Value: "gift"},
		{Key: "howmanywinner", Value: "2"},
		{Key: "member", Value: bson.A{
			bson.D{{Key: "id", Value: "user-1"}, {Key: "time", Value: int64(1700000000000)}},
			bson.D{{Key: "id", Value: "user-2"}, {Key: "time", Value: "legacy time"}},
			"malformed",
		}},
		{Key: "end", Value: "false"},
		{Key: "message_channel", Value: "channel-1"},
		{Key: "yesrole", Value: nil},
		{Key: "norole", Value: "role-2"},
		{Key: "maxNumber", Value: int32(10)},
		{Key: "owner", Value: "owner-1"},
	})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var document LotteryDocument
	if err := bson.Unmarshal(raw, &document); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	lottery := document.ToDomain()
	if lottery.EndsAtUnix != 1700000300 || lottery.WinnerCount != 2 || lottery.Ended || lottery.RequiredRoleID != "" || lottery.ForbiddenRoleID != "role-2" || lottery.MaxParticipants != 10 {
		t.Fatalf("lottery = %#v", lottery)
	}
	if len(lottery.Participants) != 2 || lottery.Participants[0].JoinedAtMillis != 1700000000000 || lottery.Participants[1].JoinedAtRaw != "legacy time" {
		t.Fatalf("participants = %#v", lottery.Participants)
	}
}
