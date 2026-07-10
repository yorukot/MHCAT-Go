package customid_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/customid"
)

type legacyComponentFixture struct {
	Name       string    `json:"name"`
	Kind       string    `json:"kind"`
	Raw        string    `json:"raw"`
	Want       routeWant `json:"want"`
	Confidence string    `json:"confidence"`
	Notes      string    `json:"notes"`
}

func TestLegacyComponentGolden(t *testing.T) {
	var fixtures []legacyComponentFixture
	readFixture(t, "../../../testdata/customid/legacy_components_golden.json", &fixtures)
	for _, fixture := range fixtures {
		t.Run(fixture.Name, func(t *testing.T) {
			parsed, err := customid.ParseComponent(fixture.Raw)
			if err != nil {
				t.Fatalf("parse %q: %v", fixture.Raw, err)
			}
			assertRoute(t, parsed, fixture.Want)
			if !parsed.Legacy {
				t.Fatalf("expected legacy route")
			}
		})
	}
}

func TestLegacyAmbiguousGolden(t *testing.T) {
	var fixtures []invalidFixture
	readFixture(t, "../../../testdata/customid/ambiguous_legacy.json", &fixtures)
	for _, fixture := range fixtures {
		t.Run(fixture.Name, func(t *testing.T) {
			_, err := customid.ParseComponent(fixture.Raw)
			if !matchesNamedError(err, fixture.Error) {
				t.Fatalf("expected %s, got %v", fixture.Error, err)
			}
		})
	}
}

func TestLegacyUnknownComponent(t *testing.T) {
	_, err := customid.ParseComponent("unknown:id")
	if !errors.Is(err, customid.ErrUnknownLegacyID) {
		t.Fatalf("expected ErrUnknownLegacyID, got %v", err)
	}
}

func TestLegacyPositionalFieldCountValidated(t *testing.T) {
	_, err := customid.ParseComponent("[123456789012345678]{}text_rank")
	if !errors.Is(err, customid.ErrUnknownLegacyID) {
		t.Fatalf("expected unknown legacy id for malformed positional fields, got %v", err)
	}
}

func TestLegacySnowflakeShapeValidated(t *testing.T) {
	_, err := customid.ParseComponent("[123]{2}text_rank")
	if !errors.Is(err, customid.ErrUnknownLegacyID) {
		t.Fatalf("expected ErrUnknownLegacyID for invalid snowflake shape, got %v", err)
	}
	if !customid.IsSnowflake("123456789012345678") {
		t.Fatal("expected public test snowflake shape to be accepted")
	}
}

func TestLegacyParserPriorityDeterministic(t *testing.T) {
	first, err := customid.ParseComponent("poll_Yes")
	if err != nil {
		t.Fatalf("first parse: %v", err)
	}
	second, err := customid.ParseComponent("poll_Yes")
	if err != nil {
		t.Fatalf("second parse: %v", err)
	}
	if first.RouteKey() != second.RouteKey() {
		t.Fatalf("route changed between parses: %#v %#v", first.RouteKey(), second.RouteKey())
	}
}

func TestLegacyPollChoiceAllowsLegacyEightyCharacters(t *testing.T) {
	raw := "poll_" + strings.Repeat("選", 80)
	parsed, err := customid.ParseComponent(raw)
	if err != nil {
		t.Fatalf("parse 80-character poll choice: %v", err)
	}
	if parsed.Feature != "poll" || parsed.Action != "vote" || !parsed.Legacy {
		t.Fatalf("parsed = %#v", parsed)
	}
}

func TestLegacyKnowledgeAnswerRoutesKnownUnicodeAndASCIIAnswers(t *testing.T) {
	for _, raw := range []string{"亞洲", "RAV4"} {
		parsed, err := customid.ParseComponent(raw)
		if err != nil {
			t.Fatalf("parse %q: %v", raw, err)
		}
		if parsed.Feature != "game" || parsed.Action != "knowledge_answer" || !parsed.Legacy {
			t.Fatalf("parsed %q = %#v", raw, parsed)
		}
	}
}

func TestLegacyVerificationComponentContract(t *testing.T) {
	parsed, err := customid.ParseComponent("ABCDEFverification")
	if err != nil {
		t.Fatalf("parse legacy verification component: %v", err)
	}
	if parsed.Feature != "verification" || parsed.Action != "prompt" || !parsed.Legacy {
		t.Fatalf("parsed = %#v", parsed)
	}

	for _, raw := range []string{
		"verification",
		"ABCDEFVerification",
		"ABC-DEFverification",
		strings.Repeat("A", 17) + "verification",
		"ABCDEFverificationtail",
	} {
		t.Run(raw, func(t *testing.T) {
			if parsed, err := customid.ParseComponent(raw); err == nil {
				t.Fatalf("malformed verification id routed as %#v", parsed)
			}
		})
	}
}

func TestLegacyRoleButtonAcceptsJavaScriptRandomNumberShape(t *testing.T) {
	for _, tc := range []struct {
		raw    string
		action string
	}{
		{raw: "2026071101341234567890.1234567add", action: "add"},
		{raw: "2026071101341e-8delete", action: "remove"},
	} {
		parsed, err := customid.ParseComponent(tc.raw)
		if err != nil {
			t.Fatalf("parse %q: %v", tc.raw, err)
		}
		if parsed.Feature != "role_button" || parsed.Action != tc.action || !parsed.Legacy {
			t.Fatalf("parsed %q = %#v", tc.raw, parsed)
		}
	}
}
