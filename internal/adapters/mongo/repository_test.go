package mongo_test

import (
	"context"
	"testing"

	mhcatmongo "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/adapters/mongo"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakemongo"
)

func TestRepositoryHealthContractWithFake(t *testing.T) {
	var health ports.RepositoryHealth = fakemongo.RepositoryHealth{}
	if err := health.Ping(context.Background()); err != nil {
		t.Fatalf("ping fake repository health: %v", err)
	}
}

func TestNewBaseRepositoryRequiresCollection(t *testing.T) {
	if _, err := mhcatmongo.NewBaseRepository(nil); err == nil {
		t.Fatal("expected collection validation error")
	}
}
