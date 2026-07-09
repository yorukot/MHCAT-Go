package repositories

import (
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
)

func TestDeleteDataCollectionNames(t *testing.T) {
	want := map[domain.DeleteDataTarget]string{
		domain.DeleteDataTargetJoinMessage:  "join_messages",
		domain.DeleteDataTargetLeaveMessage: "leave_messages",
		domain.DeleteDataTargetLogging:      "loggings",
		domain.DeleteDataTargetStats:        "numbers",
		domain.DeleteDataTargetAutoChat:     "chats",
		domain.DeleteDataTargetVerification: "verifications",
		domain.DeleteDataTargetTextXP:       "text_xp_channels",
		domain.DeleteDataTargetVoiceXP:      "voice_xp_channels",
		domain.DeleteDataTargetTicket:       "tickets",
	}
	for target, collection := range want {
		got, ok := DeleteDataCollectionName(target)
		if !ok || got != collection {
			t.Fatalf("target %s collection = %q ok=%v, want %q", target, got, ok, collection)
		}
	}
	if _, ok := DeleteDataCollectionName("bad"); ok {
		t.Fatal("unknown target should not map to a collection")
	}
}
