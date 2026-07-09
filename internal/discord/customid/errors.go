package customid

import (
	"errors"
	"fmt"
)

var (
	ErrEmptyID              = errors.New("custom id is empty")
	ErrTooLong              = errors.New("custom id is too long")
	ErrInvalidNamespace     = errors.New("custom id namespace is invalid")
	ErrUnsupportedVersion   = errors.New("custom id version is unsupported")
	ErrInvalidFeature       = errors.New("custom id feature is invalid")
	ErrInvalidAction        = errors.New("custom id action is invalid")
	ErrInvalidPayload       = errors.New("custom id payload is invalid")
	ErrUnknownLegacyID      = errors.New("legacy custom id is unknown")
	ErrAmbiguousID          = errors.New("custom id is ambiguous")
	ErrUnsafePayload        = errors.New("custom id payload is unsafe")
	ErrUnsupportedComponent = errors.New("discord component type is unsupported")
)

type ParseError struct {
	Err     error
	Context string
}

func (e *ParseError) Error() string {
	if e == nil {
		return ""
	}
	if e.Context == "" {
		return e.Err.Error()
	}
	return fmt.Sprintf("%s: %s", e.Err, e.Context)
}

func (e *ParseError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

func safeError(err error, context string) error {
	return &ParseError{Err: err, Context: context}
}
