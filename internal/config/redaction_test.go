package config

import (
	"strings"
	"testing"
)

func TestRedactDiscordToken(t *testing.T) {
	secret := strings.Repeat("a", 24) + "." + strings.Repeat("b", 6) + "." + strings.Repeat("c", 32)
	redacted := RedactValue("MHCAT_DISCORD_TOKEN", secret)
	if redacted == secret || strings.Contains(redacted, secret[:20]) {
		t.Fatal("discord token was not redacted")
	}
	if !strings.Contains(redacted, "last4=cccc") {
		t.Fatalf("expected last4 marker, got %q", redacted)
	}
}

func TestRedactWebhookURL(t *testing.T) {
	secretPart := strings.Repeat("w", 28) + "abcd"
	webhook := "https://discord.com/api/webhooks/" + strings.Repeat("1", 18) + "/" + secretPart
	redacted := RedactValue("REPORT_WEBHOOK_URL", webhook)
	if redacted == webhook || strings.Contains(redacted, secretPart) {
		t.Fatal("webhook was not redacted")
	}
	if !strings.Contains(redacted, "last4=abcd") {
		t.Fatalf("expected last4 marker, got %q", redacted)
	}
}

func TestRedactMongoSRVURI(t *testing.T) {
	password := "pass" + "word" + "123"
	uri := "mongodb+srv://user:" + password + "@cluster.example/mhcat"
	redacted := Redact(uri)
	if strings.Contains(redacted, password) {
		t.Fatalf("mongo password was not redacted: %q", redacted)
	}
	expected := "mongodb+srv://user:<redacted>@cluster.example/mhcat"
	if redacted != expected {
		t.Fatalf("unexpected redacted uri: %q", redacted)
	}
}

func TestRedactMongoURI(t *testing.T) {
	password := "pass" + "word" + "123"
	uri := "mongodb://user:" + password + "@localhost:27017/mhcat"
	redacted := Redact(uri)
	if strings.Contains(redacted, password) {
		t.Fatalf("mongo password was not redacted: %q", redacted)
	}
	expected := "mongodb://user:<redacted>@localhost:27017/mhcat"
	if redacted != expected {
		t.Fatalf("unexpected redacted uri: %q", redacted)
	}
}

func TestRedactEmptyShortAndAlreadyRedacted(t *testing.T) {
	if Redact("") != "" {
		t.Fatal("empty string should remain empty")
	}
	if Redact("hello") != "hello" {
		t.Fatal("short harmless string should remain unchanged")
	}
	already := "<redacted:secret:last4=abcd>"
	if Redact(already) != already {
		t.Fatal("already redacted value should remain unchanged")
	}
}

func TestRedactGenericLongSecret(t *testing.T) {
	secret := "sk_" + strings.Repeat("x1", 24)
	redacted := RedactValue("API_KEY", secret)
	if redacted == secret || strings.Contains(redacted, secret[:12]) {
		t.Fatal("generic API key was not redacted")
	}
}
