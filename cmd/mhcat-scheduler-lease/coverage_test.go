package main

import (
	"bytes"
	"context"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/config"
)

func TestSchedulerLeaseProductionEntryValidation(t *testing.T) {
	lookup := func(string) (string, bool) { return "", false }
	if code := run(context.Background(), nil, lookup, &bytes.Buffer{}, &bytes.Buffer{}); code != 1 {
		t.Fatalf("run exit code = %d", code)
	}
	if _, _, err := defaultLeaseStoreFactory(context.Background(), config.SchedulerLeaseConfig{}); err == nil {
		t.Fatal("default factory accepted empty configuration")
	}
}
