package poll

import (
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakemongo"
)

func TestPollModuleMetadata(t *testing.T) {
	module := NewModule(fakemongo.NewPollRepository())
	if module.Name() != "poll" || len(module.Commands()) == 0 {
		t.Fatalf("poll metadata name=%q commands=%d", module.Name(), len(module.Commands()))
	}
}
