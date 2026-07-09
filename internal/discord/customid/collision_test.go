package customid_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/customid"
)

type collisionFixture struct {
	Name      string          `json:"name"`
	Rules     []customid.Rule `json:"rules"`
	WantIssue bool            `json:"want_issue"`
}

func TestCollisionGolden(t *testing.T) {
	var fixtures []collisionFixture
	readFixture(t, "../../../testdata/customid/collision_cases.json", &fixtures)
	for _, fixture := range fixtures {
		t.Run(fixture.Name, func(t *testing.T) {
			report := customid.AnalyzeRules(fixture.Rules)
			gotIssue := len(report.Issues) > 0
			if gotIssue != fixture.WantIssue {
				t.Fatalf("issue = %t, want %t: %#v", gotIssue, fixture.WantIssue, report.Issues)
			}
		})
	}
}

func TestCollisionDetectsDelimiterlessOverlap(t *testing.T) {
	report := customid.AnalyzeRules([]customid.Rule{
		{Name: "short", Pattern: "rank"},
		{Name: "long", Pattern: "text_rank"},
	})
	if len(report.Issues) == 0 || report.Issues[0].Reason != "delimiterless overlap" {
		t.Fatalf("expected delimiterless overlap, got %#v", report.Issues)
	}
}

func TestCollisionDetectsIncludesStyleAmbiguity(t *testing.T) {
	report := customid.AnalyzeRules([]customid.Rule{
		{Name: "role-add", Pattern: "*add", Broad: true},
		{Name: "role-delete", Pattern: "*delete", Broad: true},
	})
	if len(report.Issues) == 0 || report.Issues[0].Reason != "includes-style broad rule" {
		t.Fatalf("expected broad issue, got %#v", report.Issues)
	}
}

func TestCollisionReportDeterministic(t *testing.T) {
	report := customid.AnalyzeRules([]customid.Rule{
		{Name: "b", Pattern: "tic"},
		{Name: "a", Pattern: "tic"},
	})
	var first bytes.Buffer
	var second bytes.Buffer
	if err := customid.FormatCollisionReport(&first, report, "json"); err != nil {
		t.Fatalf("format first: %v", err)
	}
	if err := customid.FormatCollisionReport(&second, report, "json"); err != nil {
		t.Fatalf("format second: %v", err)
	}
	if first.String() != second.String() {
		t.Fatalf("non-deterministic output:\n%s\n%s", first.String(), second.String())
	}
	var decoded customid.CollisionReport
	if err := json.Unmarshal(first.Bytes(), &decoded); err != nil {
		t.Fatalf("json output invalid: %v", err)
	}
	if !strings.Contains(first.String(), "duplicate pattern") {
		t.Fatalf("expected reason in report: %s", first.String())
	}
}

func TestDocumentedLegacyRulesExposeKnownAmbiguity(t *testing.T) {
	report := customid.AnalyzeRules(customid.DocumentedLegacyRules())
	if len(report.Issues) == 0 {
		t.Fatal("expected documented legacy rules to include ambiguity notes")
	}
}
