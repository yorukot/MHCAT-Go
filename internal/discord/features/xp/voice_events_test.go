package xp

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
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
	NewVoiceEventModule(fakemongo.NewXPAdminRepository()).
		WithVoiceStateReader(staticVoiceStateReader{}).
		RegisterEventRoutes(dispatcher)
	if !dispatcher.HasHandlers(events.TypeVoiceState) {
		t.Fatal("expected voice XP event handler")
	}
	if !dispatcher.HasHandlers(events.TypeGuildAvailable) {
		t.Fatal("expected voice XP guild snapshot handler")
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

func TestVoiceXPGuildSnapshotRebuildsSessionsAndWorkers(t *testing.T) {
	repo := fakemongo.NewXPAdminRepository()
	repo.VoiceProfiles["guild-1/user-1"] = domain.XPProfile{GuildID: "guild-1", UserID: "user-1", XP: 75, Level: 2, LeaveJoin: domain.VoiceXPSessionJoined}
	repo.VoiceProfiles["guild-1/stale"] = domain.XPProfile{GuildID: "guild-1", UserID: "stale", LeaveJoin: domain.VoiceXPSessionJoined}
	repo.VoiceProfiles["guild-2/user-1"] = domain.XPProfile{GuildID: "guild-2", UserID: "user-1", LeaveJoin: domain.VoiceXPSessionJoined}
	reader := staticVoiceStateReader{states: map[string][]ports.DiscordVoiceState{
		"guild-1": {
			{UserID: "user-1", ChannelID: "voice-1", RoleIDs: []string{" role-2 ", "role-2"}},
			{UserID: "user-4", ChannelID: "voice-2", RoleIDs: []string{"role-4"}},
			{UserID: "bot-1", ChannelID: "voice-1", IsBot: true},
			{UserID: "not-connected"},
		},
	}}
	module := NewVoiceEventModule(repo).WithVoiceStateReader(reader).WithRuntimeWorker(time.Hour, nil)
	t.Cleanup(func() { _ = module.StopRuntimeWorker(context.Background()) })
	module.worker.Start("guild-1", "user-1", []string{"old-role"})
	module.worker.Start("guild-1", "stale", nil)
	module.worker.Start("guild-2", "user-1", nil)
	module.worker.mu.Lock()
	nextTick := module.worker.active[voiceXPWorkerKey("guild-1", "user-1")].nextTick
	module.worker.mu.Unlock()

	if err := module.GuildAvailableHandler()(context.Background(), events.Event{Type: events.TypeGuildAvailable, GuildID: "guild-1"}); err != nil {
		t.Fatalf("reconcile guild snapshot: %v", err)
	}
	if profile := repo.VoiceProfiles["guild-1/user-1"]; profile.LeaveJoin != domain.VoiceXPSessionJoined || profile.XP != 75 || profile.Level != 2 {
		t.Fatalf("preserved active profile = %#v", profile)
	}
	if profile := repo.VoiceProfiles["guild-1/stale"]; profile.LeaveJoin != domain.VoiceXPSessionLeft {
		t.Fatalf("stale profile = %#v", profile)
	}
	if profile := repo.VoiceProfiles["guild-1/user-4"]; profile.LeaveJoin != domain.VoiceXPSessionJoined {
		t.Fatalf("new active profile = %#v", profile)
	}
	if profile := repo.VoiceProfiles["guild-2/user-1"]; profile.LeaveJoin != domain.VoiceXPSessionJoined {
		t.Fatalf("other guild profile = %#v", profile)
	}
	module.worker.mu.Lock()
	active := module.worker.active
	userOne := active[voiceXPWorkerKey("guild-1", "user-1")]
	_, staleActive := active[voiceXPWorkerKey("guild-1", "stale")]
	_, otherGuildActive := active[voiceXPWorkerKey("guild-2", "user-1")]
	_, newActive := active[voiceXPWorkerKey("guild-1", "user-4")]
	module.worker.mu.Unlock()
	if userOne == nil || !userOne.nextTick.Equal(nextTick) || len(userOne.currentRoleIDs) != 1 || userOne.currentRoleIDs[0] != "role-2" {
		t.Fatalf("existing worker = %#v", userOne)
	}
	if staleActive || !otherGuildActive || !newActive || module.worker.ActiveCount() != 3 {
		t.Fatalf("workers stale=%t other=%t new=%t active=%d", staleActive, otherGuildActive, newActive, module.worker.ActiveCount())
	}
}

func TestVoiceXPGuildSnapshotFailureStopsOnlyThatGuild(t *testing.T) {
	failure := errors.New("cache failed")
	repo := fakemongo.NewXPAdminRepository()
	module := NewVoiceEventModule(repo).
		WithVoiceStateReader(staticVoiceStateReader{err: failure}).
		WithRuntimeWorker(time.Hour, nil)
	t.Cleanup(func() { _ = module.StopRuntimeWorker(context.Background()) })
	module.worker.Start("guild-1", "user-1", nil)
	module.worker.Start("guild-2", "user-1", nil)

	err := module.GuildAvailableHandler()(context.Background(), events.Event{Type: events.TypeGuildAvailable, GuildID: "guild-1"})
	if !errors.Is(err, failure) {
		t.Fatalf("snapshot error = %v", err)
	}
	module.worker.mu.Lock()
	_, failedGuildActive := module.worker.active[voiceXPWorkerKey("guild-1", "user-1")]
	_, otherGuildActive := module.worker.active[voiceXPWorkerKey("guild-2", "user-1")]
	module.worker.mu.Unlock()
	if failedGuildActive || !otherGuildActive {
		t.Fatalf("workers failed_guild=%t other_guild=%t", failedGuildActive, otherGuildActive)
	}
}

func TestVoiceXPSnapshotConcurrencyIsBounded(t *testing.T) {
	coordinator := newVoiceXPGuildCoordinator()
	releases := make([]func(), 0, voiceXPReconcileConcurrency)
	for range voiceXPReconcileConcurrency {
		release, err := coordinator.acquireSnapshot(context.Background())
		if err != nil {
			t.Fatalf("acquire snapshot slot: %v", err)
		}
		releases = append(releases, release)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	if _, err := coordinator.acquireSnapshot(ctx); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("fifth snapshot slot error = %v", err)
	}
	for _, release := range releases {
		release()
	}
	release, err := coordinator.acquireSnapshot(context.Background())
	if err != nil {
		t.Fatalf("reacquire released snapshot slot: %v", err)
	}
	release()
}

func TestVoiceXPLeaveWriteFailureStillStopsWorker(t *testing.T) {
	failure := errors.New("mongo failed")
	repo := fakemongo.NewXPAdminRepository()
	module := NewVoiceEventModule(repo).WithRuntimeWorker(time.Hour, nil)
	t.Cleanup(func() { _ = module.StopRuntimeWorker(context.Background()) })
	module.worker.Start("guild-1", "user-1", nil)
	repo.Err = failure

	err := module.VoiceStateHandler()(context.Background(), voiceXPEvent("", "voice-1"))
	if !errors.Is(err, failure) {
		t.Fatalf("leave error = %v", err)
	}
	if module.worker.ActiveCount() != 0 {
		t.Fatalf("active workers = %d", module.worker.ActiveCount())
	}
}

func TestVoiceXPJoinWriteFailureDoesNotStartWorker(t *testing.T) {
	failure := errors.New("mongo failed")
	repo := fakemongo.NewXPAdminRepository()
	repo.Err = failure
	module := NewVoiceEventModule(repo).WithRuntimeWorker(time.Hour, nil)
	t.Cleanup(func() { _ = module.StopRuntimeWorker(context.Background()) })

	err := module.VoiceStateHandler()(context.Background(), voiceXPEvent("voice-1", ""))
	if !errors.Is(err, failure) {
		t.Fatalf("join error = %v", err)
	}
	if module.worker.ActiveCount() != 0 {
		t.Fatalf("active workers = %d", module.worker.ActiveCount())
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

func TestVoiceXPWorkerProcessesManySessionsWithOneScheduler(t *testing.T) {
	var calls atomic.Int64
	var inFlight atomic.Int64
	var maxInFlight atomic.Int64
	worker := NewVoiceXPWorker(time.Millisecond, func(context.Context, string, string, []string) (coreservice.VoiceAccrualResult, error) {
		current := inFlight.Add(1)
		for {
			previous := maxInFlight.Load()
			if current <= previous || maxInFlight.CompareAndSwap(previous, current) {
				break
			}
		}
		time.Sleep(2 * time.Millisecond)
		inFlight.Add(-1)
		calls.Add(1)
		return coreservice.VoiceAccrualResult{}, nil
	}, nil)
	t.Cleanup(func() { _ = worker.StopAll(context.Background()) })
	for i := range 16 {
		if !worker.Start("guild-1", fmt.Sprintf("user-%d", i), nil) {
			t.Fatalf("start user %d", i)
		}
	}
	waitUntil(t, 250*time.Millisecond, func() bool { return worker.ActiveCount() == 0 })
	if calls.Load() != 16 || maxInFlight.Load() != 1 {
		t.Fatalf("calls=%d max_in_flight=%d", calls.Load(), maxInFlight.Load())
	}
}

func TestVoiceXPWorkerDoesNotCatchUpAfterSlowTick(t *testing.T) {
	started := make(chan struct{}, 2)
	release := make(chan struct{})
	worker := NewVoiceXPWorker(5*time.Millisecond, func(context.Context, string, string, []string) (coreservice.VoiceAccrualResult, error) {
		started <- struct{}{}
		<-release
		return coreservice.VoiceAccrualResult{Active: true}, nil
	}, nil)
	t.Cleanup(func() {
		close(release)
		_ = worker.StopAll(context.Background())
	})
	if !worker.Start("guild-1", "user-1", nil) {
		t.Fatal("expected worker to start")
	}
	select {
	case <-started:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timed out waiting for first voice XP tick")
	}
	time.Sleep(15 * time.Millisecond)
	release <- struct{}{}
	select {
	case <-started:
		t.Fatal("slow tick was followed by an immediate catch-up tick")
	case <-time.After(3 * time.Millisecond):
	}
	select {
	case <-started:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timed out waiting for the normally scheduled next tick")
	}
}

func TestVoiceXPWorkerReconcileKeepsWorkerAfterStaleInflightTick(t *testing.T) {
	started := make(chan struct{})
	release := make(chan struct{})
	var calls atomic.Int64
	worker := NewVoiceXPWorker(time.Hour, func(context.Context, string, string, []string) (coreservice.VoiceAccrualResult, error) {
		if calls.Add(1) == 1 {
			close(started)
			<-release
			return coreservice.VoiceAccrualResult{}, nil
		}
		return coreservice.VoiceAccrualResult{Active: true}, nil
	}, nil)
	t.Cleanup(func() { _ = worker.StopAll(context.Background()) })
	worker.Start("guild-1", "user-1", []string{"old-role"})
	worker.mu.Lock()
	worker.active[voiceXPWorkerKey("guild-1", "user-1")].nextTick = time.Now().Add(-time.Second)
	worker.mu.Unlock()
	runDone := make(chan struct{})
	go func() {
		worker.runDue(time.Now())
		close(runDone)
	}()
	<-started
	worker.ReconcileGuild("guild-1", []ports.DiscordVoiceState{{UserID: "user-1", ChannelID: "voice-1", RoleIDs: []string{"new-role"}}})
	close(release)
	<-runDone

	worker.mu.Lock()
	entry := worker.active[voiceXPWorkerKey("guild-1", "user-1")]
	worker.mu.Unlock()
	if entry == nil || len(entry.currentRoleIDs) != 1 || entry.currentRoleIDs[0] != "new-role" || time.Until(entry.nextTick) < 50*time.Minute {
		t.Fatalf("reconciled worker = %#v", entry)
	}
}

func TestVoiceXPSnapshotAndLiveEventAreSerializedPerGuild(t *testing.T) {
	readerStarted := make(chan struct{})
	releaseReader := make(chan struct{})
	repo := fakemongo.NewXPAdminRepository()
	reader := blockingVoiceStateReader{
		started: readerStarted,
		release: releaseReader,
		states:  []ports.DiscordVoiceState{{UserID: "user-1", ChannelID: "voice-1"}},
	}
	module := NewVoiceEventModule(repo).WithVoiceStateReader(reader).WithRuntimeWorker(time.Hour, nil)
	t.Cleanup(func() { _ = module.StopRuntimeWorker(context.Background()) })

	snapshotDone := make(chan error, 1)
	go func() {
		snapshotDone <- module.GuildAvailableHandler()(context.Background(), events.Event{Type: events.TypeGuildAvailable, GuildID: "guild-1"})
	}()
	<-readerStarted
	liveDone := make(chan error, 1)
	go func() {
		liveDone <- module.VoiceStateHandler()(context.Background(), voiceXPEvent("", "voice-1"))
	}()
	select {
	case err := <-liveDone:
		t.Fatalf("live event bypassed snapshot lock: %v", err)
	case <-time.After(10 * time.Millisecond):
	}
	close(releaseReader)
	if err := <-snapshotDone; err != nil {
		t.Fatalf("snapshot: %v", err)
	}
	if err := <-liveDone; err != nil {
		t.Fatalf("live leave: %v", err)
	}
	if profile := repo.VoiceProfiles["guild-1/user-1"]; profile.LeaveJoin != domain.VoiceXPSessionLeft {
		t.Fatalf("final profile = %#v", profile)
	}
	if module.worker.ActiveCount() != 0 {
		t.Fatalf("active workers = %d", module.worker.ActiveCount())
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

func TestVoiceXPTickRejectsAnnouncementChannelFromAnotherGuild(t *testing.T) {
	repo := fakemongo.NewXPAdminRepository()
	repo.VoiceProfiles["guild-1/user-1"] = domain.XPProfile{GuildID: "guild-1", UserID: "user-1", XP: 96, LeaveJoin: domain.VoiceXPSessionJoined}
	configs := fakemongo.NewVoiceXPConfigRepository()
	configs.Configs["guild-1"] = domain.VoiceXPConfig{GuildID: "guild-1", ChannelID: "level-channel"}
	economy := fakemongo.NewEconomyRepository()
	economy.PutConfig(domain.EconomyConfig{GuildID: "guild-1", XPMultiple: 3})
	sideEffects := fakediscord.NewSideEffects()
	sideEffects.Channels = []ports.ChannelRef{{GuildID: "guild-2", ChannelID: "level-channel", Name: "wrong guild"}}
	module := NewVoiceEventModule(repo).
		WithAccrual(repo, configs, sideEffects).
		WithAnnouncementFallbacks(sideEffects, sideEffects, &fakebotinfo.DiscordInfoProvider{Guild: ports.DiscordGuildInfo{OwnerID: "owner-1"}}).
		WithCoinRewards(economy)

	result, err := module.TickVoiceXP(context.Background(), "guild-1", "user-1", nil)
	if err != nil || !result.Leveled {
		t.Fatalf("tick result=%#v err=%v", result, err)
	}
	if len(sideEffects.Sent) != 0 {
		t.Fatalf("cross-guild messages = %#v", sideEffects.Sent)
	}
	if len(economy.Balances) != 0 {
		t.Fatalf("cross-guild coin rewards = %#v", economy.Balances)
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

type staticVoiceStateReader struct {
	states map[string][]ports.DiscordVoiceState
	err    error
}

type blockingVoiceStateReader struct {
	started chan<- struct{}
	release <-chan struct{}
	states  []ports.DiscordVoiceState
}

func (r blockingVoiceStateReader) GuildVoiceStates(ctx context.Context, _ string) ([]ports.DiscordVoiceState, error) {
	close(r.started)
	select {
	case <-r.release:
		return append([]ports.DiscordVoiceState(nil), r.states...), ctx.Err()
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (r staticVoiceStateReader) GuildVoiceStates(ctx context.Context, guildID string) ([]ports.DiscordVoiceState, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if r.err != nil {
		return nil, r.err
	}
	return append([]ports.DiscordVoiceState(nil), r.states[guildID]...), nil
}
