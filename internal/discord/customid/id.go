package customid

import (
	"context"
	"unicode/utf8"
)

const (
	Namespace         = "mhcat"
	VersionV1         = "v1"
	LegacyVersion     = "legacy"
	MaxCustomIDLength = 100
)

type InteractionKind string

const (
	InteractionKindComponent InteractionKind = "component"
	InteractionKindModal     InteractionKind = "modal"
)

type RouteKey struct {
	Kind    InteractionKind `json:"kind"`
	Feature string          `json:"feature"`
	Action  string          `json:"action"`
	Version string          `json:"version"`
	Legacy  bool            `json:"legacy"`
}

func (k RouteKey) IsZero() bool {
	return k.Kind == "" && k.Feature == "" && k.Action == "" && k.Version == ""
}

type ID struct {
	Namespace string
	Version   string
	Feature   string
	Action    string
	Payload   Payload
	Legacy    bool
	Raw       string
	Kind      InteractionKind
	Fields    map[string]string
}

func (id ID) RouteKey() RouteKey {
	return RouteKey{
		Kind:    id.Kind,
		Feature: id.Feature,
		Action:  id.Action,
		Version: id.Version,
		Legacy:  id.Legacy,
	}
}

type StateReference struct {
	ID string
}

type StateStore interface {
	CreateState(ctx context.Context, data any) (StateReference, error)
	LoadState(ctx context.Context, ref StateReference) (any, error)
}

func customIDLength(value string) int {
	return utf8.RuneCountInString(value)
}
