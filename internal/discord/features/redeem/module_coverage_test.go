package redeem

import (
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakemongo"
)

func TestRedeemModuleMetadata(t *testing.T) {
	module := NewModule(fakemongo.NewRedeemRepository(), ports.SystemClock{}, nil)
	if module.Name() != "redeem" || len(module.Commands()) == 0 {
		t.Fatalf("redeem metadata name=%q commands=%d", module.Name(), len(module.Commands()))
	}
}
