package notifications

import (
	"errors"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakemongo"
)

func TestNotificationModuleMetadata(t *testing.T) {
	module := NewModule(fakemongo.NewAutoNotificationScheduleRepository(), nil, nil)
	if module.Name() != "auto-notification-config" || len(module.Commands()) == 0 {
		t.Fatalf("notification metadata name=%q commands=%d", module.Name(), len(module.Commands()))
	}
}

func TestNotificationErrorMapping(t *testing.T) {
	message := autoNotificationErrorFromError(errors.New("hidden"))
	if len(message.Embeds) != 1 || message.Embeds[0].Title == "" {
		t.Fatalf("notification error message = %#v", message)
	}
}
