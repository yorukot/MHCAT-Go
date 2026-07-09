package domain_test

import (
	"errors"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
)

func TestWarningSettingsValidate(t *testing.T) {
	valid := domain.WarningSettings{GuildID: "guild-1", Threshold: 3, Action: domain.WarningSettingsActionBan}
	if err := valid.Validate(); err != nil {
		t.Fatalf("valid settings: %v", err)
	}
	for _, settings := range []domain.WarningSettings{
		{GuildID: "", Threshold: 3, Action: domain.WarningSettingsActionBan},
		{GuildID: "guild-1", Threshold: 0, Action: domain.WarningSettingsActionBan},
		{GuildID: "guild-1", Threshold: 3, Action: "mute"},
	} {
		if err := settings.Validate(); !errors.Is(err, domain.ErrInvalidWarningSettings) {
			t.Fatalf("settings %#v err = %v", settings, err)
		}
	}
}
