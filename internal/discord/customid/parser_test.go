package customid_test

import (
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/customid"
)

type routeWant struct {
	Feature string `json:"feature"`
	Action  string `json:"action"`
	Version string `json:"version"`
	Legacy  bool   `json:"legacy"`
}

type versionedFixture struct {
	Name string    `json:"name"`
	Kind string    `json:"kind"`
	Raw  string    `json:"raw"`
	Want routeWant `json:"want"`
}

type invalidFixture struct {
	Name  string `json:"name"`
	Raw   string `json:"raw"`
	Error string `json:"error"`
}

func TestParseVersionedGolden(t *testing.T) {
	var fixtures []versionedFixture
	readFixture(t, "../../../testdata/customid/versioned_valid.json", &fixtures)
	for _, fixture := range fixtures {
		t.Run(fixture.Name, func(t *testing.T) {
			parsed, err := parseByKind(fixture.Raw, fixture.Kind)
			if err != nil {
				t.Fatalf("parse: %v", err)
			}
			assertRoute(t, parsed, fixture.Want)
			if len(parsed.Raw) > customid.MaxCustomIDLength {
				t.Fatalf("raw length = %d", len(parsed.Raw))
			}
		})
	}
}

func TestParseVersionedInvalidGolden(t *testing.T) {
	var fixtures []invalidFixture
	readFixture(t, "../../../testdata/customid/versioned_invalid.json", &fixtures)
	for _, fixture := range fixtures {
		t.Run(fixture.Name, func(t *testing.T) {
			_, err := customid.ParseVersioned(fixture.Raw, customid.InteractionKindComponent)
			if !matchesNamedError(err, fixture.Error) {
				t.Fatalf("expected %s, got %v", fixture.Error, err)
			}
		})
	}
}

func TestParseVersionedTooLong(t *testing.T) {
	raw := "mhcat:v1:ticket:close:" + strings.Repeat("a", customid.MaxCustomIDLength)
	_, err := customid.ParseVersioned(raw, customid.InteractionKindComponent)
	if !errors.Is(err, customid.ErrTooLong) {
		t.Fatalf("expected ErrTooLong, got %v", err)
	}
}

func TestParseVersionedDeterministic(t *testing.T) {
	raw := "mhcat:v1:poll:vote:option_1"
	first, err := customid.ParseComponent(raw)
	if err != nil {
		t.Fatalf("first parse: %v", err)
	}
	second, err := customid.ParseComponent(raw)
	if err != nil {
		t.Fatalf("second parse: %v", err)
	}
	if first.RouteKey() != second.RouteKey() || first.Payload.Raw != second.Payload.Raw {
		t.Fatalf("parser is not deterministic: %#v != %#v", first, second)
	}
}

func TestParseVersionedErrorDoesNotExposeRawSecret(t *testing.T) {
	secret := "abc.def.ghi1234567890123456789012345678901234567890"
	_, err := customid.ParseComponent("mhcat:v1:ticket:close:" + secret)
	if !errors.Is(err, customid.ErrUnsafePayload) {
		t.Fatalf("expected ErrUnsafePayload, got %v", err)
	}
	if strings.Contains(err.Error(), secret) {
		t.Fatalf("error leaked raw payload: %v", err)
	}
}

func parseByKind(raw string, kind string) (customid.ID, error) {
	if kind == "modal" {
		return customid.ParseModal(raw, nil)
	}
	return customid.ParseComponent(raw)
}

func readFixture(t *testing.T, path string, out any) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture %s: %v", path, err)
	}
	if err := json.Unmarshal(data, out); err != nil {
		t.Fatalf("decode fixture %s: %v", path, err)
	}
}

func assertRoute(t *testing.T, parsed customid.ID, want routeWant) {
	t.Helper()
	if parsed.Feature != want.Feature || parsed.Action != want.Action || parsed.Version != want.Version || parsed.Legacy != want.Legacy {
		t.Fatalf("route = feature=%q action=%q version=%q legacy=%t, want %#v", parsed.Feature, parsed.Action, parsed.Version, parsed.Legacy, want)
	}
}

func matchesNamedError(err error, name string) bool {
	switch name {
	case "ErrEmptyID":
		return errors.Is(err, customid.ErrEmptyID)
	case "ErrTooLong":
		return errors.Is(err, customid.ErrTooLong)
	case "ErrInvalidNamespace":
		return errors.Is(err, customid.ErrInvalidNamespace)
	case "ErrUnsupportedVersion":
		return errors.Is(err, customid.ErrUnsupportedVersion)
	case "ErrInvalidFeature":
		return errors.Is(err, customid.ErrInvalidFeature)
	case "ErrInvalidAction":
		return errors.Is(err, customid.ErrInvalidAction)
	case "ErrInvalidPayload":
		return errors.Is(err, customid.ErrInvalidPayload)
	case "ErrUnknownLegacyID":
		return errors.Is(err, customid.ErrUnknownLegacyID)
	case "ErrAmbiguousID":
		return errors.Is(err, customid.ErrAmbiguousID)
	case "ErrUnsafePayload":
		return errors.Is(err, customid.ErrUnsafePayload)
	default:
		return false
	}
}
