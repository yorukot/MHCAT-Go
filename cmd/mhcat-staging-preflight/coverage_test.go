package main

import (
	"errors"
	"testing"
)

func TestPositiveDurationMessage(t *testing.T) {
	if got := positiveDurationMessage("TIMEOUT", nil); got != "TIMEOUT must be positive" {
		t.Fatalf("positive duration message = %q", got)
	}
	if got := positiveDurationMessage("TIMEOUT", errors.New("invalid duration")); got != "invalid duration" {
		t.Fatalf("duration parse message = %q", got)
	}
}
