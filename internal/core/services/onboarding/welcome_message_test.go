package onboarding

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakediscord"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakemongo"
)

func TestWelcomeMessageDeliveryMissingConfigNoop(t *testing.T) {
	repo := fakemongo.NewJoinMessageConfigRepository()
	sideEffects := fakediscord.NewSideEffects()
	service := WelcomeMessageDeliveryService{Repository: repo, Messages: sideEffects, Channels: sideEffects}

	if err := service.SendOnJoin(context.Background(), WelcomeMemberEvent{GuildID: "guild-1", UserID: "user-1"}); err != nil {
		t.Fatalf("send: %v", err)
	}
	if len(sideEffects.Sent) != 0 {
		t.Fatalf("sent = %#v", sideEffects.Sent)
	}
}

func TestWelcomeMessageDeliveryDisabledConfigNoop(t *testing.T) {
	repo := fakemongo.NewJoinMessageConfigRepository()
	repo.Configs["guild-1"] = domain.JoinMessageConfig{
		GuildID:        "guild-1",
		Enabled:        false,
		ChannelID:      "channel-1",
		MessageContent: "hello",
		Color:          "#53FF53",
	}
	sideEffects := fakediscord.NewSideEffects()
	service := WelcomeMessageDeliveryService{Repository: repo, Messages: sideEffects, Channels: sideEffects}

	if err := service.SendOnJoin(context.Background(), WelcomeMemberEvent{GuildID: "guild-1", UserID: "user-1"}); err != nil {
		t.Fatalf("send: %v", err)
	}
	if len(sideEffects.Sent) != 0 {
		t.Fatalf("sent = %#v", sideEffects.Sent)
	}
}

func TestWelcomeMessageDeliveryPreservesAllSpaceContent(t *testing.T) {
	repo := fakemongo.NewJoinMessageConfigRepository()
	repo.Configs["guild-1"] = domain.JoinMessageConfig{
		GuildID:        "guild-1",
		Enabled:        true,
		ChannelID:      "channel-1",
		MessageContent: "   ",
		Color:          "#53FF53",
	}
	sideEffects := fakediscord.NewSideEffects()
	cacheLegacyDeliveryChannel(sideEffects, "guild-1", "channel-1")
	service := WelcomeMessageDeliveryService{Repository: repo, Messages: sideEffects, Channels: sideEffects}

	if err := service.SendOnJoin(context.Background(), WelcomeMemberEvent{GuildID: "guild-1", UserID: "user-1"}); err != nil {
		t.Fatalf("send: %v", err)
	}
	if len(sideEffects.Sent) != 1 || sideEffects.Sent[0].Message.Embeds[0].Description != "   " {
		t.Fatalf("sent = %#v", sideEffects.Sent)
	}
}

func TestWelcomeMessageDeliveryGenericLegacyEmbed(t *testing.T) {
	now := time.Unix(2_000_000, 0)
	repo := fakemongo.NewJoinMessageConfigRepository()
	repo.Configs["guild-1"] = domain.JoinMessageConfig{
		GuildID:        "guild-1",
		Enabled:        true,
		ChannelID:      "channel-1",
		MessageContent: "歡迎 (MEMBERNAME) {MEMBERNAME} {membername} (TAG) {TAG} {tag}",
		Color:          "#53FF53",
		ImageURL:       "https://example.test/welcome.png",
	}
	sideEffects := fakediscord.NewSideEffects()
	cacheLegacyDeliveryChannel(sideEffects, "guild-1", "channel-1")
	service := WelcomeMessageDeliveryService{Repository: repo, Messages: sideEffects, Channels: sideEffects}

	err := service.SendOnJoin(context.Background(), WelcomeMemberEvent{
		GuildID:      "guild-1",
		GuildName:    "測試伺服器",
		GuildIconURL: "https://example.test/guild.png",
		BotAvatarURL: "https://example.test/bot.png",
		UserID:       "user-1",
		Username:     "Yoru",
		UserTag:      "Yoru#0001",
		AvatarURL:    "https://example.test/avatar.png",
		Now:          now,
	})
	if err != nil {
		t.Fatalf("send: %v", err)
	}
	if len(sideEffects.Sent) != 1 || sideEffects.Sent[0].ChannelID != "channel-1" {
		t.Fatalf("sent = %#v", sideEffects.Sent)
	}
	message := sideEffects.Sent[0].Message
	if len(message.AllowedMentions.UserIDs) != 1 || message.AllowedMentions.UserIDs[0] != "user-1" || message.AllowedMentions.ParseEveryone || message.AllowedMentions.ParseRoles {
		t.Fatalf("allowed mentions = %#v", message.AllowedMentions)
	}
	embed := message.Embeds[0]
	if embed.AuthorName != "🪂 歡迎加入 測試伺服器" || embed.AuthorIconURL != "https://example.test/guild.png" {
		t.Fatalf("author = %#v", embed)
	}
	if embed.Description != "歡迎 Yoru Yoru Yoru <@user-1> <@user-1> <@user-1>" {
		t.Fatalf("description = %q", embed.Description)
	}
	if embed.Color != 0x53FF53 || embed.ThumbnailURL != "https://example.test/avatar.png" || embed.ImageURL != "https://example.test/welcome.png" || !embed.Timestamp.Equal(now) {
		t.Fatalf("embed = %#v", embed)
	}
}

func TestWelcomeMessageDeliverySpecialLegacyEmbed(t *testing.T) {
	now := time.Unix(2_000_000, 0)
	sideEffects := fakediscord.NewSideEffects()
	cacheLegacyDeliveryChannel(sideEffects, "special-guild", "special-channel")
	service := WelcomeMessageDeliveryService{
		Messages: sideEffects,
		Channels: sideEffects,
		Special: SpecialWelcomeConfig{
			GuildID:          "special-guild",
			BotID:            "special-bot",
			ChannelID:        "special-channel",
			ChatChannelID:    "111111111111111111",
			HelpChannelID:    "222222222222222222",
			BugChannelID:     "333333333333333333",
			SupportChannelID: "444444444444444444",
		},
	}

	err := service.SendOnJoin(context.Background(), WelcomeMemberEvent{
		GuildID:       "special-guild",
		BotUserID:     "special-bot",
		BotAvatarURL:  "https://example.test/bot.png",
		UserID:        "user-1",
		Username:      "Yoru",
		Discriminator: "0001",
		UserTag:       "Yoru#0001",
		AvatarURL:     "https://example.test/avatar.png",
		Now:           now,
	})
	if err != nil {
		t.Fatalf("send: %v", err)
	}
	if len(sideEffects.Sent) != 1 || sideEffects.Sent[0].ChannelID != "special-channel" {
		t.Fatalf("sent = %#v", sideEffects.Sent)
	}
	embed := sideEffects.Sent[0].Message.Embeds[0]
	if embed.AuthorName != "🪂 歡迎加入 MHCAT!" || embed.AuthorURL != "https://dsc.gg/MHCAT" || embed.AuthorIconURL != "https://example.test/bot.png" {
		t.Fatalf("author = %#v", embed)
	}
	if !strings.Contains(embed.Description, "歡迎 __Yoru#0001__ 的加入!") ||
		!strings.Contains(embed.Description, "<#111111111111111111>想要聊天") ||
		!strings.Contains(embed.Description, "<#222222222222222222>對指令") ||
		!strings.Contains(embed.Description, "<#333333333333333333>有任何bug") ||
		!strings.Contains(embed.Description, "<#444444444444444444>開啟客服頻道") ||
		embed.ImageURL != "https://i.imgur.com/cLCPRNq.png" ||
		embed.Color == 0 {
		t.Fatalf("embed = %#v", embed)
	}
}

func TestWelcomeMessageDeliverySkipsUncachedLegacyChannels(t *testing.T) {
	t.Run("generic", func(t *testing.T) {
		repo := fakemongo.NewJoinMessageConfigRepository()
		repo.Configs["guild-1"] = domain.JoinMessageConfig{
			GuildID: "guild-1", Enabled: true, ChannelID: "missing", MessageContent: "welcome", Color: "not-a-color",
		}
		sideEffects := fakediscord.NewSideEffects()
		service := WelcomeMessageDeliveryService{Repository: repo, Messages: sideEffects, Channels: sideEffects}
		if err := service.SendOnJoin(context.Background(), WelcomeMemberEvent{GuildID: "guild-1", UserID: "user-1"}); err != nil {
			t.Fatalf("send: %v", err)
		}
		if len(sideEffects.Sent) != 0 {
			t.Fatalf("sent = %#v", sideEffects.Sent)
		}
	})

	t.Run("special", func(t *testing.T) {
		sideEffects := fakediscord.NewSideEffects()
		service := WelcomeMessageDeliveryService{
			Messages: sideEffects,
			Channels: sideEffects,
			Special: SpecialWelcomeConfig{
				GuildID: "guild-1", BotID: "bot-1", ChannelID: "missing", ChatChannelID: "chat", HelpChannelID: "help", BugChannelID: "bug", SupportChannelID: "support",
			},
		}
		if err := service.SendOnJoin(context.Background(), WelcomeMemberEvent{GuildID: "guild-1", BotUserID: "bot-1", UserID: "user-1"}); err != nil {
			t.Fatalf("send: %v", err)
		}
		if len(sideEffects.Sent) != 0 {
			t.Fatalf("sent = %#v", sideEffects.Sent)
		}
	})
}

func TestWelcomeMessageDeliveryRejectsInvalidStoredColorWithoutSending(t *testing.T) {
	for _, color := range []string{"not-a-color", "   "} {
		t.Run(color, func(t *testing.T) {
			repo := fakemongo.NewJoinMessageConfigRepository()
			repo.Configs["guild-1"] = domain.JoinMessageConfig{
				GuildID: "guild-1", Enabled: true, ChannelID: "channel-1", MessageContent: "welcome", Color: color,
			}
			sideEffects := fakediscord.NewSideEffects()
			cacheLegacyDeliveryChannel(sideEffects, "guild-1", "channel-1")
			service := WelcomeMessageDeliveryService{Repository: repo, Messages: sideEffects, Channels: sideEffects}
			err := service.SendOnJoin(context.Background(), WelcomeMemberEvent{GuildID: "guild-1", UserID: "user-1"})
			if !errors.Is(err, domain.ErrInvalidJoinMessageConfig) {
				t.Fatalf("error = %v", err)
			}
			if len(sideEffects.Sent) != 0 {
				t.Fatalf("sent = %#v", sideEffects.Sent)
			}
		})
	}
}

func cacheLegacyDeliveryChannel(sideEffects *fakediscord.SideEffects, guildID string, channelID string) {
	sideEffects.Channels = append(sideEffects.Channels, ports.ChannelRef{GuildID: guildID, ChannelID: channelID})
}

func TestSpecialWelcomePreservesLegacyMigratedUsernameTag(t *testing.T) {
	description := specialWelcomeDescription(WelcomeMemberEvent{
		UserID:        "user-1",
		Username:      "yoru",
		Discriminator: "0",
		UserTag:       "yoru",
	}, SpecialWelcomeConfig{})
	if !strings.Contains(description, "歡迎 __yoru#0__ 的加入!") {
		t.Fatalf("description = %q", description)
	}
}

func TestWelcomeMessagePlaceholdersPreserveJavaScriptReplacementTokens(t *testing.T) {
	tests := []struct {
		name     string
		username string
		want     string
	}{
		{name: "dollar", username: "$$", want: "pre$post"},
		{name: "matched text", username: "$&", want: "pre(MEMBERNAME)post"},
		{name: "prefix", username: "$`", want: "preprepost"},
		{name: "suffix", username: "$'", want: "prepostpost"},
		{name: "raw whitespace", username: "  Yoru  ", want: "pre  Yoru  post"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := replaceWelcomeMessagePlaceholders("pre(MEMBERNAME)post", WelcomeMemberEvent{UserID: "user", Username: tc.username})
			if got != tc.want {
				t.Fatalf("description = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestLegacyRandomColorNames(t *testing.T) {
	for _, value := range []string{"Random", "RANDOM"} {
		if !legacyRandomColor(value) {
			t.Fatalf("expected %q to be random", value)
		}
	}
	for _, value := range []string{"random", " Random ", "red"} {
		if legacyRandomColor(value) {
			t.Fatalf("did not expect %q to be random", value)
		}
	}
	if got, ok := welcomeMessageColor("Green"); !ok || got != 0x57F287 {
		t.Fatalf("Discord.js Green = %#x/%t", got, ok)
	}
	if got, ok := leaveMessageDeliveryColor("Red"); !ok || got != 0xED4245 {
		t.Fatalf("Discord.js Red = %#x/%t", got, ok)
	}
	for _, value := range []string{"red", "invalid", " Red "} {
		if _, ok := welcomeMessageColor(value); ok {
			t.Fatalf("unexpected valid Discord.js color %q", value)
		}
	}
}
