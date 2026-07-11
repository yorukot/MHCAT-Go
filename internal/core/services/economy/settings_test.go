package economy_test

import (
	"context"
	"errors"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/services/economy"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakemongo"
)

func TestSettingsSaveConvertsHoursToSeconds(t *testing.T) {
	repo := fakemongo.NewEconomyRepository()
	service := economy.SettingsService{Repository: repo}

	config, err := service.Save(context.Background(), domain.EconomySettingsCommand{
		GuildID:           " guild-1 ",
		GachaCost:         700,
		SignCooldownHours: 2,
		SignCoins:         30,
		NotificationID:    " channel-1 ",
		XPMultiple:        2.5,
	})
	if err != nil {
		t.Fatalf("save settings: %v", err)
	}
	if config.GuildID != "guild-1" || config.GachaCost != 700 || config.SignCoins != 30 || config.ChannelID != "channel-1" || config.XPMultiple != 2.5 || config.ResetMarker != 7200 {
		t.Fatalf("config = %#v", config)
	}
	stored, err := repo.GetEconomyConfig(context.Background(), "guild-1")
	if err != nil {
		t.Fatalf("get stored config: %v", err)
	}
	if stored != config {
		t.Fatalf("stored config = %#v want %#v", stored, config)
	}
}

func TestSettingsSaveRejectsInvalidValues(t *testing.T) {
	repo := fakemongo.NewEconomyRepository()
	service := economy.SettingsService{Repository: repo}
	cases := []domain.EconomySettingsCommand{
		{GuildID: "", GachaCost: 1, SignCoins: 1, NotificationID: "channel", XPMultiple: 1},
		{GuildID: "guild", GachaCost: economy.MaxLegacyCoinBalance + 1, SignCoins: 1, NotificationID: "channel", XPMultiple: 1},
		{GuildID: "guild", GachaCost: 1, SignCoins: economy.MaxLegacyCoinBalance + 1, NotificationID: "channel", XPMultiple: 1},
		{GuildID: "guild", GachaCost: 1, SignCoins: 1, SignCooldownHours: -1, NotificationID: "channel", XPMultiple: 1},
		{GuildID: "guild", GachaCost: 1, SignCoins: 1, SignCooldownHours: int64(1<<63-1)/3600 + 1, NotificationID: "channel", XPMultiple: 1},
		{GuildID: "guild", GachaCost: 1, SignCoins: 1, NotificationID: "", XPMultiple: 1},
	}
	for _, command := range cases {
		if _, err := service.Save(context.Background(), command); !errors.Is(err, domain.ErrInvalidEconomySettings) {
			t.Fatalf("Save(%#v) error = %v, want ErrInvalidEconomySettings", command, err)
		}
	}
}
