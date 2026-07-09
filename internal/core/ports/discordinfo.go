package ports

import (
	"context"
	"time"
)

type DiscordUserInfo struct {
	ID        string
	Username  string
	AvatarURL string
	CreatedAt time.Time
	JoinedAt  time.Time
}

type DiscordGuildInfo struct {
	ID                       string
	Name                     string
	IconURL                  string
	BannerURL                string
	MemberCount              int
	PremiumSubscriptionCount int
	PremiumTier              int
	CreatedAt                time.Time
	OwnerID                  string
	EmojiCount               int
	PreferredLocale          string
	VerificationLevel        int
}

type DiscordInfoProvider interface {
	UserInfo(ctx context.Context, guildID string, userID string) (DiscordUserInfo, error)
	GuildInfo(ctx context.Context, guildID string) (DiscordGuildInfo, error)
}
