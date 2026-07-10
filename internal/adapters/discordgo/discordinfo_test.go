package discordgo

import (
	"context"
	"testing"
	"time"

	dgo "github.com/bwmarrin/discordgo"
)

func TestDiscordInfoProviderGuildInfoFromState(t *testing.T) {
	createdID := "4194304"
	session := &Session{session: &dgo.Session{State: dgo.NewState()}, ready: make(chan struct{})}
	session.session.State.User = &dgo.User{ID: "bot-1"}
	session.session.State.GuildAdd(&dgo.Guild{
		ID:                       createdID,
		Name:                     "Guild",
		Icon:                     "icon-hash",
		Banner:                   "banner-hash",
		OwnerID:                  "owner-1",
		MemberCount:              42,
		PremiumSubscriptionCount: 3,
		PremiumTier:              dgo.PremiumTier2,
		PreferredLocale:          "zh-TW",
		VerificationLevel:        dgo.VerificationLevel(2),
		Emojis:                   []*dgo.Emoji{{ID: "emoji-1"}, {ID: "emoji-2"}},
		Roles: []*dgo.Role{
			{ID: "colored-low", Position: 1, Color: 0x123456},
			{ID: "uncolored-high", Position: 2},
		},
		Members: []*dgo.Member{{
			GuildID: createdID,
			User:    &dgo.User{ID: "bot-1"},
			Roles:   []string{"colored-low", "uncolored-high"},
		}},
	})
	info, err := (DiscordInfoProvider{Session: session}).GuildInfo(context.Background(), createdID)
	if err != nil {
		t.Fatalf("guild info: %v", err)
	}
	if info.ID != createdID || info.Name != "Guild" || info.OwnerID != "owner-1" || info.MemberCount != 42 || info.EmojiCount != 2 || info.BotDisplayColor != 0x123456 {
		t.Fatalf("info = %#v", info)
	}
	if info.IconURL == "" || info.BannerURL == "" || info.CreatedAt.IsZero() {
		t.Fatalf("missing URLs or timestamp: %#v", info)
	}
}

func TestDiscordInfoProviderUserInfoFromState(t *testing.T) {
	joinedAt := time.Unix(1_700_000_000, 0)
	session := &Session{session: &dgo.Session{State: dgo.NewState()}, ready: make(chan struct{})}
	if err := session.session.State.GuildAdd(&dgo.Guild{
		ID: "guild-1",
		Members: []*dgo.Member{{
			GuildID:  "guild-1",
			JoinedAt: joinedAt,
			User: &dgo.User{
				ID:            "4194304",
				Username:      "Yoru",
				Discriminator: "1234",
				Avatar:        "avatar-hash",
			},
			Nick: "YoruNick",
		}},
	}); err != nil {
		t.Fatalf("guild add: %v", err)
	}
	info, err := (DiscordInfoProvider{Session: session}).UserInfo(context.Background(), "guild-1", "4194304")
	if err != nil {
		t.Fatalf("user info: %v", err)
	}
	if info.ID != "4194304" || info.Username != "Yoru" || !info.JoinedAt.Equal(joinedAt) || info.AvatarURL == "" || info.CreatedAt.IsZero() {
		t.Fatalf("info = %#v", info)
	}
	if info.Nickname != "YoruNick" || info.Discriminator != "1234" {
		t.Fatalf("missing legacy name fields: %#v", info)
	}
}

func TestDiscordInfoProviderCachedUserInfoDoesNotFetchMissingMembers(t *testing.T) {
	session := &Session{session: &dgo.Session{State: dgo.NewState()}, ready: make(chan struct{})}
	if err := session.session.State.GuildAdd(&dgo.Guild{
		ID: "guild-1",
		Members: []*dgo.Member{{
			GuildID: "guild-1",
			User:    &dgo.User{ID: "user-1", Username: "Yoru", Discriminator: "0"},
		}},
	}); err != nil {
		t.Fatalf("guild add: %v", err)
	}
	provider := DiscordInfoProvider{Session: session}
	info, ok, err := provider.CachedUserInfo(context.Background(), "guild-1", "user-1")
	if err != nil {
		t.Fatalf("cached user info: %v", err)
	}
	if !ok || info.Username != "Yoru" || info.Discriminator != "0" {
		t.Fatalf("info = %#v, ok = %t", info, ok)
	}
	if _, ok, err := provider.CachedUserInfo(context.Background(), "guild-1", "missing"); err != nil || ok {
		t.Fatalf("missing cached user: ok=%t err=%v", ok, err)
	}
}
