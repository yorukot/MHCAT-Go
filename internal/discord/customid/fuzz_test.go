package customid_test

import (
	"errors"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/customid"
)

func FuzzParseCustomID(f *testing.F) {
	f.Add("")
	f.Add("tic")
	f.Add("mhcat:v1:ticket:close:")
	f.Add("mhcat:v1:poll:vote:opt_1")
	f.Add("announcement_yes")
	f.Fuzz(func(t *testing.T, raw string) {
		_, err := customid.ParseComponent(raw)
		if len(raw) > customid.MaxCustomIDLength && err != nil && !errors.Is(err, customid.ErrTooLong) {
			t.Fatalf("overlong input returned %v", err)
		}
	})
}

func FuzzParseVersionedID(f *testing.F) {
	f.Add("mhcat:v1:ticket:close:")
	f.Add("mhcat:v1:poll:vote:opt_1")
	f.Add("mhcat:v2:ticket:close:")
	f.Fuzz(func(t *testing.T, raw string) {
		parsed, err := customid.ParseVersioned(raw, customid.InteractionKindComponent)
		if err == nil {
			encoded, err := parsed.Encode()
			if err != nil {
				t.Fatalf("encode parsed id: %v", err)
			}
			again, err := customid.ParseVersioned(encoded, customid.InteractionKindComponent)
			if err != nil {
				t.Fatalf("parse encoded id: %v", err)
			}
			if parsed.RouteKey() != again.RouteKey() {
				t.Fatalf("route changed after round trip")
			}
		}
	})
}

func FuzzParseLegacyID(f *testing.F) {
	f.Add("tic")
	f.Add("poll_Yes")
	f.Add("[123456789012345678]{2}text_rank")
	f.Add("123456789012345678my-profile")
	f.Fuzz(func(t *testing.T, raw string) {
		_, _ = customid.ParseLegacyComponent(raw)
	})
}
