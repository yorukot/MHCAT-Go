package safety

import (
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakemongo"
)

func TestSafetyModuleMetadataAndCombinedConstructor(t *testing.T) {
	module := NewModuleWithReport(fakemongo.NewAntiScamConfigRepository(), nil, nil)
	if module.Name() != "anti-scam-config" || len(module.Commands()) == 0 {
		t.Fatalf("safety metadata name=%q commands=%d", module.Name(), len(module.Commands()))
	}
}
