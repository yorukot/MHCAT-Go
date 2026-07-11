package economy

import (
	"context"
	"strings"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/interactions"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakediscord"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakemongo"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakeusage"
)

func TestSettingsRequiresManageMessages(t *testing.T) {
	module := NewSettingsModule(fakemongo.NewEconomyRepository(), nil, nil)
	responder := fakediscord.NewResponder()
	interaction := settingsInteraction()

	if err := module.SettingsHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handle settings: %v", err)
	}
	if len(responder.Defers) != 1 || len(responder.Edits) != 1 {
		t.Fatalf("defers=%#v edits=%#v", responder.Defers, responder.Edits)
	}
	if !strings.Contains(responder.Edits[0].Embeds[0].Title, "訊息管理") {
		t.Fatalf("permission embed = %#v", responder.Edits[0].Embeds)
	}
}

func TestSettingsSavesLegacyConfigShape(t *testing.T) {
	repo := fakemongo.NewEconomyRepository()
	usage := &fakeusage.Tracker{}
	module := NewSettingsModule(repo, nil, usage)
	module.color = func() int { return 0x123456 }
	responder := fakediscord.NewResponder()
	interaction := settingsInteraction()
	interaction.Actor.PermissionBits = economySettingsManageMessagesPermission
	interaction.Actor.AvatarURL = "https://cdn.example/avatar.png"

	if err := module.SettingsHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handle settings: %v", err)
	}
	if len(repo.SavedConfigs) != 1 {
		t.Fatalf("saved configs = %#v", repo.SavedConfigs)
	}
	config := repo.SavedConfigs[0]
	if config.GuildID != "guild-1" || config.GachaCost != 700 || config.SignCoins != 30 || config.ChannelID != "222222222222222222" || config.XPMultiple != 2.5 || config.ResetMarker != 7200 {
		t.Fatalf("saved config = %#v", config)
	}
	if len(responder.Edits) != 1 || len(responder.Edits[0].Embeds) != 1 {
		t.Fatalf("edits = %#v", responder.Edits)
	}
	embed := responder.Edits[0].Embeds[0]
	if embed.Color != 0x123456 {
		t.Fatalf("success color = %#x", embed.Color)
	}
	if !strings.Contains(embed.Title, "扭蛋所需代幣:`700`") || !strings.Contains(embed.Title, "等級提升給予倍數:`2.5`") || embed.Description != "通知頻道:<#222222222222222222>" {
		t.Fatalf("success embed = %#v", embed)
	}
	if embed.Footer == nil || embed.Footer.Text != "MHCAT" || embed.Footer.IconURL != "https://cdn.example/avatar.png" {
		t.Fatalf("footer = %#v", embed.Footer)
	}
	if len(usage.Events) != 1 || usage.Events[0].CommandName != EconomySettingsCommandName || usage.Events[0].Feature != "economy-settings" {
		t.Fatalf("usage = %#v", usage.Events)
	}
}

func TestSettingsRejectsLegacyMaxAndNegativeValues(t *testing.T) {
	module := NewSettingsModule(fakemongo.NewEconomyRepository(), nil, nil)
	cases := []struct {
		name    string
		options map[string]string
		want    string
	}{
		{name: "max", options: map[string]string{"coin-raffle-takes": "1000000000"}, want: economySettingsMaxError},
		{name: "cooldown", options: map[string]string{"check-in-cooldown-time": "-1"}, want: economySettingsCooldownError},
		{name: "negative gacha", options: map[string]string{"coin-raffle-takes": "-1"}, want: economySettingsNonNegativeError},
		{name: "negative xp", options: map[string]string{"level-up-multiply-amount": "-0.5"}, want: economySettingsNonNegativeError},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			interaction := settingsInteraction()
			interaction.Actor.PermissionBits = economySettingsManageMessagesPermission
			for key, value := range tc.options {
				interaction.Options[key] = value
			}
			responder := fakediscord.NewResponder()
			if err := module.SettingsHandler()(context.Background(), interaction, responder); err != nil {
				t.Fatalf("handle settings: %v", err)
			}
			if len(responder.Edits) != 1 || !strings.Contains(responder.Edits[0].Embeds[0].Title, tc.want) {
				t.Fatalf("edits = %#v want %q", responder.Edits, tc.want)
			}
		})
	}
}

func TestSettingsModuleRegistersOnlySettingsRoute(t *testing.T) {
	repo := fakemongo.NewEconomyRepository()
	module := NewSettingsModule(repo, nil, nil)
	router := interactions.NewRouter()
	if err := module.RegisterRoutes(router); err != nil {
		t.Fatalf("register routes: %v", err)
	}
	responder := fakediscord.NewResponder()
	if err := router.Handle(context.Background(), fakediscord.SlashInteraction("代幣查詢"), responder); err == nil {
		t.Fatal("settings-only module should not register coin query")
	}
	interaction := settingsInteraction()
	interaction.Actor.PermissionBits = economySettingsManageMessagesPermission
	responder = fakediscord.NewResponder()
	if err := router.Handle(context.Background(), interaction, responder); err != nil {
		t.Fatalf("route settings: %v", err)
	}
	if len(repo.SavedConfigs) != 1 {
		t.Fatalf("saved configs = %#v", repo.SavedConfigs)
	}
}

func settingsInteraction() interactions.Interaction {
	return fakediscord.SlashInteractionWithOptions(EconomySettingsCommandName, "", map[string]string{
		"coin-raffle-takes":        "700",
		"check-in-cooldown-time":   "2",
		"check-in-give-coins":      "30",
		"notification-channel":     "222222222222222222",
		"level-up-multiply-amount": "2.5",
	})
}
