package domain

import (
	"testing"
	"time"
)

func TestLotteryLegacyStateHelpers(t *testing.T) {
	lottery := Lottery{
		EndsAtUnix:      200,
		MaxParticipants: 2,
		Participants: []LotteryParticipant{
			{UserID: " user-1 "},
			{UserID: "user-2"},
			{UserID: ""},
		},
	}.Normalized()
	if !lottery.HasParticipant("user-1") || !lottery.AtCapacity() {
		t.Fatalf("lottery = %#v", lottery)
	}
	if lottery.IsExpired(time.Unix(200, 0)) {
		t.Fatal("legacy lottery remains enterable through its ending second")
	}
	if !lottery.IsExpired(time.Unix(201, 0)) {
		t.Fatal("lottery should expire after its ending second")
	}
}

func TestLotteryJoinRequestValidation(t *testing.T) {
	valid := LotteryJoinRequest{GuildID: "guild-1", ID: "id-1", UserID: "user-1", JoinedAtMillis: 1000, NowUnix: 1}
	if err := valid.Validate(); err != nil {
		t.Fatalf("valid request: %v", err)
	}
	valid.UserID = ""
	if err := valid.Validate(); err == nil {
		t.Fatal("expected invalid request")
	}
}
