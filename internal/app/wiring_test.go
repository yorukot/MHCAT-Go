package app

import (
	"context"
	"io"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/config"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	discordevents "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/events"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakebotinfo"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakediscord"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakemongo"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/faketranslate"
)

func TestBuildRuntimeRoutesUtilityCommands(t *testing.T) {
	dispatcher, err := BuildRuntime(RuntimeOptions{
		Config: validTestConfig(),
		Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
	})
	if err != nil {
		t.Fatalf("build runtime: %v", err)
	}
	responder := fakediscord.NewResponder()
	if err := dispatcher.Dispatch(context.Background(), fakediscord.SlashInteraction("ping"), responder); err != nil {
		t.Fatalf("dispatch ping: %v", err)
	}
	if len(responder.Replies) != 1 || !strings.Contains(responder.Replies[0].Content, "Pong!") {
		t.Fatalf("reply = %#v", responder.Replies)
	}
}

func TestDefaultEventRuntimeFactoryHasNoHandlersWhenRelayDisabled(t *testing.T) {
	dispatcher, err := defaultEventRuntimeFactory(validTestConfig(), slog.New(slog.NewTextHandler(io.Discard, nil)), nil, nil)
	if err != nil {
		t.Fatalf("default event runtime: %v", err)
	}
	if dispatcher.HasHandlers(discordevents.TypeMessageCreate) {
		t.Fatal("message event handler should not be registered when relay is disabled")
	}
}

func TestBuildRuntimeRoutesHelpDetail(t *testing.T) {
	dispatcher, err := BuildRuntime(RuntimeOptions{Config: validTestConfig()})
	if err != nil {
		t.Fatalf("build runtime: %v", err)
	}
	responder := fakediscord.NewResponder()
	interaction := fakediscord.SlashInteractionWithOptions("help", "", map[string]string{"指令名稱": "ping"})
	if err := dispatcher.Dispatch(context.Background(), interaction, responder); err != nil {
		t.Fatalf("dispatch help detail: %v", err)
	}
	if len(responder.Edits) != 1 || len(responder.Edits[0].Embeds) != 1 || !strings.Contains(responder.Edits[0].Embeds[0].Title, "指令資料") {
		t.Fatalf("edits = %#v replies=%#v", responder.Edits, responder.Replies)
	}
}

func TestBuildRuntimeRoutesInfoBot(t *testing.T) {
	dispatcher, err := BuildRuntime(RuntimeOptions{Config: validTestConfig()})
	if err != nil {
		t.Fatalf("build runtime: %v", err)
	}
	responder := fakediscord.NewResponder()
	interaction := fakediscord.SlashInteractionWithOptions("info", "bot", nil)
	if err := dispatcher.Dispatch(context.Background(), interaction, responder); err != nil {
		t.Fatalf("dispatch info bot: %v", err)
	}
	if len(responder.Follow) != 1 || len(responder.Follow[0].Embeds) != 1 || !strings.Contains(responder.Follow[0].Embeds[0].Title, "錯誤") {
		t.Fatalf("follow=%#v edits=%#v replies=%#v", responder.Follow, responder.Edits, responder.Replies)
	}
}

func TestBuildRuntimeRoutesInfoShard(t *testing.T) {
	dispatcher, err := BuildRuntime(RuntimeOptions{Config: validTestConfig()})
	if err != nil {
		t.Fatalf("build runtime: %v", err)
	}
	responder := fakediscord.NewResponder()
	interaction := fakediscord.SlashInteractionWithOptions("info", "shard", nil)
	if err := dispatcher.Dispatch(context.Background(), interaction, responder); err != nil {
		t.Fatalf("dispatch info shard: %v", err)
	}
	if len(responder.Follow) != 1 || len(responder.Follow[0].Embeds) != 1 || !strings.Contains(responder.Follow[0].Embeds[0].Title, "錯誤") {
		t.Fatalf("follow=%#v edits=%#v replies=%#v", responder.Follow, responder.Edits, responder.Replies)
	}
}

func TestBuildRuntimeRoutesHelpComponent(t *testing.T) {
	dispatcher, err := BuildRuntime(RuntimeOptions{Config: validTestConfig()})
	if err != nil {
		t.Fatalf("build runtime: %v", err)
	}
	responder := fakediscord.NewResponder()
	interaction := fakediscord.ComponentInteractionFromID("helphelphelphelpmenu")
	interaction.Values = []string{"實用工具"}
	if err := dispatcher.Dispatch(context.Background(), interaction, responder); err != nil {
		t.Fatalf("dispatch help component: %v", err)
	}
	if len(responder.Edits) != 1 || len(responder.Edits[0].Embeds) != 1 || !strings.Contains(responder.Edits[0].Embeds[0].Title, "實用工具 指令") {
		t.Fatalf("edits = %#v replies=%#v", responder.Edits, responder.Replies)
	}
}

func TestBuildRuntimeUsesConfiguredInteractionTimeout(t *testing.T) {
	cfg := validTestConfig()
	cfg.DiscordInteractionTimeout = config.DefaultInteractionTimeout
	dispatcher, err := BuildRuntime(RuntimeOptions{Config: cfg})
	if err != nil {
		t.Fatalf("build runtime: %v", err)
	}
	if dispatcher == nil {
		t.Fatal("dispatcher is nil")
	}
}

func TestBuildRuntimeRoutesTicketOnlyWithRepository(t *testing.T) {
	dispatcher, err := BuildRuntime(RuntimeOptions{Config: validTestConfig()})
	if err != nil {
		t.Fatalf("build runtime: %v", err)
	}
	responder := fakediscord.NewResponder()
	interaction := fakediscord.SlashInteractionWithOptions("私人頻道設置", "", map[string]string{
		"類別":     "222222222222222222",
		"管理員身分組": "333333333333333333",
	})
	interaction.Actor.PermissionBits = 8192
	if err := dispatcher.Dispatch(context.Background(), interaction, responder); err == nil {
		t.Fatal("ticket route should not be available without repository")
	}

	dispatcher, err = BuildRuntime(RuntimeOptions{
		Config:                 validTestConfig(),
		TicketConfigRepository: fakemongo.NewTicketConfigRepository(),
	})
	if err != nil {
		t.Fatalf("build runtime with ticket repo: %v", err)
	}
	responder = fakediscord.NewResponder()
	if err := dispatcher.Dispatch(context.Background(), interaction, responder); err != nil {
		t.Fatalf("dispatch ticket setup: %v", err)
	}
	if len(responder.Modals) != 1 || responder.Modals[0].Title != "私人頻道系統!" {
		t.Fatalf("modals = %#v", responder.Modals)
	}
}

func TestBuildRuntimeRoutesTicketOpenWithExplicitSideEffects(t *testing.T) {
	repo := fakemongo.NewTicketConfigRepository()
	if err := repo.SaveTicketConfig(context.Background(), domain.TicketConfig{
		GuildID:        "guild-1",
		CategoryID:     "222222222222222222",
		AdminRoleID:    "333333333333333333",
		EveryoneRoleID: "guild-1",
	}); err != nil {
		t.Fatalf("seed ticket config: %v", err)
	}
	sideEffects := fakediscord.NewSideEffects()
	dispatcher, err := BuildRuntime(RuntimeOptions{
		Config:                 validTestConfig(),
		TicketConfigRepository: repo,
		TicketChannelPort:      sideEffects,
		TicketMessagePort:      sideEffects,
		BotUserID:              "444444444444444444",
	})
	if err != nil {
		t.Fatalf("build runtime with ticket side effects: %v", err)
	}
	responder := fakediscord.NewResponder()
	interaction := fakediscord.ComponentInteractionFromID("tic")
	interaction.ChannelID = "panel-channel"
	interaction.MessageID = "panel-message"
	if err := dispatcher.Dispatch(context.Background(), interaction, responder); err != nil {
		t.Fatalf("dispatch ticket open: %v", err)
	}
	if len(sideEffects.Created) != 1 || sideEffects.Created[0].Name != "user-1" {
		t.Fatalf("created channels = %#v", sideEffects.Created)
	}
	if len(responder.Replies) != 1 || !strings.Contains(responder.Replies[0].Embeds[0].Description, "成功開啟") {
		t.Fatalf("replies = %#v", responder.Replies)
	}
}

func TestBuildRuntimeRoutesPollOnlyWithExplicitRepository(t *testing.T) {
	dispatcher, err := BuildRuntime(RuntimeOptions{Config: validTestConfig()})
	if err != nil {
		t.Fatalf("build runtime: %v", err)
	}
	responder := fakediscord.NewResponder()
	interaction := fakediscord.SlashInteractionWithOptions("投票創建", "", map[string]string{
		"問題": "問題",
		"選項": "A^B",
	})
	interaction.Actor.PermissionBits = 8192
	interaction.ChannelID = "channel-1"
	if err := dispatcher.Dispatch(context.Background(), interaction, responder); err == nil {
		t.Fatal("poll route should not be available without repository")
	}

	sideEffects := fakediscord.NewSideEffects()
	sideEffects.NonBotMembers = 2
	dispatcher, err = BuildRuntime(RuntimeOptions{
		Config:            validTestConfig(),
		PollRepository:    fakemongo.NewPollRepository(),
		PollMessagePort:   sideEffects,
		PollMemberCounter: sideEffects,
	})
	if err != nil {
		t.Fatalf("build runtime with poll repo: %v", err)
	}
	responder = fakediscord.NewResponder()
	if err := dispatcher.Dispatch(context.Background(), interaction, responder); err != nil {
		t.Fatalf("dispatch poll create: %v", err)
	}
	if len(sideEffects.Sent) != 1 || len(responder.Edits) != 1 {
		t.Fatalf("sent=%#v edits=%#v", sideEffects.Sent, responder.Edits)
	}
	if len(sideEffects.Sent[0].Message.Components) != 2 {
		t.Fatalf("poll components = %#v", sideEffects.Sent[0].Message.Components)
	}
}

func TestBuildRuntimeRoutesEconomyQueryOnlyWithRepository(t *testing.T) {
	dispatcher, err := BuildRuntime(RuntimeOptions{Config: validTestConfig()})
	if err != nil {
		t.Fatalf("build runtime: %v", err)
	}
	responder := fakediscord.NewResponder()
	interaction := fakediscord.SlashInteraction("代幣查詢")
	if err := dispatcher.Dispatch(context.Background(), interaction, responder); err == nil {
		t.Fatal("economy query route should not be available without repository")
	}

	repo := fakemongo.NewEconomyRepository()
	repo.PutBalance(domain.CoinBalance{GuildID: "guild-1", UserID: "user-1", Coins: 42})
	dispatcher, err = BuildRuntime(RuntimeOptions{
		Config:                 validTestConfig(),
		EconomyQueryRepository: repo,
	})
	if err != nil {
		t.Fatalf("build runtime with economy query repo: %v", err)
	}
	responder = fakediscord.NewResponder()
	if err := dispatcher.Dispatch(context.Background(), interaction, responder); err != nil {
		t.Fatalf("dispatch economy query: %v", err)
	}
	if len(responder.Edits) != 1 || len(responder.Edits[0].Embeds) != 1 || !strings.Contains(responder.Edits[0].Embeds[0].Title, "`42`個代幣") {
		t.Fatalf("economy response = %#v", responder.Edits)
	}
}

func TestBuildRuntimeRoutesEconomySignInWithoutPublishingQuery(t *testing.T) {
	repo := fakemongo.NewEconomyRepository()
	repo.PutSignInResult(domain.SignInResult{
		Balance:  domain.CoinBalance{GuildID: "guild-1", UserID: "user-1", Coins: 25, Today: 1},
		Calendar: domain.SignCalendar{GuildID: "guild-1", UserID: "user-1", Date: map[string]map[string][]string{}},
		Reward:   25,
		SignedAt: timeNowForTest(),
	})
	dispatcher, err := BuildRuntime(RuntimeOptions{
		Config:                  validTestConfig(),
		EconomySignInRepository: repo,
	})
	if err != nil {
		t.Fatalf("build runtime with sign-in repo: %v", err)
	}
	responder := fakediscord.NewResponder()
	if err := dispatcher.Dispatch(context.Background(), fakediscord.SlashInteraction("代幣查詢"), responder); err == nil {
		t.Fatal("economy query route should not be available with sign-in repository alone")
	}
	responder = fakediscord.NewResponder()
	if err := dispatcher.Dispatch(context.Background(), fakediscord.SlashInteraction("簽到"), responder); err != nil {
		t.Fatalf("dispatch sign-in: %v", err)
	}
	if len(responder.Edits) != 1 || len(responder.Edits[0].Files) != 1 {
		t.Fatalf("sign-in response = %#v", responder.Edits)
	}
}

func TestBuildRuntimeRoutesEconomySettingsWithoutPublishingQueryOrSignIn(t *testing.T) {
	repo := fakemongo.NewEconomyRepository()
	dispatcher, err := BuildRuntime(RuntimeOptions{
		Config:                    validTestConfig(),
		EconomySettingsRepository: repo,
	})
	if err != nil {
		t.Fatalf("build runtime with economy settings repo: %v", err)
	}
	for _, commandName := range []string{"代幣查詢", "簽到"} {
		responder := fakediscord.NewResponder()
		if err := dispatcher.Dispatch(context.Background(), fakediscord.SlashInteraction(commandName), responder); err == nil {
			t.Fatalf("%s route should not be available with settings repository alone", commandName)
		}
	}
	interaction := fakediscord.SlashInteractionWithOptions("coin-related-settings", "", map[string]string{
		"coin-raffle-takes":        "700",
		"check-in-cooldown-time":   "2",
		"check-in-give-coins":      "30",
		"notification-channel":     "222222222222222222",
		"level-up-multiply-amount": "2.5",
	})
	interaction.Actor.PermissionBits = 8192
	responder := fakediscord.NewResponder()
	if err := dispatcher.Dispatch(context.Background(), interaction, responder); err != nil {
		t.Fatalf("dispatch economy settings: %v", err)
	}
	if len(repo.SavedConfigs) != 1 || repo.SavedConfigs[0].ResetMarker != 7200 {
		t.Fatalf("saved configs = %#v", repo.SavedConfigs)
	}
}

func TestBuildRuntimeRoutesWorkOnlyWhenEnabled(t *testing.T) {
	dispatcher, err := BuildRuntime(RuntimeOptions{Config: validTestConfig()})
	if err != nil {
		t.Fatalf("build runtime: %v", err)
	}
	interaction := fakediscord.SlashInteractionWithOptions("打工系統", "新增打工事項", nil)
	responder := fakediscord.NewResponder()
	if err := dispatcher.Dispatch(context.Background(), interaction, responder); err == nil {
		t.Fatal("work route should not be available by default")
	}

	dispatcher, err = BuildRuntime(RuntimeOptions{
		Config:             validTestConfig(),
		WorkFeatureEnabled: true,
	})
	if err != nil {
		t.Fatalf("build runtime with work feature: %v", err)
	}
	responder = fakediscord.NewResponder()
	if err := dispatcher.Dispatch(context.Background(), interaction, responder); err != nil {
		t.Fatalf("dispatch work dashboard redirect: %v", err)
	}
	if len(responder.Replies) != 1 || len(responder.Replies[0].Components) != 1 {
		t.Fatalf("work response = %#v", responder.Replies)
	}
}

func TestBuildRuntimeRoutesWorkInterfaceOnlyWithRepository(t *testing.T) {
	dispatcher, err := BuildRuntime(RuntimeOptions{
		Config:             validTestConfig(),
		WorkFeatureEnabled: true,
	})
	if err != nil {
		t.Fatalf("build runtime with work feature: %v", err)
	}
	interaction := fakediscord.SlashInteractionWithOptions("打工系統", "打工介面", nil)
	responder := fakediscord.NewResponder()
	if err := dispatcher.Dispatch(context.Background(), interaction, responder); err != nil {
		t.Fatalf("dispatch work interface without repo: %v", err)
	}
	if len(responder.Replies) != 1 || !responder.Replies[0].Ephemeral {
		t.Fatalf("expected safe unimplemented reply without repo, got %#v", responder.Replies)
	}

	repo := fakemongo.NewWorkInterfaceRepository()
	repo.PutConfig(domain.WorkConfig{GuildID: "guild-1", MaxEnergy: 10})
	repo.PutItems("guild-1", domain.WorkItem{GuildID: "guild-1", Name: "礦坑", DurationSec: 3600, EnergyCost: 2, CoinReward: 7})
	repo.PutUser(domain.WorkUserState{GuildID: "guild-1", UserID: "user-1", State: domain.WorkIdleState, Energy: 5, Initialized: true})
	dispatcher, err = BuildRuntime(RuntimeOptions{
		Config:                  validTestConfig(),
		WorkFeatureEnabled:      true,
		WorkInterfaceRepository: repo,
	})
	if err != nil {
		t.Fatalf("build runtime with work repo: %v", err)
	}
	responder = fakediscord.NewResponder()
	if err := dispatcher.Dispatch(context.Background(), interaction, responder); err != nil {
		t.Fatalf("dispatch work interface with repo: %v", err)
	}
	if len(responder.Edits) != 1 || len(responder.Edits[0].Embeds) != 1 || !strings.Contains(responder.Edits[0].Embeds[0].Title, "打工簡章") {
		t.Fatalf("work interface response = %#v", responder.Edits)
	}
	detailID := responder.Edits[0].Components[0].Components[0].CustomID
	responder = fakediscord.NewResponder()
	if err := dispatcher.Dispatch(context.Background(), fakediscord.ComponentInteractionFromID(detailID), responder); err != nil {
		t.Fatalf("dispatch read-only work detail: %v", err)
	}
	if len(responder.Updates) != 1 || !responder.Updates[0].Components[0].Components[0].Disabled {
		t.Fatalf("expected read-only disabled confirm, got %#v", responder.Updates)
	}

	dispatcher, err = BuildRuntime(RuntimeOptions{
		Config:                  validTestConfig(),
		WorkFeatureEnabled:      true,
		WorkInterfaceRepository: repo,
		WorkStartRepository:     repo,
	})
	if err != nil {
		t.Fatalf("build runtime with work start repo: %v", err)
	}
	responder = fakediscord.NewResponder()
	if err := dispatcher.Dispatch(context.Background(), fakediscord.ComponentInteractionFromID(detailID), responder); err != nil {
		t.Fatalf("dispatch start-enabled work detail: %v", err)
	}
	if len(responder.Updates) != 1 || responder.Updates[0].Components[0].Components[0].Disabled {
		t.Fatalf("expected active confirm, got %#v", responder.Updates)
	}
	startID := responder.Updates[0].Components[0].Components[0].CustomID
	responder = fakediscord.NewResponder()
	if err := dispatcher.Dispatch(context.Background(), fakediscord.ComponentInteractionFromID(startID), responder); err != nil {
		t.Fatalf("dispatch work start: %v", err)
	}
	if len(responder.Updates) != 1 || !strings.Contains(responder.Updates[0].Embeds[0].Title, "成功取得該工作") {
		t.Fatalf("expected work start success, got %#v", responder.Updates)
	}
}

func TestBuildRuntimeRoutesWorkAdminOnlyWithAdminRepository(t *testing.T) {
	repo := fakemongo.NewWorkInterfaceRepository()
	dispatcher, err := BuildRuntime(RuntimeOptions{
		Config:              validTestConfig(),
		WorkFeatureEnabled:  true,
		WorkAdminRepository: repo,
	})
	if err != nil {
		t.Fatalf("build runtime with work admin repo: %v", err)
	}
	interaction := fakediscord.SlashInteractionWithOptions("打工系統", "打工系統設定", map[string]string{
		"每天可獲得多少精力": "5",
		"精力上限為多少":   "20",
		"是否需要驗證":    "true",
	})
	interaction.Actor.PermissionBits = 8192
	responder := fakediscord.NewResponder()
	if err := dispatcher.Dispatch(context.Background(), interaction, responder); err != nil {
		t.Fatalf("dispatch work settings: %v", err)
	}
	if len(responder.Edits) != 1 || !strings.Contains(responder.Edits[0].Embeds[0].Description, "成功設定打工系統") {
		t.Fatalf("work admin settings response = %#v", responder.Edits)
	}
	config, err := repo.GetWorkConfig(context.Background(), "guild-1")
	if err != nil {
		t.Fatalf("get saved work config: %v", err)
	}
	if config.DailyEnergy != 5 || config.MaxEnergy != 20 || !config.Captcha {
		t.Fatalf("saved config = %#v", config)
	}
}

func TestBuildRuntimeRoutesWarningsOnlyWithRepository(t *testing.T) {
	dispatcher, err := BuildRuntime(RuntimeOptions{Config: validTestConfig(), WarningsFeatureEnabled: true})
	if err != nil {
		t.Fatalf("build runtime with warnings feature: %v", err)
	}
	interaction := fakediscord.SlashInteractionWithOptions("警告紀錄", "", map[string]string{"使用者": "user-2"})
	interaction.Actor.PermissionBits = 8192
	responder := fakediscord.NewResponder()
	if err := dispatcher.Dispatch(context.Background(), interaction, responder); err == nil {
		t.Fatal("warning history route should not be available without repository")
	}

	repo := fakemongo.NewWarningHistoryRepository()
	repo.Put(domain.WarningHistory{
		GuildID: "guild-1",
		UserID:  "user-2",
		Entries: []domain.WarningEntry{{ModeratorID: "mod-1", Reason: "reason", Time: "time"}},
	})
	members := fakediscord.NewSideEffects()
	members.MemberTagValues["mod-1"] = "admin#0001"
	dispatcher, err = BuildRuntime(RuntimeOptions{
		Config:                   validTestConfig(),
		WarningsFeatureEnabled:   true,
		WarningHistoryRepository: repo,
		WarningMemberReader:      members,
	})
	if err != nil {
		t.Fatalf("build runtime with warning repo: %v", err)
	}
	responder = fakediscord.NewResponder()
	if err := dispatcher.Dispatch(context.Background(), interaction, responder); err != nil {
		t.Fatalf("dispatch warning history: %v", err)
	}
	if len(responder.Edits) != 1 || !strings.Contains(responder.Edits[0].Embeds[0].Title, "警告紀錄") {
		t.Fatalf("warning response = %#v", responder.Edits)
	}
}

func TestBuildRuntimeRoutesLoggingConfigOnlyWithRepository(t *testing.T) {
	dispatcher, err := BuildRuntime(RuntimeOptions{Config: validTestConfig()})
	if err != nil {
		t.Fatalf("build runtime: %v", err)
	}
	interaction := fakediscord.SlashInteractionWithOptions("set-log-channel", "", map[string]string{"channel": "channel-1"})
	interaction.Actor.PermissionBits = 8192
	responder := fakediscord.NewResponder()
	if err := dispatcher.Dispatch(context.Background(), interaction, responder); err == nil {
		t.Fatal("logging config route should not be available without repository")
	}

	repo := &fakemongo.LoggingConfigRepository{}
	dispatcher, err = BuildRuntime(RuntimeOptions{
		Config:                  validTestConfig(),
		LoggingConfigRepository: repo,
	})
	if err != nil {
		t.Fatalf("build runtime with logging config repo: %v", err)
	}
	responder = fakediscord.NewResponder()
	if err := dispatcher.Dispatch(context.Background(), interaction, responder); err != nil {
		t.Fatalf("dispatch logging config prompt: %v", err)
	}
	if len(responder.Edits) != 1 || len(responder.Edits[0].Components) != 1 {
		t.Fatalf("logging prompt response = %#v", responder.Edits)
	}
	componentID := responder.Edits[0].Components[0].Components[0].CustomID
	component := fakediscord.ComponentInteractionFromID(componentID)
	component.Values = []string{"訊息刪除", "頻道更新"}
	responder = fakediscord.NewResponder()
	if err := dispatcher.Dispatch(context.Background(), component, responder); err != nil {
		t.Fatalf("dispatch logging config select: %v", err)
	}
	saved, ok := repo.Last()
	if !ok || saved.ChannelID != "channel-1" || !saved.MessageDelete || !saved.ChannelUpdate || saved.MessageUpdate || saved.MemberVoiceUpdate {
		t.Fatalf("saved logging config = %#v ok=%v", saved, ok)
	}
}

func TestBuildRuntimeRoutesGachaPrizeListOnlyWithRepository(t *testing.T) {
	dispatcher, err := BuildRuntime(RuntimeOptions{Config: validTestConfig()})
	if err != nil {
		t.Fatalf("build runtime: %v", err)
	}
	interaction := fakediscord.SlashInteraction("扭蛋獎池查詢")
	responder := fakediscord.NewResponder()
	if err := dispatcher.Dispatch(context.Background(), interaction, responder); err == nil {
		t.Fatal("gacha prize-list route should not be available without repository")
	}

	repo := fakemongo.NewGachaRepository()
	repo.Prizes["guild-1"] = []domain.GachaPrize{{GuildID: "guild-1", Name: "大獎", Chance: 10, Count: 1}}
	repo.Configs["guild-1"] = domain.EconomyConfig{GuildID: "guild-1", GachaCost: 700, SignCoins: 40, XPMultiple: 2}
	dispatcher, err = BuildRuntime(RuntimeOptions{
		Config:                   validTestConfig(),
		GachaPrizePoolRepository: repo,
	})
	if err != nil {
		t.Fatalf("build runtime with gacha prize-list repo: %v", err)
	}
	responder = fakediscord.NewResponder()
	if err := dispatcher.Dispatch(context.Background(), interaction, responder); err != nil {
		t.Fatalf("dispatch gacha prize-list: %v", err)
	}
	if len(responder.Edits) != 1 || len(responder.Edits[0].Embeds) != 1 || !strings.Contains(responder.Edits[0].Embeds[0].Title, "以下是") {
		t.Fatalf("gacha prize-list response = %#v", responder.Edits)
	}
}

func TestBuildRuntimeRoutesLotteryDisabledCommandOnlyWhenEnabled(t *testing.T) {
	dispatcher, err := BuildRuntime(RuntimeOptions{Config: validTestConfig()})
	if err != nil {
		t.Fatalf("build runtime: %v", err)
	}
	interaction := fakediscord.SlashInteraction("抽獎設置")
	responder := fakediscord.NewResponder()
	if err := dispatcher.Dispatch(context.Background(), interaction, responder); err == nil {
		t.Fatal("lottery disabled command route should not be available by default")
	}

	dispatcher, err = BuildRuntime(RuntimeOptions{
		Config:                        validTestConfig(),
		LotteryDisabledCommandEnabled: true,
	})
	if err != nil {
		t.Fatalf("build runtime with lottery disabled command: %v", err)
	}
	responder = fakediscord.NewResponder()
	if err := dispatcher.Dispatch(context.Background(), interaction, responder); err != nil {
		t.Fatalf("dispatch lottery disabled command: %v", err)
	}
	if len(responder.Defers) != 1 || !responder.Defers[0].Ephemeral {
		t.Fatalf("defers = %#v", responder.Defers)
	}
	if len(responder.Edits) != 1 || len(responder.Edits[0].Embeds) != 1 || !strings.Contains(responder.Edits[0].Embeds[0].Title, "暫時無法使用") {
		t.Fatalf("lottery disabled command response = %#v", responder.Edits)
	}
}

func TestBuildRuntimeRoutesStatsQueryOnlyWhenEnabled(t *testing.T) {
	dispatcher, err := BuildRuntime(RuntimeOptions{Config: validTestConfig()})
	if err != nil {
		t.Fatalf("build runtime: %v", err)
	}
	interaction := fakediscord.SlashInteraction("統計系統查詢")
	responder := fakediscord.NewResponder()
	if err := dispatcher.Dispatch(context.Background(), interaction, responder); err == nil {
		t.Fatal("stats query route should not be available by default")
	}

	dispatcher, err = BuildRuntime(RuntimeOptions{
		Config:            validTestConfig(),
		StatsQueryEnabled: true,
	})
	if err != nil {
		t.Fatalf("build runtime with stats query: %v", err)
	}
	responder = fakediscord.NewResponder()
	if err := dispatcher.Dispatch(context.Background(), interaction, responder); err != nil {
		t.Fatalf("dispatch stats query: %v", err)
	}
	if len(responder.Replies) != 1 || len(responder.Replies[0].Embeds) != 1 || !strings.Contains(responder.Replies[0].Embeds[0].Description, "我的統計系統是每**10分鐘更新一次**") {
		t.Fatalf("stats query response = %#v", responder.Replies)
	}
}

func TestBuildRuntimeRoutesAutoChatConfigOnlyWithRepository(t *testing.T) {
	dispatcher, err := BuildRuntime(RuntimeOptions{Config: validTestConfig()})
	if err != nil {
		t.Fatalf("build runtime: %v", err)
	}
	interaction := fakediscord.SlashInteractionWithOptions("自動聊天頻道", "", map[string]string{
		"頻道": "channel-1",
	})
	interaction.Actor.PermissionBits = 8192
	responder := fakediscord.NewResponder()
	if err := dispatcher.Dispatch(context.Background(), interaction, responder); err == nil {
		t.Fatal("autochat config route should not be available without repository")
	}

	repo := fakemongo.NewAutoChatConfigRepository()
	dispatcher, err = BuildRuntime(RuntimeOptions{
		Config:                   validTestConfig(),
		AutoChatConfigRepository: repo,
	})
	if err != nil {
		t.Fatalf("build runtime with autochat repo: %v", err)
	}
	responder = fakediscord.NewResponder()
	if err := dispatcher.Dispatch(context.Background(), interaction, responder); err != nil {
		t.Fatalf("dispatch autochat config: %v", err)
	}
	saved, ok := repo.Last()
	if !ok || saved.GuildID != "guild-1" || saved.ChannelID != "channel-1" {
		t.Fatalf("saved = %#v ok=%v", saved, ok)
	}
	if len(responder.Edits) != 1 || !strings.Contains(responder.Edits[0].Embeds[0].Description, "自動聊天頻道成功創建") {
		t.Fatalf("edits = %#v", responder.Edits)
	}
}

func TestBuildRuntimeRoutesAntiScamConfigOnlyWithRepository(t *testing.T) {
	dispatcher, err := BuildRuntime(RuntimeOptions{Config: validTestConfig()})
	if err != nil {
		t.Fatalf("build runtime: %v", err)
	}
	interaction := fakediscord.SlashInteraction("防詐騙網址")
	interaction.Actor.PermissionBits = 8192
	responder := fakediscord.NewResponder()
	if err := dispatcher.Dispatch(context.Background(), interaction, responder); err == nil {
		t.Fatal("anti-scam config route should not be available without repository")
	}

	repo := fakemongo.NewAntiScamConfigRepository()
	dispatcher, err = BuildRuntime(RuntimeOptions{
		Config:                   validTestConfig(),
		AntiScamConfigRepository: repo,
	})
	if err != nil {
		t.Fatalf("build runtime with anti-scam config repo: %v", err)
	}
	responder = fakediscord.NewResponder()
	if err := dispatcher.Dispatch(context.Background(), interaction, responder); err != nil {
		t.Fatalf("dispatch anti-scam config: %v", err)
	}
	saved, ok := repo.Last()
	if !ok || saved.GuildID != "guild-1" || !saved.Open {
		t.Fatalf("saved = %#v ok=%v", saved, ok)
	}
	if len(responder.Edits) != 1 || !strings.Contains(responder.Edits[0].Embeds[0].Description, "您的防詐騙啟用狀態已改為:\ntrue") {
		t.Fatalf("edits = %#v", responder.Edits)
	}
}

func TestBuildRuntimeRoutesBirthdayConfigOnlyWithRepository(t *testing.T) {
	dispatcher, err := BuildRuntime(RuntimeOptions{Config: validTestConfig()})
	if err != nil {
		t.Fatalf("build runtime: %v", err)
	}
	interaction := fakediscord.SlashInteractionWithOptions("生日系統", "祝福語設定", map[string]string{
		"祝福語":        "{user} 生日快樂",
		"頻道":         "channel-1",
		"是否可以自行設定生日": "true",
		"時區":         "+08:00",
	})
	interaction.Actor.PermissionBits = 8192
	responder := fakediscord.NewResponder()
	if err := dispatcher.Dispatch(context.Background(), interaction, responder); err == nil {
		t.Fatal("birthday config route should not be available without repository")
	}

	repo := &fakemongo.BirthdayConfigRepository{}
	dispatcher, err = BuildRuntime(RuntimeOptions{
		Config:                   validTestConfig(),
		BirthdayConfigRepository: repo,
	})
	if err != nil {
		t.Fatalf("build runtime with birthday repo: %v", err)
	}
	responder = fakediscord.NewResponder()
	if err := dispatcher.Dispatch(context.Background(), interaction, responder); err != nil {
		t.Fatalf("dispatch birthday config: %v", err)
	}
	saved, ok := repo.Last()
	if !ok || saved.GuildID != "guild-1" || saved.ChannelID != "channel-1" {
		t.Fatalf("saved = %#v ok=%v", saved, ok)
	}
	if len(responder.Edits) != 1 || !strings.Contains(responder.Edits[0].Embeds[0].Title, "生日系統祝福語設定") {
		t.Fatalf("edits = %#v", responder.Edits)
	}
}

func TestBuildRuntimeRoutesBirthdayAddComponentsWithRepository(t *testing.T) {
	repo := &fakemongo.BirthdayConfigRepository{Configs: map[string]domain.BirthdayConfig{
		"guild-1": {
			GuildID:                    "guild-1",
			Message:                    "{user} 生日快樂",
			UTCOffset:                  "+08:00",
			ChannelID:                  "channel-1",
			EveryoneCanSetBirthdayDate: true,
		},
	}}
	dispatcher, err := BuildRuntime(RuntimeOptions{
		Config:                   validTestConfig(),
		BirthdayConfigRepository: repo,
		Clock:                    appFixedClock{now: time.Unix(1700000000, 0)},
	})
	if err != nil {
		t.Fatalf("build runtime with birthday repo: %v", err)
	}
	interaction := fakediscord.SlashInteractionWithOptions("生日系統", "增加", map[string]string{
		"生日月份": "7",
		"生日日期": "9",
		"生日年份": "2000",
	})
	responder := fakediscord.NewResponder()
	if err := dispatcher.Dispatch(context.Background(), interaction, responder); err != nil {
		t.Fatalf("dispatch birthday add: %v", err)
	}
	if len(responder.Edits) != 1 || len(responder.Edits[0].Components) != 1 {
		t.Fatalf("birthday add prompt = %#v", responder.Edits)
	}
	hourCustomID := responder.Edits[0].Components[0].Components[0].CustomID
	componentResponder := fakediscord.NewResponder()
	component := fakediscord.ComponentInteractionFromID(hourCustomID)
	component.Values = []string{"8"}
	if err := dispatcher.Dispatch(context.Background(), component, componentResponder); err != nil {
		t.Fatalf("dispatch birthday hour component: %v", err)
	}
	if len(componentResponder.Updates) != 1 || !strings.Contains(componentResponder.Updates[0].Embeds[0].Description, "請選取你的生日通知要在幾分發送") {
		t.Fatalf("birthday hour update = %#v", componentResponder.Updates)
	}
}

func TestBuildRuntimeRoutesAnnouncementConfigOnlyWithRepository(t *testing.T) {
	dispatcher, err := BuildRuntime(RuntimeOptions{Config: validTestConfig()})
	if err != nil {
		t.Fatalf("build runtime: %v", err)
	}
	interaction := fakediscord.SlashInteractionWithOptions("公告頻道設置", "一次性公告頻道", map[string]string{
		"頻道": "channel-1",
	})
	interaction.Actor.PermissionBits = 8192
	responder := fakediscord.NewResponder()
	if err := dispatcher.Dispatch(context.Background(), interaction, responder); err == nil {
		t.Fatal("announcement config route should not be available without repository")
	}

	repo := fakemongo.NewAnnouncementConfigRepository()
	dispatcher, err = BuildRuntime(RuntimeOptions{
		Config:                       validTestConfig(),
		AnnouncementConfigRepository: repo,
	})
	if err != nil {
		t.Fatalf("build runtime with announcement config repo: %v", err)
	}
	responder = fakediscord.NewResponder()
	if err := dispatcher.Dispatch(context.Background(), interaction, responder); err != nil {
		t.Fatalf("dispatch announcement config: %v", err)
	}
	if repo.AnnouncementChannels["guild-1"] != "channel-1" {
		t.Fatalf("announcement config = %#v", repo.AnnouncementChannels)
	}
	if len(responder.Edits) != 1 || len(responder.Edits[0].Embeds) != 1 || !strings.Contains(responder.Edits[0].Embeds[0].Description, "成功__創建__!!") {
		t.Fatalf("announcement response = %#v", responder.Edits)
	}
}

func TestBuildRuntimeRoutesAnnouncementSendOnlyWithRepositoryAndMessagePort(t *testing.T) {
	dispatcher, err := BuildRuntime(RuntimeOptions{Config: validTestConfig()})
	if err != nil {
		t.Fatalf("build runtime: %v", err)
	}
	interaction := fakediscord.SlashInteraction("公告發送")
	interaction.Actor.PermissionBits = 8192
	responder := fakediscord.NewResponder()
	if err := dispatcher.Dispatch(context.Background(), interaction, responder); err == nil {
		t.Fatal("announcement send route should not be available by default")
	}

	repo := fakemongo.NewAnnouncementConfigRepository()
	repo.AnnouncementChannels["guild-1"] = "announcement-channel"
	sideEffects := fakediscord.NewSideEffects()
	dispatcher, err = BuildRuntime(RuntimeOptions{
		Config:                     validTestConfig(),
		AnnouncementSendRepository: repo,
		AnnouncementMessagePort:    sideEffects,
	})
	if err != nil {
		t.Fatalf("build runtime with announcement send: %v", err)
	}
	responder = fakediscord.NewResponder()
	if err := dispatcher.Dispatch(context.Background(), interaction, responder); err != nil {
		t.Fatalf("dispatch announcement send slash: %v", err)
	}
	if len(responder.Modals) != 1 || responder.Modals[0].Title != "公告系統" {
		t.Fatalf("modal = %#v", responder.Modals)
	}
}

func TestBuildRuntimeRoutesTextXPConfigOnlyWithRepository(t *testing.T) {
	dispatcher, err := BuildRuntime(RuntimeOptions{Config: validTestConfig()})
	if err != nil {
		t.Fatalf("build runtime: %v", err)
	}
	interaction := fakediscord.SlashInteractionWithOptions("聊天經驗設定", "", map[string]string{
		"頻道": "channel-1",
		"訊息": "恭喜 {user} 升到 {level}",
		"顏色": "#00ff00",
	})
	interaction.ChannelID = "command-channel"
	interaction.Actor.PermissionBits = 8192
	responder := fakediscord.NewResponder()
	if err := dispatcher.Dispatch(context.Background(), interaction, responder); err == nil {
		t.Fatal("text XP config route should not be available without repository")
	}

	repo := fakemongo.NewTextXPConfigRepository()
	sideEffects := fakediscord.NewSideEffects()
	dispatcher, err = BuildRuntime(RuntimeOptions{
		Config:                 validTestConfig(),
		TextXPConfigRepository: repo,
		TextXPMessagePort:      sideEffects,
	})
	if err != nil {
		t.Fatalf("build runtime with text XP config repo: %v", err)
	}
	responder = fakediscord.NewResponder()
	if err := dispatcher.Dispatch(context.Background(), interaction, responder); err != nil {
		t.Fatalf("dispatch text XP config: %v", err)
	}
	saved, ok := repo.Configs["guild-1"]
	if !ok || saved.ChannelID != "channel-1" || saved.Color != "#00ff00" || saved.Message != "恭喜 {user} 升到 {level}" {
		t.Fatalf("saved text XP config = %#v ok=%v", saved, ok)
	}
	if len(responder.Edits) != 1 || len(responder.Edits[0].Embeds) != 1 || !strings.Contains(responder.Edits[0].Embeds[0].Description, "<#channel-1>") {
		t.Fatalf("text XP response = %#v", responder.Edits)
	}
	if len(sideEffects.Sent) != 1 || !strings.Contains(sideEffects.Sent[0].Message.Content, "以下為你的訊息預覽") {
		t.Fatalf("preview messages = %#v", sideEffects.Sent)
	}

	deleteInteraction := fakediscord.SlashInteraction("聊天經驗刪除")
	deleteInteraction.Actor.PermissionBits = 8192
	responder = fakediscord.NewResponder()
	if err := dispatcher.Dispatch(context.Background(), deleteInteraction, responder); err != nil {
		t.Fatalf("dispatch text XP delete: %v", err)
	}
	if _, ok := repo.Configs["guild-1"]; ok {
		t.Fatalf("text XP config was not deleted: %#v", repo.Configs)
	}
	if len(responder.Edits) != 1 || len(responder.Edits[0].Embeds) != 1 || !strings.Contains(responder.Edits[0].Embeds[0].Description, "成功刪除") {
		t.Fatalf("text XP delete response = %#v", responder.Edits)
	}
}

func TestBuildRuntimeRoutesVoiceXPConfigOnlyWithRepository(t *testing.T) {
	dispatcher, err := BuildRuntime(RuntimeOptions{Config: validTestConfig()})
	if err != nil {
		t.Fatalf("build runtime: %v", err)
	}
	interaction := fakediscord.SlashInteractionWithOptions("語音經驗設定", "", map[string]string{
		"頻道": "voice-channel-1",
		"訊息": "恭喜 {user} 升到 {level}",
		"顏色": "#00ff00",
		"背景": "https://example.invalid/background.png",
	})
	interaction.ChannelID = "command-channel"
	interaction.Actor.PermissionBits = 8192
	responder := fakediscord.NewResponder()
	if err := dispatcher.Dispatch(context.Background(), interaction, responder); err == nil {
		t.Fatal("voice XP config route should not be available without repository")
	}

	repo := fakemongo.NewVoiceXPConfigRepository()
	sideEffects := fakediscord.NewSideEffects()
	dispatcher, err = BuildRuntime(RuntimeOptions{
		Config:                  validTestConfig(),
		VoiceXPConfigRepository: repo,
		VoiceXPMessagePort:      sideEffects,
	})
	if err != nil {
		t.Fatalf("build runtime with voice XP config repo: %v", err)
	}
	responder = fakediscord.NewResponder()
	if err := dispatcher.Dispatch(context.Background(), interaction, responder); err != nil {
		t.Fatalf("dispatch voice XP config: %v", err)
	}
	saved, ok := repo.Configs["guild-1"]
	if !ok || saved.ChannelID != "voice-channel-1" || saved.Color != "#00ff00" || saved.Message != "恭喜 {user} 升到 {level}" {
		t.Fatalf("saved voice XP config = %#v ok=%v", saved, ok)
	}
	if len(responder.Edits) != 1 || len(responder.Edits[0].Embeds) != 1 || !strings.Contains(responder.Edits[0].Embeds[0].Description, "<#voice-channel-1>") {
		t.Fatalf("voice XP response = %#v", responder.Edits)
	}
	if len(sideEffects.Sent) != 1 || !strings.Contains(sideEffects.Sent[0].Message.Content, "以下為你的訊息預覽") {
		t.Fatalf("preview messages = %#v", sideEffects.Sent)
	}

	deleteInteraction := fakediscord.SlashInteraction("語音經驗刪除")
	deleteInteraction.Actor.PermissionBits = 8192
	responder = fakediscord.NewResponder()
	if err := dispatcher.Dispatch(context.Background(), deleteInteraction, responder); err != nil {
		t.Fatalf("dispatch voice XP delete: %v", err)
	}
	if _, ok := repo.Configs["guild-1"]; ok {
		t.Fatalf("voice XP config was not deleted: %#v", repo.Configs)
	}
	if len(responder.Edits) != 1 || len(responder.Edits[0].Embeds) != 1 || !strings.Contains(responder.Edits[0].Embeds[0].Description, "成功刪除") {
		t.Fatalf("voice XP delete response = %#v", responder.Edits)
	}
}

func TestBuildRuntimeRoutesJoinRoleConfigOnlyWithRepository(t *testing.T) {
	dispatcher, err := BuildRuntime(RuntimeOptions{Config: validTestConfig()})
	if err != nil {
		t.Fatalf("build runtime: %v", err)
	}
	interaction := fakediscord.SlashInteractionWithOptions("加入身份組設置", "", map[string]string{
		"身分組":      "role-1",
		"給人還是給機器人": domain.JoinRoleGiveMembers,
	})
	interaction.Actor.PermissionBits = 8192
	responder := fakediscord.NewResponder()
	if err := dispatcher.Dispatch(context.Background(), interaction, responder); err == nil {
		t.Fatal("join-role config route should not be available without repository")
	}

	repo := fakemongo.NewJoinRoleConfigRepository()
	sideEffects := fakediscord.NewSideEffects()
	sideEffects.AssignableRoles["guild-1/role-1"] = true
	dispatcher, err = BuildRuntime(RuntimeOptions{
		Config:                   validTestConfig(),
		JoinRoleConfigRepository: repo,
		JoinRoleInspector:        sideEffects,
	})
	if err != nil {
		t.Fatalf("build runtime with join-role config repo: %v", err)
	}
	responder = fakediscord.NewResponder()
	if err := dispatcher.Dispatch(context.Background(), interaction, responder); err != nil {
		t.Fatalf("dispatch join-role config: %v", err)
	}
	saved, ok := repo.Configs["guild-1/role-1"]
	if !ok || saved.RoleID != "role-1" || saved.GiveTo != domain.JoinRoleGiveMembers {
		t.Fatalf("saved join-role config = %#v ok=%v", saved, ok)
	}
	if len(responder.Edits) != 1 || len(responder.Edits[0].Embeds) != 1 || !strings.Contains(responder.Edits[0].Embeds[0].Description, "成功創建加入給身分組") {
		t.Fatalf("join-role response = %#v", responder.Edits)
	}

	deleteInteraction := fakediscord.SlashInteractionWithOptions("加入身份組刪除", "", map[string]string{"身分組": "role-1"})
	deleteInteraction.Actor.PermissionBits = 8192
	responder = fakediscord.NewResponder()
	if err := dispatcher.Dispatch(context.Background(), deleteInteraction, responder); err != nil {
		t.Fatalf("dispatch join-role delete: %v", err)
	}
	if _, ok := repo.Configs["guild-1/role-1"]; ok {
		t.Fatalf("join-role config was not deleted: %#v", repo.Configs)
	}
	if len(responder.Edits) != 1 || len(responder.Edits[0].Embeds) != 1 || !strings.Contains(responder.Edits[0].Embeds[0].Description, "成功刪除") {
		t.Fatalf("join-role delete response = %#v", responder.Edits)
	}
}

func TestBuildRuntimeRoutesVerificationConfigOnlyWithRepository(t *testing.T) {
	dispatcher, err := BuildRuntime(RuntimeOptions{Config: validTestConfig()})
	if err != nil {
		t.Fatalf("build runtime: %v", err)
	}
	interaction := fakediscord.SlashInteractionWithOptions("驗證設置", "", map[string]string{
		"身分組": "role-1",
		"改名":  "{name} | MHCAT",
	})
	interaction.Actor.PermissionBits = 8192
	responder := fakediscord.NewResponder()
	if err := dispatcher.Dispatch(context.Background(), interaction, responder); err == nil {
		t.Fatal("verification config route should not be available without repository")
	}

	repo := fakemongo.NewVerificationConfigRepository()
	sideEffects := fakediscord.NewSideEffects()
	sideEffects.AssignableRoles["guild-1/role-1"] = true
	dispatcher, err = BuildRuntime(RuntimeOptions{
		Config:                       validTestConfig(),
		VerificationConfigRepository: repo,
		VerificationRoleInspector:    sideEffects,
	})
	if err != nil {
		t.Fatalf("build runtime with verification config repo: %v", err)
	}
	responder = fakediscord.NewResponder()
	if err := dispatcher.Dispatch(context.Background(), interaction, responder); err != nil {
		t.Fatalf("dispatch verification config: %v", err)
	}
	saved, ok := repo.Configs["guild-1"]
	if !ok || saved.RoleID != "role-1" || saved.RenameTemplate != "{name} | MHCAT" {
		t.Fatalf("saved verification config = %#v ok=%v", saved, ok)
	}
	if len(responder.Edits) != 1 || len(responder.Edits[0].Embeds) != 1 || !strings.Contains(responder.Edits[0].Embeds[0].Title, "設置成功") {
		t.Fatalf("verification response = %#v", responder.Edits)
	}
}

func TestBuildRuntimeRoutesVerificationFlowOnlyWithRepository(t *testing.T) {
	dispatcher, err := BuildRuntime(RuntimeOptions{Config: validTestConfig()})
	if err != nil {
		t.Fatalf("build runtime: %v", err)
	}
	responder := fakediscord.NewResponder()
	if err := dispatcher.Dispatch(context.Background(), fakediscord.SlashInteraction("驗證"), responder); err == nil {
		t.Fatal("verification flow route should not be available without repository")
	}

	repo := fakemongo.NewVerificationConfigRepository()
	repo.Configs["guild-1"] = domain.VerificationConfig{GuildID: "guild-1", RoleID: "role-1"}
	dispatcher, err = BuildRuntime(RuntimeOptions{
		Config:                     validTestConfig(),
		VerificationFlowRepository: repo,
	})
	if err != nil {
		t.Fatalf("build runtime with verification repo only: %v", err)
	}
	responder = fakediscord.NewResponder()
	if err := dispatcher.Dispatch(context.Background(), fakediscord.SlashInteraction("驗證"), responder); err == nil {
		t.Fatal("verification flow route should not be available without role side-effect port")
	}

	sideEffects := fakediscord.NewSideEffects()
	sideEffects.AssignableRoles["guild-1/role-1"] = true
	dispatcher, err = BuildRuntime(RuntimeOptions{
		Config:                     validTestConfig(),
		VerificationFlowRepository: repo,
		VerificationRolePort:       sideEffects,
		VerificationMemberPort:     sideEffects,
		VerificationRoleInspector:  sideEffects,
		VerificationGuildInfo:      &fakebotinfo.DiscordInfoProvider{Guild: ports.DiscordGuildInfo{OwnerID: "owner-1"}},
	})
	if err != nil {
		t.Fatalf("build runtime with verification flow repo: %v", err)
	}
	responder = fakediscord.NewResponder()
	if err := dispatcher.Dispatch(context.Background(), fakediscord.SlashInteraction("驗證"), responder); err != nil {
		t.Fatalf("dispatch verification flow: %v", err)
	}
	if len(responder.Edits) != 1 || len(responder.Edits[0].Files) != 1 || len(responder.Edits[0].Components) != 1 {
		t.Fatalf("verification prompt = %#v", responder.Edits)
	}
}

func TestBuildRuntimeRoutesTranslateOnlyWithProvider(t *testing.T) {
	dispatcher, err := BuildRuntime(RuntimeOptions{Config: validTestConfig(), TranslateFeatureEnabled: true})
	if err != nil {
		t.Fatalf("build runtime with translate feature: %v", err)
	}
	responder := fakediscord.NewResponder()
	interaction := fakediscord.SlashInteractionWithOptions("翻譯", "", map[string]string{
		"要的翻譯": "你好",
		"目標語言": "en",
	})
	if err := dispatcher.Dispatch(context.Background(), interaction, responder); err == nil {
		t.Fatal("translate route should not be available without provider")
	}

	dispatcher, err = BuildRuntime(RuntimeOptions{
		Config:                  validTestConfig(),
		TranslateFeatureEnabled: true,
		TranslateProvider:       &faketranslate.Translator{Result: ports.TranslationResult{Text: "hello"}},
	})
	if err != nil {
		t.Fatalf("build runtime with translate provider: %v", err)
	}
	responder = fakediscord.NewResponder()
	if err := dispatcher.Dispatch(context.Background(), interaction, responder); err != nil {
		t.Fatalf("dispatch translate: %v", err)
	}
	if len(responder.Edits) != 2 || !strings.Contains(responder.Edits[1].Embeds[0].Title, "翻譯系統") {
		t.Fatalf("edits = %#v", responder.Edits)
	}
}

func timeNowForTest() time.Time {
	return time.Date(2026, 7, 4, 10, 30, 0, 0, time.FixedZone("Asia/Taipei", 8*60*60))
}

type appFixedClock struct {
	now time.Time
}

func (c appFixedClock) Now() time.Time {
	return c.now
}
