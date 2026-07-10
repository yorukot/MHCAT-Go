package roles

import (
	"testing"
	"time"
)

func TestLegacyRoleButtonIDPreservesJavaScriptRandomNumberShape(t *testing.T) {
	now := time.Date(2026, time.July, 10, 17, 34, 0, 0, time.UTC)
	if got := legacyRoleButtonIDAt(now, 0.12345678901234567); got != "2026071101341234567890.1234567" {
		t.Fatalf("role button id = %q", got)
	}
	if got := legacyRoleButtonIDAt(now, 1e-18); got != "2026071101341e-8" {
		t.Fatalf("small role button id = %q", got)
	}
}
