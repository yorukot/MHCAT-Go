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

func TestCreateStatsRoleRequiresBaseStatsConfig(t *testing.T) {
	repo := fakemongo.NewStatsConfigRepository()
	discord := fakediscord.NewSideEffects()
	discord.RoleNames["guild-1/role-1"] = "VIP"
	service := RoleCreateService{StatsRepository: repo, RoleRepository: repo, Channels: discord, Roles: discord}

	_, err := service.Create(context.Background(), RoleCreateRequest{
		GuildID:     "guild-1",
		ChannelType: domain.StatsChannelTypeText,
		RoleID:      "role-1",
	})
	if !errors.Is(err, ports.ErrStatsConfigMissing) {
		t.Fatalf("expected missing stats config, got %v", err)
	}
	if len(discord.Created) != 0 {
		t.Fatalf("created channels = %#v", discord.Created)
	}
}

func TestCreateStatsRoleCreatesLegacyTextChannelAndSavesRoleNumber(t *testing.T) {
	repo := fakemongo.NewStatsConfigRepository()
	repo.Put(domain.StatsConfig{GuildID: "guild-1", ParentID: "parent-1"})
	discord := fakediscord.NewSideEffects()
	discord.Channels = append(discord.Channels, ports.ChannelRef{GuildID: "guild-1", ChannelID: "parent-1", Name: "stats", Type: discordChannelTypeGuildCategory})
	discord.RoleNames["guild-1/role-1"] = "VIP"
	discord.RoleMemberCounts["guild-1/role-1"] = 4
	service := RoleCreateService{StatsRepository: repo, RoleRepository: repo, Channels: discord, Roles: discord}

	config, err := service.Create(context.Background(), RoleCreateRequest{
		GuildID:     "guild-1",
		ChannelType: domain.StatsChannelTypeText,
		RoleID:      "role-1",
		BotUserID:   "bot-1",
	})
	if err != nil {
		t.Fatalf("create role stats: %v", err)
	}
	if len(discord.Created) != 1 {
		t.Fatalf("created channels = %#v", discord.Created)
	}
	request := discord.Created[0]
	if request.Name != "VIP: 4" || request.Type != discordChannelTypeGuildText || request.ParentID != "parent-1" {
		t.Fatalf("created channel = %#v", request)
	}
	if len(request.PermissionOverwrites) != 2 || request.PermissionOverwrites[0].ID != "bot-1" || request.PermissionOverwrites[0].Allow&permissionManageChannels == 0 || request.PermissionOverwrites[1].Deny != permissionSendMessages {
		t.Fatalf("permission overwrites = %#v", request.PermissionOverwrites)
	}
	saved := repo.RoleConfigs["guild-1/role-1"]
	if saved.ChannelID != config.ChannelID || saved.ChannelName != "4" || saved.RoleID != "role-1" {
		t.Fatalf("saved role config = %#v", saved)
	}
}

func TestRoleStatsPermissionOverwritesMatchLegacy(t *testing.T) {
	tests := []struct {
		name        string
		channelType int
		want        []ports.PermissionOverwrite
	}{
		{
			name:        "text",
			channelType: discordChannelTypeGuildText,
			want: []ports.PermissionOverwrite{
				{ID: "bot-1", Type: permissionOverwriteMember, Allow: permissionViewChannel | permissionManageChannels | permissionSendMessages},
				{ID: "guild-1", Type: permissionOverwriteRole, Allow: permissionViewChannel, Deny: permissionSendMessages},
			},
		},
		{
			name:        "voice",
			channelType: discordChannelTypeGuildVoice,
			want: []ports.PermissionOverwrite{
				{ID: "bot-1", Type: permissionOverwriteMember, Allow: permissionViewChannel | permissionManageMessages | permissionConnect},
				{ID: "guild-1", Type: permissionOverwriteRole, Allow: permissionViewChannel, Deny: permissionConnect},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := roleStatsPermissionOverwrites("guild-1", "bot-1", test.channelType)
			if len(got) != len(test.want) {
				t.Fatalf("overwrites = %#v", got)
			}
			for index := range test.want {
				if got[index] != test.want[index] {
					t.Fatalf("overwrite[%d] = %#v, want %#v", index, got[index], test.want[index])
				}
			}
		})
	}
}

func TestCreateStatsRoleCreatesLegacyVoiceChannelWithoutMissingParent(t *testing.T) {
	repo := fakemongo.NewStatsConfigRepository()
	repo.Put(domain.StatsConfig{GuildID: "guild-1", ParentID: "missing-parent"})
	discord := fakediscord.NewSideEffects()
	discord.RoleNames["guild-1/role-1"] = "VIP"
	discord.RoleMemberCounts["guild-1/role-1"] = 3
	service := RoleCreateService{StatsRepository: repo, RoleRepository: repo, Channels: discord, Roles: discord}

	_, err := service.Create(context.Background(), RoleCreateRequest{
		GuildID:     "guild-1",
		ChannelType: domain.StatsChannelTypeVoice,
		RoleID:      "role-1",
		BotUserID:   "bot-1",
	})
	if err != nil {
		t.Fatalf("create role stats: %v", err)
	}
	if len(discord.Created) != 1 {
		t.Fatalf("created channels = %#v", discord.Created)
	}
	request := discord.Created[0]
	if request.Name != "VIP: 3" || request.Type != discordChannelTypeGuildVoice || request.ParentID != "" {
		t.Fatalf("created channel = %#v", request)
	}
	if len(request.PermissionOverwrites) != 2 || request.PermissionOverwrites[0].ID != "bot-1" || request.PermissionOverwrites[1].Deny != permissionConnect {
		t.Fatalf("voice overwrites = %#v", request.PermissionOverwrites)
	}
}

func TestCreateStatsRoleRejectsMissingRole(t *testing.T) {
	repo := fakemongo.NewStatsConfigRepository()
	repo.Put(domain.StatsConfig{GuildID: "guild-1", ParentID: "parent-1"})
	discord := fakediscord.NewSideEffects()
	service := RoleCreateService{StatsRepository: repo, RoleRepository: repo, Channels: discord, Roles: discord}

	_, err := service.Create(context.Background(), RoleCreateRequest{
		GuildID:     "guild-1",
		ChannelType: domain.StatsChannelTypeText,
		RoleID:      "role-1",
	})
	if !errors.Is(err, ports.ErrDiscordRoleMissing) {
		t.Fatalf("expected missing role, got %v", err)
	}
	if len(discord.Created) != 0 {
		t.Fatalf("created channels = %#v", discord.Created)
	}
}

func TestCreateStatsRolePreservesSpacedParentAsCacheMiss(t *testing.T) {
	repo := fakemongo.NewStatsConfigRepository()
	repo.Put(domain.StatsConfig{GuildID: "guild-1", ParentID: " parent-1 "})
	discord := fakediscord.NewSideEffects()
	discord.Channels = append(discord.Channels, ports.ChannelRef{GuildID: "guild-1", ChannelID: "parent-1", Type: discordChannelTypeGuildCategory})
	discord.RoleNames["guild-1/role-1"] = "VIP"
	service := RoleCreateService{StatsRepository: repo, RoleRepository: repo, Channels: discord, Roles: discord}

	_, err := service.Create(context.Background(), RoleCreateRequest{
		GuildID: "guild-1", ChannelType: domain.StatsChannelTypeText, RoleID: "role-1",
	})
	if err != nil {
		t.Fatalf("create role stats: %v", err)
	}
	if len(discord.Created) != 1 || discord.Created[0].ParentID != "" {
		t.Fatalf("created channels = %#v", discord.Created)
	}
}
