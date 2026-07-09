package domain

import (
	"errors"
	"testing"
)

func TestParseDeleteDataTargetPreservesLegacyLabels(t *testing.T) {
	target, ok := ParseDeleteDataTarget(" 驗證設置 ")
	if !ok || target != DeleteDataTargetVerification {
		t.Fatalf("target = %q ok=%v", target, ok)
	}
	if _, ok := ParseDeleteDataTarget("unknown"); ok {
		t.Fatal("unknown target should be rejected")
	}
}

func TestDeleteDataRequestValidate(t *testing.T) {
	request := DeleteDataRequest{GuildID: " guild-1 ", Target: DeleteDataTargetAutoChat}
	if err := request.Validate(); err != nil {
		t.Fatalf("validate: %v", err)
	}
	if normalized := request.Normalize(); normalized.GuildID != "guild-1" || normalized.Target != DeleteDataTargetAutoChat {
		t.Fatalf("normalized = %#v", normalized)
	}
	if err := (DeleteDataRequest{GuildID: "guild-1", Target: "bad"}).Validate(); !errors.Is(err, ErrInvalidDeleteDataRequest) {
		t.Fatalf("invalid target err = %v", err)
	}
	if err := (DeleteDataRequest{Target: DeleteDataTargetTicket}).Validate(); !errors.Is(err, ErrInvalidDeleteDataRequest) {
		t.Fatalf("missing guild err = %v", err)
	}
}
