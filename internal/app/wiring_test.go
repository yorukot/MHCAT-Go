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
	repo.PutBalance(domain.CoinBalance{GuildID: "guild-1", UserID: "user-1", Today: 1})
	responder = fakediscord.NewResponder()
	if err := dispatcher.Dispatch(context.Background(), fakediscord.SlashInteraction("簽到列表"), responder); err != nil {
		t.Fatalf("dispatch sign-in list: %v", err)
	}
	if len(responder.Edits) != 1 || len(responder.Edits[0].Embeds) != 1 || !strings.Contains(responder.Edits[0].Embeds[0].Description, "`1`**人已經簽到") {
		t.Fatalf("sign-in list response = %#v", responder.Edits)
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

func TestBuildRuntimeRoutesEconomyCoinAdminWithoutPublishingQueryOrSignIn(t *testing.T) {
	repo := fakemongo.NewEconomyRepository()
	dispatcher, err := BuildRuntime(RuntimeOptions{
		Config:                     validTestConfig(),
		EconomyCoinAdminRepository: repo,
	})
	if err != nil {
		t.Fatalf("build runtime with economy coin-admin repo: %v", err)
	}
	for _, commandName := range []string{"代幣查詢", "簽到", "coin-related-settings"} {
		responder := fakediscord.NewResponder()
		if err := dispatcher.Dispatch(context.Background(), fakediscord.SlashInteraction(commandName), responder); err == nil {
			t.Fatalf("%s route should not be available with coin-admin repository alone", commandName)
		}
	}
	interaction := fakediscord.SlashInteractionWithOptions("代幣增加", "", map[string]string{
		"使用者":   "user-2",
		"增加或減少": "add",
		"數量":    "25",
	})
	interaction.Actor.PermissionBits = 8192
	responder := fakediscord.NewResponder()
	if err := dispatcher.Dispatch(context.Background(), interaction, responder); err != nil {
		t.Fatalf("dispatch coin admin: %v", err)
	}
	balance, err := repo.GetCoinBalance(context.Background(), "guild-1", "user-2")
	if err != nil || balance.Coins != 25 {
		t.Fatalf("balance=%#v err=%v", balance, err)
	}
}

func TestBuildRuntimeRoutesEconomyCoinRankWithoutPublishingOtherEconomyCommands(t *testing.T) {
	repo := fakemongo.NewEconomyRepository()
	viewerID := "123456789012345678"
	repo.PutBalance(domain.CoinBalance{GuildID: "guild-1", UserID: viewerID, Coins: 42})
	dispatcher, err := BuildRuntime(RuntimeOptions{
		Config:                    validTestConfig(),
		EconomyCoinRankRepository: repo,
	})
	if err != nil {
		t.Fatalf("build runtime with economy coin-rank repo: %v", err)
	}
	for _, commandName := range []string{"代幣查詢", "簽到", "coin-related-settings", "代幣增加"} {
		responder := fakediscord.NewResponder()
		if err := dispatcher.Dispatch(context.Background(), fakediscord.SlashInteraction(commandName), responder); err == nil {
			t.Fatalf("%s route should not be available with coin-rank repository alone", commandName)
		}
	}
	interaction := fakediscord.SlashInteraction("代幣排行榜")
	interaction.Actor.UserID = viewerID
	responder := fakediscord.NewResponder()
	if err := dispatcher.Dispatch(context.Background(), interaction, responder); err != nil {
		t.Fatalf("dispatch coin rank: %v", err)
	}
	if len(responder.Edits) != 1 || len(responder.Edits[0].Files) != 1 {
		t.Fatalf("coin rank response = %#v", responder.Edits)
	}
}

func TestBuildRuntimeRoutesEconomyRPSWithoutPublishingOtherEconomyCommands(t *testing.T) {
	repo := fakemongo.NewEconomyRepository()
	repo.PutBalance(domain.CoinBalance{GuildID: "guild-1", UserID: "user-1", Coins: 42})
	dispatcher, err := BuildRuntime(RuntimeOptions{
		Config:               validTestConfig(),
		EconomyRPSRepository: repo,
	})
	if err != nil {
		t.Fatalf("build runtime with economy RPS repo: %v", err)
	}
	for _, commandName := range []string{"代幣查詢", "簽到", "coin-related-settings", "代幣增加", "代幣排行榜", "代幣遊戲"} {
		responder := fakediscord.NewResponder()
		if err := dispatcher.Dispatch(context.Background(), fakediscord.SlashInteraction(commandName), responder); err == nil {
			t.Fatalf("%s route should not be available with RPS repository alone", commandName)
		}
	}
	interaction := fakediscord.SlashInteractionWithOptions("剪刀石頭布", "", map[string]string{
		"使用多少代幣來進行": "5",
		"剪刀石頭或布":    "剪刀",
	})
	responder := fakediscord.NewResponder()
	if err := dispatcher.Dispatch(context.Background(), interaction, responder); err != nil {
		t.Fatalf("dispatch economy RPS: %v", err)
	}
	if len(responder.Edits) != 1 || len(responder.Edits[0].Embeds) != 1 {
		t.Fatalf("RPS response = %#v", responder.Edits)
	}
}

func TestBuildRuntimeRoutesEconomyGameWithoutPublishingOtherEconomyCommands(t *testing.T) {
	repo := fakemongo.NewEconomyRepository()
	repo.PutBalance(domain.CoinBalance{GuildID: "guild-1", UserID: "user-1", Coins: 42})
	repo.PutBalance(domain.CoinBalance{GuildID: "guild-1", UserID: "user-2", Coins: 42})
	dispatcher, err := BuildRuntime(RuntimeOptions{
		Config:                validTestConfig(),
		EconomyGameRepository: repo,
	})
	if err != nil {
		t.Fatalf("build runtime with economy game repo: %v", err)
	}
	for _, commandName := range []string{"代幣查詢", "簽到", "coin-related-settings", "代幣增加", "代幣排行榜", "剪刀石頭布"} {
		responder := fakediscord.NewResponder()
		if err := dispatcher.Dispatch(context.Background(), fakediscord.SlashInteraction(commandName), responder); err == nil {
			t.Fatalf("%s route should not be available with economy game repository alone", commandName)
		}
	}
	interaction := fakediscord.SlashInteractionWithOptions("代幣遊戲", "比大小", map[string]string{
		"跟誰玩": "user-2",
		"賭注":  "5",
	})
	responder := fakediscord.NewResponder()
	if err := dispatcher.Dispatch(context.Background(), interaction, responder); err != nil {
		t.Fatalf("dispatch economy game: %v", err)
	}
	if len(responder.Follow) != 1 || len(responder.Follow[0].Components) != 1 {
		t.Fatalf("game response = %#v", responder.Follow)
	}
}

func TestBuildRuntimeRoutesEconomyShopWithoutPublishingOtherEconomyCommands(t *testing.T) {
	repo := fakemongo.NewEconomyRepository()
	repo.PutShopItem(domain.ShopItem{GuildID: "guild-1", CommodityID: 1001, Name: "VIP", NeedCoins: 50, Description: "role reward", Count: 1})
	dispatcher, err := BuildRuntime(RuntimeOptions{
		Config:                validTestConfig(),
		EconomyShopRepository: repo,
	})
	if err != nil {
		t.Fatalf("build runtime with economy shop repo: %v", err)
	}
	for _, commandName := range []string{"代幣查詢", "簽到", "coin-related-settings", "代幣增加", "代幣排行榜", "剪刀石頭布", "代幣遊戲", "my-profile"} {
		responder := fakediscord.NewResponder()
		if err := dispatcher.Dispatch(context.Background(), fakediscord.SlashInteraction(commandName), responder); err == nil {
			t.Fatalf("%s route should not be available with shop repository alone", commandName)
		}
	}
	interaction := fakediscord.SlashInteractionWithOptions("代幣商店", "商品查詢", nil)
	responder := fakediscord.NewResponder()
	if err := dispatcher.Dispatch(context.Background(), interaction, responder); err != nil {
		t.Fatalf("dispatch economy shop: %v", err)
	}
	if len(responder.Edits) != 1 || len(responder.Edits[0].Embeds) != 1 || !strings.Contains(responder.Edits[0].Embeds[0].Fields[0].Value, "**商品id:**`1001`") {
		t.Fatalf("shop response = %#v", responder.Edits)
	}
}

func TestBuildRuntimeRoutesEconomyProfileWithoutPublishingOtherEconomyCommands(t *testing.T) {
	repo := fakemongo.NewEconomyProfileRepository()
	userID := "123456789012345678"
	repo.PutBalance(domain.CoinBalance{GuildID: "guild-1", UserID: userID, Coins: 42})
	dispatcher, err := BuildRuntime(RuntimeOptions{
		Config:                   validTestConfig(),
		EconomyProfileRepository: repo,
		Clock:                    appFixedClock{now: time.Unix(1_000, 0)},
	})
	if err != nil {
		t.Fatalf("build runtime with economy profile repo: %v", err)
	}
	for _, commandName := range []string{"代幣查詢", "簽到", "coin-related-settings", "代幣增加", "代幣排行榜", "剪刀石頭布", "代幣遊戲", "代幣商店"} {
		responder := fakediscord.NewResponder()
		if err := dispatcher.Dispatch(context.Background(), fakediscord.SlashInteraction(commandName), responder); err == nil {
			t.Fatalf("%s route should not be available with profile repository alone", commandName)
		}
	}
	interaction := fakediscord.SlashInteraction("my-profile")
	interaction.Actor.UserID = userID
	responder := fakediscord.NewResponder()
	if err := dispatcher.Dispatch(context.Background(), interaction, responder); err != nil {
		t.Fatalf("dispatch economy profile: %v", err)
	}
	if len(responder.Edits) != 1 || len(responder.Edits[0].Files) != 1 {
		t.Fatalf("profile response = %#v", responder.Edits)
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

func TestBuildRuntimeRoutesWarningSettingsOnlyWithRepository(t *testing.T) {
	dispatcher, err := BuildRuntime(RuntimeOptions{Config: validTestConfig(), WarningSettingsFeatureEnabled: true})
	if err != nil {
		t.Fatalf("build runtime with warning settings feature: %v", err)
	}
	interaction := fakediscord.SlashInteractionWithOptions("警告設定", "", map[string]string{
		"執行的動作":     "停權",
		"幾次警告後執行動作": "3",
	})
	interaction.Actor.PermissionBits = 8192
	responder := fakediscord.NewResponder()
	if err := dispatcher.Dispatch(context.Background(), interaction, responder); err == nil {
		t.Fatal("warning settings route should not be available without repository")
	}

	repo := fakemongo.NewWarningSettingsRepository()
	dispatcher, err = BuildRuntime(RuntimeOptions{
		Config:                        validTestConfig(),
		WarningSettingsFeatureEnabled: true,
		WarningSettingsRepository:     repo,
	})
	if err != nil {
		t.Fatalf("build runtime with warning settings repo: %v", err)
	}
	responder = fakediscord.NewResponder()
	if err := dispatcher.Dispatch(context.Background(), interaction, responder); err != nil {
		t.Fatalf("dispatch warning settings: %v", err)
	}
	if len(responder.Edits) != 1 || responder.Edits[0].Embeds[0].Description != "警告成功設為警告3次後\n執行停權" {
		t.Fatalf("warning settings response = %#v", responder.Edits)
	}
	if got := repo.Settings["guild-1"]; got.Threshold != 3 || got.Action != "停權" {
		t.Fatalf("saved warning settings = %#v", got)
	}
}

func TestBuildRuntimeRoutesWarningRemovalOnlyWithRepository(t *testing.T) {
	dispatcher, err := BuildRuntime(RuntimeOptions{Config: validTestConfig(), WarningRemovalFeatureEnabled: true})
	if err != nil {
		t.Fatalf("build runtime with warning removal feature: %v", err)
	}
	interaction := fakediscord.SlashInteractionWithOptions("警告清除", "", map[string]string{
		"使用者": "user-2",
		"第幾項": "1",
	})
	interaction.Actor.PermissionBits = 8192
	responder := fakediscord.NewResponder()
	if err := dispatcher.Dispatch(context.Background(), interaction, responder); err == nil {
		t.Fatal("warning removal route should not be available without repository")
	}

	repo := fakemongo.NewWarningRemovalRepository()
	repo.Put(domain.WarningHistory{GuildID: "guild-1", UserID: "user-2", Entries: []domain.WarningEntry{{Reason: "first"}}})
	sideEffects := fakediscord.NewSideEffects()
	dispatcher, err = BuildRuntime(RuntimeOptions{
		Config:                       validTestConfig(),
		WarningRemovalFeatureEnabled: true,
		WarningRemovalRepository:     repo,
		WarningRemovalDirectMessage:  sideEffects,
	})
	if err != nil {
		t.Fatalf("build runtime with warning removal repo: %v", err)
	}
	responder = fakediscord.NewResponder()
	if err := dispatcher.Dispatch(context.Background(), interaction, responder); err != nil {
		t.Fatalf("dispatch warning removal: %v", err)
	}
	if len(responder.Edits) != 1 || responder.Edits[0].Embeds[0].Title != "<a:greentick:980496858445135893> | 這位使用者的警告成功移除!" {
		t.Fatalf("warning removal response = %#v", responder.Edits)
	}
	if len(repo.Histories["guild-1\x00user-2"].Entries) != 0 {
		t.Fatalf("saved warning history = %#v", repo.Histories)
	}
}

func TestBuildRuntimeRoutesWarningIssueOnlyWithRepository(t *testing.T) {
	dispatcher, err := BuildRuntime(RuntimeOptions{Config: validTestConfig(), WarningIssueFeatureEnabled: true})
	if err != nil {
		t.Fatalf("build runtime with warning issue feature: %v", err)
	}
	interaction := fakediscord.SlashInteractionWithOptions("警告", "", map[string]string{
		"使用者": "user-2",
		"原因":  "洗版",
	})
	interaction.Actor.PermissionBits = 8192
	responder := fakediscord.NewResponder()
	if err := dispatcher.Dispatch(context.Background(), interaction, responder); err == nil {
		t.Fatal("warning issue route should not be available without repository")
	}

	repo := fakemongo.NewWarningHistoryRepository()
	settings := fakemongo.NewWarningSettingsRepository()
	sideEffects := fakediscord.NewSideEffects()
	dispatcher, err = BuildRuntime(RuntimeOptions{
		Config:                     validTestConfig(),
		WarningIssueFeatureEnabled: true,
		WarningIssueRepository:     repo,
		WarningSettingsRepository:  settings,
		WarningIssueDirectMessage:  sideEffects,
		WarningIssueHierarchy:      sideEffects,
		WarningIssueMemberPort:     sideEffects,
		WarningIssueMessagePort:    sideEffects,
		Clock:                      appFixedClock{now: time.Date(2026, 7, 4, 10, 30, 0, 0, time.UTC)},
	})
	if err != nil {
		t.Fatalf("build runtime with warning issue repo: %v", err)
	}
	responder = fakediscord.NewResponder()
	if err := dispatcher.Dispatch(context.Background(), interaction, responder); err != nil {
		t.Fatalf("dispatch warning issue: %v", err)
	}
	if len(responder.Edits) != 1 || responder.Edits[0].Embeds[0].Title != "<a:greentick:980496858445135893> | 成功警告這位使用者!" {
		t.Fatalf("warning issue response = %#v", responder.Edits)
	}
	history := repo.Histories["guild-1\x00user-2"]
	if len(history.Entries) != 1 || history.Entries[0].Reason != "洗版" || history.Entries[0].Time != "2026年07月04日 18點30分" {
		t.Fatalf("saved warning history = %#v", history)
	}
}

func TestBuildRuntimeRoutesMessageCleanupOnlyWithCleaner(t *testing.T) {
	dispatcher, err := BuildRuntime(RuntimeOptions{Config: validTestConfig(), MessageCleanupFeatureEnabled: true})
	if err != nil {
		t.Fatalf("build runtime with message cleanup feature: %v", err)
	}
	interaction := fakediscord.SlashInteractionWithOptions("刪除訊息", "", map[string]string{"刪除數量": "5"})
	interaction.Actor.PermissionBits = 8192
	interaction.ChannelID = "channel-1"
	responder := fakediscord.NewResponder()
	if err := dispatcher.Dispatch(context.Background(), interaction, responder); err == nil {
		t.Fatal("message cleanup route should not be available without cleaner")
	}

	sideEffects := fakediscord.NewSideEffects()
	sideEffects.CleanupDeleted = 3
	dispatcher, err = BuildRuntime(RuntimeOptions{
		Config:                       validTestConfig(),
		MessageCleanupFeatureEnabled: true,
		MessageCleaner:               sideEffects,
	})
	if err != nil {
		t.Fatalf("build runtime with message cleanup cleaner: %v", err)
	}
	responder = fakediscord.NewResponder()
	if err := dispatcher.Dispatch(context.Background(), interaction, responder); err != nil {
		t.Fatalf("dispatch message cleanup: %v", err)
	}
	if len(responder.Edits) != 1 || !strings.Contains(responder.Edits[0].Embeds[0].Description, "`3`/`5`") {
		t.Fatalf("cleanup response = %#v", responder.Edits)
	}
	if len(sideEffects.CleanupRequests) != 1 || sideEffects.CleanupRequests[0].ChannelID != "channel-1" || sideEffects.CleanupRequests[0].Limit != 5 {
		t.Fatalf("cleanup requests = %#v", sideEffects.CleanupRequests)
	}
}

func TestBuildRuntimeRoutesDeleteDataOnlyWithRepository(t *testing.T) {
	dispatcher, err := BuildRuntime(RuntimeOptions{Config: validTestConfig(), DeleteDataFeatureEnabled: true})
	if err != nil {
		t.Fatalf("build runtime with delete data feature: %v", err)
	}
	interaction := fakediscord.SlashInteraction("刪除資料")
	interaction.Actor.PermissionBits = 8192
	responder := fakediscord.NewResponder()
	if err := dispatcher.Dispatch(context.Background(), interaction, responder); err == nil {
		t.Fatal("delete data route should not be available without repository")
	}

	repo := fakemongo.NewDeleteDataRepository()
	repo.Put("guild-1", domain.DeleteDataTargetAutoChat)
	dispatcher, err = BuildRuntime(RuntimeOptions{
		Config:                   validTestConfig(),
		DeleteDataFeatureEnabled: true,
		DeleteDataRepository:     repo,
	})
	if err != nil {
		t.Fatalf("build runtime with delete data repo: %v", err)
	}
	responder = fakediscord.NewResponder()
	if err := dispatcher.Dispatch(context.Background(), interaction, responder); err != nil {
		t.Fatalf("dispatch delete data slash: %v", err)
	}
	if len(responder.Follow) != 1 || responder.Follow[0].Embeds[0].Title != "<:trashbin:995991389043163257> 刪除資料" {
		t.Fatalf("delete data prompt = %#v", responder.Follow)
	}

	component := fakediscord.ComponentInteractionFromID("delete-data")
	component.Actor.PermissionBits = 8192
	component.Values = []string{string(domain.DeleteDataTargetAutoChat)}
	responder = fakediscord.NewResponder()
	if err := dispatcher.Dispatch(context.Background(), component, responder); err != nil {
		t.Fatalf("dispatch delete data component: %v", err)
	}
	if len(repo.Deleted) != 1 || repo.Deleted[0].Target != domain.DeleteDataTargetAutoChat {
		t.Fatalf("deleted = %#v", repo.Deleted)
	}
	if len(responder.Edits) != 1 || !strings.Contains(responder.Edits[0].Content, "成功刪除該設定") {
		t.Fatalf("delete data response = %#v", responder.Edits)
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

func TestBuildRuntimeRoutesGachaDrawOnlyWithRepository(t *testing.T) {
	dispatcher, err := BuildRuntime(RuntimeOptions{Config: validTestConfig()})
	if err != nil {
		t.Fatalf("build runtime: %v", err)
	}
	interaction := fakediscord.SlashInteraction("扭蛋")
	responder := fakediscord.NewResponder()
	if err := dispatcher.Dispatch(context.Background(), interaction, responder); err == nil {
		t.Fatal("gacha draw route should not be available without repository")
	}

	repo := fakemongo.NewGachaRepository()
	repo.Balances["guild-1/user-1"] = domain.CoinBalance{GuildID: "guild-1", UserID: "user-1", Coins: 1000}
	repo.Prizes["guild-1"] = []domain.GachaPrize{{GuildID: "guild-1", Name: "大獎", Chance: 100, Count: 1}}
	repo.PrizeConfigs["guild-1"] = []domain.GachaPrizeConfig{{GuildID: "guild-1", Name: "大獎", Chance: 100, AutoDelete: true, Count: 1}}
	dispatcher, err = BuildRuntime(RuntimeOptions{
		Config:              validTestConfig(),
		GachaDrawRepository: repo,
	})
	if err != nil {
		t.Fatalf("build runtime with gacha draw repo: %v", err)
	}
	responder = fakediscord.NewResponder()
	if err := dispatcher.Dispatch(context.Background(), interaction, responder); err != nil {
		t.Fatalf("dispatch gacha draw: %v", err)
	}
	if len(responder.Edits) != 2 || len(responder.Edits[1].Embeds) != 1 || !strings.Contains(responder.Edits[1].Embeds[0].Title, "扭蛋系統") {
		t.Fatalf("gacha draw response = %#v", responder.Edits)
	}

	responder = fakediscord.NewResponder()
	listInteraction := fakediscord.SlashInteraction("扭蛋獎池查詢")
	if err := dispatcher.Dispatch(context.Background(), listInteraction, responder); err == nil {
		t.Fatal("gacha prize-list route should not be available with draw-only repository")
	}
}

func TestBuildRuntimeRoutesGachaPrizeDeleteOnlyWithRepository(t *testing.T) {
	dispatcher, err := BuildRuntime(RuntimeOptions{Config: validTestConfig()})
	if err != nil {
		t.Fatalf("build runtime: %v", err)
	}
	interaction := fakediscord.SlashInteractionWithOptions("扭蛋獎池刪除", "", map[string]string{"獎品名稱": "大獎"})
	interaction.Actor.PermissionBits = 8192
	responder := fakediscord.NewResponder()
	if err := dispatcher.Dispatch(context.Background(), interaction, responder); err == nil {
		t.Fatal("gacha prize-delete route should not be available without repository")
	}

	repo := fakemongo.NewGachaRepository()
	repo.Prizes["guild-1"] = []domain.GachaPrize{{GuildID: "guild-1", Name: "大獎", Chance: 10, Count: 1}}
	dispatcher, err = BuildRuntime(RuntimeOptions{
		Config:                     validTestConfig(),
		GachaPrizeDeleteRepository: repo,
	})
	if err != nil {
		t.Fatalf("build runtime with gacha prize-delete repo: %v", err)
	}
	responder = fakediscord.NewResponder()
	if err := dispatcher.Dispatch(context.Background(), interaction, responder); err != nil {
		t.Fatalf("dispatch gacha prize-delete: %v", err)
	}
	if len(responder.Edits) != 1 || len(responder.Edits[0].Embeds) != 1 || !strings.Contains(responder.Edits[0].Embeds[0].Title, "成功刪除") {
		t.Fatalf("gacha prize-delete response = %#v", responder.Edits)
	}
	if len(repo.Prizes["guild-1"]) != 0 {
		t.Fatalf("prize should be deleted: %#v", repo.Prizes["guild-1"])
	}

	responder = fakediscord.NewResponder()
	listInteraction := fakediscord.SlashInteraction("扭蛋獎池查詢")
	if err := dispatcher.Dispatch(context.Background(), listInteraction, responder); err == nil {
		t.Fatal("gacha prize-list route should not be available with delete-only repository")
	}
}

func TestBuildRuntimeRoutesGachaPrizeCreateOnlyWithRepository(t *testing.T) {
	dispatcher, err := BuildRuntime(RuntimeOptions{Config: validTestConfig()})
	if err != nil {
		t.Fatalf("build runtime: %v", err)
	}
	interaction := fakediscord.SlashInteractionWithOptions("扭蛋獎池增加", "", map[string]string{
		"獎品名稱": "大獎",
		"機率":   "10",
	})
	interaction.Actor.PermissionBits = 8192
	responder := fakediscord.NewResponder()
	if err := dispatcher.Dispatch(context.Background(), interaction, responder); err == nil {
		t.Fatal("gacha prize-create route should not be available without repository")
	}

	repo := fakemongo.NewGachaRepository()
	dispatcher, err = BuildRuntime(RuntimeOptions{
		Config:                     validTestConfig(),
		GachaPrizeCreateRepository: repo,
	})
	if err != nil {
		t.Fatalf("build runtime with gacha prize-create repo: %v", err)
	}
	responder = fakediscord.NewResponder()
	if err := dispatcher.Dispatch(context.Background(), interaction, responder); err != nil {
		t.Fatalf("dispatch gacha prize-create: %v", err)
	}
	if len(responder.Edits) != 1 || len(responder.Edits[0].Embeds) != 1 || !strings.Contains(responder.Edits[0].Embeds[0].Title, "設置成功") {
		t.Fatalf("gacha prize-create response = %#v", responder.Edits)
	}
	if len(repo.Prizes["guild-1"]) != 1 || repo.Prizes["guild-1"][0].Name != "大獎" {
		t.Fatalf("prize should be created: %#v", repo.Prizes["guild-1"])
	}

	responder = fakediscord.NewResponder()
	listInteraction := fakediscord.SlashInteraction("扭蛋獎池查詢")
	if err := dispatcher.Dispatch(context.Background(), listInteraction, responder); err == nil {
		t.Fatal("gacha prize-list route should not be available with create-only repository")
	}
}

func TestBuildRuntimeRoutesGachaPrizeEditOnlyWithRepository(t *testing.T) {
	dispatcher, err := BuildRuntime(RuntimeOptions{Config: validTestConfig()})
	if err != nil {
		t.Fatalf("build runtime: %v", err)
	}
	interaction := fakediscord.SlashInteractionWithOptions("扭蛋獎品編輯", "", map[string]string{
		"獎品名稱": "大獎",
		"機率":   "12.5",
	})
	interaction.Actor.PermissionBits = 8192
	responder := fakediscord.NewResponder()
	if err := dispatcher.Dispatch(context.Background(), interaction, responder); err == nil {
		t.Fatal("gacha prize-edit route should not be available without repository")
	}

	repo := fakemongo.NewGachaRepository()
	repo.Prizes["guild-1"] = []domain.GachaPrize{{GuildID: "guild-1", Name: "大獎", Chance: 10, Count: 2}}
	dispatcher, err = BuildRuntime(RuntimeOptions{
		Config:                   validTestConfig(),
		GachaPrizeEditRepository: repo,
	})
	if err != nil {
		t.Fatalf("build runtime with gacha prize-edit repo: %v", err)
	}
	responder = fakediscord.NewResponder()
	if err := dispatcher.Dispatch(context.Background(), interaction, responder); err != nil {
		t.Fatalf("dispatch gacha prize-edit: %v", err)
	}
	if len(responder.Edits) != 1 || len(responder.Edits[0].Embeds) != 1 || !strings.Contains(responder.Edits[0].Embeds[0].Title, "編輯成功成功") {
		t.Fatalf("gacha prize-edit response = %#v", responder.Edits)
	}
	if len(repo.Prizes["guild-1"]) != 1 || repo.Prizes["guild-1"][0].Chance != 12.5 || repo.Prizes["guild-1"][0].Count != 1 {
		t.Fatalf("prize should be edited: %#v", repo.Prizes["guild-1"])
	}

	responder = fakediscord.NewResponder()
	listInteraction := fakediscord.SlashInteraction("扭蛋獎池查詢")
	if err := dispatcher.Dispatch(context.Background(), listInteraction, responder); err == nil {
		t.Fatal("gacha prize-list route should not be available with edit-only repository")
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

func TestBuildRuntimeRoutesLotteryComponentsSeparatelyFromDisabledCommand(t *testing.T) {
	const id = "1700000000000999lotter"
	now := time.Unix(1_700_000_000, 0)
	repo := fakemongo.NewLotteryRepository()
	repo.Lotteries["guild-1:"+id] = domain.Lottery{GuildID: "guild-1", ID: id, EndsAtUnix: now.Add(time.Hour).Unix()}
	sideEffects := fakediscord.NewSideEffects()
	dispatcher, err := BuildRuntime(RuntimeOptions{
		Config:                   validTestConfig(),
		LotteryComponentsEnabled: true,
		LotteryRepository:        repo,
		LotteryMemberReader:      sideEffects,
		LotteryMessagePort:       sideEffects,
		Clock:                    appFixedClock{now: now},
	})
	if err != nil {
		t.Fatalf("build runtime: %v", err)
	}
	responder := fakediscord.NewResponder()
	if err := dispatcher.Dispatch(context.Background(), fakediscord.ComponentInteractionFromID(id), responder); err != nil {
		t.Fatalf("dispatch lottery enter: %v", err)
	}
	if len(responder.Edits) != 1 || len(responder.Edits[0].Embeds) != 1 || !strings.Contains(responder.Edits[0].Embeds[0].Title, "成功參加抽獎") {
		t.Fatalf("edits = %#v", responder.Edits)
	}
	if len(repo.Lotteries["guild-1:"+id].Participants) != 1 {
		t.Fatalf("lottery = %#v", repo.Lotteries["guild-1:"+id])
	}
	if err := dispatcher.Dispatch(context.Background(), fakediscord.SlashInteraction("抽獎設置"), fakediscord.NewResponder()); err == nil {
		t.Fatal("component gate must not enable the disabled slash command route")
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

func TestBuildRuntimeRoutesStatsDeleteOnlyWithRepository(t *testing.T) {
	dispatcher, err := BuildRuntime(RuntimeOptions{Config: validTestConfig()})
	if err != nil {
		t.Fatalf("build runtime: %v", err)
	}
	interaction := fakediscord.SlashInteraction("統計系統刪除")
	interaction.Actor.PermissionBits = 8192
	responder := fakediscord.NewResponder()
	if err := dispatcher.Dispatch(context.Background(), interaction, responder); err == nil {
		t.Fatal("stats delete route should not be available without repository")
	}

	repo := fakemongo.NewStatsConfigRepository()
	repo.Put(domain.StatsConfig{GuildID: "guild-1", ParentID: "parent-1"})
	dispatcher, err = BuildRuntime(RuntimeOptions{
		Config:                validTestConfig(),
		StatsDeleteRepository: repo,
	})
	if err != nil {
		t.Fatalf("build runtime with stats delete: %v", err)
	}
	responder = fakediscord.NewResponder()
	if err := dispatcher.Dispatch(context.Background(), interaction, responder); err != nil {
		t.Fatalf("dispatch stats delete: %v", err)
	}
	if len(responder.Edits) != 1 || len(responder.Edits[0].Embeds) != 1 || !strings.Contains(responder.Edits[0].Embeds[0].Title, "成功刪除") {
		t.Fatalf("stats delete response = %#v", responder.Edits)
	}
}

func TestBuildRuntimeRoutesStatsCreateOnlyWithDependencies(t *testing.T) {
	dispatcher, err := BuildRuntime(RuntimeOptions{Config: validTestConfig()})
	if err != nil {
		t.Fatalf("build runtime: %v", err)
	}
	interaction := fakediscord.SlashInteractionWithOptions("統計系統創建", "", map[string]string{"統計頻道類型": "文字頻道"})
	interaction.Actor.PermissionBits = 8192
	responder := fakediscord.NewResponder()
	if err := dispatcher.Dispatch(context.Background(), interaction, responder); err == nil {
		t.Fatal("stats create route should not be available without dependencies")
	}

	repo := fakemongo.NewStatsConfigRepository()
	sideEffects := fakediscord.NewSideEffects()
	sideEffects.TotalMembers = 11
	sideEffects.NonBotMembers = 9
	dispatcher, err = BuildRuntime(RuntimeOptions{
		Config:                 validTestConfig(),
		StatsCreateRepository:  repo,
		StatsCreateChannelPort: sideEffects,
		StatsCreateGuildStats:  sideEffects,
		BotUserID:              "bot-1",
	})
	if err != nil {
		t.Fatalf("build runtime with stats create: %v", err)
	}
	responder = fakediscord.NewResponder()
	if err := dispatcher.Dispatch(context.Background(), interaction, responder); err != nil {
		t.Fatalf("dispatch stats create: %v", err)
	}
	if len(responder.Follow) != 1 || len(responder.Follow[0].Embeds) != 1 || !strings.Contains(responder.Follow[0].Embeds[0].Title, "成功創建") {
		t.Fatalf("stats create response = %#v", responder.Follow)
	}
	if len(sideEffects.Created) != 4 || repo.Configs["guild-1"].MemberNumberName != "11" {
		t.Fatalf("created=%#v repo=%#v", sideEffects.Created, repo.Configs)
	}
}

func TestBuildRuntimeRoutesStatsRoleOnlyWithDependencies(t *testing.T) {
	dispatcher, err := BuildRuntime(RuntimeOptions{Config: validTestConfig()})
	if err != nil {
		t.Fatalf("build runtime: %v", err)
	}
	interaction := fakediscord.SlashInteractionWithOptions("統計身分組人數", "", map[string]string{"統計頻道類型": "文字頻道", "身分組": "role-1"})
	interaction.Actor.PermissionBits = 8192
	responder := fakediscord.NewResponder()
	if err := dispatcher.Dispatch(context.Background(), interaction, responder); err == nil {
		t.Fatal("stats role route should not be available without dependencies")
	}

	repo := fakemongo.NewStatsConfigRepository()
	repo.Put(domain.StatsConfig{GuildID: "guild-1", ParentID: "parent-1"})
	sideEffects := fakediscord.NewSideEffects()
	sideEffects.Channels = append(sideEffects.Channels, ports.ChannelRef{GuildID: "guild-1", ChannelID: "parent-1", Name: "stats", Type: 4})
	sideEffects.RoleNames["guild-1/role-1"] = "VIP"
	sideEffects.RoleMemberCounts["guild-1/role-1"] = 5
	dispatcher, err = BuildRuntime(RuntimeOptions{
		Config:                    validTestConfig(),
		StatsRoleStatsRepository:  repo,
		StatsRoleConfigRepository: repo,
		StatsRoleChannelPort:      sideEffects,
		StatsRoleStatsReader:      sideEffects,
		BotUserID:                 "bot-1",
	})
	if err != nil {
		t.Fatalf("build runtime with stats role: %v", err)
	}
	responder = fakediscord.NewResponder()
	if err := dispatcher.Dispatch(context.Background(), interaction, responder); err != nil {
		t.Fatalf("dispatch stats role: %v", err)
	}
	if len(responder.Edits) != 1 || len(responder.Edits[0].Embeds) != 1 || !strings.Contains(responder.Edits[0].Embeds[0].Title, "統計特定身分組成功創建") {
		t.Fatalf("stats role response = %#v", responder.Edits)
	}
	if len(sideEffects.Created) != 1 || repo.RoleConfigs["guild-1/role-1"].ChannelName != "5" {
		t.Fatalf("created=%#v repo=%#v", sideEffects.Created, repo.RoleConfigs)
	}
}

func TestBuildRuntimeRoutesXPRoleConfigOnlyWithRepositories(t *testing.T) {
	dispatcher, err := BuildRuntime(RuntimeOptions{Config: validTestConfig()})
	if err != nil {
		t.Fatalf("build runtime: %v", err)
	}
	interaction := fakediscord.SlashInteractionWithOptions("聊天經驗身分組設定", "設定查詢", nil)
	interaction.Actor.PermissionBits = 8192
	responder := fakediscord.NewResponder()
	if err := dispatcher.Dispatch(context.Background(), interaction, responder); err == nil {
		t.Fatal("XP role config route should not be available without repositories")
	}

	textRepo := fakemongo.NewTextXPRewardRoleRepository()
	textRepo.Configs = []domain.XPRewardRoleConfig{{GuildID: "guild-1", Level: 5, RoleID: "role-1"}}
	voiceRepo := fakemongo.NewVoiceXPRewardRoleRepository()
	dispatcher, err = BuildRuntime(RuntimeOptions{
		Config:                      validTestConfig(),
		TextXPRewardRoleRepository:  textRepo,
		VoiceXPRewardRoleRepository: voiceRepo,
	})
	if err != nil {
		t.Fatalf("build runtime with XP role config: %v", err)
	}
	responder = fakediscord.NewResponder()
	if err := dispatcher.Dispatch(context.Background(), interaction, responder); err != nil {
		t.Fatalf("dispatch XP role config: %v", err)
	}
	if len(responder.Edits) != 1 || len(responder.Edits[0].Embeds) != 1 || !strings.Contains(responder.Edits[0].Embeds[0].Title, "聊天經驗身分組") {
		t.Fatalf("XP role config response = %#v", responder.Edits)
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

func TestBuildRuntimeRoutesAutoNotificationConfigOnlyWithRepository(t *testing.T) {
	dispatcher, err := BuildRuntime(RuntimeOptions{Config: validTestConfig()})
	if err != nil {
		t.Fatalf("build runtime: %v", err)
	}
	interaction := fakediscord.SlashInteraction("自動通知列表")
	interaction.Actor.PermissionBits = 8192
	responder := fakediscord.NewResponder()
	if err := dispatcher.Dispatch(context.Background(), interaction, responder); err == nil {
		t.Fatal("auto-notification config route should not be available without repository")
	}

	repo := fakemongo.NewAutoNotificationScheduleRepository()
	repo.Schedules["guild-1"] = []domain.AutoNotificationSchedule{{GuildID: "guild-1", ID: "schedule-1", Cron: "0 9 * * 1", ChannelID: "channel-1"}}
	dispatcher, err = BuildRuntime(RuntimeOptions{
		Config:                     validTestConfig(),
		AutoNotificationRepository: repo,
	})
	if err != nil {
		t.Fatalf("build runtime with auto-notification repo: %v", err)
	}
	responder = fakediscord.NewResponder()
	if err := dispatcher.Dispatch(context.Background(), interaction, responder); err != nil {
		t.Fatalf("dispatch auto-notification list: %v", err)
	}
	if len(responder.Replies) != 1 || !strings.Contains(responder.Replies[0].Embeds[0].Description, "schedule-1") {
		t.Fatalf("replies = %#v", responder.Replies)
	}
}

func TestBuildRuntimeRoutesBalanceQueryOnlyWithRepository(t *testing.T) {
	dispatcher, err := BuildRuntime(RuntimeOptions{Config: validTestConfig()})
	if err != nil {
		t.Fatalf("build runtime: %v", err)
	}
	interaction := fakediscord.SlashInteraction("查看餘額")
	responder := fakediscord.NewResponder()
	if err := dispatcher.Dispatch(context.Background(), interaction, responder); err == nil {
		t.Fatal("balance query route should not be available without repository")
	}

	repo := fakemongo.NewBalanceRepository()
	repo.Balances["guild-1"] = domain.Balance{GuildID: "guild-1", Amount: "18"}
	dispatcher, err = BuildRuntime(RuntimeOptions{
		Config:            validTestConfig(),
		BalanceRepository: repo,
	})
	if err != nil {
		t.Fatalf("build runtime with balance repo: %v", err)
	}
	responder = fakediscord.NewResponder()
	if err := dispatcher.Dispatch(context.Background(), interaction, responder); err != nil {
		t.Fatalf("dispatch balance query: %v", err)
	}
	if len(responder.Edits) != 1 || len(responder.Edits[0].Embeds) != 1 || responder.Edits[0].Embeds[0].Author == nil || !strings.Contains(responder.Edits[0].Embeds[0].Author.Name, "18") {
		t.Fatalf("balance query response = %#v", responder.Edits)
	}
}

func TestBuildRuntimeRoutesRedeemOnlyWithRepository(t *testing.T) {
	dispatcher, err := BuildRuntime(RuntimeOptions{Config: validTestConfig()})
	if err != nil {
		t.Fatalf("build runtime: %v", err)
	}
	interaction := fakediscord.SlashInteractionWithOptions("兌換", "", map[string]string{"代碼": "abc"})
	responder := fakediscord.NewResponder()
	if err := dispatcher.Dispatch(context.Background(), interaction, responder); err == nil {
		t.Fatal("redeem route should not be available without repository")
	}

	repo := fakemongo.NewRedeemRepository()
	repo.Codes["abc"] = domain.RedeemCode{Code: "abc", Price: 3, CreatedAtMillis: time.Now().UnixMilli()}
	dispatcher, err = BuildRuntime(RuntimeOptions{
		Config:           validTestConfig(),
		RedeemRepository: repo,
	})
	if err != nil {
		t.Fatalf("build runtime with redeem repo: %v", err)
	}
	responder = fakediscord.NewResponder()
	if err := dispatcher.Dispatch(context.Background(), interaction, responder); err != nil {
		t.Fatalf("dispatch redeem: %v", err)
	}
	if len(responder.Edits) != 1 || len(responder.Edits[0].Embeds) != 1 || responder.Edits[0].Embeds[0].Author == nil || responder.Edits[0].Embeds[0].Author.Name != "成功兌換代碼!" {
		t.Fatalf("redeem response = %#v", responder.Edits)
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

func TestBuildRuntimeRoutesAntiScamReportOnlyWithCatalogAndSender(t *testing.T) {
	dispatcher, err := BuildRuntime(RuntimeOptions{Config: validTestConfig()})
	if err != nil {
		t.Fatalf("build runtime: %v", err)
	}
	interaction := fakediscord.SlashInteractionWithOptions("詐騙網址回報", "", map[string]string{"網址": "https://bad.example"})
	responder := fakediscord.NewResponder()
	if err := dispatcher.Dispatch(context.Background(), interaction, responder); err == nil {
		t.Fatal("anti-scam report route should not be available without catalog and sender")
	}

	catalog := fakemongo.NewScamURLCatalogRepository()
	sender := &fakeScamReportSender{}
	dispatcher, err = BuildRuntime(RuntimeOptions{
		Config:                   validTestConfig(),
		ScamURLCatalogRepository: catalog,
		ScamReportSender:         sender,
	})
	if err != nil {
		t.Fatalf("build runtime with anti-scam report ports: %v", err)
	}
	responder = fakediscord.NewResponder()
	if err := dispatcher.Dispatch(context.Background(), interaction, responder); err != nil {
		t.Fatalf("dispatch anti-scam report: %v", err)
	}
	if len(sender.Sent) != 1 || sender.Sent[0].URL != "https://bad.example" {
		t.Fatalf("sent = %#v", sender.Sent)
	}
	if len(responder.Edits) != 1 || !strings.Contains(responder.Edits[0].Embeds[0].Description, "成功回報https://bad.example") {
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

func TestBuildRuntimeRoutesDisabledXPProfilesOnlyWhenEnabled(t *testing.T) {
	dispatcher, err := BuildRuntime(RuntimeOptions{Config: validTestConfig()})
	if err != nil {
		t.Fatalf("build runtime: %v", err)
	}
	interaction := fakediscord.SlashInteraction("聊天經驗")
	responder := fakediscord.NewResponder()
	if err := dispatcher.Dispatch(context.Background(), interaction, responder); err == nil {
		t.Fatal("disabled XP profile route should not be available by default")
	}

	dispatcher, err = BuildRuntime(RuntimeOptions{
		Config:                   validTestConfig(),
		XPProfileDisabledEnabled: true,
	})
	if err != nil {
		t.Fatalf("build runtime with disabled XP profiles: %v", err)
	}
	responder = fakediscord.NewResponder()
	if err := dispatcher.Dispatch(context.Background(), interaction, responder); err != nil {
		t.Fatalf("dispatch text XP profile: %v", err)
	}
	if len(responder.Edits) != 1 || len(responder.Edits[0].Embeds) != 1 || !strings.Contains(responder.Edits[0].Embeds[0].Title, "該指令即將被移除") {
		t.Fatalf("text XP profile response = %#v", responder.Edits)
	}

	responder = fakediscord.NewResponder()
	if err := dispatcher.Dispatch(context.Background(), fakediscord.SlashInteraction("語音經驗"), responder); err != nil {
		t.Fatalf("dispatch voice XP profile: %v", err)
	}
	if len(responder.Edits) != 1 || len(responder.Edits[0].Embeds) != 1 || !strings.Contains(responder.Edits[0].Embeds[0].Title, "/我的檔案") {
		t.Fatalf("voice XP profile response = %#v", responder.Edits)
	}
}

func TestBuildRuntimeRoutesXPAdminOnlyWithRepository(t *testing.T) {
	dispatcher, err := BuildRuntime(RuntimeOptions{Config: validTestConfig()})
	if err != nil {
		t.Fatalf("build runtime: %v", err)
	}
	interaction := fakediscord.SlashInteractionWithOptions("經驗值改變", "聊天經驗改變", map[string]string{
		"使用者": "user-2",
		"經驗值": "150",
	})
	interaction.Actor.PermissionBits = 2
	responder := fakediscord.NewResponder()
	if err := dispatcher.Dispatch(context.Background(), interaction, responder); err == nil {
		t.Fatal("XP admin route should not be available without repository")
	}

	repo := fakemongo.NewXPAdminRepository()
	dispatcher, err = BuildRuntime(RuntimeOptions{
		Config:            validTestConfig(),
		XPAdminRepository: repo,
	})
	if err != nil {
		t.Fatalf("build runtime with XP admin: %v", err)
	}
	responder = fakediscord.NewResponder()
	if err := dispatcher.Dispatch(context.Background(), interaction, responder); err != nil {
		t.Fatalf("dispatch XP admin: %v", err)
	}
	profile := repo.TextProfiles["guild-1/user-2"]
	if profile.Level != 1 || profile.XP != 50 {
		t.Fatalf("profile = %#v", profile)
	}
	if len(responder.Edits) != 1 || len(responder.Edits[0].Embeds) != 1 || !strings.Contains(responder.Edits[0].Embeds[0].Title, "經驗系統") {
		t.Fatalf("XP admin response = %#v", responder.Edits)
	}
}

func TestBuildRuntimeRoutesXPResetOnlyWithDependencies(t *testing.T) {
	dispatcher, err := BuildRuntime(RuntimeOptions{Config: validTestConfig()})
	if err != nil {
		t.Fatalf("build runtime: %v", err)
	}
	interaction := fakediscord.SlashInteractionWithOptions("經驗值重製", "重製個人聊天經驗", map[string]string{"使用者": "user-2"})
	responder := fakediscord.NewResponder()
	if err := dispatcher.Dispatch(context.Background(), interaction, responder); err == nil {
		t.Fatal("XP reset route should not be available without dependencies")
	}

	repo := fakemongo.NewXPAdminRepository()
	repo.TextProfiles["guild-1/user-2"] = domain.XPProfile{GuildID: "guild-1", UserID: "user-2", XP: 10, Level: 1}
	dispatcher, err = BuildRuntime(RuntimeOptions{
		Config:             validTestConfig(),
		XPResetRepository:  repo,
		XPResetMessagePort: fakediscord.NewSideEffects(),
		XPResetGuildInfo:   &fakebotinfo.DiscordInfoProvider{Guild: ports.DiscordGuildInfo{OwnerID: "user-1"}},
	})
	if err != nil {
		t.Fatalf("build runtime with XP reset: %v", err)
	}
	responder = fakediscord.NewResponder()
	if err := dispatcher.Dispatch(context.Background(), interaction, responder); err != nil {
		t.Fatalf("dispatch XP reset: %v", err)
	}
	if _, ok := repo.TextProfiles["guild-1/user-2"]; ok {
		t.Fatal("XP reset route did not delete text profile")
	}
	if len(responder.Edits) != 1 || !strings.Contains(responder.Edits[0].Content, "成功清除<@user-2>的聊天經驗") {
		t.Fatalf("XP reset response = %#v", responder.Edits)
	}
}

func TestBuildRuntimeRoutesEconomyCoinResetOnlyWithDependencies(t *testing.T) {
	interaction := fakediscord.SlashInteraction("代幣重製")
	interaction.ChannelID = "channel-1"
	dispatcher, err := BuildRuntime(RuntimeOptions{
		Config:                     validTestConfig(),
		EconomyCoinResetRepository: fakemongo.NewEconomyRepository(),
	})
	if err != nil {
		t.Fatalf("build runtime: %v", err)
	}
	responder := fakediscord.NewResponder()
	if err := dispatcher.Dispatch(context.Background(), interaction, responder); err == nil {
		t.Fatal("economy coin-reset route should not be available without message and guild-info dependencies")
	}

	dispatcher, err = BuildRuntime(RuntimeOptions{
		Config:                      validTestConfig(),
		EconomyCoinResetRepository:  fakemongo.NewEconomyRepository(),
		EconomyCoinResetMessagePort: fakediscord.NewSideEffects(),
		EconomyCoinResetGuildInfo:   &fakebotinfo.DiscordInfoProvider{Guild: ports.DiscordGuildInfo{OwnerID: "user-1"}},
	})
	if err != nil {
		t.Fatalf("build runtime with economy coin reset: %v", err)
	}
	responder = fakediscord.NewResponder()
	if err := dispatcher.Dispatch(context.Background(), interaction, responder); err != nil {
		t.Fatalf("dispatch economy coin reset: %v", err)
	}
	if len(responder.Replies) != 1 || responder.Replies[0].Content != ":warning: | 一但重製，___**將無法復原**___，如確定要還原請於60秒內輸入`^確認^`(只有一次機會)!!!" {
		t.Fatalf("economy coin-reset response = %#v", responder.Replies)
	}
}

func TestBuildRuntimeRoutesRoleSelectionOnlyWithDependencies(t *testing.T) {
	interaction := fakediscord.SlashInteractionWithOptions("選取身分組-表情符號", "", map[string]string{
		"訊息url": "https://discord.com/channels/guild-1/channel-1/message-1",
		"身分組":   "role-1",
		"表情符號":  "👍",
	})
	interaction.Actor.PermissionBits = 8192
	dispatcher, err := BuildRuntime(RuntimeOptions{Config: validTestConfig()})
	if err != nil {
		t.Fatalf("build runtime: %v", err)
	}
	responder := fakediscord.NewResponder()
	if err := dispatcher.Dispatch(context.Background(), interaction, responder); err == nil {
		t.Fatal("role-selection route should not be available without repository")
	}

	repo := fakemongo.NewRoleSelectionRepository()
	sideEffects := fakediscord.NewSideEffects()
	sideEffects.AssignableRoles["guild-1/role-1"] = true
	dispatcher, err = BuildRuntime(RuntimeOptions{
		Config:                     validTestConfig(),
		RoleSelectionRepository:    repo,
		RoleSelectionRolePort:      sideEffects,
		RoleSelectionRoleInspector: sideEffects,
		RoleSelectionReactionPort:  sideEffects,
		RoleSelectionMessagePort:   sideEffects,
		RoleSelectionDirectMessage: sideEffects,
	})
	if err != nil {
		t.Fatalf("build runtime with role-selection: %v", err)
	}
	responder = fakediscord.NewResponder()
	if err := dispatcher.Dispatch(context.Background(), interaction, responder); err != nil {
		t.Fatalf("dispatch role-selection: %v", err)
	}
	saved, ok := repo.Reactions["guild-1/message-1/👍"]
	if !ok || saved.RoleID != "role-1" {
		t.Fatalf("saved role-selection config = %#v ok=%v", saved, ok)
	}
	if len(sideEffects.Reactions) != 1 || sideEffects.Reactions[0].ChannelID != "channel-1" || sideEffects.Reactions[0].MessageID != "message-1" {
		t.Fatalf("reaction side effect = %#v", sideEffects.Reactions)
	}
	if len(responder.Edits) != 1 || !strings.Contains(responder.Edits[0].Embeds[0].Title, "表情符號選取身分組成功設定") {
		t.Fatalf("role-selection response = %#v", responder.Edits)
	}
}

func TestBuildRuntimeRoutesXPRankOnlyWithRepository(t *testing.T) {
	dispatcher, err := BuildRuntime(RuntimeOptions{Config: validTestConfig()})
	if err != nil {
		t.Fatalf("build runtime: %v", err)
	}
	viewerID := "123456789012345678"
	interaction := fakediscord.SlashInteraction("聊天排行榜")
	interaction.Actor.UserID = viewerID
	responder := fakediscord.NewResponder()
	if err := dispatcher.Dispatch(context.Background(), interaction, responder); err == nil {
		t.Fatal("XP rank route should not be available without repository")
	}

	repo := fakemongo.NewXPAdminRepository()
	_ = repo.SaveTextXPProfile(context.Background(), domain.XPProfile{GuildID: "guild-1", UserID: viewerID, Level: 1, XP: 100})
	_ = repo.SaveTextXPProfile(context.Background(), domain.XPProfile{GuildID: "guild-1", UserID: "222222222222222222", Level: 2, XP: 0})
	dispatcher, err = BuildRuntime(RuntimeOptions{
		Config:           validTestConfig(),
		XPRankRepository: repo,
	})
	if err != nil {
		t.Fatalf("build runtime with XP rank: %v", err)
	}
	responder = fakediscord.NewResponder()
	if err := dispatcher.Dispatch(context.Background(), interaction, responder); err != nil {
		t.Fatalf("dispatch XP rank: %v", err)
	}
	if len(responder.Replies) != 1 || len(responder.Replies[0].Embeds) != 1 {
		t.Fatalf("XP rank loading reply = %#v", responder.Replies)
	}
	if len(responder.Edits) != 1 || len(responder.Edits[0].Files) != 1 || responder.Edits[0].Files[0].Name != "user-info.png" {
		t.Fatalf("XP rank response = %#v", responder.Edits)
	}
	if len(responder.Edits[0].Components) != 2 {
		t.Fatalf("XP rank components = %#v", responder.Edits[0].Components)
	}

	responder = fakediscord.NewResponder()
	component := fakediscord.ComponentInteractionFromID("[" + viewerID + "]{0}text_rank")
	if err := dispatcher.Dispatch(context.Background(), component, responder); err != nil {
		t.Fatalf("dispatch XP rank pagination: %v", err)
	}
	if len(responder.Updates) != 1 || len(responder.Updates[0].Files) != 1 || responder.Updates[0].Files[0].Name != "user-info.png" {
		t.Fatalf("XP rank pagination update = %#v", responder.Updates)
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

func TestBuildRuntimeRoutesVoiceRoomConfigOnlyWithRepository(t *testing.T) {
	dispatcher, err := BuildRuntime(RuntimeOptions{Config: validTestConfig()})
	if err != nil {
		t.Fatalf("build runtime: %v", err)
	}
	interaction := fakediscord.SlashInteractionWithOptions("語音包廂設置", "", map[string]string{
		"語音頻道":     "voice-1",
		"設定頻道名稱":   "{name} 的包廂",
		"是否予許房主上鎖": "true",
		"設定人數上限":   "4",
	})
	interaction.Actor.PermissionBits = 8192
	responder := fakediscord.NewResponder()
	if err := dispatcher.Dispatch(context.Background(), interaction, responder); err == nil {
		t.Fatal("voice-room config route should not be available without repository")
	}

	repo := fakemongo.NewVoiceRoomConfigRepository()
	dispatcher, err = BuildRuntime(RuntimeOptions{
		Config:                    validTestConfig(),
		VoiceRoomConfigRepository: repo,
	})
	if err != nil {
		t.Fatalf("build runtime with voice-room config repo: %v", err)
	}
	responder = fakediscord.NewResponder()
	if err := dispatcher.Dispatch(context.Background(), interaction, responder); err != nil {
		t.Fatalf("dispatch voice-room config: %v", err)
	}
	saved, ok := repo.Last()
	if !ok || saved.TriggerChannelID != "voice-1" || saved.Name != "{name} 的包廂" || saved.Limit != 4 || !saved.Lock {
		t.Fatalf("saved voice-room config = %#v ok=%v", saved, ok)
	}
	if len(responder.Edits) != 1 || len(responder.Edits[0].Embeds) != 1 || !strings.Contains(responder.Edits[0].Embeds[0].Description, "語音包廂") {
		t.Fatalf("voice-room response = %#v", responder.Edits)
	}
}

func TestBuildRuntimeRoutesVoiceRoomLockOnlyWithRepository(t *testing.T) {
	dispatcher, err := BuildRuntime(RuntimeOptions{Config: validTestConfig()})
	if err != nil {
		t.Fatalf("build runtime: %v", err)
	}
	interaction := fakediscord.SlashInteractionWithOptions("上鎖頻道", "", map[string]string{"密碼": "secret"})
	interaction.Actor.VoiceChannelID = "voice-1"
	interaction.ChannelID = "text-1"
	responder := fakediscord.NewResponder()
	if err := dispatcher.Dispatch(context.Background(), interaction, responder); err == nil {
		t.Fatal("voice-room lock route should not be available without repository")
	}

	repo := fakemongo.NewVoiceRoomLockRepository()
	repo.Locks["guild-1\x00voice-1"] = domain.VoiceRoomLock{
		GuildID:       "guild-1",
		ChannelID:     "voice-1",
		OwnerID:       "user-1",
		TextChannelID: "old-text",
	}
	dispatcher, err = BuildRuntime(RuntimeOptions{
		Config:                  validTestConfig(),
		VoiceRoomLockRepository: repo,
	})
	if err != nil {
		t.Fatalf("build runtime with voice-room lock repo: %v", err)
	}
	responder = fakediscord.NewResponder()
	if err := dispatcher.Dispatch(context.Background(), interaction, responder); err != nil {
		t.Fatalf("dispatch voice-room lock: %v", err)
	}
	saved, ok := repo.Last()
	if !ok || saved.ChannelID != "voice-1" || saved.OwnerID != "user-1" || saved.TextChannelID != "text-1" || saved.Password != "secret" {
		t.Fatalf("saved voice-room lock = %#v ok=%v", saved, ok)
	}
	if len(responder.Edits) != 1 || len(responder.Edits[0].Embeds) != 1 || !strings.Contains(responder.Edits[0].Embeds[0].Description, "secret") {
		t.Fatalf("voice-room lock response = %#v", responder.Edits)
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

type fakeScamReportSender struct {
	Sent []domain.ScamURLReport
}

func (s *fakeScamReportSender) SendScamURLReport(ctx context.Context, report domain.ScamURLReport) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s.Sent = append(s.Sent, report)
	return nil
}
