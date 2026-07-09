package customid

import (
	"strings"
)

func ParseVersioned(raw string, kind InteractionKind) (ID, error) {
	if raw == "" {
		return ID{}, ErrEmptyID
	}
	if customIDLength(raw) > MaxCustomIDLength {
		return ID{}, safeError(ErrTooLong, "custom id")
	}
	parts := strings.Split(raw, ":")
	if len(parts) != 5 {
		if len(parts) > 5 && len(parts) >= 2 && parts[0] == Namespace && parts[1] == VersionV1 {
			return ID{}, safeError(ErrInvalidPayload, "payload delimiter")
		}
		return ID{}, safeError(ErrInvalidNamespace, "format")
	}
	if parts[0] != Namespace {
		return ID{}, safeError(ErrInvalidNamespace, "namespace")
	}
	if parts[1] != VersionV1 {
		return ID{}, safeError(ErrUnsupportedVersion, "version")
	}
	if !tokenRe.MatchString(parts[2]) {
		return ID{}, safeError(ErrInvalidFeature, "feature")
	}
	if !tokenRe.MatchString(parts[3]) {
		return ID{}, safeError(ErrInvalidAction, "action")
	}
	payload, err := ParsePayload(parts[4])
	if err != nil {
		return ID{}, err
	}
	return ID{
		Namespace: Namespace,
		Version:   VersionV1,
		Feature:   parts[2],
		Action:    parts[3],
		Payload:   payload,
		Raw:       raw,
		Kind:      kind,
	}, nil
}

func ParseComponent(raw string) (ID, error) {
	if strings.HasPrefix(raw, Namespace+":") {
		return ParseVersioned(raw, InteractionKindComponent)
	}
	return ParseLegacyComponent(raw)
}

func ParseModal(raw string, fields []ModalField) (ID, error) {
	if strings.HasPrefix(raw, Namespace+":") {
		return ParseVersioned(raw, InteractionKindModal)
	}
	return ParseLegacyModal(raw, fields)
}
