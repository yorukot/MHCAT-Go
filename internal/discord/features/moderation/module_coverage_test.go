package moderation

import (
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakemongo"
)

func TestModerationModuleMetadata(t *testing.T) {
	module := NewModule(fakemongo.NewWarningHistoryRepository(), nil, nil, nil)
	if module.Name() != "warnings" || len(module.Commands()) == 0 {
		t.Fatalf("moderation metadata name=%q commands=%d", module.Name(), len(module.Commands()))
	}
}
