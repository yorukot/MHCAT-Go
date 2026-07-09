package documents_test

import (
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/adapters/mongo/documents"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
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

func TestWarningSettingsDocumentFromDomainStoresLegacyShape(t *testing.T) {
	doc := documents.WarningSettingsDocumentFromDomain(domain.WarningSettings{
		GuildID:   "guild-1",
		Threshold: 3,
		Action:    domain.WarningSettingsActionKick,
	})
	if doc.Guild != "guild-1" || doc.BanCount != "3" || doc.Move != domain.WarningSettingsActionKick {
		t.Fatalf("document = %#v", doc)
	}
}
