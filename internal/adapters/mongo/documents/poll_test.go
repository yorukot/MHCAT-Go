package documents

import (
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestPollDocumentLegacyBSONDecodes(t *testing.T) {
	payload, err := bson.Marshal(bson.D{
		{Key: "guild", Value: "guild-1"},
		{Key: "messageid", Value: "message-1"},
		{Key: "question", Value: "問題"},
		{Key: "create_member_id", Value: "owner-1"},
		{Key: "many_choose", Value: 2},
		{Key: "can_change_choose", Value: true},
		{Key: "can_see_result", Value: false},
		{Key: "end", Value: false},
		{Key: "anonymous", Value: false},
		{Key: "choose_data", Value: bson.A{"A", "B"}},
		{Key: "join_member", Value: bson.A{bson.D{
			{Key: "id", Value: "user-1"},
			{Key: "choise", Value: "A"},
			{Key: "time", Value: "1700000000000"},
		}}},
	})
	if err != nil {
		t.Fatalf("marshal legacy poll: %v", err)
	}
	var document PollDocument
	if err := bson.Unmarshal(payload, &document); err != nil {
		t.Fatalf("decode legacy poll: %v", err)
	}
	poll := document.ToDomain()
	if poll.GuildID != "guild-1" || poll.MessageID != "message-1" || poll.MaxChoices != 2 || len(poll.Votes) != 1 || poll.Votes[0].Choice != "A" {
		t.Fatalf("poll = %#v", poll)
	}
}

func TestPollDocumentMissingFieldsDecodeSafe(t *testing.T) {
	payload, err := bson.Marshal(bson.D{{Key: "guild", Value: "guild-1"}, {Key: "messageid", Value: "message-1"}})
	if err != nil {
		t.Fatalf("marshal partial poll: %v", err)
	}
	var document PollDocument
	if err := bson.Unmarshal(payload, &document); err != nil {
		t.Fatalf("decode partial poll: %v", err)
	}
	poll := document.ToDomain()
	if poll.MaxChoices != 1 || poll.GuildID != "guild-1" || poll.MessageID != "message-1" {
		t.Fatalf("partial poll = %#v", poll)
	}
}

func TestPollDocumentRoundTripDomainPreservesChoiseField(t *testing.T) {
	poll := domain.Poll{
		GuildID:    "guild-1",
		MessageID:  "message-1",
		Question:   "問題",
		CreatorID:  "owner-1",
		MaxChoices: 1,
		Choices:    []string{"A", "B"},
		Votes:      []domain.PollVote{{UserID: "user-1", Choice: "A", Time: "1"}},
	}
	document := PollDocumentFromDomain(poll)
	if len(document.JoinMember) != 1 || document.JoinMember[0].Choice != "A" {
		t.Fatalf("document = %#v", document)
	}
	if got := document.ToDomain(); got.Votes[0].Choice != poll.Votes[0].Choice || got.Choices[0] != "A" {
		t.Fatalf("round trip = %#v", got)
	}
}
