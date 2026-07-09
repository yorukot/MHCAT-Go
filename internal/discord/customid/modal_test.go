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
