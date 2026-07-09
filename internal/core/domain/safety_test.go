package domain

import (
	"errors"
	"testing"
)

func TestAntiScamConfigValidateRequiresGuildID(t *testing.T) {
	config := AntiScamConfig{GuildID: " "}
	if err := config.Validate(); !errors.Is(err, ErrInvalidAntiScamConfig) {
		t.Fatalf("expected invalid anti-scam config, got %v", err)
	}
}

func TestAntiScamConfigValidateAllowsOpenFalse(t *testing.T) {
	config := AntiScamConfig{GuildID: "guild-1", Open: false}
	if err := config.Validate(); err != nil {
		t.Fatalf("validate anti-scam config: %v", err)
	}
}

func TestScamURLReportValidateRequiresHTTPURLAndReporter(t *testing.T) {
	for _, report := range []ScamURLReport{
		{URL: "not-a-url", ReporterUserID: "user-1"},
		{URL: "ftp://example.com", ReporterUserID: "user-1"},
		{URL: "https://example.com/path"},
	} {
		if err := report.Validate(); !errors.Is(err, ErrInvalidScamURLReport) {
			t.Fatalf("expected invalid report for %#v, got %v", report, err)
		}
	}
	valid := ScamURLReport{URL: "https://example.com/path", ReporterUserID: "user-1"}
	if err := valid.Validate(); err != nil {
		t.Fatalf("validate valid report: %v", err)
	}
}
