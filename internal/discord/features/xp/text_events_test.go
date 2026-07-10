package xp

import (
	"context"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/events"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakediscord"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakemongo"
)

func TestTextXPEventAccruesMessageXP(t *testing.T) {
	repo := fakemongo.NewXPAdminRepository()
	module := NewTextEventModule(repo, nil, nil)
	module.service.RandomMultiplier = fixedTextEventMultiplier(500)

	if err := module.MessageCreateHandler()(context.Background(), textXPEvent("hello")); err != nil {
		t.Fatalf("message create: %v", err)
	}
	profile := repo.TextProfiles["guild-1/user-1"]
	if profile.GuildID != "guild-1" || profile.UserID != "user-1" || profile.XP != 5 || profile.Level != 0 {
		t.Fatalf("profile = %#v", profile)
	}
}

func TestTextXPEventIgnoresBotDMAndNonMessageEvents(t *testing.T) {
	repo := fakemongo.NewXPAdminRepository()
	module := NewTextEventModule(repo, nil, nil)
	module.service.RandomMultiplier = fixedTextEventMultiplier(500)

	bot := textXPEvent("bot")
	bot.IsBot = true
	for _, event := range []events.Event{
		{Type: events.TypeMessageUpdate, GuildID: "guild-1", UserID: "user-1", Content: "edit"},
		{Type: events.TypeMessageCreate, UserID: "user-1", Content: "dm"},
		{Type: events.TypeMessageCreate, GuildID: "guild-1", Content: "missing user"},
		bot,
	} {
		if err := module.MessageCreateHandler()(context.Background(), event); err != nil {
			t.Fatalf("ignored event returned error: %v", err)
		}
	}
	if len(repo.TextProfiles) != 0 {
		t.Fatalf("unexpected profiles = %#v", repo.TextProfiles)
	}
}

func TestTextXPEventRegisteredOnlyWithRepository(t *testing.T) {
	dispatcher := events.NewDispatcher(nil)
	NewTextEventModule(fakemongo.NewXPAdminRepository(), nil, nil).RegisterEventRoutes(dispatcher)
	if !dispatcher.HasHandlers(events.TypeMessageCreate) {
		t.Fatal("expected text XP message handler")
	}

	empty := events.NewDispatcher(nil)
	TextEventModule{}.RegisterEventRoutes(empty)
	if empty.HasHandlers(events.TypeMessageCreate) {
		t.Fatal("unexpected text XP message handler")
	}
}

func TestTextXPEventSendsConfiguredLevelUpAnnouncement(t *testing.T) {
	repo := fakemongo.NewXPAdminRepository()
	repo.TextProfiles["guild-1/user-1"] = domain.XPProfile{GuildID: "guild-1", UserID: "user-1", XP: 96, Level: 0}
	configs := fakemongo.NewTextXPConfigRepository()
	configs.Configs["guild-1"] = domain.TextXPConfig{GuildID: "guild-1", ChannelID: "level-channel", Message: "(user) 升到了 {level}"}
	sideEffects := fakediscord.NewSideEffects()
	module := NewTextEventModule(repo, configs, sideEffects)
	module.service.RandomMultiplier = fixedTextEventMultiplier(500)

	if err := module.MessageCreateHandler()(context.Background(), textXPEvent("hello")); err != nil {
		t.Fatalf("message create: %v", err)
	}
	if len(sideEffects.Sent) != 1 {
		t.Fatalf("sent messages = %#v", sideEffects.Sent)
	}
	sent := sideEffects.Sent[0]
	if sent.ChannelID != "level-channel" || sent.Message.Content != "<@user-1> 升到了 1" {
		t.Fatalf("sent = %#v", sent)
	}
	if sent.Message.AllowedMentions.ParseUsers || sent.Message.AllowedMentions.ParseRoles || sent.Message.AllowedMentions.ParseEveryone {
		t.Fatalf("announcement should only allow explicit user mention: %#v", sent.Message.AllowedMentions)
	}
	if len(sent.Message.AllowedMentions.UserIDs) != 1 || sent.Message.AllowedMentions.UserIDs[0] != "user-1" {
		t.Fatalf("allowed users = %#v", sent.Message.AllowedMentions)
	}
	profile := repo.TextProfiles["guild-1/user-1"]
	if profile.Level != 1 || profile.XP != 0 {
		t.Fatalf("profile = %#v", profile)
	}
}

func TestTextXPEventUsesCurrentChannelAnnouncementSentinel(t *testing.T) {
	repo := fakemongo.NewXPAdminRepository()
	repo.TextProfiles["guild-1/user-1"] = domain.XPProfile{GuildID: "guild-1", UserID: "user-1", XP: 96, Level: 0}
	configs := fakemongo.NewTextXPConfigRepository()
	configs.Configs["guild-1"] = domain.TextXPConfig{GuildID: "guild-1", ChannelID: "ONCHANEL", Message: "level {level}"}
	sideEffects := fakediscord.NewSideEffects()
	module := NewTextEventModule(repo, configs, sideEffects)
	module.service.RandomMultiplier = fixedTextEventMultiplier(500)

	if err := module.MessageCreateHandler()(context.Background(), textXPEvent("hello")); err != nil {
		t.Fatalf("message create: %v", err)
	}
	if len(sideEffects.Sent) != 1 || sideEffects.Sent[0].ChannelID != "channel-1" || sideEffects.Sent[0].Message.Content != "level 1" {
		t.Fatalf("sent messages = %#v", sideEffects.Sent)
	}
}

func TestTextXPEventSkipsLevelUpAnnouncementWithoutConfig(t *testing.T) {
	repo := fakemongo.NewXPAdminRepository()
	repo.TextProfiles["guild-1/user-1"] = domain.XPProfile{GuildID: "guild-1", UserID: "user-1", XP: 96, Level: 0}
	configs := fakemongo.NewTextXPConfigRepository()
	sideEffects := fakediscord.NewSideEffects()
	module := NewTextEventModule(repo, configs, sideEffects)
	module.service.RandomMultiplier = fixedTextEventMultiplier(500)

	if err := module.MessageCreateHandler()(context.Background(), textXPEvent("hello")); err != nil {
		t.Fatalf("message create: %v", err)
	}
	if len(sideEffects.Sent) != 0 {
		t.Fatalf("sent messages = %#v", sideEffects.Sent)
	}
}

func TestTextXPEventReturnsAnnouncementRepositoryError(t *testing.T) {
	repo := fakemongo.NewXPAdminRepository()
	repo.TextProfiles["guild-1/user-1"] = domain.XPProfile{GuildID: "guild-1", UserID: "user-1", XP: 96, Level: 0}
	configs := fakemongo.NewTextXPConfigRepository()
	configs.Err = ports.ErrChannelNotFound
	module := NewTextEventModule(repo, configs, fakediscord.NewSideEffects())
	module.service.RandomMultiplier = fixedTextEventMultiplier(500)

	if err := module.MessageCreateHandler()(context.Background(), textXPEvent("hello")); err != ports.ErrChannelNotFound {
		t.Fatalf("expected config error, got %v", err)
	}
}

func TestTextXPEventAppliesRewardRolesOnLevelUp(t *testing.T) {
	repo := fakemongo.NewXPAdminRepository()
	repo.TextProfiles["guild-1/user-1"] = domain.XPProfile{GuildID: "guild-1", UserID: "user-1", XP: 96, Level: 0}
	rewardRoles := fakemongo.NewTextXPRewardRoleRepository()
	rewardRoles.Configs = []domain.XPRewardRoleConfig{
		{GuildID: "guild-1", Level: 0, RoleID: "old-role", DeleteWhenNot: true},
		{GuildID: "guild-1", Level: 1, RoleID: "new-role"},
		{GuildID: "guild-1", Level: 2, RoleID: "future-role", DeleteWhenNot: true},
	}
	sideEffects := fakediscord.NewSideEffects()
	module := NewTextEventModule(repo, fakemongo.NewTextXPConfigRepository(), sideEffects).WithRewardRoles(rewardRoles, sideEffects)
	module.service.RandomMultiplier = fixedTextEventMultiplier(500)
	event := textXPEvent("hello")
	event.Member = &events.Member{UserID: "user-1", RoleIDs: []string{"old-role"}}

	if err := module.MessageCreateHandler()(context.Background(), event); err != nil {
		t.Fatalf("message create: %v", err)
	}
	if len(sideEffects.RemovedRoles) != 1 || sideEffects.RemovedRoles[0].RoleID != "old-role" {
		t.Fatalf("removed roles = %#v", sideEffects.RemovedRoles)
	}
	if len(sideEffects.AddedRoles) != 1 || sideEffects.AddedRoles[0].RoleID != "new-role" {
		t.Fatalf("added roles = %#v", sideEffects.AddedRoles)
	}
}

func TestTextXPEventDoesNotApplyRewardRolesWithoutLevelUp(t *testing.T) {
	repo := fakemongo.NewXPAdminRepository()
	rewardRoles := fakemongo.NewTextXPRewardRoleRepository()
	rewardRoles.Configs = []domain.XPRewardRoleConfig{{GuildID: "guild-1", Level: 1, RoleID: "new-role"}}
	sideEffects := fakediscord.NewSideEffects()
	module := NewTextEventModule(repo, nil, nil).WithRewardRoles(rewardRoles, sideEffects)
	module.service.RandomMultiplier = fixedTextEventMultiplier(500)

	if err := module.MessageCreateHandler()(context.Background(), textXPEvent("hello")); err != nil {
		t.Fatalf("message create: %v", err)
	}
	if len(sideEffects.AddedRoles) != 0 || len(sideEffects.RemovedRoles) != 0 {
		t.Fatalf("role changes added=%#v removed=%#v", sideEffects.AddedRoles, sideEffects.RemovedRoles)
	}
}

func textXPEvent(content string) events.Event {
	return events.Event{
		Type:      events.TypeMessageCreate,
		GuildID:   "guild-1",
		ChannelID: "channel-1",
		UserID:    "user-1",
		Content:   content,
	}
}

func fixedTextEventMultiplier(value int64) func() int64 {
	return func() int64 { return value }
}
