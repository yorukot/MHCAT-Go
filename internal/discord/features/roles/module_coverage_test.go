package roles

import (
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakemongo"
)

func TestRoleModuleMetadataAndEventRegistration(t *testing.T) {
	module := NewModule(fakemongo.NewRoleSelectionRepository(), nil, nil, nil, nil, nil)
	if module.Name() != "role-selection" || len(module.Commands()) == 0 {
		t.Fatalf("role metadata name=%q commands=%d", module.Name(), len(module.Commands()))
	}
	module.RegisterEventRoutes(nil)
	if id := legacyRoleButtonID(); id == "" {
		t.Fatal("legacy role button id is empty")
	}
}
