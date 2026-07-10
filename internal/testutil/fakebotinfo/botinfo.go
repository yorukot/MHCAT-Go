package fakebotinfo

import (
	"context"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
)

type Provider struct {
	Info ports.BotInfo
	Err  error
}

func (p Provider) BotInfo(context.Context) (ports.BotInfo, error) {
	return p.Info, p.Err
}

type DiscordInfoProvider struct {
	User            ports.DiscordUserInfo
	Users           map[string]ports.DiscordUserInfo
	CachedUsers     map[string]ports.DiscordUserInfo
	Guild           ports.DiscordGuildInfo
	UserErr         error
	CachedUserErr   error
	GuildErr        error
	UserCalls       []string
	CachedUserCalls []string
	GuildCalls      []string
}

func (p *DiscordInfoProvider) CachedUserInfo(_ context.Context, guildID string, userID string) (ports.DiscordUserInfo, bool, error) {
	p.CachedUserCalls = append(p.CachedUserCalls, guildID+":"+userID)
	user, ok := p.CachedUsers[userID]
	return user, ok, p.CachedUserErr
}

func (p *DiscordInfoProvider) UserInfo(_ context.Context, guildID string, userID string) (ports.DiscordUserInfo, error) {
	p.UserCalls = append(p.UserCalls, guildID+":"+userID)
	if p.Users != nil {
		if user, ok := p.Users[userID]; ok {
			return user, p.UserErr
		}
	}
	return p.User, p.UserErr
}

func (p *DiscordInfoProvider) GuildInfo(_ context.Context, guildID string) (ports.DiscordGuildInfo, error) {
	p.GuildCalls = append(p.GuildCalls, guildID)
	return p.Guild, p.GuildErr
}

var _ ports.DiscordInfoProvider = (*DiscordInfoProvider)(nil)
var _ ports.DiscordCachedUserInfoProvider = (*DiscordInfoProvider)(nil)
