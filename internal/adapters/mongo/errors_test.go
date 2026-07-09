package mongo

import (
	"context"
	"errors"
	"strings"
	"testing"

	drivermongo "go.mongodb.org/mongo-driver/v2/mongo"
)

func TestMapErrorNotFound(t *testing.T) {
	err := MapError(drivermongo.ErrNoDocuments)
	if !ErrorIs(err, ErrorKindNotFound) {
		t.Fatalf("expected not found, got %v", err)
	}
}

func TestMapErrorDuplicateKey(t *testing.T) {
	err := MapError(drivermongo.CommandError{Code: 11000, Message: "duplicate key"})
	if !ErrorIs(err, ErrorKindConflict) {
		t.Fatalf("expected conflict, got %v", err)
	}
}

func TestMapErrorTimeoutAndCanceled(t *testing.T) {
	if err := MapError(context.DeadlineExceeded); !ErrorIs(err, ErrorKindTimeout) {
		t.Fatalf("expected timeout, got %v", err)
	}
	if err := MapError(context.Canceled); !ErrorIs(err, ErrorKindCanceled) {
		t.Fatalf("expected canceled, got %v", err)
	}
}

func TestMapErrorValidation(t *testing.T) {
	err := MapError(drivermongo.CommandError{Code: 121, Name: "DocumentValidationFailure"})
	if !ErrorIs(err, ErrorKindInvalid) {
		t.Fatalf("expected invalid, got %v", err)
	}
}

func TestMapErrorTransient(t *testing.T) {
	err := MapError(drivermongo.CommandError{Labels: []string{"RetryableWriteError"}})
	if !ErrorIs(err, ErrorKindTransient) {
		t.Fatalf("expected transient, got %v", err)
	}
}

func TestMapErrorUnknownWrapsOriginal(t *testing.T) {
	original := errors.New("plain failure")
	err := MapError(original)
	if !ErrorIs(err, ErrorKindUnknown) {
		t.Fatalf("expected unknown, got %v", err)
	}
	if !errors.Is(err, original) {
		t.Fatalf("expected wrapped original")
	}
}

func TestSafeMessageDoesNotExposeRawURI(t *testing.T) {
	password := "pass" + "word"
	original := errors.New("mongodb://user:" + password + "@example.invalid/db failed")
	err := MapError(original)
	var mapped *Error
	if !errors.As(err, &mapped) {
		t.Fatalf("expected mapped error")
	}
	if strings.Contains(mapped.SafeMessage(), password) || strings.Contains(mapped.SafeMessage(), "mongodb://") {
		t.Fatalf("safe message leaked URI: %q", mapped.SafeMessage())
	}
}
