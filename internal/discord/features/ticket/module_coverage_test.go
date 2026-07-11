package ticket

import (
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakemongo"
)

func TestTicketModuleMetadata(t *testing.T) {
	module := NewModule(fakemongo.NewTicketConfigRepository())
	if module.Name() != "ticket" || len(module.Commands()) == 0 {
		t.Fatalf("ticket metadata name=%q commands=%d", module.Name(), len(module.Commands()))
	}
}
