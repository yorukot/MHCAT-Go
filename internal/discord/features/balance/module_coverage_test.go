package balance

import (
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakemongo"
)

func TestBalanceModuleMetadata(t *testing.T) {
	module := NewModule(fakemongo.NewBalanceRepository(), nil)
	if module.Name() != "balance-query" || len(module.Commands()) == 0 {
		t.Fatalf("balance metadata name=%q commands=%d", module.Name(), len(module.Commands()))
	}
}
