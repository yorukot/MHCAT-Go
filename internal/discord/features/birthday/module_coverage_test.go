package birthday

import (
	"errors"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/responses"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakemongo"
)

func TestBirthdayModuleMetadata(t *testing.T) {
	module := NewModule(&fakemongo.BirthdayConfigRepository{})
	if module.Name() != "birthday-config" || len(module.Commands()) == 0 {
		t.Fatalf("birthday metadata name=%q commands=%d", module.Name(), len(module.Commands()))
	}
}

func TestBirthdayErrorMessageHelpers(t *testing.T) {
	message := birthdayProfileErrorMessage(ports.ErrBirthdayProfileMissing, "missing")
	if len(message.Embeds) != 1 {
		t.Fatalf("profile error message = %#v", message)
	}
	ephemeral := birthdayAddEphemeralErrorMessage(responses.Message{Content: "failure"})
	if !ephemeral.Ephemeral || ephemeral.Content != "failure" {
		t.Fatalf("ephemeral message = %#v", ephemeral)
	}
	_ = birthdayProfileErrorMessage(errors.New("unknown"), "missing")
}
