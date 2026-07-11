package observability

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"
)

func TestLoggerRedactsStringAttributes(t *testing.T) {
	var out bytes.Buffer
	logger := NewLogger(LoggerOptions{
		Level:  "info",
		Format: "text",
		Writer: &out,
	})
	secret := strings.Repeat("a", 24) + "." + strings.Repeat("b", 6) + "." + strings.Repeat("c", 32)
	password := "pass" + "word" + "123"
	uri := "mongodb://user:" + password + "@localhost:27017/mhcat"

	logger.Info("test", "token", secret, "mongo_uri", uri, "safe", "hello")

	logged := out.String()
	if strings.Contains(logged, secret) || strings.Contains(logged, password) {
		t.Fatalf("logger output contains secret: %q", logged)
	}
	if !strings.Contains(logged, "hello") {
		t.Fatalf("logger output missing safe value: %q", logged)
	}
}

func TestLoggerRedactsBoundAndGroupedAttributes(t *testing.T) {
	var out bytes.Buffer
	logger := NewLogger(LoggerOptions{Level: "debug", Format: "json", Writer: &out}).
		With("token", "secret-token").
		WithGroup("request")
	logger.Info("grouped", slog.Group("database", slog.String("mongo_uri", "mongodb://user:password@example/db")))
	logged := out.String()
	if strings.Contains(logged, "secret-token") || strings.Contains(logged, "password") {
		t.Fatalf("grouped logger leaked secret: %s", logged)
	}
	if !strings.Contains(logged, "request") || !strings.Contains(logged, "database") {
		t.Fatalf("grouped logger output = %s", logged)
	}
}
