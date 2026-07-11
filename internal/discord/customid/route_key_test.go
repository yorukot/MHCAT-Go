package customid

import "testing"

func TestRouteKeyIsZero(t *testing.T) {
	if !(RouteKey{}).IsZero() {
		t.Fatal("empty route key must be zero")
	}
	if !(RouteKey{Legacy: true}).IsZero() {
		t.Fatal("legacy marker alone is not a route identity")
	}
	if (RouteKey{Feature: "ticket"}).IsZero() {
		t.Fatal("feature route key must not be zero")
	}
}
