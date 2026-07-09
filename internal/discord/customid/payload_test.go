package customid_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/customid"
)

func TestPayloadToken(t *testing.T) {
	payload, err := customid.TokenPayload("opt_1")
	if err != nil {
		t.Fatalf("token payload: %v", err)
	}
	encoded, err := payload.Encode()
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	if encoded != "opt_1" {
		t.Fatalf("encoded = %q", encoded)
	}
}

func TestPayloadKeyValueDeterministic(t *testing.T) {
	payload, err := customid.KeyValuePayload(map[string]string{"user": "123", "page": "2"})
	if err != nil {
		t.Fatalf("kv payload: %v", err)
	}
	encoded, err := payload.Encode()
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	if encoded != "page=2,user=123" {
		t.Fatalf("encoded = %q", encoded)
	}
}

func TestPayloadState(t *testing.T) {
	payload, err := customid.StateIDPayload("abc123")
	if err != nil {
		t.Fatalf("state payload: %v", err)
	}
	encoded, err := payload.Encode()
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	if encoded != "state=abc123" {
		t.Fatalf("encoded = %q", encoded)
	}
	parsed, err := customid.ParsePayload(encoded)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if parsed.Kind != customid.PayloadState || parsed.StateID != "abc123" {
		t.Fatalf("parsed = %#v", parsed)
	}
}

func TestPayloadRejectsInvalidCharacters(t *testing.T) {
	_, err := customid.ParsePayload("bad:value")
	if !errors.Is(err, customid.ErrInvalidPayload) {
		t.Fatalf("expected ErrInvalidPayload, got %v", err)
	}
}

func TestPayloadRejectsOverlong(t *testing.T) {
	_, err := customid.ParsePayload(strings.Repeat("a", customid.MaxPayloadLength+1))
	if !errors.Is(err, customid.ErrTooLong) {
		t.Fatalf("expected ErrTooLong, got %v", err)
	}
}

func TestPayloadRejectsSecretLikeValues(t *testing.T) {
	_, err := customid.TokenPayload("abc.def.ghi1234567890123456789012345678901234567890")
	if !errors.Is(err, customid.ErrUnsafePayload) {
		t.Fatalf("expected ErrUnsafePayload, got %v", err)
	}
	_, err = customid.StateIDPayload("abc.def.ghi1234567890123456789012345678901234567890")
	if !errors.Is(err, customid.ErrUnsafePayload) {
		t.Fatalf("expected ErrUnsafePayload for state payload, got %v", err)
	}
}
