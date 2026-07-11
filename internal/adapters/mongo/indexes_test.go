package mongo

import (
	"testing"
	"time"
)

func TestIndexPlanValidateValid(t *testing.T) {
	plan := IndexPlan{Indexes: []IndexSpec{plainIndex("coin", "guild_member_idx")}}
	if err := plan.Validate(); err != nil {
		t.Fatalf("validate index plan: %v", err)
	}
}

func TestIndexPlanValidateUniqueRequiresDuplicateAudit(t *testing.T) {
	spec := plainIndex("coin", "guild_member_idx")
	spec.Unique = true
	plan := IndexPlan{Indexes: []IndexSpec{spec}}
	if err := plan.Validate(); err == nil {
		t.Fatal("expected validation error")
	}
}

func TestIndexPlanValidateTTLRequiresRetentionADR(t *testing.T) {
	ttlSeconds := int32(3600)
	spec := plainIndex("sessions", "expire_idx")
	spec.TTLSeconds = &ttlSeconds
	plan := IndexPlan{Indexes: []IndexSpec{spec}}
	if err := plan.Validate(); err == nil {
		t.Fatal("expected validation error")
	}
}

func TestIndexDurationHelpers(t *testing.T) {
	if value := ttl(60); value == nil || *value != 60 {
		t.Fatalf("ttl = %#v", value)
	}
	if value := secondsFromDuration(2500 * time.Millisecond); value == nil || *value != 2 {
		t.Fatalf("duration seconds = %#v", value)
	}
}
