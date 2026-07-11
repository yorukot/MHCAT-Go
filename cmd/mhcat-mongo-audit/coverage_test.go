package main

import "testing"

func TestAliasAttrsFlattensFields(t *testing.T) {
	attrs := aliasAttrs(map[string]string{"source": "legacy"})
	if len(attrs) != 2 || attrs[0] != "source" || attrs[1] != "legacy" {
		t.Fatalf("alias attrs = %#v", attrs)
	}
}
