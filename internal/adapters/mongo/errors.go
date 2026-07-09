package mongo

import (
	"context"
	"errors"
	"fmt"
	"strings"

	drivermongo "go.mongodb.org/mongo-driver/v2/mongo"
)

type ErrorKind string

const (
	ErrorKindNotFound  ErrorKind = "not_found"
	ErrorKindConflict  ErrorKind = "conflict"
	ErrorKindTimeout   ErrorKind = "timeout"
	ErrorKindCanceled  ErrorKind = "canceled"
	ErrorKindInvalid   ErrorKind = "invalid"
	ErrorKindTransient ErrorKind = "transient"
	ErrorKindUnknown   ErrorKind = "unknown"
)

type Error struct {
	Kind    ErrorKind
	Message string
	Err     error
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	if e.Message != "" {
		return e.Message
	}
	return string(e.Kind)
}

func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

func (e *Error) SafeMessage() string {
	switch e.Kind {
	case ErrorKindNotFound:
		return "requested document was not found"
	case ErrorKindConflict:
		return "document conflict"
	case ErrorKindTimeout:
		return "mongo operation timed out"
	case ErrorKindCanceled:
		return "mongo operation was canceled"
	case ErrorKindInvalid:
		return "invalid mongo request"
	case ErrorKindTransient:
		return "temporary mongo error"
	default:
		return "mongo operation failed"
	}
}

func MapError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, drivermongo.ErrNoDocuments) {
		return wrapError(ErrorKindNotFound, "mongo document not found", err)
	}
	if drivermongo.IsDuplicateKeyError(err) {
		return wrapError(ErrorKindConflict, "mongo duplicate key conflict", err)
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return wrapError(ErrorKindTimeout, "mongo operation timed out", err)
	}
	if errors.Is(err, context.Canceled) {
		return wrapError(ErrorKindCanceled, "mongo operation canceled", err)
	}
	if isValidationError(err) {
		return wrapError(ErrorKindInvalid, "mongo validation error", err)
	}
	if isTransientError(err) {
		return wrapError(ErrorKindTransient, "mongo transient error", err)
	}
	return wrapError(ErrorKindUnknown, "mongo operation failed", err)
}

func wrapError(kind ErrorKind, message string, err error) error {
	return &Error{Kind: kind, Message: message, Err: err}
}

func ErrorIs(err error, kind ErrorKind) bool {
	var mapped *Error
	if !errors.As(err, &mapped) {
		return false
	}
	return mapped.Kind == kind
}

func isValidationError(err error) bool {
	var commandError drivermongo.CommandError
	if errors.As(err, &commandError) {
		return commandError.HasErrorCode(121) || strings.Contains(strings.ToLower(commandError.Name), "validation")
	}
	return strings.Contains(strings.ToLower(err.Error()), "validation")
}

func isTransientError(err error) bool {
	var commandError drivermongo.CommandError
	if errors.As(err, &commandError) {
		return commandError.HasErrorLabel("TransientTransactionError") ||
			commandError.HasErrorLabel("RetryableWriteError") ||
			commandError.HasErrorLabel("NetworkError")
	}
	text := strings.ToLower(err.Error())
	return strings.Contains(text, "connection reset") ||
		strings.Contains(text, "connection refused") ||
		strings.Contains(text, "server selection")
}

func safeErrorForLog(err error) string {
	if err == nil {
		return ""
	}
	return fmt.Sprintf("%v", MapError(err))
}
