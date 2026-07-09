package external

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
)

func TestDiscordWebhookReporterPostsLegacyReportContent(t *testing.T) {
	var payload map[string]string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s", r.Method)
		}
		if contentType := r.Header.Get("Content-Type"); contentType != "application/json" {
			t.Fatalf("content type = %q", contentType)
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode payload: %v", err)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	reporter := NewDiscordWebhookReporter(server.URL)
	if err := reporter.SendScamURLReport(context.Background(), domain.ScamURLReport{
		URL:            "https://bad.example",
		ReporterUserID: "user-1",
	}); err != nil {
		t.Fatalf("send report: %v", err)
	}
	if payload["content"] != "```https://bad.example```\nby:<@user-1>" {
		t.Fatalf("payload = %#v", payload)
	}
}

func TestDiscordWebhookReporterSanitizesCodeBlockBreakout(t *testing.T) {
	var payload map[string]string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode payload: %v", err)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	reporter := NewDiscordWebhookReporter(server.URL)
	if err := reporter.SendScamURLReport(context.Background(), domain.ScamURLReport{
		URL:            "https://bad.example/```",
		ReporterUserID: "user-1",
	}); err != nil {
		t.Fatalf("send report: %v", err)
	}
	if strings.Contains(payload["content"], "https://bad.example/``````") {
		t.Fatalf("payload did not sanitize code block breakout: %q", payload["content"])
	}
}

func TestDiscordWebhookReporterReturnsStatusError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer server.Close()

	reporter := NewDiscordWebhookReporter(server.URL)
	err := reporter.SendScamURLReport(context.Background(), domain.ScamURLReport{
		URL:            "https://bad.example",
		ReporterUserID: "user-1",
	})
	if err == nil || !strings.Contains(err.Error(), "status 502") {
		t.Fatalf("expected status error, got %v", err)
	}
}
