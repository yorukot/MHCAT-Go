package mongo_test

import (
	"context"
	"errors"
	"testing"

	mhcatmongo "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/adapters/mongo"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakemongo"
)

func TestFakeTransactionRunnerCommitsSuccessfulCallback(t *testing.T) {
	var runner ports.TransactionRunner = &fakemongo.TransactionRunner{}
	err := runner.WithinTransaction(context.Background(), func(context.Context) error {
		return nil
	})
	if err != nil {
		t.Fatalf("transaction: %v", err)
	}
	fake := runner.(*fakemongo.TransactionRunner)
	if !fake.Committed || fake.RolledBack {
		t.Fatalf("fake runner state = %#v", fake)
	}
}

func TestFakeTransactionRunnerRollsBackFailedCallback(t *testing.T) {
	expected := errors.New("callback failed")
	fake := &fakemongo.TransactionRunner{}
	err := fake.WithinTransaction(context.Background(), func(context.Context) error {
		return expected
	})
	if !errors.Is(err, expected) {
		t.Fatalf("expected callback error, got %v", err)
	}
	if !fake.RolledBack || fake.Committed {
		t.Fatalf("fake runner state = %#v", fake)
	}
}

func TestFakeTransactionRunnerPropagatesContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	fake := &fakemongo.TransactionRunner{}
	err := fake.WithinTransaction(ctx, func(context.Context) error {
		t.Fatal("callback should not run")
		return nil
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context canceled, got %v", err)
	}
}

func TestNewTransactionRunnerRequiresConnectedClient(t *testing.T) {
	if _, err := mhcatmongo.NewTransactionRunner(nil); err == nil {
		t.Fatal("expected client validation error")
	}
}
