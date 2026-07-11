package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestWriteReportFormatsTextAndJSON(t *testing.T) {
	result := report{Environment: "production", MongoDatabase: "mhcat"}
	var text bytes.Buffer
	writeReport(&text, result, "text")
	if !strings.Contains(text.String(), "environment=production") {
		t.Fatalf("text report = %q", text.String())
	}
	var json bytes.Buffer
	writeReport(&json, result, "json")
	if !strings.Contains(json.String(), `"environment": "production"`) {
		t.Fatalf("json report = %q", json.String())
	}
}
