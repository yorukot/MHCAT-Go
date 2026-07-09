package customid_test

import (
	"errors"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/customid"
)

func TestEncodeEmptyPayload(t *testing.T) {
	encoded, err := customid.Encode(customid.InteractionKindComponent, "ticket", "close", customid.EmptyPayload())
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	if encoded != "mhcat:v1:ticket:close:" {
		t.Fatalf("encoded = %q", encoded)
	}
}

func TestEncodeOutputAtMost100Characters(t *testing.T) {
	payload, err := customid.TokenPayload("abc123")
	if err != nil {
		t.Fatalf("payload: %v", err)
	}
	encoded, err := customid.Encode(customid.InteractionKindComponent, "ticket", "close", payload)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	if len(encoded) > customid.MaxCustomIDLength {
		t.Fatalf("encoded id length = %d", len(encoded))
	}
}

func TestEncodeTooLongRecommendsStateID(t *testing.T) {
	payload, err := customid.TokenPayload("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	if err != nil {
		t.Fatalf("payload: %v", err)
	}
	_, err = customid.Encode(customid.InteractionKindComponent, "very_long_feature_name_for_tests", "very_long_action_name_for_tests", payload)
	if !errors.Is(err, customid.ErrTooLong) {
		t.Fatalf("expected ErrTooLong, got %v", err)
	}
}

func TestEncodeParseRoundTrip(t *testing.T) {
	payload, err := customid.KeyValuePayload(map[string]string{"page": "2", "user": "123"})
	if err != nil {
		t.Fatalf("payload: %v", err)
	}
	encoded, err := customid.Encode(customid.InteractionKindComponent, "rank", "page", payload)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	parsed, err := customid.ParseComponent(encoded)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if parsed.Feature != "rank" || parsed.Action != "page" || parsed.Payload.Kind != customid.PayloadKV {
		t.Fatalf("parsed = %#v", parsed)
	}
}
