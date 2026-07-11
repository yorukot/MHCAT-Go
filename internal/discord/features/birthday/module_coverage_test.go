package birthday

import (
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakemongo"
)

func TestBirthdayModuleMetadata(t *testing.T) {
	module := NewModule(&fakemongo.BirthdayConfigRepository{})
	if module.Name() != "birthday-config" || len(module.Commands()) == 0 {
		t.Fatalf("birthday metadata name=%q commands=%d", module.Name(), len(module.Commands()))
	}
}
