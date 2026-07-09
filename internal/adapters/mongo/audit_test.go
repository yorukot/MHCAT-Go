package mongo

import (
	"bytes"
	"strings"
	"testing"
)

func TestAnalyzeAuditEmptyCatalogDeterministic(t *testing.T) {
	report := AnalyzeAudit("mhcat", nil, nil, AuditOptions{LargeDocumentBytes: 1024})
	var first bytes.Buffer
	var second bytes.Buffer
	if err := FormatAuditReport(&first, report, "json"); err != nil {
		t.Fatalf("format first: %v", err)
	}
	if err := FormatAuditReport(&second, report, "json"); err != nil {
		t.Fatalf("format second: %v", err)
	}
	if first.String() != second.String() {
		t.Fatalf("audit output is not deterministic:\n%s\n---\n%s", first.String(), second.String())
	}
}

func TestAnalyzeAuditUnknownCollectionDetected(t *testing.T) {
	report := AnalyzeAudit("mhcat", nil, []CollectionSnapshot{{Name: "unknown", DocumentCount: 1}}, AuditOptions{LargeDocumentBytes: 1024})
	if len(report.UnknownCollections) != 1 || report.UnknownCollections[0] != "unknown" {
		t.Fatalf("unknown collections = %#v", report.UnknownCollections)
	}
}

func TestAnalyzeAuditMissingCollectionDetected(t *testing.T) {
	catalog := []CollectionSpec{{Name: "coin"}}
	report := AnalyzeAudit("mhcat", catalog, nil, AuditOptions{LargeDocumentBytes: 1024})
	if len(report.MissingCollections) != 1 || report.MissingCollections[0] != "coin" {
		t.Fatalf("missing collections = %#v", report.MissingCollections)
	}
}

func TestAnalyzeAuditMixedFieldTypesDetected(t *testing.T) {
	catalog := []CollectionSpec{{Name: "coin"}}
	snapshots := []CollectionSnapshot{{
		Name: "coin",
		Samples: []SampleDocument{
			{Fields: map[string]string{"coin": "int64"}},
			{Fields: map[string]string{"coin": "string"}},
		},
	}}
	report := AnalyzeAudit("mhcat", catalog, snapshots, AuditOptions{LargeDocumentBytes: 1024})
	if got := report.Collections[0].FieldTypes["coin"]; len(got) != 2 {
		t.Fatalf("field types = %#v", got)
	}
	if len(report.Warnings) != 1 {
		t.Fatalf("warnings = %#v", report.Warnings)
	}
}

func TestAnalyzeAuditLargeDocumentIssueDetected(t *testing.T) {
	catalog := []CollectionSpec{{Name: "coin"}}
	snapshots := []CollectionSnapshot{{
		Name: "coin",
		Samples: []SampleDocument{
			{Ref: "doc-1", SizeBytes: 2048, Fields: map[string]string{"guild": "string"}},
		},
	}}
	report := AnalyzeAudit("mhcat", catalog, snapshots, AuditOptions{LargeDocumentBytes: 1024})
	if len(report.Collections[0].LargeDocuments) != 1 {
		t.Fatalf("large docs = %#v", report.Collections[0].LargeDocuments)
	}
}

func TestAnalyzeAuditMissingRequiredFieldIssueDetected(t *testing.T) {
	catalog := []CollectionSpec{{
		Name:           "coin",
		RequiredFields: []FieldSpec{{Name: "guild", Required: true}},
	}}
	snapshots := []CollectionSnapshot{{
		Name: "coin",
		Samples: []SampleDocument{
			{Fields: map[string]string{"member": "string"}},
		},
	}}
	report := AnalyzeAudit("mhcat", catalog, snapshots, AuditOptions{LargeDocumentBytes: 1024})
	issues := report.Collections[0].MissingRequiredFields
	if len(issues) != 1 || issues[0].Field != "guild" || issues[0].Count != 1 {
		t.Fatalf("missing field issues = %#v", issues)
	}
}

func TestAuditReportDoesNotIncludeRawDocumentValuesByDefault(t *testing.T) {
	catalog := []CollectionSpec{{Name: "coin"}}
	secretValue := "raw-user-value-should-not-appear"
	report := AnalyzeAudit("mhcat", catalog, []CollectionSnapshot{{
		Name: "coin",
		Samples: []SampleDocument{
			{Fields: map[string]string{"note": "string"}, Ref: "redacted-ref", SizeBytes: len(secretValue)},
		},
	}}, AuditOptions{LargeDocumentBytes: 1024})
	var output bytes.Buffer
	if err := FormatAuditReport(&output, report, "json"); err != nil {
		t.Fatalf("format audit: %v", err)
	}
	if strings.Contains(output.String(), secretValue) {
		t.Fatalf("audit output included raw value: %s", output.String())
	}
}

func TestAnalyzeAuditDuplicateKeyRiskWarns(t *testing.T) {
	catalog := []CollectionSpec{{
		Name: "coins",
		LogicalKeys: []LogicalKeySpec{{
			Name:   "coins_guild_member",
			Fields: []string{"guild", "member"},
			Unique: true,
		}},
	}}
	report := AnalyzeAudit("mhcat", catalog, []CollectionSnapshot{{
		Name: "coins",
		DuplicateRisks: []DuplicateKeyRisk{{
			KeyName:         "coins_guild_member",
			Fields:          []string{"guild", "member"},
			DuplicateGroups: 2,
		}},
	}}, AuditOptions{LargeDocumentBytes: 1024})
	if len(report.Collections) != 1 || len(report.Collections[0].DuplicateKeyRisks) != 1 {
		t.Fatalf("duplicate risks = %#v", report.Collections)
	}
	if len(report.Warnings) != 1 || !strings.Contains(report.Warnings[0].Message, "coins_guild_member") {
		t.Fatalf("warnings = %#v", report.Warnings)
	}
}

func TestFormatAuditReportTextIncludesDuplicateKeyRisk(t *testing.T) {
	report := AuditReport{
		Database: "mhcat",
		Collections: []CollectionAudit{{
			Name:          "sign_lists",
			DocumentCount: 3,
			DuplicateKeyRisks: []DuplicateKeyRisk{{
				KeyName:         "sign_lists_guild_member",
				Fields:          []string{"guild", "member"},
				DuplicateGroups: 1,
			}},
		}},
	}
	var output bytes.Buffer
	if err := FormatAuditReport(&output, report, "text"); err != nil {
		t.Fatalf("format report: %v", err)
	}
	text := output.String()
	if !strings.Contains(text, "duplicate_key_risk collection=sign_lists key=sign_lists_guild_member fields=guild,member groups=1") {
		t.Fatalf("duplicate risk missing from text output: %s", text)
	}
}
