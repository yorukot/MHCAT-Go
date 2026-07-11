package voice

import (
	"errors"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/interactions"
)

func TestVoiceModuleMetadataAndRegistration(t *testing.T) {
	module := Module{}
	lockModule := LockModule{}
	if module.Name() != "voice-room-config" || lockModule.Name() != "voice-room-lock" {
		t.Fatalf("module names = %q, %q", module.Name(), lockModule.Name())
	}
	if len(module.Commands()) == 0 || len(lockModule.Commands()) == 0 {
		t.Fatal("voice module commands must not be empty")
	}
	router := interactions.NewRouter()
	if err := module.RegisterRoutes(router); err != nil {
		t.Fatalf("register voice routes: %v", err)
	}
	if err := lockModule.RegisterRoutes(router); err != nil {
		t.Fatalf("register voice lock routes: %v", err)
	}
	LockEventModule{}.RegisterEventRoutes(nil)
	RoomEventModule{}.RegisterEventRoutes(nil)
}

func TestVoiceFallbackHelpers(t *testing.T) {
	if color := legacyVoiceRandomColor(); color < 0 || color > 0xFFFFFF {
		t.Fatalf("voice color = %#x", color)
	}
	message := voiceUnknownError(errors.New("hidden"))
	if len(message.Embeds) != 1 || message.Embeds[0].Title == "" {
		t.Fatalf("voice error message = %#v", message)
	}
	interaction := interactions.Interaction{
		Options: map[string]string{"legacy": " legacy-value "},
		CommandOptions: map[string]interactions.CommandOptionValue{
			"typed": {String: " typed-value "},
		},
	}
	if firstOption(interaction, "legacy") != "legacy-value" || firstOption(interaction, "typed") != "typed-value" || firstOption(interaction, "missing") != "" {
		t.Fatalf("unexpected first option values")
	}
}
