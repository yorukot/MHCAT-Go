package utility_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/commands"
	featureutility "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/features/utility"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/interactions"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/responses"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakebotinfo"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakediscord"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakeusage"
)

func TestInfoBotHandlerReturnsLegacyEmbed(t *testing.T) {
	provider := fakebotinfo.Provider{Info: ports.BotInfo{
		Name:            "MHCAT",
		ShardCount:      2,
		GuildCount:      12,
		UserCount:       345,
		Uptime:          time.Minute,
		CPUModel:        "test-cpu",
		CPUUsagePercent: 12.34,
		MemoryUsedMB:    128,
		MemoryTotalMB:   512,
	}}
	module := featureutility.NewModule(commands.BuiltinRegistry(commands.Scope{Kind: commands.ScopeGlobal}), provider, nil, nil)
	responder := fakediscord.NewResponder()
	interaction := fakediscord.SlashInteractionWithOptions("info", "bot", nil)
	if err := module.InfoHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Defers) != 1 || len(responder.Follow) != 1 {
		t.Fatalf("defers=%d follow=%d edits=%d", len(responder.Defers), len(responder.Follow), len(responder.Edits))
	}
	msg := responder.Follow[0]
	if len(msg.Embeds) != 1 || !strings.Contains(msg.Embeds[0].Title, "MHCAT目前系統使用量") {
		t.Fatalf("legacy info embed = %#v", msg.Embeds)
	}
	for _, want := range []string{"CPU型號", "CPU使用量", "分片數量", "RAM使用量", "開機時間", "總伺服器", "總使用者"} {
		if !embedHasFieldContaining(msg.Embeds[0], want) {
			t.Fatalf("legacy info embed missing field %q: %#v", want, msg.Embeds[0].Fields)
		}
	}
	if len(msg.Components) != 1 || len(msg.Components[0].Components) != 1 {
		t.Fatalf("legacy info components = %#v", msg.Components)
	}
	button := msg.Components[0].Components[0]
	if button.CustomID != "botinfoupdate" || button.Label != "更新" || button.Style != responses.ButtonStyleSuccess {
		t.Fatalf("legacy refresh button = %#v", button)
	}
}

func TestInfoBotHandlerDegradedProviderSafe(t *testing.T) {
	module := featureutility.NewModule(commands.BuiltinRegistry(commands.Scope{Kind: commands.ScopeGlobal}), fakebotinfo.Provider{Err: errors.New("secret internal failure")}, nil, nil)
	responder := fakediscord.NewResponder()
	if err := module.InfoHandler()(context.Background(), fakediscord.SlashInteractionWithOptions("info", "bot", nil), responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	msg := responder.Follow[0]
	if len(msg.Embeds) != 1 || !strings.Contains(msg.Embeds[0].Title, "錯誤") || strings.Contains(msg.Embeds[0].Description, "secret internal failure") {
		t.Fatalf("unsafe degraded response: %#v", msg)
	}
}

func TestInfoShardHandlerReturnsLegacyEmbed(t *testing.T) {
	provider := fakebotinfo.Provider{Info: ports.BotInfo{
		Name:          "MHCAT",
		ShardID:       1,
		ShardCount:    2,
		GuildCount:    22,
		UserCount:     330,
		Latency:       12 * time.Millisecond,
		Uptime:        3 * time.Minute,
		MemoryUsedMB:  256,
		MemoryTotalMB: 1024,
	}}
	module := featureutility.NewModule(commands.BuiltinRegistry(commands.Scope{Kind: commands.ScopeGlobal}), provider, nil, nil)
	responder := fakediscord.NewResponder()
	if err := module.InfoHandler()(context.Background(), fakediscord.SlashInteractionWithOptions("info", "shard", nil), responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Defers) != 1 || len(responder.Follow) != 1 {
		t.Fatalf("defers=%#v follow=%#v", responder.Defers, responder.Follow)
	}
	msg := responder.Follow[0]
	if len(msg.Embeds) != 1 || !strings.Contains(msg.Embeds[0].Title, "以下是每個分片的資訊") {
		t.Fatalf("shard embed = %#v", msg.Embeds)
	}
	if len(msg.Embeds[0].Fields) != 0 {
		t.Fatalf("legacy initial shard embed must remain empty: %#v", msg.Embeds[0].Fields)
	}
	if len(msg.Components) != 1 || len(msg.Components[0].Components) != 1 {
		t.Fatalf("shard components = %#v", msg.Components)
	}
	button := msg.Components[0].Components[0]
	if button.CustomID != "shardinfoupdate" || button.Label != "更新" || button.Style != responses.ButtonStyleSuccess {
		t.Fatalf("shard refresh button = %#v", button)
	}
}

func TestInfoUnsupportedSubcommandSafe(t *testing.T) {
	module := featureutility.NewModule(commands.BuiltinRegistry(commands.Scope{Kind: commands.ScopeGlobal}), nil, nil, nil)
	responder := fakediscord.NewResponder()
	if err := module.InfoHandler()(context.Background(), fakediscord.SlashInteractionWithOptions("info", "unknown", nil), responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Replies) != 1 || !strings.Contains(responder.Replies[0].Content, "尚未") {
		t.Fatalf("reply = %#v", responder.Replies)
	}
}

func TestInfoUserHandlerReturnsLegacyEmbed(t *testing.T) {
	createdAt := time.Unix(1_700_000_000, 0)
	joinedAt := time.Unix(1_700_100_000, 0)
	discordInfo := &fakebotinfo.DiscordInfoProvider{User: ports.DiscordUserInfo{
		ID:        "target-user",
		Username:  "Yoru",
		AvatarURL: "https://cdn.example/avatar.png",
		CreatedAt: createdAt,
		JoinedAt:  joinedAt,
	}}
	module := featureutility.NewModuleWithDiscordInfo(commands.BuiltinRegistry(commands.Scope{Kind: commands.ScopeGlobal}), nil, discordInfo, nil, nil)
	responder := fakediscord.NewResponder()
	interaction := fakediscord.SlashInteractionWithOptions("info", "user", map[string]string{"user": "target-user"})
	if err := module.InfoHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Defers) != 1 || len(responder.Edits) != 1 {
		t.Fatalf("defers=%#v edits=%#v", responder.Defers, responder.Edits)
	}
	if len(discordInfo.UserCalls) != 1 || discordInfo.UserCalls[0] != "guild-1:target-user" {
		t.Fatalf("user calls = %#v", discordInfo.UserCalls)
	}
	embed := responder.Edits[0].Embeds[0]
	if !strings.Contains(embed.Title, "以下是Yoru的資料") || embed.Thumbnail == nil || embed.Thumbnail.URL == "" {
		t.Fatalf("user embed = %#v", embed)
	}
	for _, want := range []string{"使用者ID", "創建時間", "加入時間"} {
		if !embedHasFieldContaining(embed, want) {
			t.Fatalf("user embed missing field %q: %#v", want, embed.Fields)
		}
	}
}

func TestInfoUserDefaultsToActor(t *testing.T) {
	discordInfo := &fakebotinfo.DiscordInfoProvider{User: ports.DiscordUserInfo{ID: "user-1", Username: "Actor"}}
	module := featureutility.NewModuleWithDiscordInfo(commands.BuiltinRegistry(commands.Scope{Kind: commands.ScopeGlobal}), nil, discordInfo, nil, nil)
	responder := fakediscord.NewResponder()
	if err := module.InfoHandler()(context.Background(), fakediscord.SlashInteractionWithOptions("info", "user", nil), responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(discordInfo.UserCalls) != 1 || discordInfo.UserCalls[0] != "guild-1:user-1" {
		t.Fatalf("user calls = %#v", discordInfo.UserCalls)
	}
}

func TestInfoGuildHandlerReturnsLegacyEmbed(t *testing.T) {
	createdAt := time.Unix(1_650_000_000, 0)
	discordInfo := &fakebotinfo.DiscordInfoProvider{Guild: ports.DiscordGuildInfo{
		ID:                       "guild-1",
		Name:                     "MHCAT Test",
		IconURL:                  "https://cdn.example/icon.png",
		BannerURL:                "https://cdn.example/banner.png",
		MemberCount:              123,
		PremiumSubscriptionCount: 2,
		PremiumTier:              1,
		CreatedAt:                createdAt,
		OwnerID:                  "owner-1",
		EmojiCount:               9,
		PreferredLocale:          "zh-TW",
		VerificationLevel:        2,
	}}
	module := featureutility.NewModuleWithDiscordInfo(commands.BuiltinRegistry(commands.Scope{Kind: commands.ScopeGlobal}), nil, discordInfo, nil, nil)
	responder := fakediscord.NewResponder()
	if err := module.InfoHandler()(context.Background(), fakediscord.SlashInteractionWithOptions("info", "guild", nil), responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Defers) != 1 || len(responder.Edits) != 1 {
		t.Fatalf("defers=%#v edits=%#v", responder.Defers, responder.Edits)
	}
	embed := responder.Edits[0].Embeds[0]
	if !strings.Contains(embed.Title, "以下是MHCAT Test的資料") || embed.Thumbnail == nil || embed.Image == nil {
		t.Fatalf("guild embed = %#v", embed)
	}
	for _, want := range []string{"伺服器ID", "成員數量", "加成狀態", "創建時間", "擁有者", "Emoji數量", "伺服器語言", "伺服器驗證等級"} {
		if !embedHasFieldContaining(embed, want) {
			t.Fatalf("guild embed missing field %q: %#v", want, embed.Fields)
		}
	}
}

func TestInfoUserAndGuildLookupErrorsAreSafe(t *testing.T) {
	discordInfo := &fakebotinfo.DiscordInfoProvider{
		UserErr:  errors.New("secret user lookup failure"),
		GuildErr: errors.New("secret guild lookup failure"),
	}
	module := featureutility.NewModuleWithDiscordInfo(commands.BuiltinRegistry(commands.Scope{Kind: commands.ScopeGlobal}), nil, discordInfo, nil, nil)
	for _, subcommand := range []string{"user", "guild"} {
		responder := fakediscord.NewResponder()
		if err := module.InfoHandler()(context.Background(), fakediscord.SlashInteractionWithOptions("info", subcommand, nil), responder); err != nil {
			t.Fatalf("handler %s: %v", subcommand, err)
		}
		if len(responder.Edits) != 1 || len(responder.Edits[0].Embeds) != 1 {
			t.Fatalf("%s edits = %#v", subcommand, responder.Edits)
		}
		got := responder.Edits[0].Embeds[0].Description
		if !strings.Contains(got, "無法獲取資訊") || strings.Contains(got, "secret") {
			t.Fatalf("unsafe lookup error for %s: %#v", subcommand, responder.Edits[0])
		}
	}
}

func TestInfoHandlerTracksUsage(t *testing.T) {
	tracker := &fakeusage.Tracker{}
	module := featureutility.NewModule(commands.BuiltinRegistry(commands.Scope{Kind: commands.ScopeGlobal}), nil, nil, tracker)
	if err := module.InfoHandler()(context.Background(), fakediscord.SlashInteractionWithOptions("info", "bot", nil), fakediscord.NewResponder()); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(tracker.Events) != 1 || tracker.Events[0].CommandName != "info" {
		t.Fatalf("usage events = %#v", tracker.Events)
	}
}

func TestInfoBotRefreshUpdatesMessageAndFollowsUp(t *testing.T) {
	provider := fakebotinfo.Provider{Info: ports.BotInfo{
		Name:            "MHCAT",
		ShardCount:      1,
		GuildCount:      2,
		UserCount:       30,
		CPUModel:        "test-cpu",
		MemoryUsedMB:    128,
		MemoryTotalMB:   512,
		CPUUsagePercent: 4.56,
	}}
	module := featureutility.NewModule(commands.BuiltinRegistry(commands.Scope{Kind: commands.ScopeGlobal}), provider, nil, nil)
	responder := fakediscord.NewResponder()
	if err := module.InfoBotRefreshHandler()(context.Background(), fakediscord.ComponentInteractionFromID("botinfoupdate"), responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Updates) != 1 || len(responder.Follow) != 1 {
		t.Fatalf("updates=%#v follow=%#v", responder.Updates, responder.Follow)
	}
	if len(responder.Updates[0].Embeds) != 1 || !strings.Contains(responder.Updates[0].Embeds[0].Title, "MHCAT目前系統使用量") {
		t.Fatalf("update = %#v", responder.Updates)
	}
	if !embedHasFieldContaining(responder.Updates[0].Embeds[0], "集群數量") {
		t.Fatalf("refresh fields = %#v", responder.Updates[0].Embeds[0].Fields)
	}
	if !responder.Follow[0].Ephemeral || !strings.Contains(responder.Follow[0].Content, "成功更新") {
		t.Fatalf("follow-up = %#v", responder.Follow)
	}
}

func TestInfoBotRefreshRoutesByParsedLegacyID(t *testing.T) {
	provider := fakebotinfo.Provider{Info: ports.BotInfo{
		Name:          "MHCAT",
		ShardCount:    1,
		GuildCount:    2,
		UserCount:     30,
		CPUModel:      "test-cpu",
		MemoryUsedMB:  128,
		MemoryTotalMB: 512,
	}}
	module := featureutility.NewModule(commands.BuiltinRegistry(commands.Scope{Kind: commands.ScopeGlobal}), provider, nil, nil)
	router := interactions.NewRouter()
	router.SetCustomIDParser(interactions.DefaultCustomIDParser{})
	if err := module.RegisterRoutes(router); err != nil {
		t.Fatalf("register routes: %v", err)
	}
	responder := fakediscord.NewResponder()
	if err := router.Handle(context.Background(), fakediscord.ComponentInteractionFromID("botinfoupdate"), responder); err != nil {
		t.Fatalf("handle: %v", err)
	}
	if len(responder.Updates) != 1 || len(responder.Follow) != 1 {
		t.Fatalf("updates=%#v follow=%#v", responder.Updates, responder.Follow)
	}
}

func TestInfoShardRefreshUpdatesMessage(t *testing.T) {
	provider := fakebotinfo.Provider{Info: ports.BotInfo{
		Name:          "MHCAT",
		ShardID:       3,
		ShardCount:    4,
		GuildCount:    22,
		UserCount:     330,
		Latency:       12 * time.Millisecond,
		Uptime:        3 * time.Minute,
		MemoryUsedMB:  256,
		MemoryTotalMB: 1024,
	}}
	module := featureutility.NewModule(commands.BuiltinRegistry(commands.Scope{Kind: commands.ScopeGlobal}), provider, nil, nil)
	responder := fakediscord.NewResponder()
	if err := module.InfoShardRefreshHandler()(context.Background(), fakediscord.ComponentInteractionFromID("shardinfoupdate"), responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Updates) != 1 || len(responder.Follow) != 0 {
		t.Fatalf("updates=%#v follow=%#v", responder.Updates, responder.Follow)
	}
	embed := responder.Updates[0].Embeds[0]
	if !strings.Contains(embed.Title, "以下是每個分片的資訊") || len(embed.Fields) != 1 {
		t.Fatalf("shard update embed = %#v", embed)
	}
	if !strings.Contains(embed.Fields[0].Name, "分片ID: 3") ||
		!strings.Contains(embed.Fields[0].Value, "公會數量: 22") ||
		!strings.Contains(embed.Fields[0].Value, "使用者數量: 330") ||
		!strings.Contains(embed.Fields[0].Value, "上線時間:00h03m00s") ||
		!strings.Contains(embed.Fields[0].Value, "延遲: 12```") {
		t.Fatalf("shard field = %#v", embed.Fields[0])
	}
}

func TestInfoShardRefreshRoutesByParsedLegacyID(t *testing.T) {
	provider := fakebotinfo.Provider{Info: ports.BotInfo{Name: "MHCAT", ShardCount: 1}}
	module := featureutility.NewModule(commands.BuiltinRegistry(commands.Scope{Kind: commands.ScopeGlobal}), provider, nil, nil)
	router := interactions.NewRouter()
	router.SetCustomIDParser(interactions.DefaultCustomIDParser{})
	if err := module.RegisterRoutes(router); err != nil {
		t.Fatalf("register routes: %v", err)
	}
	responder := fakediscord.NewResponder()
	if err := router.Handle(context.Background(), fakediscord.ComponentInteractionFromID("shardinfoupdate"), responder); err != nil {
		t.Fatalf("handle: %v", err)
	}
	if len(responder.Updates) != 1 {
		t.Fatalf("updates=%#v", responder.Updates)
	}
}

func embedHasFieldContaining(embed responses.Embed, value string) bool {
	for _, field := range embed.Fields {
		if strings.Contains(field.Name, value) {
			return true
		}
	}
	return false
}
