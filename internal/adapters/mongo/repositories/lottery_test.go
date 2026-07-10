package repositories

import (
	"strings"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestNewLotteryRepositoryRequiresCollection(t *testing.T) {
	if _, err := NewLotteryRepository(nil); err == nil {
		t.Fatal("expected nil collection error")
	}
	if _, err := NewLotteryRepositoryFromDatabase(nil); err == nil {
		t.Fatal("expected nil database error")
	}
}

func TestLotteryJoinFilterOwnsLegacyEntryGuards(t *testing.T) {
	request := domain.LotteryJoinRequest{GuildID: "guild-1", ID: "id-1", UserID: "user-1", JoinedAtMillis: 1700000000000, NowUnix: 1700000000}
	encoded, err := bson.MarshalExtJSON(lotteryJoinFilter(request), false, false)
	if err != nil {
		t.Fatalf("marshal filter: %v", err)
	}
	text := string(encoded)
	for _, expected := range []string{`"guild":"guild-1"`, `"id":"id-1"`, `"end":{"$ne":true}`, `"member":{"$not":{"$elemMatch":{"id":"user-1"}}}`, `"$date"`, `"$maxNumber"`, `"$isArray":"$member"`} {
		if !strings.Contains(text, expected) {
			t.Fatalf("filter missing %s: %s", expected, text)
		}
	}
}

func TestLotteryJoinUpdateAppendsRollbackCompatibleParticipant(t *testing.T) {
	request := domain.LotteryJoinRequest{UserID: "user-1", JoinedAtMillis: 1700000000000}
	encoded, err := bson.MarshalExtJSON(bson.D{{Key: "update", Value: lotteryJoinUpdate(request)}}, false, false)
	if err != nil {
		t.Fatalf("marshal update: %v", err)
	}
	text := string(encoded)
	for _, expected := range []string{`"$concatArrays"`, `"id":"user-1"`, `"time":1700000000000`} {
		if !strings.Contains(text, expected) {
			t.Fatalf("update missing %s: %s", expected, text)
		}
	}
}
