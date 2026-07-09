package customid_test

import (
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/customid"
)

func TestParseComponentInput(t *testing.T) {
	parsed, err := customid.ParseComponentInput(customid.ComponentInput{
		CustomID: "mhcat:v1:poll:vote:opt_1",
		Values:   []string{"ignored-by-parser"},
	})
	if err != nil {
		t.Fatalf("parse input: %v", err)
	}
	assertRoute(t, parsed, routeWant{Feature: "poll", Action: "vote", Version: "v1"})
}
