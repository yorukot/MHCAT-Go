package documents_test

import (
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/adapters/mongo/documents"
)

func TestWarningDocumentToDomain(t *testing.T) {
	doc := documents.WarningDocument{
		Guild: "guild-1",
		User:  "user-1",
		Content: []documents.WarningEntryDocument{{
			Moderator: "mod-1",
			Reason:    "reason",
			Time:      "2026-07-04",
		}},
	}
	got := doc.ToDomain()
	if got.GuildID != "guild-1" || got.UserID != "user-1" || len(got.Entries) != 1 {
		t.Fatalf("history = %#v", got)
	}
	if got.Entries[0].ModeratorID != "mod-1" || got.Entries[0].Reason != "reason" || got.Entries[0].Time != "2026-07-04" {
		t.Fatalf("entry = %#v", got.Entries[0])
	}
}
