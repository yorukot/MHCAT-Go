package logging

import (
	"errors"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakemongo"
)

func TestLoggingModuleMetadata(t *testing.T) {
	module := NewModule(&fakemongo.LoggingConfigRepository{}, nil)
	if module.Name() != "logging" || len(module.Commands()) == 0 {
		t.Fatalf("logging metadata name=%q commands=%d", module.Name(), len(module.Commands()))
	}
}

func TestLoggingErrorMapping(t *testing.T) {
	message := loggingErrorFromError(errors.New("hidden"))
	if len(message.Embeds) != 1 || message.Embeds[0].Title == "" {
		t.Fatalf("logging error message = %#v", message)
	}
}
