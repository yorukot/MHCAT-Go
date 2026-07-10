package discordgo

import (
	"context"
	"errors"
	"strings"

	dgo "github.com/bwmarrin/discordgo"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
)

type CachedRoleInspector struct {
	Session *Session
}

func NewCachedRoleInspector(client SideEffectClient) CachedRoleInspector {
	return CachedRoleInspector{Session: client.Session}
}

func (c CachedRoleInspector) CanAssignRole(ctx context.Context, guildID string, roleID string) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	session, err := (SideEffectClient{Session: c.Session}).session()
	if err != nil {
		return false, err
	}
	if session.State == nil {
		return false, errors.New("discord role cache is unavailable")
	}
	guildID = strings.TrimSpace(guildID)
	guild, err := session.State.Guild(guildID)
	if err != nil || guild == nil {
		return false, errors.New("discord role cache is missing guild")
	}
	session.State.RLock()
	botID := ""
	if session.State.User != nil {
		botID = session.State.User.ID
	}
	session.State.RUnlock()
	if botID == "" {
		return false, errors.New("discord role cache is missing bot user")
	}
	botMember, err := session.State.Member(guildID, botID)
	if err != nil || botMember == nil {
		return false, errors.New("discord role cache is missing bot member")
	}

	session.State.RLock()
	defer session.State.RUnlock()
	targetPosition, ok := rolePosition(guild.Roles, roleID)
	if !ok {
		return false, ports.ErrDiscordRoleMissing
	}
	botHighest := highestCachedRolePosition(guild.Roles, botMember.Roles)
	return targetPosition < botHighest, ctx.Err()
}

func highestCachedRolePosition(roles []*dgo.Role, roleIDs []string) int {
	highest := -1
	for _, roleID := range roleIDs {
		if position, ok := rolePosition(roles, roleID); ok && position > highest {
			highest = position
		}
	}
	return highest
}

var _ ports.DiscordRoleInspector = CachedRoleInspector{}
