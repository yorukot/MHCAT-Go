package config

import (
	"regexp"
	"strings"
	"unicode"
)

var (
	discordWebhookPattern = regexp.MustCompile(`(?i)^https://(?:canary\.|ptb\.)?discord(?:app)?\.com/api/webhooks/[0-9]+/([A-Za-z0-9_-]+)`)
	mongoURIPattern       = regexp.MustCompile(`^(mongodb(?:\+srv)?://[^:/@]+:)([^@]+)(@.+)$`)
)

func RedactValue(name, value string) string {
	if value == "" || strings.Contains(value, "<redacted") {
		return value
	}

	lowerName := strings.ToLower(name)
	if redacted, ok := redactMongoURI(value); ok {
		return redacted
	}
	if discordWebhookPattern.MatchString(value) {
		return "<redacted:discord-webhook:last4=" + last4(value) + ">"
	}
	if strings.Contains(lowerName, "discord") && strings.Contains(lowerName, "token") {
		return "<redacted:discord-token:last4=" + last4(value) + ">"
	}
	if looksLikeDiscordToken(value) {
		return "<redacted:discord-token:last4=" + last4(value) + ">"
	}
	if isSecretName(lowerName) || looksLikeLongSecret(value) {
		return "<redacted:secret:last4=" + last4(value) + ">"
	}
	return value
}

func Redact(value string) string {
	return RedactValue("", value)
}

func redactMongoURI(value string) (string, bool) {
	matches := mongoURIPattern.FindStringSubmatch(value)
	if len(matches) != 4 {
		return "", false
	}
	return matches[1] + "<redacted>" + matches[3], true
}

func isSecretName(name string) bool {
	for _, marker := range []string{"token", "secret", "password", "api_key", "apikey", "webhook", "connection_string"} {
		if strings.Contains(name, marker) {
			return true
		}
	}
	return false
}

func looksLikeDiscordToken(value string) bool {
	if len(value) < 50 || strings.Count(value, ".") < 2 || strings.ContainsAny(value, " \t\n\r") {
		return false
	}
	parts := strings.Split(value, ".")
	return len(parts) >= 3 && len(parts[0]) >= 10 && len(parts[1]) >= 5 && len(parts[2]) >= 20
}

func looksLikeLongSecret(value string) bool {
	if len(value) < 32 || strings.ContainsAny(value, " \t\n\r") {
		return false
	}
	hasAlpha := false
	hasDigit := false
	for _, r := range value {
		if unicode.IsLetter(r) {
			hasAlpha = true
		}
		if unicode.IsDigit(r) {
			hasDigit = true
		}
		if !(unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' || r == '-' || r == '.' || r == ':') {
			return false
		}
	}
	return hasAlpha && hasDigit
}

func last4(value string) string {
	if len(value) <= 4 {
		return value
	}
	return value[len(value)-4:]
}
