package onboarding

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakebotinfo"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakediscord"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakemongo"
)

type accountAgeClock struct {
	now time.Time
}

func (c accountAgeClock) Now() time.Time {
	return c.now
}

func TestAccountAgeConfigSetRequirementPreservesChannel(t *testing.T) {
	repo := fakemongo.NewAccountAgeConfigRepository()
	repo.Configs["guild-1"] = domain.AccountAgeConfig{GuildID: "guild-1", RequiredSeconds: 3600, ChannelID: "log-1"}
	service := AccountAgeConfigService{Repository: repo}

	config, err := service.SetRequirement(context.Background(), "guild-1", 48)
	if err != nil {
		t.Fatalf("set requirement: %v", err)
	}
	if config.RequiredSeconds != 172800 || config.ChannelID != "log-1" {
		t.Fatalf("config = %#v", config)
	}
}

func TestAccountAgePolicyKicksTooNewMemberAndLogs(t *testing.T) {
	now := time.Unix(2_000_000, 0)
	repo := fakemongo.NewAccountAgeConfigRepository()
	repo.Configs["guild-1"] = domain.AccountAgeConfig{GuildID: "guild-1", RequiredSeconds: 24 * 3600, ChannelID: "log-1"}
	sideEffects := fakediscord.NewSideEffects()
	info := &fakebotinfo.DiscordInfoProvider{Guild: ports.DiscordGuildInfo{Name: "測試伺服器"}}
	service := AccountAgePolicyService{
		Repository:     repo,
		DirectMessages: sideEffects,
		Members:        sideEffects,
		Messages:       sideEffects,
		Guilds:         info,
		Clock:          accountAgeClock{now: now},
	}

	result, err := service.GateMemberAdd(context.Background(), AccountAgeMemberEvent{
		GuildID:          "guild-1",
		UserID:           "user-1",
		UserTag:          "Tester#0001",
		AvatarURL:        "https://example.test/avatar.png",
		AccountCreatedAt: now.Add(-time.Hour),
	})
	if err != nil {
		t.Fatalf("gate: %v", err)
	}
	if !result.Matched || !result.Kicked || !result.Logged {
		t.Fatalf("result = %#v", result)
	}
	if len(sideEffects.DirectMessages) != 1 {
		t.Fatalf("direct messages = %#v", sideEffects.DirectMessages)
	}
	dmEmbed := sideEffects.DirectMessages[0].Message.Embeds[0]
	if dmEmbed.Title != "<a:warn:1000814885506129990> | 帳號創建時數過低!" {
		t.Fatalf("dm title = %#v", dmEmbed)
	}
	if !strings.Contains(dmEmbed.Description, "已將您踢出`測試伺服器`") || dmEmbed.FooterText != "管理員所設定時間: 24 小時" {
		t.Fatalf("dm embed = %#v", dmEmbed)
	}
	if len(sideEffects.Kicked) != 1 || sideEffects.Kicked[0].Reason != accountAgeKickReason {
		t.Fatalf("kicks = %#v", sideEffects.Kicked)
	}
	if len(sideEffects.Sent) != 1 || sideEffects.Sent[0].ChannelID != "log-1" {
		t.Fatalf("sent = %#v", sideEffects.Sent)
	}
	logEmbed := sideEffects.Sent[0].Message.Embeds[0]
	if logEmbed.Title != "低於管理員所設定的時數" || logEmbed.Fields[0].Name != "該使用者帳號創建時間:" {
		t.Fatalf("log embed = %#v", logEmbed)
	}
	if logEmbed.FooterText != "BAN:Tester#0001" {
		t.Fatalf("footer = %q", logEmbed.FooterText)
	}
}

func TestAccountAgePolicyDoesNotLogBanWhenKickFails(t *testing.T) {
	now := time.Unix(2_000_000, 0)
	repo := fakemongo.NewAccountAgeConfigRepository()
	repo.Configs["guild-1"] = domain.AccountAgeConfig{GuildID: "guild-1", RequiredSeconds: 24 * 3600, ChannelID: "log-1"}
	sideEffects := fakediscord.NewSideEffects()
	sideEffects.KickErr = errors.New("kick denied")
	service := AccountAgePolicyService{
		Repository:     repo,
		DirectMessages: sideEffects,
		Members:        sideEffects,
		Messages:       sideEffects,
		Clock:          accountAgeClock{now: now},
	}

	result, err := service.GateMemberAdd(context.Background(), AccountAgeMemberEvent{
		GuildID:          "guild-1",
		GuildName:        "測試伺服器",
		UserID:           "user-1",
		UserTag:          "Tester#0001",
		AccountCreatedAt: now.Add(-time.Hour),
	})
	if err == nil {
		t.Fatal("expected kick failure")
	}
	if !result.Matched || result.Kicked || result.Logged {
		t.Fatalf("result = %#v", result)
	}
	if len(sideEffects.DirectMessages) != 1 {
		t.Fatalf("direct messages = %#v", sideEffects.DirectMessages)
	}
	if len(sideEffects.Sent) != 0 {
		t.Fatalf("BAN log should not be sent when kick failed: %#v", sideEffects.Sent)
	}
}

func TestAccountAgePolicyUsesGatewayCreatedAtWithoutUserLookup(t *testing.T) {
	now := time.Unix(2_000_000, 0)
	repo := fakemongo.NewAccountAgeConfigRepository()
	repo.Configs["guild-1"] = domain.AccountAgeConfig{GuildID: "guild-1", RequiredSeconds: 24 * 3600}
	sideEffects := fakediscord.NewSideEffects()
	info := &fakebotinfo.DiscordInfoProvider{GuildErr: errors.New("guild lookup unavailable")}
	service := AccountAgePolicyService{
		Repository:     repo,
		DirectMessages: sideEffects,
		Members:        sideEffects,
		Guilds:         info,
		Clock:          accountAgeClock{now: now},
	}

	result, err := service.GateMemberAdd(context.Background(), AccountAgeMemberEvent{
		GuildID:          "guild-1",
		UserID:           "user-1",
		UserTag:          "Tester#0001",
		AccountCreatedAt: now.Add(-time.Hour),
	})
	if err != nil {
		t.Fatalf("gate: %v", err)
	}
	if !result.Matched || !result.Kicked {
		t.Fatalf("result = %#v", result)
	}
	if len(info.UserCalls) != 0 {
		t.Fatalf("account creation time was already available; user lookup calls = %#v", info.UserCalls)
	}
	if len(sideEffects.Kicked) != 1 {
		t.Fatalf("kicks = %#v", sideEffects.Kicked)
	}
	if len(sideEffects.DirectMessages) != 1 || !strings.Contains(sideEffects.DirectMessages[0].Message.Embeds[0].Description, "已將您踢出`guild-1`") {
		t.Fatalf("direct messages = %#v", sideEffects.DirectMessages)
	}
}

func TestAccountAgePolicyAllowsOldEnoughMember(t *testing.T) {
	now := time.Unix(2_000_000, 0)
	repo := fakemongo.NewAccountAgeConfigRepository()
	repo.Configs["guild-1"] = domain.AccountAgeConfig{GuildID: "guild-1", RequiredSeconds: 24 * 3600, ChannelID: "log-1"}
	sideEffects := fakediscord.NewSideEffects()
	service := AccountAgePolicyService{
		Repository: repo,
		Members:    sideEffects,
		Messages:   sideEffects,
		Clock:      accountAgeClock{now: now},
	}

	result, err := service.GateMemberAdd(context.Background(), AccountAgeMemberEvent{
		GuildID:          "guild-1",
		UserID:           "user-1",
		AccountCreatedAt: now.Add(-48 * time.Hour),
	})
	if err != nil {
		t.Fatalf("gate: %v", err)
	}
	if result.Matched || len(sideEffects.Kicked) != 0 || len(sideEffects.Sent) != 0 {
		t.Fatalf("result=%#v sideEffects=%#v", result, sideEffects)
	}
}
