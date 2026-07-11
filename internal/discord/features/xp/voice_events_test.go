package xp

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	coreservice "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/services/xp"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/events"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakebotinfo"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakediscord"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakemongo"
)

func TestVoiceXPEventMarksJoinMoveAndLeave(t *testing.T) {
	repo := fakemongo.NewXPAdminRepository()
	module := NewVoiceEventModule(repo)

	if err := module.VoiceStateHandler()(context.Background(), voiceXPEvent("voice-1", "")); err != nil {
		t.Fatalf("join: %v", err)
	}
	profile := repo.VoiceProfiles["guild-1/user-1"]
	if profile.LeaveJoin != domain.VoiceXPSessionJoined || profile.XP != 0 || profile.Level != 0 {
		t.Fatalf("joined profile = %#v", profile)
	}

	profile.XP = 75
	profile.Level = 2
	repo.VoiceProfiles["guild-1/user-1"] = profile
	if err := module.VoiceStateHandler()(context.Background(), voiceXPEvent("voice-2", "voice-1")); err != nil {
		t.Fatalf("move: %v", err)
	}
	profile = repo.VoiceProfiles["guild-1/user-1"]
	if profile.LeaveJoin != domain.VoiceXPSessionJoined || profile.XP != 75 || profile.Level != 2 {
		t.Fatalf("moved profile = %#v", profile)
	}

	if err := module.VoiceStateHandler()(context.Background(), voiceXPEvent("", "voice-2")); err != nil {
		t.Fatalf("leave: %v", err)
	}
	profile = repo.VoiceProfiles["guild-1/user-1"]
	if profile.LeaveJoin != domain.VoiceXPSessionLeft || profile.XP != 75 || profile.Level != 2 {
		t.Fatalf("left profile = %#v", profile)
	}
}

func TestVoiceXPEventIgnoresBotSameChannelAndMissingPayload(t *testing.T) {
	repo := fakemongo.NewXPAdminRepository()
	module := NewVoiceEventModule(repo)
	for _, event := range []events.Event{
		{Type: events.TypeMessageCreate},
		func() events.Event {
			event := voiceXPEvent("voice-1", "")
			event.IsBot = true
			return event
		}(),
		voiceXPEvent("voice-1", "voice-1"),
		{Type: events.TypeVoiceState, GuildID: "guild-1", VoiceState: &events.VoiceState{ChannelID: "voice-1"}},
	} {
		if err := module.VoiceStateHandler()(context.Background(), event); err != nil {
			t.Fatalf("ignored event returned error: %v", err)
		}
	}
	if len(repo.VoiceProfiles) != 0 {
		t.Fatalf("unexpected profiles = %#v", repo.VoiceProfiles)
	}
}

func TestVoiceXPEventRegisteredOnlyWithRepository(t *testing.T) {
	dispatcher := events.NewDispatcher(nil)
	NewVoiceEventModule(fakemongo.NewXPAdminRepository()).RegisterEventRoutes(dispatcher)
	if !dispatcher.HasHandlers(events.TypeVoiceState) {
		t.Fatal("expected voice XP event handler")
	}

	empty := events.NewDispatcher(nil)
	VoiceEventModule{}.RegisterEventRoutes(empty)
	if empty.HasHandlers(events.TypeVoiceState) {
		t.Fatal("unexpected voice XP event handler")
	}
}

func TestVoiceXPEventStartsAndStopsRuntimeWorker(t *testing.T) {
	repo := fakemongo.NewXPAdminRepository()
	module := NewVoiceEventModule(repo).WithRuntimeWorker(time.Hour, nil)
	t.Cleanup(func() { _ = module.StopRuntimeWorker(context.Background()) })

	event := voiceXPEvent("voice-1", "")
	event.Member = &events.Member{UserID: "user-1", RoleIDs: []string{"role-1"}}
	if err := module.VoiceStateHandler()(context.Background(), event); err != nil {
		t.Fatalf("join: %v", err)
	}
	if module.worker.ActiveCount() != 1 {
		t.Fatalf("active workers after join = %d", module.worker.ActiveCount())
	}
	if err := module.VoiceStateHandler()(context.Background(), voiceXPEvent("voice-2", "voice-1")); err != nil {
		t.Fatalf("move: %v", err)
	}
	if module.worker.ActiveCount() != 1 {
		t.Fatalf("active workers after move = %d", module.worker.ActiveCount())
	}
	if err := module.VoiceStateHandler()(context.Background(), voiceXPEvent("", "voice-2")); err != nil {
		t.Fatalf("leave: %v", err)
	}
	if module.worker.ActiveCount() != 0 {
		t.Fatalf("active workers after leave = %d", module.worker.ActiveCount())
	}
}

func TestVoiceXPStartJoinedSessionsRestoresRuntimeWorkers(t *testing.T) {
	repo := fakemongo.NewXPAdminRepository()
	repo.VoiceProfiles["guild-1/user-1"] = domain.XPProfile{GuildID: "guild-1", UserID: "user-1", LeaveJoin: domain.VoiceXPSessionJoined}
	repo.VoiceProfiles["guild-1/user-2"] = domain.XPProfile{GuildID: "guild-1", UserID: "user-2", LeaveJoin: domain.VoiceXPSessionLeft}
	repo.VoiceProfiles["guild-2/user-3"] = domain.XPProfile{GuildID: "guild-2", UserID: "user-3", LeaveJoin: domain.VoiceXPSessionJoined}
	module := NewVoiceEventModule(repo).WithRuntimeWorker(time.Hour, nil)
	t.Cleanup(func() { _ = module.StopRuntimeWorker(context.Background()) })

	started, err := module.StartJoinedSessions(context.Background())
	if err != nil {
		t.Fatalf("start joined sessions: %v", err)
	}
	if started != 2 || module.worker.ActiveCount() != 2 {
		t.Fatalf("started=%d active=%d", started, module.worker.ActiveCount())
	}
	started, err = module.StartJoinedSessions(context.Background())
	if err != nil {
		t.Fatalf("start joined sessions again: %v", err)
	}
	if started != 0 || module.worker.ActiveCount() != 2 {
		t.Fatalf("duplicate reconciliation started=%d active=%d", started, module.worker.ActiveCount())
	}
}

func TestVoiceXPStartJoinedSessionsRestoresOnlyOwnedShard(t *testing.T) {
	repo := fakemongo.NewXPAdminRepository()
	for shardID := 0; shardID < 3; shardID++ {
		guildID := fmt.Sprintf("%d", uint64(shardID)<<22)
		repo.VoiceProfiles[guildID+"/user"] = domain.XPProfile{GuildID: guildID, UserID: "user", LeaveJoin: domain.VoiceXPSessionJoined}
	}
	module := NewVoiceEventModule(repo).WithRuntimeWorker(time.Hour, nil).WithShard(1, 3)
	t.Cleanup(func() { _ = module.StopRuntimeWorker(context.Background()) })

	started, err := module.StartJoinedSessions(context.Background())
	if err != nil {
		t.Fatalf("start joined sessions: %v", err)
	}
	if started != 1 || module.worker.ActiveCount() != 1 {
		t.Fatalf("started=%d active=%d", started, module.worker.ActiveCount())
	}
}

func TestVoiceXPWorkerTicksAndStopsWhenProfileInactive(t *testing.T) {
	calls := make(chan []string, 1)
	worker := NewVoiceXPWorker(time.Millisecond, func(ctx context.Context, guildID string, userID string, currentRoleIDs []string) (coreservice.VoiceAccrualResult, error) {
		calls <- append([]string(nil), currentRoleIDs...)
		return coreservice.VoiceAccrualResult{}, nil
	}, nil)
	t.Cleanup(func() { _ = worker.StopAll(context.Background()) })

	if !worker.Start(" guild-1 ", " user-1 ", []string{" role-1 ", ""}) {
		t.Fatal("expected worker to start")
	}
	if worker.Start("guild-1", "user-1", nil) {
		t.Fatal("duplicate worker should not start")
	}
	select {
	case roles := <-calls:
		if len(roles) != 1 || roles[0] != "role-1" {
			t.Fatalf("roles = %#v", roles)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timed out waiting for voice XP tick")
	}
	waitUntil(t, 100*time.Millisecond, func() bool { return worker.ActiveCount() == 0 })
}

func TestVoiceXPWorkerStopAllPreventsNewStarts(t *testing.T) {
	worker := NewVoiceXPWorker(time.Hour, func(ctx context.Context, guildID string, userID string, currentRoleIDs []string) (coreservice.VoiceAccrualResult, error) {
		return coreservice.VoiceAccrualResult{Active: true}, nil
	}, nil)
	if !worker.Start("guild-1", "user-1", nil) {
		t.Fatal("expected worker to start")
	}
	if err := worker.StopAll(context.Background()); err != nil {
		t.Fatalf("stop all: %v", err)
	}
	if worker.ActiveCount() != 0 {
		t.Fatalf("active workers = %d", worker.ActiveCount())
	}
	if worker.Start("guild-1", "user-2", nil) {
		t.Fatal("worker should not start after StopAll")
	}
}

func TestVoiceXPTickAppliesAnnouncementRolesAndCoinReward(t *testing.T) {
	repo := fakemongo.NewXPAdminRepository()
	repo.VoiceProfiles["guild-1/user-1"] = domain.XPProfile{GuildID: "guild-1", UserID: "user-1", XP: 96, Level: 0, LeaveJoin: domain.VoiceXPSessionJoined}
	configs := fakemongo.NewVoiceXPConfigRepository()
	configs.Configs["guild-1"] = domain.VoiceXPConfig{GuildID: "guild-1", ChannelID: "level-channel", Message: "(user) voice {level}"}
	economy := fakemongo.NewEconomyRepository()
	economy.PutConfig(domain.EconomyConfig{GuildID: "guild-1", XPMultiple: 2.5})
	rewardRoles := fakemongo.NewVoiceXPRewardRoleRepository()
	rewardRoles.Configs = []domain.XPRewardRoleConfig{
		{GuildID: "guild-1", Level: 0, RoleID: "old-role", DeleteWhenNot: true},
		{GuildID: "guild-1", Level: 1, RoleID: "new-role"},
	}
	sideEffects := fakediscord.NewSideEffects()
	sideEffects.Channels = []ports.ChannelRef{{GuildID: "guild-1", ChannelID: "level-channel", Name: "語音等級"}}
	module := NewVoiceEventModule(repo).
		WithAccrual(repo, configs, sideEffects).
		WithAnnouncementFallbacks(sideEffects, sideEffects, &fakebotinfo.DiscordInfoProvider{Guild: ports.DiscordGuildInfo{OwnerID: "owner-1"}}).
		WithRewardRoles(rewardRoles, sideEffects).
		WithCoinRewards(economy)

	result, err := module.TickVoiceXP(context.Background(), "guild-1", "user-1", []string{"old-role"})
	if err != nil {
		t.Fatalf("tick voice xp: %v", err)
	}
	if !result.Leveled || result.Profile.Level != 1 || result.Profile.XP != 5 {
		t.Fatalf("result = %#v", result)
	}
	if len(sideEffects.Sent) != 1 || sideEffects.Sent[0].ChannelID != "level-channel" || sideEffects.Sent[0].Message.Content != "<@user-1> voice 1" {
		t.Fatalf("sent messages = %#v", sideEffects.Sent)
	}
	if len(sideEffects.Sent[0].Message.AllowedMentions.UserIDs) != 1 || sideEffects.Sent[0].Message.AllowedMentions.UserIDs[0] != "user-1" {
		t.Fatalf("allowed mentions = %#v", sideEffects.Sent[0].Message.AllowedMentions)
	}
	if len(sideEffects.RemovedRoles) != 1 || sideEffects.RemovedRoles[0].RoleID != "old-role" {
		t.Fatalf("removed roles = %#v", sideEffects.RemovedRoles)
	}
	if len(sideEffects.AddedRoles) != 1 || sideEffects.AddedRoles[0].RoleID != "new-role" {
		t.Fatalf("added roles = %#v", sideEffects.AddedRoles)
	}
	balance := economy.Balances["guild-1\x00user-1"]
	if balance.Coins != 2 || balance.Today != 0 {
		t.Fatalf("balance = %#v", balance)
	}
}

func TestVoiceXPTickAppliesRolesButNoCoinsWithoutAnnouncementConfig(t *testing.T) {
	repo := fakemongo.NewXPAdminRepository()
	repo.VoiceProfiles["guild-1/user-1"] = domain.XPProfile{GuildID: "guild-1", UserID: "user-1", XP: 96, Level: 0, LeaveJoin: domain.VoiceXPSessionJoined}
	configs := fakemongo.NewVoiceXPConfigRepository()
	economy := fakemongo.NewEconomyRepository()
	economy.PutConfig(domain.EconomyConfig{GuildID: "guild-1", XPMultiple: 3})
	rewardRoles := fakemongo.NewVoiceXPRewardRoleRepository()
	rewardRoles.Configs = []domain.XPRewardRoleConfig{{GuildID: "guild-1", Level: 1, RoleID: "new-role"}}
	sideEffects := fakediscord.NewSideEffects()
	module := NewVoiceEventModule(repo).
		WithAccrual(repo, configs, sideEffects).
		WithRewardRoles(rewardRoles, sideEffects).
		WithCoinRewards(economy)

	if _, err := module.TickVoiceXP(context.Background(), "guild-1", "user-1", nil); err != nil {
		t.Fatalf("tick voice xp: %v", err)
	}
	if len(sideEffects.AddedRoles) != 1 || sideEffects.AddedRoles[0].RoleID != "new-role" {
		t.Fatalf("added roles = %#v", sideEffects.AddedRoles)
	}
	if len(economy.Balances) != 0 {
		t.Fatalf("balances = %#v", economy.Balances)
	}
}

func TestVoiceXPTickDMsOwnerWhenAnnouncementChannelMissing(t *testing.T) {
	repo := fakemongo.NewXPAdminRepository()
	repo.VoiceProfiles["guild-1/user-1"] = domain.XPProfile{GuildID: "guild-1", UserID: "user-1", XP: 96, Level: 0, LeaveJoin: domain.VoiceXPSessionJoined}
	configs := fakemongo.NewVoiceXPConfigRepository()
	configs.Configs["guild-1"] = domain.VoiceXPConfig{GuildID: "guild-1", ChannelID: "missing-channel"}
	economy := fakemongo.NewEconomyRepository()
	economy.PutConfig(domain.EconomyConfig{GuildID: "guild-1", XPMultiple: 3})
	sideEffects := fakediscord.NewSideEffects()
	module := NewVoiceEventModule(repo).
		WithAccrual(repo, configs, sideEffects).
		WithAnnouncementFallbacks(sideEffects, sideEffects, &fakebotinfo.DiscordInfoProvider{Guild: ports.DiscordGuildInfo{OwnerID: "owner-1"}}).
		WithCoinRewards(economy)

	if _, err := module.TickVoiceXP(context.Background(), "guild-1", "user-1", nil); err != nil {
		t.Fatalf("tick voice xp: %v", err)
	}
	if len(sideEffects.DirectMessages) != 1 || sideEffects.DirectMessages[0].UserID != "owner-1" {
		t.Fatalf("direct messages = %#v", sideEffects.DirectMessages)
	}
	if got := sideEffects.DirectMessages[0].Message.Content; got != ":x: 有人的語音頻道等級升級了，但升等頻道已經被刪除了!" {
		t.Fatalf("direct message = %q", got)
	}
	if len(economy.Balances) != 0 {
		t.Fatalf("balances = %#v", economy.Balances)
	}
}

func TestVoiceXPTickDMsOwnerWhenAnnouncementSendFails(t *testing.T) {
	repo := fakemongo.NewXPAdminRepository()
	repo.VoiceProfiles["guild-1/user-1"] = domain.XPProfile{GuildID: "guild-1", UserID: "user-1", XP: 96, Level: 0, LeaveJoin: domain.VoiceXPSessionJoined}
	configs := fakemongo.NewVoiceXPConfigRepository()
	configs.Configs["guild-1"] = domain.VoiceXPConfig{GuildID: "guild-1", ChannelID: "level-channel"}
	economy := fakemongo.NewEconomyRepository()
	economy.PutConfig(domain.EconomyConfig{GuildID: "guild-1", XPMultiple: 3})
	sideEffects := fakediscord.NewSideEffects()
	sideEffects.Channels = []ports.ChannelRef{{GuildID: "guild-1", ChannelID: "level-channel", Name: "語音等級"}}
	module := NewVoiceEventModule(repo).
		WithAccrual(repo, configs, failingTextXPMessagePort{err: errors.New("send failed")}).
		WithAnnouncementFallbacks(sideEffects, sideEffects, &fakebotinfo.DiscordInfoProvider{Guild: ports.DiscordGuildInfo{OwnerID: "owner-1"}}).
		WithCoinRewards(economy)

	if _, err := module.TickVoiceXP(context.Background(), "guild-1", "user-1", nil); err != nil {
		t.Fatalf("tick voice xp: %v", err)
	}
	if len(sideEffects.DirectMessages) != 1 || sideEffects.DirectMessages[0].UserID != "owner-1" {
		t.Fatalf("direct messages = %#v", sideEffects.DirectMessages)
	}
	if got := sideEffects.DirectMessages[0].Message.Content; got != ":x: 有人的語音頻道等級升級了，但是我沒有權限在語音等級發送消息!\n因為你是該伺服器擁有者，所以我找你報告: P" {
		t.Fatalf("direct message = %q", got)
	}
	if len(economy.Balances) != 0 {
		t.Fatalf("balances = %#v", economy.Balances)
	}
}

func voiceXPEvent(channelID string, beforeChannelID string) events.Event {
	return events.Event{
		Type:    events.TypeVoiceState,
		GuildID: "guild-1",
		UserID:  "user-1",
		VoiceState: &events.VoiceState{
			GuildID:       "guild-1",
			UserID:        "user-1",
			ChannelID:     channelID,
			BeforeChannel: beforeChannelID,
		},
	}
}

func waitUntil(t *testing.T, timeout time.Duration, condition func() bool) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if condition() {
			return
		}
		time.Sleep(time.Millisecond)
	}
	if !condition() {
		t.Fatal("condition was not met before timeout")
	}
}
