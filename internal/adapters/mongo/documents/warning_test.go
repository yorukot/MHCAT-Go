package documents_test

import (
	"errors"
	"math"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/adapters/mongo/documents"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
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

func TestWarningReadDocumentUsesMongooseMixedArrayCoercion(t *testing.T) {
	moderatorID := bson.NewObjectID()
	payload, err := bson.Marshal(bson.D{
		{Key: "guild", Value: int64(123)},
		{Key: "user", Value: true},
		{Key: "content", Value: bson.A{
			bson.D{{Key: "moderator", Value: moderatorID}, {Key: "reason", Value: int32(7)}, {Key: "time", Value: false}},
			bson.D{{Key: "reason", Value: bson.D{{Key: "nested", Value: true}}}, {Key: "time", Value: bson.A{1, nil, "x"}}},
			"malformed-entry",
		}},
	})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var document documents.WarningReadDocument
	if err := bson.Unmarshal(payload, &document); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	history := document.ToDomain()
	if history.GuildID != "123" || history.UserID != "true" || len(history.Entries) != 3 {
		t.Fatalf("history = %#v", history)
	}
	if got := history.Entries[0]; got.ModeratorID != moderatorID.Hex() || got.Reason != "7" || got.Time != "false" {
		t.Fatalf("scalar entry = %#v", got)
	}
	if got := history.Entries[1]; got.ModeratorID != "undefined" || got.Reason != "[object Object]" || got.Time != "1,,x" {
		t.Fatalf("compound entry = %#v", got)
	}
	if got := history.Entries[2]; got.ModeratorID != "undefined" || got.Reason != "undefined" || got.Time != "undefined" {
		t.Fatalf("malformed entry = %#v", got)
	}
}

func TestWarningReadDocumentWrapsMongooseArrayScalar(t *testing.T) {
	payload, err := bson.Marshal(bson.D{{Key: "guild", Value: "guild-1"}, {Key: "user", Value: "user-1"}, {Key: "content", Value: "malformed-entry"}})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var document documents.WarningReadDocument
	if err := bson.Unmarshal(payload, &document); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if history := document.ToDomain(); len(history.Entries) != 1 || history.Entries[0].Reason != "undefined" {
		t.Fatalf("history = %#v", history)
	}
}

func TestWarningReadDocumentContentValuesPreserveMixedElements(t *testing.T) {
	payload, err := bson.Marshal(bson.D{
		{Key: "_id", Value: bson.NewObjectID()},
		{Key: "content", Value: bson.A{bson.D{{Key: "reason", Value: 7}}, "raw", nil}},
	})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var document documents.WarningReadDocument
	if err := bson.Unmarshal(payload, &document); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	values, isArray, err := document.ContentValues()
	if err != nil {
		t.Fatalf("content values: %v", err)
	}
	if !isArray || len(values) != 3 || values[1] != "raw" || values[2] != nil {
		t.Fatalf("values = %#v isArray=%v", values, isArray)
	}

	scalarPayload, err := bson.Marshal(bson.D{{Key: "content", Value: "raw"}})
	if err != nil {
		t.Fatalf("marshal scalar: %v", err)
	}
	if err := bson.Unmarshal(scalarPayload, &document); err != nil {
		t.Fatalf("unmarshal scalar: %v", err)
	}
	values, isArray, err = document.ContentValues()
	if err != nil || isArray || len(values) != 1 || values[0] != "raw" {
		t.Fatalf("scalar values = %#v isArray=%v err=%v", values, isArray, err)
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

func TestWarningEntryDocumentFromIssueStoresLegacyShape(t *testing.T) {
	doc := documents.WarningEntryDocumentFromIssue(domain.WarningIssue{
		ModeratorID: "mod-1",
		Reason:      "spam",
		Time:        "2026年07月04日 18點30分",
	})
	if doc.Moderator != "mod-1" || doc.Reason != "spam" || doc.Time != "2026年07月04日 18點30分" {
		t.Fatalf("document = %#v", doc)
	}
}

func TestWarningSettingsDocumentToDomain(t *testing.T) {
	got, err := (documents.WarningSettingsDocument{
		Guild:    "guild-1",
		BanCount: "2",
		Move:     domain.WarningSettingsActionBan,
	}).ToDomain()
	if err != nil {
		t.Fatalf("to domain: %v", err)
	}
	if got.GuildID != "guild-1" || got.Threshold != 2 || got.Action != domain.WarningSettingsActionBan {
		t.Fatalf("settings = %#v", got)
	}
	invalid, err := (documents.WarningSettingsDocument{Guild: "guild-1", BanCount: "bad", Move: domain.WarningSettingsActionBan}).ToDomain()
	if err != nil || !math.IsNaN(invalid.Threshold) {
		t.Fatalf("invalid threshold settings = %#v err=%v", invalid, err)
	}
}

func TestWarningSettingsReadDocumentUsesMongooseStringCoercion(t *testing.T) {
	payload, err := bson.Marshal(bson.D{
		{Key: "guild", Value: bson.NewObjectID()},
		{Key: "ban_count", Value: int32(2)},
		{Key: "move", Value: domain.WarningSettingsActionBan},
	})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var document documents.WarningSettingsReadDocument
	if err := bson.Unmarshal(payload, &document); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	settings, err := document.ToDomain()
	if err != nil {
		t.Fatalf("to domain: %v", err)
	}
	if settings.GuildID == "" || settings.Threshold != 2 || settings.Action != domain.WarningSettingsActionBan {
		t.Fatalf("settings = %#v", settings)
	}
}

func TestWarningSettingsReadDocumentRejectsInvalidIdentityAndAction(t *testing.T) {
	tests := []bson.D{
		{{Key: "guild", Value: bson.D{{Key: "bad", Value: true}}}, {Key: "ban_count", Value: "2"}, {Key: "move", Value: domain.WarningSettingsActionBan}},
		{{Key: "guild", Value: "guild-1"}, {Key: "ban_count", Value: "2"}, {Key: "move", Value: true}},
	}
	for index, input := range tests {
		payload, err := bson.Marshal(input)
		if err != nil {
			t.Fatalf("case %d marshal: %v", index, err)
		}
		var document documents.WarningSettingsReadDocument
		if err := bson.Unmarshal(payload, &document); err != nil {
			t.Fatalf("case %d unmarshal: %v", index, err)
		}
		if _, err := document.ToDomain(); !errors.Is(err, domain.ErrInvalidWarningSettings) {
			t.Fatalf("case %d error = %v", index, err)
		}
	}
}

func TestWarningSettingsReadDocumentPreservesJavaScriptNumberThresholds(t *testing.T) {
	tests := []struct {
		name    string
		value   any
		want    float64
		wantNaN bool
	}{
		{name: "zero", value: nil, want: 0},
		{name: "negative", value: "-2", want: -2},
		{name: "decimal", value: "2.5", want: 2.5},
		{name: "hex", value: "0x10", want: 16},
		{name: "malformed", value: "bad", wantNaN: true},
		{name: "boolean string cast", value: true, wantNaN: true},
		{name: "compound", value: bson.A{2}, wantNaN: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			payload, err := bson.Marshal(bson.D{{Key: "guild", Value: "guild-1"}, {Key: "ban_count", Value: test.value}, {Key: "move", Value: domain.WarningSettingsActionKick}})
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			var document documents.WarningSettingsReadDocument
			if err := bson.Unmarshal(payload, &document); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			settings, err := document.ToDomain()
			if err != nil {
				t.Fatalf("to domain: %v", err)
			}
			if test.wantNaN {
				if !math.IsNaN(settings.Threshold) {
					t.Fatalf("threshold = %v", settings.Threshold)
				}
			} else if settings.Threshold != test.want {
				t.Fatalf("threshold = %v, want %v", settings.Threshold, test.want)
			}
		})
	}
}
