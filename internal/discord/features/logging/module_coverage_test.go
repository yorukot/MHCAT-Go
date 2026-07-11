package logging

import (
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakemongo"
)

func TestLoggingModuleMetadata(t *testing.T) {
	module := NewModule(&fakemongo.LoggingConfigRepository{}, nil)
	if module.Name() != "logging" || len(module.Commands()) == 0 {
		t.Fatalf("logging metadata name=%q commands=%d", module.Name(), len(module.Commands()))
	}
}
