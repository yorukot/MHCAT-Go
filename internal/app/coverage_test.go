package app

import (
	"bytes"
	"context"
	"testing"
	"time"
)

func TestDefaultFactoriesBuildConfiguredAdapters(t *testing.T) {
	cfg := validTestConfig()
	mongoClient, err := defaultMongoFactory(cfg)
	if err != nil || mongoClient == nil {
		t.Fatalf("default mongo factory: client=%#v err=%v", mongoClient, err)
	}
	discordSession, err := defaultDiscordFactory(cfg)
	if err != nil || discordSession == nil {
		t.Fatalf("default discord factory: session=%#v err=%v", discordSession, err)
	}
}

func TestMongoRepositoryFactoriesRejectNonDefaultClient(t *testing.T) {
	client := &fakeMongo{}
	tests := []struct {
		name string
		run  func() error
	}{
		{name: "ticket", run: func() error { _, err := ticketConfigRepositoryFromMongo(client); return err }},
		{name: "poll", run: func() error { _, err := pollRepositoryFromMongo(client); return err }},
		{name: "economy", run: func() error { _, err := economyRepositoryFromMongo(client); return err }},
		{name: "role selection", run: func() error { _, err := roleSelectionRepositoryFromMongo(client); return err }},
		{name: "economy profile", run: func() error { _, err := economyProfileRepositoryFromMongo(client); return err }},
		{name: "stats", run: func() error { _, err := statsConfigRepositoryFromMongo(client); return err }},
		{name: "work", run: func() error { _, err := workInterfaceRepositoryFromMongo(client); return err }},
		{name: "warning history", run: func() error { _, err := warningHistoryRepositoryFromMongo(client); return err }},
		{name: "warning settings", run: func() error { _, err := warningSettingsRepositoryFromMongo(client); return err }},
		{name: "delete data", run: func() error { _, err := deleteDataRepositoryFromMongo(client); return err }},
		{name: "logging", run: func() error { _, err := loggingConfigRepositoryFromMongo(client); return err }},
		{name: "autochat", run: func() error { _, err := autoChatConfigRepositoryFromMongo(client); return err }},
		{name: "autochat fallback", run: func() error { _, _, err := autoChatFallbackRepositoriesFromMongo(client); return err }},
		{name: "autochat paid", run: func() error { _, _, _, err := autoChatPaidRepositoriesFromMongo(client); return err }},
		{name: "auto notification", run: func() error { _, err := autoNotificationScheduleRepositoryFromMongo(client); return err }},
		{name: "scheduler lease", run: func() error { _, err := schedulerLeaseStoreFromMongo(client, "scheduler"); return err }},
		{name: "daily reset", run: func() error { _, err := dailyResetRepositoryFromMongo(client, "daily reset"); return err }},
		{name: "work payout", run: func() error { _, err := workPayoutRepositoryFromMongo(client, "work payout"); return err }},
		{name: "balance", run: func() error { _, err := balanceRepositoryFromMongo(client); return err }},
		{name: "usage", run: func() error { _, err := usageTrackerFromMongo(client); return err }},
		{name: "redeem", run: func() error { _, err := redeemRepositoryFromMongo(client); return err }},
		{name: "anti scam", run: func() error { _, err := antiScamConfigRepositoryFromMongo(client); return err }},
		{name: "scam catalog", run: func() error { _, err := scamURLCatalogRepositoryFromMongo(client); return err }},
		{name: "gacha", run: func() error { _, err := gachaRepositoryFromMongo(client); return err }},
		{name: "lottery", run: func() error { _, err := lotteryRepositoryFromMongo(client); return err }},
		{name: "birthday", run: func() error { _, err := birthdayConfigRepositoryFromMongo(client); return err }},
		{name: "announcement", run: func() error { _, err := announcementConfigRepositoryFromMongo(client); return err }},
		{name: "text xp", run: func() error { _, err := textXPConfigRepositoryFromMongo(client); return err }},
		{name: "voice xp", run: func() error { _, err := voiceXPConfigRepositoryFromMongo(client); return err }},
		{name: "text xp rewards", run: func() error { _, err := textXPRewardRoleRepositoryFromMongo(client); return err }},
		{name: "voice xp rewards", run: func() error { _, err := voiceXPRewardRoleRepositoryFromMongo(client); return err }},
		{name: "xp admin", run: func() error { _, err := xpAdminRepositoryFromMongo(client); return err }},
		{name: "voice room", run: func() error { _, err := voiceRoomConfigRepositoryFromMongo(client); return err }},
		{name: "voice lock", run: func() error { _, err := voiceRoomLockRepositoryFromMongo(client); return err }},
		{name: "voice state", run: func() error { _, err := voiceRoomStateRepositoryFromMongo(client); return err }},
		{name: "join role", run: func() error { _, err := joinRoleConfigRepositoryFromMongo(client); return err }},
		{name: "leave message", run: func() error { _, err := leaveMessageConfigRepositoryFromMongo(client); return err }},
		{name: "join message", run: func() error { _, err := joinMessageConfigRepositoryFromMongo(client); return err }},
		{name: "verification", run: func() error { _, err := verificationConfigRepositoryFromMongo(client); return err }},
		{name: "account age", run: func() error { _, err := accountAgeConfigRepositoryFromMongo(client); return err }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := test.run(); err == nil {
				t.Fatal("non-default mongo client was accepted")
			}
		})
	}
}

func TestDiscordAdapterHelpersRejectNonDefaultSession(t *testing.T) {
	session := &fakeDiscord{}
	if _, err := ticketSideEffectsFromSession(session); err == nil {
		t.Fatal("ticket side effects accepted non-default session")
	}
	if _, err := messageSideEffectsFromSession(session, "test feature"); err == nil {
		t.Fatal("message side effects accepted non-default session")
	}
	if provider := discordCachedUserInfoProvider(session); provider != nil {
		t.Fatalf("cached user provider = %#v", provider)
	}
}

func TestShutdownContextsAndRunConfigFailure(t *testing.T) {
	signalCtx, stop := SignalContext(context.Background())
	stop()
	select {
	case <-signalCtx.Done():
	case <-time.After(time.Second):
		t.Fatal("signal context did not stop")
	}
	timeoutCtx, cancel := ContextWithShutdownTimeout(context.Background(), time.Millisecond)
	defer cancel()
	select {
	case <-timeoutCtx.Done():
	case <-time.After(time.Second):
		t.Fatal("shutdown timeout did not expire")
	}
	t.Setenv("MHCAT_MONGODB_URI", "")
	t.Setenv("MONGOOSE_CONNECTION_STRING", "")
	t.Setenv("MHCAT_MONGODB_DATABASE", "")
	if err := Run(context.Background(), &bytes.Buffer{}); err == nil {
		t.Fatal("run accepted missing MongoDB configuration")
	}
}
