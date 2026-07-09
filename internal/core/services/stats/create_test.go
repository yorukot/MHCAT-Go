package stats

import (
	"context"
	"errors"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakediscord"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakemongo"
)

func TestCreateStatsConfigCreatesLegacyTextChannels(t *testing.T) {
	repo := fakemongo.NewStatsConfigRepository()
	discord := fakediscord.NewSideEffects()
	discord.TotalMembers = 10
	discord.NonBotMembers = 8
	service := CreateService{Repository: repo, Channels: discord, GuildStats: discord}

	config, err := service.Create(context.Background(), CreateRequest{
		GuildID:     "guild-1",
		ChannelType: domain.StatsChannelTypeText,
		BotUserID:   "bot-1",
	})
	if err != nil {
		t.Fatalf("create stats: %v", err)
	}
	if len(discord.Created) != 4 {
		t.Fatalf("created channels = %#v", discord.Created)
	}
	if discord.Created[0].Name != "伺服器統計數據(這串可隨便改)" || discord.Created[0].Type != discordChannelTypeGuildCategory {
		t.Fatalf("category request = %#v", discord.Created[0])
	}
	for index, want := range []string{"總人數: 10", "總成員: 8", "總BOT數: 2"} {
		request := discord.Created[index+1]
		if request.Name != want || request.Type != discordChannelTypeGuildText || request.ParentID != config.ParentID {
			t.Fatalf("created[%d] = %#v", index+1, request)
		}
		if len(request.PermissionOverwrites) != 2 || request.PermissionOverwrites[0].ID != "bot-1" || request.PermissionOverwrites[1].ID != "guild-1" {
			t.Fatalf("permission overwrites = %#v", request.PermissionOverwrites)
		}
	}
	saved := repo.Configs["guild-1"]
	if saved.ParentID == "" || saved.MemberNumberName != "10" || saved.UserNumberName != "8" || saved.BotNumberName != "2" {
		t.Fatalf("saved config = %#v", saved)
	}
}

func TestCreateStatsConfigCreatesLegacyVoiceChannelNames(t *testing.T) {
	repo := fakemongo.NewStatsConfigRepository()
	discord := fakediscord.NewSideEffects()
	discord.TotalMembers = 7
	discord.NonBotMembers = 6
	service := CreateService{Repository: repo, Channels: discord, GuildStats: discord}

	_, err := service.Create(context.Background(), CreateRequest{
		GuildID:     "guild-1",
		ChannelType: domain.StatsChannelTypeVoice,
	})
	if err != nil {
		t.Fatalf("create stats: %v", err)
	}
	for index, want := range []string{"總人數:7", "總成員:6", "總BOT數:1"} {
		request := discord.Created[index+1]
		if request.Name != want || request.Type != discordChannelTypeGuildVoice {
			t.Fatalf("created[%d] = %#v", index+1, request)
		}
		if len(request.PermissionOverwrites) != 1 || request.PermissionOverwrites[0].Deny != permissionConnect {
			t.Fatalf("voice overwrites = %#v", request.PermissionOverwrites)
		}
	}
}

func TestCreateStatsConfigRequiresOptionAfterBaseExists(t *testing.T) {
	repo := fakemongo.NewStatsConfigRepository()
	repo.Put(domain.StatsConfig{GuildID: "guild-1", ParentID: "parent-1"})
	discord := fakediscord.NewSideEffects()
	service := CreateService{Repository: repo, Channels: discord, GuildStats: discord}

	_, err := service.Create(context.Background(), CreateRequest{
		GuildID:     "guild-1",
		ChannelType: domain.StatsChannelTypeText,
	})
	if !errors.Is(err, domain.ErrStatsOptionRequired) {
		t.Fatalf("expected option required, got %v", err)
	}
	if len(discord.Created) != 0 {
		t.Fatalf("created channels = %#v", discord.Created)
	}
}

func TestCreateStatsConfigAddsOptionalChannelUnderExistingParent(t *testing.T) {
	repo := fakemongo.NewStatsConfigRepository()
	repo.Put(domain.StatsConfig{GuildID: "guild-1", ParentID: "parent-1"})
	discord := fakediscord.NewSideEffects()
	discord.Channels = append(discord.Channels, ports.ChannelRef{GuildID: "guild-1", ChannelID: "parent-1", Name: "stats", Type: discordChannelTypeGuildCategory})
	discord.ChannelCount = 12
	discord.TextChannelCount = 8
	discord.VoiceChannelCount = 4
	service := CreateService{Repository: repo, Channels: discord, GuildStats: discord}

	config, err := service.Create(context.Background(), CreateRequest{
		GuildID:     "guild-1",
		ChannelType: domain.StatsChannelTypeVoice,
		Option:      domain.StatsOptionVoiceCount,
	})
	if err != nil {
		t.Fatalf("create optional stats: %v", err)
	}
	if len(discord.Created) != 1 {
		t.Fatalf("created channels = %#v", discord.Created)
	}
	request := discord.Created[0]
	if request.Name != "總語音頻道數: 4" || request.Type != discordChannelTypeGuildVoice || request.ParentID != "parent-1" {
		t.Fatalf("optional request = %#v", request)
	}
	if config.VoiceNumberID == "" || config.VoiceNumberName != "4" {
		t.Fatalf("config = %#v", config)
	}
}

func TestCreateStatsConfigRejectsDuplicateOptionalChannel(t *testing.T) {
	repo := fakemongo.NewStatsConfigRepository()
	repo.Put(domain.StatsConfig{GuildID: "guild-1", ParentID: "parent-1", TextNumberID: "text-stat"})
	discord := fakediscord.NewSideEffects()
	service := CreateService{Repository: repo, Channels: discord, GuildStats: discord}

	_, err := service.Create(context.Background(), CreateRequest{
		GuildID:     "guild-1",
		ChannelType: domain.StatsChannelTypeText,
		Option:      domain.StatsOptionTextCount,
	})
	if !errors.Is(err, domain.ErrStatsChannelAlreadyExists) {
		t.Fatalf("expected duplicate option, got %v", err)
	}
	if len(discord.Created) != 0 {
		t.Fatalf("created channels = %#v", discord.Created)
	}
}
