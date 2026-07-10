package customid_test

import (
	"errors"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/customid"
)

type legacyModalFixture struct {
	Name       string              `json:"name"`
	Kind       string              `json:"kind"`
	Raw        string              `json:"raw"`
	Fields     []fixtureModalField `json:"fields"`
	Want       routeWant           `json:"want"`
	Confidence string              `json:"confidence"`
	Notes      string              `json:"notes"`
}

type fixtureModalField struct {
	CustomID string `json:"custom_id"`
	Value    string `json:"value"`
}

func TestLegacyModalGolden(t *testing.T) {
	var fixtures []legacyModalFixture
	readFixture(t, "../../../testdata/customid/legacy_modals_golden.json", &fixtures)
	for _, fixture := range fixtures {
		t.Run(fixture.Name, func(t *testing.T) {
			fields := make([]customid.ModalField, 0, len(fixture.Fields))
			for _, field := range fixture.Fields {
				fields = append(fields, customid.ModalField{CustomID: field.CustomID, Value: field.Value})
			}
			parsed, err := customid.ParseModal(fixture.Raw, fields)
			if err != nil {
				t.Fatalf("parse %q: %v", fixture.Raw, err)
			}
			assertRoute(t, parsed, fixture.Want)
			if len(fields) > 0 && parsed.Fields[fields[0].CustomID] != fields[0].Value {
				t.Fatalf("modal fields not preserved: %#v", parsed.Fields)
			}
		})
	}
}

func TestLegacyModalRequiresFirstField(t *testing.T) {
	_, err := customid.ParseModal("nal", nil)
	if !errors.Is(err, customid.ErrUnknownLegacyID) {
		t.Fatalf("expected ErrUnknownLegacyID, got %v", err)
	}
}

func TestLegacyModalUnknownFirstField(t *testing.T) {
	_, err := customid.ParseModal("nal", []customid.ModalField{{CustomID: "unknown"}})
	if !errors.Is(err, customid.ErrUnknownLegacyID) {
		t.Fatalf("expected ErrUnknownLegacyID, got %v", err)
	}
}

func TestLegacyVerificationModalContract(t *testing.T) {
	raw := "ABCDEFver"
	parsed, err := customid.ParseModal(raw, []customid.ModalField{{CustomID: raw, Value: "ABCDEF"}})
	if err != nil {
		t.Fatalf("parse legacy verification modal: %v", err)
	}
	if parsed.Feature != "verification" || parsed.Action != "answer" || !parsed.Legacy {
		t.Fatalf("parsed = %#v", parsed)
	}

	tests := []struct {
		name  string
		raw   string
		field string
	}{
		{name: "missing answer", raw: "ver", field: "ver"},
		{name: "wrong suffix case", raw: "ABCDEFVer", field: "ABCDEFVer"},
		{name: "punctuation", raw: "ABC-DEFver", field: "ABC-DEFver"},
		{name: "answer too long", raw: "ABCDEFGHIJKLMNOPQver", field: "ABCDEFGHIJKLMNOPQver"},
		{name: "field mismatch", raw: "ABCDEFver", field: "ABCDEGver"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := customid.ParseModal(tc.raw, []customid.ModalField{{CustomID: tc.field}})
			if !errors.Is(err, customid.ErrUnknownLegacyID) {
				t.Fatalf("expected ErrUnknownLegacyID, got %v", err)
			}
		})
	}
}

func TestLegacyRoleButtonModalAcceptsJavaScriptRandomNumberShape(t *testing.T) {
	for _, field := range []string{
		"roleaddcontent2026071101341234567890.1234567",
		"roleaddcontent2026071101341e-8",
	} {
		parsed, err := customid.ParseModal("nal", []customid.ModalField{{CustomID: field, Value: "panel"}})
		if err != nil {
			t.Fatalf("parse %q: %v", field, err)
		}
		if parsed.Feature != "role_button" || parsed.Action != "modal_submit" || !parsed.Legacy {
			t.Fatalf("parsed %q = %#v", field, parsed)
		}
	}
}
