package main

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/parity"
)

func TestParityAuditRunPassesCurrentLegacyTree(t *testing.T) {
	legacyRoot := filepath.Join("..", "..", "..", "MHCAT")
	var stdout, stderr bytes.Buffer
	code := run([]string{"--legacy-root", legacyRoot, "--format", "json"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("code=%d stderr=%q", code, stderr.String())
	}
	var audit parity.CommandAudit
	if err := json.Unmarshal(stdout.Bytes(), &audit); err != nil {
		t.Fatalf("decode audit: %v", err)
	}
	if audit.LegacyUniqueCommands != 74 || audit.GoDefinitionCount != 74 || len(audit.MatchingCommands) != 74 || auditFailed(audit) {
		t.Fatalf("audit=%#v", audit)
	}
}

func TestParityAuditRejectsPositionalArguments(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{"unexpected"}, &stdout, &stderr)
	if code == 0 || stdout.Len() != 0 || !strings.Contains(stderr.String(), "unexpected positional arguments") {
		t.Fatalf("code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
	}
}

func TestParityAuditFailedDetectsEveryFailureClass(t *testing.T) {
	tests := []parity.CommandAudit{
		{CommandsWithDrift: []parity.CommandComparison{{Name: "drift"}}},
		{MissingGoDefinitions: []parity.LegacyCommand{{Name: "missing"}}},
		{ExtraGoDefinitions: []parity.GoCommandSummary{{Name: "extra"}}},
		{DuplicateLegacyNames: []parity.DuplicateName{{Name: "duplicate"}}},
		{DuplicateGoNames: []parity.DuplicateName{{Name: "duplicate"}}},
		{LegacyParseErrorCount: 1},
	}
	for i, audit := range tests {
		if !auditFailed(audit) {
			t.Fatalf("failure case %d was accepted", i)
		}
	}
	if auditFailed(parity.CommandAudit{}) {
		t.Fatal("empty audit should pass")
	}
}
