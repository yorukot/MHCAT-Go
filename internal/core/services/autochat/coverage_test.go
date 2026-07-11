package autochat

import "testing"

func TestLegacyEditDistance(t *testing.T) {
	if got := legacyEditDistance([]uint16{'k', 'i', 't', 't', 'e', 'n'}, []uint16{'s', 'i', 't', 't', 'i', 'n', 'g'}); got != 3 {
		t.Fatalf("edit distance = %d", got)
	}
}
