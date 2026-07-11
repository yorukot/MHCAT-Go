package autochat

import (
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakemongo"
)

func TestAutoChatModuleMetadata(t *testing.T) {
	module := NewModule(fakemongo.NewAutoChatConfigRepository(), nil)
	if module.Name() != "autochat-config" || len(module.Commands()) == 0 {
		t.Fatalf("autochat metadata name=%q commands=%d", module.Name(), len(module.Commands()))
	}
}
