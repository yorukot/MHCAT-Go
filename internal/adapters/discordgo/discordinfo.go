package discordgo

import (
	"context"
	"fmt"
	"sort"
	"time"

	dgo "github.com/bwmarrin/discordgo"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
)

type DiscordInfoProvider struct {
	Session *Session
}

func (p DiscordInfoProvider) UserInfo(ctx context.Context, guildID string, userID string) (ports.DiscordUserInfo, error) {
	if err := ctx.Err(); err != nil {
		return ports.DiscordUserInfo{}, err
	}
	session, err := p.session()
	if err != nil {
		return ports.DiscordUserInfo{}, err
	}
	member, err := stateMember(session, guildID, userID)
	if err != nil {
		member, err = session.GuildMember(guildID, userID, dgo.WithContext(ctx))
		if err != nil {
			return ports.DiscordUserInfo{}, fmt.Errorf("load guild member: %w", err)
		}
	}
	return userInfoFromMember(member), nil
}

func (p DiscordInfoProvider) GuildInfo(ctx context.Context, guildID string) (ports.DiscordGuildInfo, error) {
	if err := ctx.Err(); err != nil {
		return ports.DiscordGuildInfo{}, err
	}
	session, err := p.session()
	if err != nil {
		return ports.DiscordGuildInfo{}, err
	}
	guild, err := stateGuild(session, guildID)
	if err != nil {
		guild, err = session.GuildWithCounts(guildID, dgo.WithContext(ctx))
		if err != nil {
			return ports.DiscordGuildInfo{}, fmt.Errorf("load guild: %w", err)
		}
	}
	info := guildInfoFromGuild(guild)
	info.BotDisplayColor = discordBotDisplayColor(session, guild)
	return info, nil
}

func (p DiscordInfoProvider) session() (*dgo.Session, error) {
	if p.Session == nil {
		return nil, fmt.Errorf("discord info provider is not configured")
	}
	p.Session.mu.Lock()
	defer p.Session.mu.Unlock()
	if p.Session.session == nil {
		return nil, fmt.Errorf("discord session is not configured")
	}
	return p.Session.session, nil
}

func stateMember(session *dgo.Session, guildID string, userID string) (*dgo.Member, error) {
	if session.State == nil {
		return nil, fmt.Errorf("discord state is not configured")
	}
	return session.State.Member(guildID, userID)
}

func stateGuild(session *dgo.Session, guildID string) (*dgo.Guild, error) {
	if session.State == nil {
		return nil, fmt.Errorf("discord state is not configured")
	}
	return session.State.Guild(guildID)
}

func userInfoFromMember(member *dgo.Member) ports.DiscordUserInfo {
	if member == nil || member.User == nil {
		return ports.DiscordUserInfo{}
	}
	createdAt, _ := dgo.SnowflakeTimestamp(member.User.ID)
	return ports.DiscordUserInfo{
		ID:            member.User.ID,
		Username:      member.User.Username,
		Nickname:      member.Nick,
		Discriminator: member.User.Discriminator,
		AvatarURL:     member.AvatarURL(""),
		CreatedAt:     createdAt,
		JoinedAt:      member.JoinedAt,
	}
}

func guildInfoFromGuild(guild *dgo.Guild) ports.DiscordGuildInfo {
	if guild == nil {
		return ports.DiscordGuildInfo{}
	}
	createdAt, _ := dgo.SnowflakeTimestamp(guild.ID)
	return ports.DiscordGuildInfo{
		ID:                       guild.ID,
		Name:                     guild.Name,
		IconURL:                  guild.IconURL(""),
		BannerURL:                guild.BannerURL("1024"),
		MemberCount:              guild.MemberCount,
		PremiumSubscriptionCount: guild.PremiumSubscriptionCount,
		PremiumTier:              int(guild.PremiumTier),
		CreatedAt:                zeroIfInvalid(createdAt),
		OwnerID:                  guild.OwnerID,
		EmojiCount:               len(guild.Emojis),
		PreferredLocale:          guild.PreferredLocale,
		VerificationLevel:        int(guild.VerificationLevel),
	}
}

func discordBotDisplayColor(session *dgo.Session, guild *dgo.Guild) int {
	if session == nil || session.State == nil || session.State.User == nil || guild == nil {
		return 0
	}
	member, err := session.State.Member(guild.ID, session.State.User.ID)
	if err != nil || member == nil {
		return 0
	}
	memberRoles := make(map[string]struct{}, len(member.Roles))
	for _, roleID := range member.Roles {
		memberRoles[roleID] = struct{}{}
	}
	roles := append(dgo.Roles(nil), guild.Roles...)
	sort.Sort(roles)
	for _, role := range roles {
		if role == nil || role.Color == 0 {
			continue
		}
		if _, ok := memberRoles[role.ID]; ok {
			return role.Color
		}
	}
	return 0
}

func zeroIfInvalid(value time.Time) time.Time {
	if value.IsZero() {
		return time.Time{}
	}
	return value
}
