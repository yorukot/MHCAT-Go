package interactions_test

import (
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/interactions"
)

func TestRouteKeyStringDeterministic(t *testing.T) {
	key := interactions.RouteKey{
		Kind:    interactions.TypeComponent,
		Version: "v1",
		Feature: "ticket",
		Action:  "close",
	}
	if key.String() != "component:v1:ticket:close:false" {
		t.Fatalf("key string = %q", key.String())
	}
}

func TestRouteKeyZero(t *testing.T) {
	var key interactions.RouteKey
	if !key.IsZero() {
		t.Fatal("zero route key was not reported as zero")
	}
}
