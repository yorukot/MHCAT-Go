package customid

import (
	"regexp"
	"strings"
)

var tokenRe = regexp.MustCompile(`^[a-z][a-z0-9_-]{0,31}$`)

func Encode(kind InteractionKind, feature string, action string, payload Payload) (string, error) {
	id := ID{
		Namespace: Namespace,
		Version:   VersionV1,
		Feature:   feature,
		Action:    action,
		Payload:   payload,
		Kind:      kind,
	}
	return id.Encode()
}

func (id ID) Encode() (string, error) {
	if id.Namespace == "" {
		id.Namespace = Namespace
	}
	if id.Version == "" {
		id.Version = VersionV1
	}
	if id.Namespace != Namespace {
		return "", safeError(ErrInvalidNamespace, "encode")
	}
	if id.Version != VersionV1 {
		return "", safeError(ErrUnsupportedVersion, "encode")
	}
	if !tokenRe.MatchString(id.Feature) {
		return "", safeError(ErrInvalidFeature, "encode")
	}
	if !tokenRe.MatchString(id.Action) {
		return "", safeError(ErrInvalidAction, "encode")
	}
	payload, err := id.Payload.Encode()
	if err != nil {
		return "", err
	}
	encoded := strings.Join([]string{id.Namespace, id.Version, id.Feature, id.Action, payload}, ":")
	if len(encoded) > MaxCustomIDLength {
		return "", safeError(ErrTooLong, "custom id; use state id")
	}
	return encoded, nil
}
