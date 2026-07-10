package ports_test

import (
	"errors"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
)

func TestUsageEventValidate(t *testing.T) {
	if err := (ports.UsageEvent{CommandName: " ping "}).Validate(); err != nil {
		t.Fatalf("valid event: %v", err)
	}
	if err := (ports.UsageEvent{}).Validate(); !errors.Is(err, ports.ErrInvalidUsageEvent) {
		t.Fatalf("expected ErrInvalidUsageEvent, got %v", err)
	}
}
