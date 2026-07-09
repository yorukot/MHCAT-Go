package customid

import (
	"regexp"
	"sort"
	"strings"
)

const MaxPayloadLength = 64

type PayloadKind string

const (
	PayloadEmpty PayloadKind = "empty"
	PayloadToken PayloadKind = "token"
	PayloadKV    PayloadKind = "kv"
	PayloadState PayloadKind = "state"
)

type Payload struct {
	Kind    PayloadKind
	Raw     string
	Values  map[string]string
	StateID string
}

var (
	payloadTokenRe = regexp.MustCompile(`^[A-Za-z0-9._~-]{1,64}$`)
	payloadKeyRe   = regexp.MustCompile(`^[a-z][a-z0-9_]{0,15}$`)
	payloadValueRe = regexp.MustCompile(`^[A-Za-z0-9._~-]{1,48}$`)
)

func EmptyPayload() Payload {
	return Payload{Kind: PayloadEmpty}
}

func TokenPayload(token string) (Payload, error) {
	if !payloadTokenRe.MatchString(token) {
		return Payload{}, safeError(ErrInvalidPayload, "token")
	}
	if looksSecretLike(token) {
		return Payload{}, safeError(ErrUnsafePayload, "token")
	}
	return Payload{Kind: PayloadToken, Raw: token}, nil
}

func KeyValuePayload(values map[string]string) (Payload, error) {
	if len(values) == 0 {
		return EmptyPayload(), nil
	}
	copied := make(map[string]string, len(values))
	for key, value := range values {
		if !payloadKeyRe.MatchString(key) || !payloadValueRe.MatchString(value) {
			return Payload{}, safeError(ErrInvalidPayload, "key-value")
		}
		if looksSecretLike(value) {
			return Payload{}, safeError(ErrUnsafePayload, "key-value")
		}
		copied[key] = value
	}
	payload := Payload{Kind: PayloadKV, Values: copied}
	encoded, err := payload.Encode()
	if err != nil {
		return Payload{}, err
	}
	payload.Raw = encoded
	return payload, nil
}

func StateIDPayload(id string) (Payload, error) {
	if !payloadTokenRe.MatchString(id) {
		return Payload{}, safeError(ErrInvalidPayload, "state")
	}
	if looksSecretLike(id) {
		return Payload{}, safeError(ErrUnsafePayload, "state")
	}
	return Payload{Kind: PayloadState, StateID: id, Raw: "state=" + id}, nil
}

func ParsePayload(raw string) (Payload, error) {
	if raw == "" {
		return EmptyPayload(), nil
	}
	if len(raw) > MaxPayloadLength {
		return Payload{}, safeError(ErrTooLong, "payload")
	}
	if strings.Contains(raw, ":") {
		return Payload{}, safeError(ErrInvalidPayload, "payload delimiter")
	}
	if looksSecretLike(raw) {
		return Payload{}, safeError(ErrUnsafePayload, "payload")
	}
	if !strings.Contains(raw, "=") && !strings.Contains(raw, ",") {
		return TokenPayload(raw)
	}
	values := map[string]string{}
	for _, part := range strings.Split(raw, ",") {
		key, value, ok := strings.Cut(part, "=")
		if !ok || !payloadKeyRe.MatchString(key) || !payloadValueRe.MatchString(value) {
			return Payload{}, safeError(ErrInvalidPayload, "key-value")
		}
		if _, exists := values[key]; exists {
			return Payload{}, safeError(ErrInvalidPayload, "duplicate key")
		}
		values[key] = value
	}
	if stateID, ok := values["state"]; ok && len(values) == 1 {
		return StateIDPayload(stateID)
	}
	payload, err := KeyValuePayload(values)
	if err != nil {
		return Payload{}, err
	}
	payload.Raw = raw
	return payload, nil
}

func (p Payload) Encode() (string, error) {
	switch p.Kind {
	case "", PayloadEmpty:
		return "", nil
	case PayloadToken:
		if !payloadTokenRe.MatchString(p.Raw) {
			return "", safeError(ErrInvalidPayload, "token")
		}
		if looksSecretLike(p.Raw) {
			return "", safeError(ErrUnsafePayload, "token")
		}
		return p.Raw, nil
	case PayloadState:
		if !payloadTokenRe.MatchString(p.StateID) {
			return "", safeError(ErrInvalidPayload, "state")
		}
		if looksSecretLike(p.StateID) {
			return "", safeError(ErrUnsafePayload, "state")
		}
		return "state=" + p.StateID, nil
	case PayloadKV:
		if len(p.Values) == 0 {
			return "", nil
		}
		keys := make([]string, 0, len(p.Values))
		for key := range p.Values {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		parts := make([]string, 0, len(keys))
		for _, key := range keys {
			value := p.Values[key]
			if !payloadKeyRe.MatchString(key) || !payloadValueRe.MatchString(value) {
				return "", safeError(ErrInvalidPayload, "key-value")
			}
			if looksSecretLike(value) {
				return "", safeError(ErrUnsafePayload, "key-value")
			}
			parts = append(parts, key+"="+value)
		}
		encoded := strings.Join(parts, ",")
		if len(encoded) > MaxPayloadLength {
			return "", safeError(ErrTooLong, "payload; use state id")
		}
		return encoded, nil
	default:
		return "", safeError(ErrInvalidPayload, "payload kind")
	}
}

func looksSecretLike(value string) bool {
	lower := strings.ToLower(value)
	if strings.Contains(lower, "token") || strings.Contains(lower, "webhook") || strings.Contains(lower, "password") {
		return true
	}
	if strings.HasPrefix(lower, "mongodb") || strings.HasPrefix(lower, "http") {
		return true
	}
	return len(value) >= 48 && strings.Count(value, ".") >= 2
}
