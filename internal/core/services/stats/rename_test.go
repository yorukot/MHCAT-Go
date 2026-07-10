package stats

import (
	"context"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakediscord"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakemongo"
)

func TestRenameStatsChannelsUsesLegacyReplacementAndUpdatesCounters(t *testing.T) {
	repo := fakemongo.NewStatsConfigRepository()
	repo.Put(domain.StatsConfig{
		GuildID:          "guild-1",
		MemberNumberID:   "member-channel",
		MemberNumberName: "10",
		UserNumberID:     "user-channel",
		UserNumberName:   "8",
		BotNumberID:      "bot-channel",
		BotNumberName:    "2",
	})
	discord := fakediscord.NewSideEffects()
	discord.TotalMembers = 12
	discord.NonBotMembers = 9
	discord.Channels = append(discord.Channels,
		ports.ChannelRef{GuildID: "guild-1", ChannelID: "member-channel", Name: "總人數: 10"},
		ports.ChannelRef{GuildID: "guild-1", ChannelID: "user-channel", Name: "custom-users"},
		ports.ChannelRef{GuildID: "guild-1", ChannelID: "bot-channel", Name: "總BOT數: 2"},
	)
	channels := &cacheTrackingChannelPort{SideEffects: discord}
	service := RenameService{Repository: repo, Channels: channels, GuildStats: discord, RoleStats: discord}

	result, err := service.RunOnce(context.Background())
	if err != nil {
		t.Fatalf("rename stats: %v", err)
	}
	if result.ConfigsChecked != 1 || result.ChannelsRenamed != 3 || result.CountersUpdated != 1 {
		t.Fatalf("result = %#v", result)
	}
	if len(discord.Renamed) != 3 {
		t.Fatalf("renamed channels = %#v", discord.Renamed)
	}
	if channels.cachedCalls != 3 || channels.genericCalls != 0 {
		t.Fatalf("channel lookup calls = cached %d generic %d", channels.cachedCalls, channels.genericCalls)
	}
	wantNames := map[string]string{
		"member-channel": "總人數: 12",
		"user-channel":   "9",
		"bot-channel":    "總BOT數: 3",
	}
	for _, channel := range discord.Channels {
		if want, ok := wantNames[channel.ChannelID]; ok && channel.Name != want {
			t.Fatalf("channel %s name = %q, want %q", channel.ChannelID, channel.Name, want)
		}
	}
	saved := repo.Configs["guild-1"]
	if saved.MemberNumberName != "12" || saved.UserNumberName != "9" || saved.BotNumberName != "3" {
		t.Fatalf("saved counters = %#v", saved)
	}
}

type cacheTrackingChannelPort struct {
	*fakediscord.SideEffects
	cachedCalls  int
	genericCalls int
}

func (p *cacheTrackingChannelPort) FindChannelByID(ctx context.Context, guildID string, channelID string) (ports.ChannelRef, error) {
	p.genericCalls++
	return p.SideEffects.FindChannelByID(ctx, guildID, channelID)
}

func (p *cacheTrackingChannelPort) FindCachedChannelByID(ctx context.Context, guildID string, channelID string) (ports.ChannelRef, error) {
	p.cachedCalls++
	return p.SideEffects.FindCachedChannelByID(ctx, guildID, channelID)
}

func TestRenameStatsRoleChannelsUpdatesRoleNumber(t *testing.T) {
	repo := fakemongo.NewStatsConfigRepository()
	if err := repo.SaveStatsRoleConfig(context.Background(), domain.StatsRoleConfig{
		GuildID:     "guild-1",
		ChannelID:   "role-channel",
		ChannelName: "4",
		RoleID:      "role-1",
	}); err != nil {
		t.Fatalf("save role config: %v", err)
	}
	discord := fakediscord.NewSideEffects()
	discord.Channels = append(discord.Channels, ports.ChannelRef{GuildID: "guild-1", ChannelID: "role-channel", Name: "VIP: 4"})
	discord.RoleNames["guild-1/role-1"] = "VIP"
	discord.RoleMemberCounts["guild-1/role-1"] = 6
	service := RenameService{Repository: repo, Channels: discord, GuildStats: discord, RoleStats: discord}

	result, err := service.RunOnce(context.Background())
	if err != nil {
		t.Fatalf("rename role stats: %v", err)
	}
	if result.RoleConfigsChecked != 1 || result.ChannelsRenamed != 1 || result.CountersUpdated != 1 {
		t.Fatalf("result = %#v", result)
	}
	if discord.Channels[0].Name != "VIP: 6" {
		t.Fatalf("role channel name = %q", discord.Channels[0].Name)
	}
	if saved := repo.RoleConfigs["guild-1/role-1"]; saved.ChannelName != "6" {
		t.Fatalf("saved role config = %#v", saved)
	}
}

func TestRenameStatsSkipsMissingChannelsWithoutUpdatingCounter(t *testing.T) {
	repo := fakemongo.NewStatsConfigRepository()
	repo.Put(domain.StatsConfig{GuildID: "guild-1", MemberNumberID: "missing-channel", MemberNumberName: "10"})
	discord := fakediscord.NewSideEffects()
	discord.TotalMembers = 11
	service := RenameService{Repository: repo, Channels: discord, GuildStats: discord, RoleStats: discord}

	result, err := service.RunOnce(context.Background())
	if err != nil {
		t.Fatalf("rename stats: %v", err)
	}
	if result.ChannelsSkipped != 1 || result.ChannelsRenamed != 0 || result.CountersUpdated != 0 {
		t.Fatalf("result = %#v", result)
	}
	if saved := repo.Configs["guild-1"]; saved.MemberNumberName != "10" {
		t.Fatalf("saved config = %#v", saved)
	}
}

func TestLegacyStatsRenamedChannelName(t *testing.T) {
	if got := legacyStatsRenamedChannelName("總人數: 10 / 10", "10", "12"); got != "總人數: 12 / 10" {
		t.Fatalf("replace first old value = %q", got)
	}
	if got := legacyStatsRenamedChannelName("custom-name", "10", "12"); got != "12" {
		t.Fatalf("missing old value rename = %q", got)
	}
}
