package mongo

import (
	"bytes"
	"testing"
)

func TestDiffIndexesMissingIndexPlannedCreate(t *testing.T) {
	plan := IndexPlan{Indexes: []IndexSpec{plainIndex("coin", "guild_member_idx")}}
	diff, err := DiffIndexes(plan, nil, IndexDiffOptions{})
	if err != nil {
		t.Fatalf("diff indexes: %v", err)
	}
	assertIndexOperation(t, diff, IndexOperationCreate, "coin", "guild_member_idx")
}

func TestDiffIndexesExistingUniqueIndexCoversSafeLookupFallback(t *testing.T) {
	spec := plainIndex("coins", "coins_guild_member_lookup")
	spec.Keys = []IndexKey{{Field: "guild", Order: 1}, {Field: "member", Order: 1}}
	plan, err := DiffIndexes(IndexPlan{Indexes: []IndexSpec{spec}}, map[string][]IndexInfo{
		"coins": {{
			Collection: "coins",
			Name:       "coins_guild_member",
			Keys:       append([]IndexKey(nil), spec.Keys...),
			Unique:     true,
		}},
	}, IndexDiffOptions{})
	if err != nil {
		t.Fatalf("diff indexes: %v", err)
	}
	for _, operation := range plan.Operations {
		if operation.IndexName == spec.Name && operation.Operation != IndexOperationExists {
			t.Fatalf("fallback operation = %#v", operation)
		}
	}
}

func TestDiffIndexesExistingIdenticalMarkedExists(t *testing.T) {
	spec := plainIndex("coin", "guild_member_idx")
	diff, err := DiffIndexes(IndexPlan{Indexes: []IndexSpec{spec}}, map[string][]IndexInfo{
		"coin": {{Collection: "coin", Name: "guild_member_idx", Keys: spec.Keys}},
	}, IndexDiffOptions{})
	if err != nil {
		t.Fatalf("diff indexes: %v", err)
	}
	assertIndexOperation(t, diff, IndexOperationExists, "coin", "guild_member_idx")
}

func TestDiffIndexesChangedMarkedDangerous(t *testing.T) {
	spec := plainIndex("coin", "guild_member_idx")
	diff, err := DiffIndexes(IndexPlan{Indexes: []IndexSpec{spec}}, map[string][]IndexInfo{
		"coin": {{Collection: "coin", Name: "guild_member_idx", Keys: []IndexKey{{Field: "guild", Order: -1}}}},
	}, IndexDiffOptions{})
	if err != nil {
		t.Fatalf("diff indexes: %v", err)
	}
	op := assertIndexOperation(t, diff, IndexOperationChanged, "coin", "guild_member_idx")
	if op.Risk != IndexRiskHigh {
		t.Fatalf("changed index risk = %s", op.Risk)
	}
}

func TestDiffIndexesUnknownRemoteSkipped(t *testing.T) {
	diff, err := DiffIndexes(IndexPlan{}, map[string][]IndexInfo{
		"coin": {{Collection: "coin", Name: "legacy_idx", Keys: []IndexKey{{Field: "guild", Order: 1}}}},
	}, IndexDiffOptions{})
	if err != nil {
		t.Fatalf("diff indexes: %v", err)
	}
	assertIndexOperation(t, diff, IndexOperationUnknownRemote, "coin", "legacy_idx")
}

func TestDiffIndexesUniqueRequiresDuplicateAudit(t *testing.T) {
	spec := uniqueIndex("coin", "guild_member_idx")
	diff, err := DiffIndexes(IndexPlan{Indexes: []IndexSpec{spec}}, nil, IndexDiffOptions{AllowUnique: true})
	if err != nil {
		t.Fatalf("diff indexes: %v", err)
	}
	op := assertIndexOperation(t, diff, IndexOperationDangerous, "coin", "guild_member_idx")
	if !op.RequiresDuplicateAudit {
		t.Fatalf("expected duplicate audit requirement: %#v", op)
	}
}

func TestDiffIndexesTTLRequiresAllowFlagAndADR(t *testing.T) {
	ttlSeconds := int32(3600)
	spec := plainIndex("sessions", "expire_idx")
	spec.TTLSeconds = &ttlSeconds
	spec.RequiresRetentionADR = true
	diff, err := DiffIndexes(IndexPlan{Indexes: []IndexSpec{spec}}, nil, IndexDiffOptions{})
	if err != nil {
		t.Fatalf("diff indexes: %v", err)
	}
	op := assertIndexOperation(t, diff, IndexOperationDangerous, "sessions", "expire_idx")
	if !op.RequiresRetentionADR {
		t.Fatalf("expected retention ADR requirement: %#v", op)
	}
}

func TestSafeIndexApplyOperationsDoesNotDropIndexes(t *testing.T) {
	diff := IndexDiffPlan{Operations: []IndexOperation{
		{Operation: IndexOperationCreate, Collection: "coin", IndexName: "create_idx", Risk: IndexRiskMedium},
		{Operation: IndexOperationUnknownRemote, Collection: "coin", IndexName: "legacy_idx", Risk: IndexRiskLow},
		{Operation: IndexOperationChanged, Collection: "coin", IndexName: "changed_idx", Risk: IndexRiskHigh},
	}}
	ops := SafeIndexApplyOperations(diff)
	if len(ops) != 1 || ops[0].IndexName != "create_idx" {
		t.Fatalf("safe apply operations = %#v", ops)
	}
}

func TestFormatIndexDiffPlanDeterministic(t *testing.T) {
	diff := IndexDiffPlan{Operations: []IndexOperation{
		{Operation: IndexOperationCreate, Collection: "z", IndexName: "z_idx", Risk: IndexRiskMedium},
		{Operation: IndexOperationExists, Collection: "a", IndexName: "a_idx", Risk: IndexRiskLow},
	}}
	var first bytes.Buffer
	var second bytes.Buffer
	if err := FormatIndexDiffPlan(&first, diff, "json"); err != nil {
		t.Fatalf("format first: %v", err)
	}
	if err := FormatIndexDiffPlan(&second, diff, "json"); err != nil {
		t.Fatalf("format second: %v", err)
	}
	if first.String() != second.String() {
		t.Fatalf("index diff output not deterministic:\n%s\n---\n%s", first.String(), second.String())
	}
	var text bytes.Buffer
	if err := FormatIndexDiffPlan(&text, diff, "text"); err != nil {
		t.Fatalf("format text: %v", err)
	}
	if text.String() == "" {
		t.Fatal("expected text output")
	}
}

func plainIndex(collection, name string) IndexSpec {
	return IndexSpec{
		Collection: collection,
		Name:       name,
		Keys:       []IndexKey{{Field: "guild", Order: 1}, {Field: "member", Order: 1}},
		Reason:     "test index",
	}
}

func uniqueIndex(collection, name string) IndexSpec {
	spec := plainIndex(collection, name)
	spec.Unique = true
	spec.RequiresDuplicateAudit = true
	return spec
}

func assertIndexOperation(t *testing.T, diff IndexDiffPlan, operation IndexOperationKind, collection, name string) IndexOperation {
	t.Helper()
	for _, op := range diff.Operations {
		if op.Operation == operation && op.Collection == collection && op.IndexName == name {
			return op
		}
	}
	t.Fatalf("operation %s for %s.%s not found in %#v", operation, collection, name, diff.Operations)
	return IndexOperation{}
}
