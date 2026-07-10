package discordgo

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	dgo "github.com/bwmarrin/discordgo"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
)

type SideEffectClient struct {
	Session *Session
}

func (c SideEffectClient) SendMessage(ctx context.Context, channelID string, msg ports.OutboundMessage) (ports.MessageRef, error) {
	session, err := c.session()
	if err != nil {
		return ports.MessageRef{}, err
	}
	sent, err := session.ChannelMessageSendComplex(channelID, outboundMessageSend(channelID, msg), dgo.WithContext(ctx))
	if err != nil {
		return ports.MessageRef{}, fmt.Errorf("send discord message: %w", err)
	}
	return ports.MessageRef{ChannelID: sent.ChannelID, MessageID: sent.ID}, ctx.Err()
}

func (c SideEffectClient) SendDirectMessage(ctx context.Context, userID string, msg ports.OutboundMessage) (ports.MessageRef, error) {
	session, err := c.session()
	if err != nil {
		return ports.MessageRef{}, err
	}
	channel, err := session.UserChannelCreate(userID, dgo.WithContext(ctx))
	if err != nil {
		return ports.MessageRef{}, fmt.Errorf("create discord direct message channel: %w", err)
	}
	sent, err := session.ChannelMessageSendComplex(channel.ID, outboundMessageSend(channel.ID, msg), dgo.WithContext(ctx))
	if err != nil {
		return ports.MessageRef{}, fmt.Errorf("send discord direct message: %w", err)
	}
	return ports.MessageRef{ChannelID: sent.ChannelID, MessageID: sent.ID}, ctx.Err()
}

func (c SideEffectClient) SendTyping(ctx context.Context, channelID string) error {
	session, err := c.session()
	if err != nil {
		return err
	}
	if err := session.ChannelTyping(channelID, dgo.WithContext(ctx)); err != nil {
		return fmt.Errorf("send discord typing indicator: %w", err)
	}
	return ctx.Err()
}

func (c SideEffectClient) EditMessage(ctx context.Context, ref ports.MessageRef, msg ports.OutboundMessage) error {
	session, err := c.session()
	if err != nil {
		return err
	}
	edit := dgo.NewMessageEdit(ref.ChannelID, ref.MessageID)
	edit.Content = &msg.Content
	embeds := outboundEmbeds(msg.Embeds)
	components := outboundComponents(msg.Components)
	edit.Embeds = &embeds
	edit.Components = &components
	edit.AllowedMentions = coreAllowedMentions(msg.AllowedMentions)
	if _, err := session.ChannelMessageEditComplex(edit, dgo.WithContext(ctx)); err != nil {
		return fmt.Errorf("edit discord message: %w", err)
	}
	return ctx.Err()
}

func (c SideEffectClient) DeleteMessage(ctx context.Context, ref ports.MessageRef) error {
	session, err := c.session()
	if err != nil {
		return err
	}
	if err := session.ChannelMessageDelete(ref.ChannelID, ref.MessageID, dgo.WithContext(ctx)); err != nil {
		return fmt.Errorf("delete discord message: %w", err)
	}
	return ctx.Err()
}

func (c SideEffectClient) AddReaction(ctx context.Context, channelID string, messageID string, emoji string) error {
	session, err := c.session()
	if err != nil {
		return err
	}
	if err := session.MessageReactionAdd(channelID, messageID, emoji, dgo.WithContext(ctx)); err != nil {
		return fmt.Errorf("add discord reaction: %w", err)
	}
	return ctx.Err()
}

func (c SideEffectClient) CleanupMessages(ctx context.Context, req ports.MessageCleanupRequest) (int, error) {
	session, err := c.session()
	if err != nil {
		return 0, err
	}
	if req.Limit <= 0 {
		return 0, ctx.Err()
	}
	remaining := req.Limit
	deleted := 0
	iterations := (req.Limit + 99) / 100
	if req.UserID != "" {
		iterations = 10
	}
	cutoff := time.Now().Add(-14 * 24 * time.Hour)
	beforeID := ""
	for i := 0; i < iterations && remaining > 0; i++ {
		if err := ctx.Err(); err != nil {
			return deleted, err
		}
		limit := remaining
		if limit > 100 {
			limit = 100
		}
		messages, err := session.ChannelMessages(req.ChannelID, limit, beforeID, "", "", dgo.WithContext(ctx))
		if err != nil {
			return deleted, fmt.Errorf("fetch discord messages for cleanup: %w", err)
		}
		if len(messages) == 0 {
			break
		}
		ids := make([]string, 0, len(messages))
		lastID := ""
		hasRecentMessage := false
		for _, message := range messages {
			if message == nil || message.ID == "" {
				continue
			}
			lastID = message.ID
			if message.Timestamp.IsZero() || !message.Timestamp.Before(cutoff) {
				hasRecentMessage = true
			}
			if req.UserID != "" {
				if message.Author == nil || message.Author.ID != req.UserID {
					continue
				}
			}
			if !message.Timestamp.IsZero() && message.Timestamp.Before(cutoff) {
				continue
			}
			ids = append(ids, message.ID)
		}
		if len(ids) == 0 {
			if req.UserID != "" && hasRecentMessage && lastID != "" {
				beforeID = lastID
				continue
			}
			break
		}
		if len(ids) == 1 {
			if err := session.ChannelMessageDelete(req.ChannelID, ids[0], dgo.WithContext(ctx)); err != nil {
				return deleted, fmt.Errorf("delete discord message for cleanup: %w", err)
			}
		} else if err := session.ChannelMessagesBulkDelete(req.ChannelID, ids, dgo.WithContext(ctx)); err != nil {
			return deleted, fmt.Errorf("bulk delete discord messages: %w", err)
		}
		deleted += len(ids)
		remaining = req.Limit - deleted
		if req.UserID != "" {
			beforeID = lastID
		}
		if len(messages) < limit {
			break
		}
	}
	return deleted, ctx.Err()
}

func (c SideEffectClient) FindChannelByName(ctx context.Context, guildID string, name string, channelType int) (ports.ChannelRef, error) {
	session, err := c.session()
	if err != nil {
		return ports.ChannelRef{}, err
	}
	channels, err := session.GuildChannels(guildID, dgo.WithContext(ctx))
	if err != nil {
		return ports.ChannelRef{}, fmt.Errorf("list discord channels: %w", err)
	}
	for _, channel := range channels {
		if channel == nil || channel.Name != name {
			continue
		}
		if channelType >= 0 && int(channel.Type) != channelType {
			continue
		}
		return channelRefFromDiscord(channel), ctx.Err()
	}
	return ports.ChannelRef{}, ports.ErrChannelNotFound
}

func (c SideEffectClient) FindChannelByID(ctx context.Context, guildID string, channelID string) (ports.ChannelRef, error) {
	session, err := c.session()
	if err != nil {
		return ports.ChannelRef{}, err
	}
	channel, err := session.Channel(channelID, dgo.WithContext(ctx))
	if err != nil {
		if isDiscordNotFound(err) {
			return ports.ChannelRef{}, ports.ErrChannelNotFound
		}
		return ports.ChannelRef{}, fmt.Errorf("load discord channel: %w", err)
	}
	if channel == nil || channel.ID == "" || channel.GuildID != guildID {
		return ports.ChannelRef{}, ports.ErrChannelNotFound
	}
	return channelRefFromDiscord(channel), ctx.Err()
}

func (c SideEffectClient) GuildStats(ctx context.Context, guildID string) (domain.StatsSnapshot, error) {
	session, err := c.session()
	if err != nil {
		return domain.StatsSnapshot{}, err
	}
	guild, err := stateGuild(session, guildID)
	if err != nil {
		guild, err = session.GuildWithCounts(guildID, dgo.WithContext(ctx))
		if err != nil {
			return domain.StatsSnapshot{}, fmt.Errorf("load discord guild stats: %w", err)
		}
	}
	memberCount, userCount, botCount := legacyStatsMemberCounts(guild)
	snapshot := domain.StatsSnapshot{
		MemberCount: memberCount,
		UserCount:   userCount,
		BotCount:    botCount,
	}
	if guild != nil {
		snapshot.ChannelCount, snapshot.TextChannelCount, snapshot.VoiceChannelCount = legacyStatsChannelCounts(guild.Channels)
	}
	return snapshot, ctx.Err()
}

func legacyStatsMemberCounts(guild *dgo.Guild) (int, int, int) {
	if guild == nil {
		return 0, 0, 0
	}
	botCount := 0
	for _, member := range guild.Members {
		if member != nil && member.User != nil && member.User.Bot {
			botCount++
		}
	}
	return guild.MemberCount, guild.MemberCount - botCount, botCount
}

func legacyStatsChannelCounts(channels []*dgo.Channel) (int, int, int) {
	total := 0
	for _, channel := range channels {
		if channel != nil {
			total++
		}
	}
	// Legacy discord.js v14 compared numeric channel types to stale string names.
	return total, 0, 0
}

func (c SideEffectClient) RoleStats(ctx context.Context, guildID string, roleID string) (domain.StatsRoleSnapshot, error) {
	session, err := c.session()
	if err != nil {
		return domain.StatsRoleSnapshot{}, err
	}
	roles, err := session.GuildRoles(guildID, dgo.WithContext(ctx))
	if err != nil {
		return domain.StatsRoleSnapshot{}, fmt.Errorf("list discord guild roles: %w", err)
	}
	roleName := ""
	for _, role := range roles {
		if role != nil && role.ID == roleID {
			roleName = role.Name
			break
		}
	}
	if roleName == "" {
		return domain.StatsRoleSnapshot{}, ports.ErrDiscordRoleMissing
	}
	after := ""
	count := 0
	for {
		members, err := session.GuildMembers(guildID, after, 1000, dgo.WithContext(ctx))
		if err != nil {
			return domain.StatsRoleSnapshot{}, fmt.Errorf("list discord guild members for role stats: %w", err)
		}
		if len(members) == 0 {
			break
		}
		for _, member := range members {
			if member == nil || member.User == nil {
				continue
			}
			after = member.User.ID
			if memberHasRole(member.Roles, roleID) {
				count++
			}
		}
		if len(members) < 1000 {
			break
		}
		if err := ctx.Err(); err != nil {
			return domain.StatsRoleSnapshot{}, err
		}
	}
	return domain.StatsRoleSnapshot{RoleID: roleID, RoleName: roleName, MemberCount: count}, ctx.Err()
}

func (c SideEffectClient) CountNonBotMembers(ctx context.Context, guildID string) (int, error) {
	session, err := c.session()
	if err != nil {
		return 0, err
	}
	after := ""
	count := 0
	for {
		members, err := session.GuildMembers(guildID, after, 1000, dgo.WithContext(ctx))
		if err != nil {
			return 0, fmt.Errorf("list discord guild members: %w", err)
		}
		if len(members) == 0 {
			break
		}
		for _, member := range members {
			if member == nil || member.User == nil {
				continue
			}
			after = member.User.ID
			if !member.User.Bot {
				count++
			}
		}
		if len(members) < 1000 {
			break
		}
		if err := ctx.Err(); err != nil {
			return 0, err
		}
	}
	return count, ctx.Err()
}

func memberHasRole(roleIDs []string, roleID string) bool {
	for _, candidate := range roleIDs {
		if candidate == roleID {
			return true
		}
	}
	return false
}

func (c SideEffectClient) MemberTag(ctx context.Context, guildID string, userID string) (string, bool, error) {
	session, err := c.session()
	if err != nil {
		return "", false, err
	}
	member, err := session.GuildMember(guildID, userID, dgo.WithContext(ctx))
	if err != nil {
		if isDiscordNotFound(err) {
			return "", false, ctx.Err()
		}
		return "", false, fmt.Errorf("fetch discord guild member: %w", err)
	}
	if member == nil || member.User == nil {
		return "", false, ctx.Err()
	}
	return discordUserTag(member.User), true, ctx.Err()
}

func (c SideEffectClient) MemberTags(ctx context.Context, guildID string, userIDs []string) (map[string]string, error) {
	session, err := c.session()
	if err != nil {
		return nil, err
	}
	wanted := make(map[string]struct{}, len(userIDs))
	for _, userID := range userIDs {
		if userID != "" {
			wanted[userID] = struct{}{}
		}
	}
	out := make(map[string]string, len(wanted))
	if len(wanted) == 0 {
		return out, ctx.Err()
	}
	if len(wanted) <= 20 {
		for userID := range wanted {
			tag, ok, err := c.MemberTag(ctx, guildID, userID)
			if err != nil {
				return nil, err
			}
			if ok {
				out[userID] = tag
			}
		}
		return out, ctx.Err()
	}
	after := ""
	for {
		members, err := session.GuildMembers(guildID, after, 1000, dgo.WithContext(ctx))
		if err != nil {
			return nil, fmt.Errorf("list discord guild members for poll export: %w", err)
		}
		if len(members) == 0 {
			break
		}
		for _, member := range members {
			if member == nil || member.User == nil {
				continue
			}
			after = member.User.ID
			if _, ok := wanted[member.User.ID]; ok {
				out[member.User.ID] = discordUserTag(member.User)
				if len(out) == len(wanted) {
					return out, ctx.Err()
				}
			}
		}
		if len(members) < 1000 {
			break
		}
		if err := ctx.Err(); err != nil {
			return nil, err
		}
	}
	return out, ctx.Err()
}

func isDiscordNotFound(err error) bool {
	var restErr *dgo.RESTError
	if !errors.As(err, &restErr) || restErr.Response == nil {
		return false
	}
	return restErr.Response.StatusCode == http.StatusNotFound
}

func discordUserTag(user *dgo.User) string {
	if user.Discriminator != "" && user.Discriminator != "0" {
		return user.Username + "#" + user.Discriminator
	}
	return user.Username
}

func (c SideEffectClient) CreateChannel(ctx context.Context, req ports.ChannelCreateRequest) (ports.ChannelRef, error) {
	session, err := c.session()
	if err != nil {
		return ports.ChannelRef{}, err
	}
	overwrites := make([]*dgo.PermissionOverwrite, 0, len(req.PermissionOverwrites))
	for _, overwrite := range req.PermissionOverwrites {
		overwrites = append(overwrites, &dgo.PermissionOverwrite{
			ID:    overwrite.ID,
			Type:  dgo.PermissionOverwriteType(overwrite.Type),
			Allow: overwrite.Allow,
			Deny:  overwrite.Deny,
		})
	}
	created, err := session.GuildChannelCreateComplex(req.GuildID, dgo.GuildChannelCreateData{
		Name:                 req.Name,
		Type:                 dgo.ChannelType(req.Type),
		ParentID:             req.ParentID,
		UserLimit:            req.UserLimit,
		PermissionOverwrites: overwrites,
	}, dgo.WithContext(ctx))
	if err != nil {
		return ports.ChannelRef{}, fmt.Errorf("create discord channel: %w", err)
	}
	return channelRefFromDiscord(created), ctx.Err()
}

func (c SideEffectClient) RenameChannel(ctx context.Context, guildID string, channelID string, name string) (ports.ChannelRef, error) {
	session, err := c.session()
	if err != nil {
		return ports.ChannelRef{}, err
	}
	updated, err := session.ChannelEdit(channelID, &dgo.ChannelEdit{Name: name}, dgo.WithContext(ctx))
	if err != nil {
		if isDiscordNotFound(err) {
			return ports.ChannelRef{}, ports.ErrChannelNotFound
		}
		return ports.ChannelRef{}, fmt.Errorf("rename discord channel: %w", err)
	}
	if updated == nil || updated.ID == "" || updated.GuildID != guildID {
		return ports.ChannelRef{}, ports.ErrChannelNotFound
	}
	return channelRefFromDiscord(updated), ctx.Err()
}

func (c SideEffectClient) DeleteChannel(ctx context.Context, channelID string) error {
	session, err := c.session()
	if err != nil {
		return err
	}
	if _, err := session.ChannelDelete(channelID, dgo.WithContext(ctx)); err != nil {
		if isDiscordNotFound(err) {
			return ports.ErrChannelNotFound
		}
		return fmt.Errorf("delete discord channel: %w", err)
	}
	return ctx.Err()
}

func (c SideEffectClient) VoiceChannelMemberCount(ctx context.Context, guildID string, channelID string) (int, error) {
	session, err := c.session()
	if err != nil {
		return 0, err
	}
	if session.State == nil {
		return 0, errors.New("discord voice state cache is unavailable")
	}
	guild, err := session.State.Guild(guildID)
	if err != nil || guild == nil {
		channel, channelErr := session.Channel(channelID, dgo.WithContext(ctx))
		if channelErr != nil {
			if isDiscordNotFound(channelErr) {
				return 0, ports.ErrChannelNotFound
			}
			return 0, fmt.Errorf("load discord voice channel for member count: %w", channelErr)
		}
		if channel == nil || channel.ID == "" || channel.GuildID != guildID {
			return 0, ports.ErrChannelNotFound
		}
		return 0, errors.New("discord voice state cache is missing guild")
	}
	count := 0
	for _, voice := range guild.VoiceStates {
		if voice != nil && voice.ChannelID == channelID {
			count++
		}
	}
	return count, ctx.Err()
}

func channelRefFromDiscord(channel *dgo.Channel) ports.ChannelRef {
	if channel == nil {
		return ports.ChannelRef{}
	}
	overwrites := make([]ports.PermissionOverwrite, 0, len(channel.PermissionOverwrites))
	for _, overwrite := range channel.PermissionOverwrites {
		if overwrite == nil {
			continue
		}
		overwrites = append(overwrites, ports.PermissionOverwrite{
			ID:    overwrite.ID,
			Type:  int(overwrite.Type),
			Allow: overwrite.Allow,
			Deny:  overwrite.Deny,
		})
	}
	return ports.ChannelRef{
		GuildID:              channel.GuildID,
		ChannelID:            channel.ID,
		ParentID:             channel.ParentID,
		Name:                 channel.Name,
		Type:                 int(channel.Type),
		PermissionOverwrites: overwrites,
	}
}

func (c SideEffectClient) AddRole(ctx context.Context, guildID string, userID string, roleID string) error {
	session, err := c.session()
	if err != nil {
		return err
	}
	if err := session.GuildMemberRoleAdd(guildID, userID, roleID, dgo.WithContext(ctx)); err != nil {
		return fmt.Errorf("add discord role: %w", err)
	}
	return ctx.Err()
}

func (c SideEffectClient) RemoveRole(ctx context.Context, guildID string, userID string, roleID string) error {
	session, err := c.session()
	if err != nil {
		return err
	}
	if err := session.GuildMemberRoleRemove(guildID, userID, roleID, dgo.WithContext(ctx)); err != nil {
		return fmt.Errorf("remove discord role: %w", err)
	}
	return ctx.Err()
}

func (c SideEffectClient) CanAssignRole(ctx context.Context, guildID string, roleID string) (bool, error) {
	session, err := c.session()
	if err != nil {
		return false, err
	}
	botID, err := c.botUserID(ctx, session)
	if err != nil {
		return false, err
	}
	roles, err := session.GuildRoles(guildID, dgo.WithContext(ctx))
	if err != nil {
		return false, fmt.Errorf("list discord guild roles: %w", err)
	}
	targetPosition, ok := rolePosition(roles, roleID)
	if !ok {
		return false, ports.ErrDiscordRoleMissing
	}
	botMember, err := session.GuildMember(guildID, botID, dgo.WithContext(ctx))
	if err != nil {
		return false, fmt.Errorf("fetch discord bot member: %w", err)
	}
	botHighest := -1
	if botMember != nil {
		for _, currentRoleID := range botMember.Roles {
			if position, ok := rolePosition(roles, currentRoleID); ok && position > botHighest {
				botHighest = position
			}
		}
	}
	return targetPosition < botHighest, ctx.Err()
}

func (c SideEffectClient) ActorCanModerate(ctx context.Context, guildID string, actorRoleIDs []string, targetUserID string) (bool, error) {
	session, err := c.session()
	if err != nil {
		return false, err
	}
	roles, err := session.GuildRoles(guildID, dgo.WithContext(ctx))
	if err != nil {
		return false, fmt.Errorf("list discord guild roles: %w", err)
	}
	targetMember, err := stateMember(session, guildID, targetUserID)
	if err != nil {
		targetMember, err = session.GuildMember(guildID, targetUserID, dgo.WithContext(ctx))
	}
	if err != nil {
		return false, fmt.Errorf("fetch discord target member: %w", err)
	}
	actorHighest := highestRolePosition(roles, actorRoleIDs)
	targetHighest := -1
	if targetMember != nil {
		targetHighest = highestRolePosition(roles, targetMember.Roles)
	}
	return actorHighest > targetHighest, ctx.Err()
}

func (c SideEffectClient) MoveMember(ctx context.Context, guildID string, userID string, channelID *string) error {
	session, err := c.session()
	if err != nil {
		return err
	}
	if err := session.GuildMemberMove(guildID, userID, channelID, dgo.WithContext(ctx)); err != nil {
		return fmt.Errorf("move discord member: %w", err)
	}
	return ctx.Err()
}

func (c SideEffectClient) SetNickname(ctx context.Context, guildID string, userID string, nickname string) error {
	session, err := c.session()
	if err != nil {
		return err
	}
	if err := session.GuildMemberNickname(guildID, userID, nickname, dgo.WithContext(ctx)); err != nil {
		return fmt.Errorf("set discord nickname: %w", err)
	}
	return ctx.Err()
}

func (c SideEffectClient) KickMember(ctx context.Context, guildID string, userID string, reason string) error {
	session, err := c.session()
	if err != nil {
		return err
	}
	if err := session.GuildMemberDeleteWithReason(guildID, userID, reason, dgo.WithContext(ctx)); err != nil {
		return fmt.Errorf("kick discord member: %w", err)
	}
	return ctx.Err()
}

func (c SideEffectClient) BanMember(ctx context.Context, guildID string, userID string, reason string, deleteMessageDays int) error {
	session, err := c.session()
	if err != nil {
		return err
	}
	if err := session.GuildBanCreateWithReason(guildID, userID, reason, deleteMessageDays, dgo.WithContext(ctx)); err != nil {
		return fmt.Errorf("ban discord member: %w", err)
	}
	return ctx.Err()
}

func (c SideEffectClient) AuditLog(ctx context.Context, query ports.AuditLogQuery) ([]ports.AuditLogEntry, error) {
	session, err := c.session()
	if err != nil {
		return nil, err
	}
	logs, err := session.GuildAuditLog(query.GuildID, query.UserID, query.Before, query.Action, query.Limit, dgo.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("read discord audit log: %w", err)
	}
	return auditLogEntriesFromDiscord(logs), ctx.Err()
}

func auditLogEntriesFromDiscord(logs *dgo.GuildAuditLog) []ports.AuditLogEntry {
	if logs == nil {
		return nil
	}
	usersByID := map[string]*dgo.User{}
	for _, user := range logs.Users {
		if user != nil && user.ID != "" {
			usersByID[user.ID] = user
		}
	}
	entries := make([]ports.AuditLogEntry, 0, len(logs.AuditLogEntries))
	for _, entry := range logs.AuditLogEntries {
		if entry == nil {
			continue
		}
		action := 0
		if entry.ActionType != nil {
			action = int(*entry.ActionType)
		}
		out := ports.AuditLogEntry{
			ID:        entry.ID,
			UserID:    entry.UserID,
			TargetID:  entry.TargetID,
			ChannelID: auditLogChannelID(entry),
			Reason:    entry.Reason,
			Action:    action,
		}
		if user := usersByID[entry.UserID]; user != nil {
			out.Username = user.Username
			out.UserTag = userTag(user)
			out.AvatarURL = user.AvatarURL("")
		}
		entries = append(entries, out)
	}
	return entries
}

func auditLogChannelID(entry *dgo.AuditLogEntry) string {
	if entry == nil || entry.Options == nil {
		return ""
	}
	return entry.Options.ChannelID
}

func (c SideEffectClient) session() (*dgo.Session, error) {
	if c.Session == nil {
		return nil, fmt.Errorf("discord side-effect client is not configured")
	}
	c.Session.mu.Lock()
	defer c.Session.mu.Unlock()
	if c.Session.session == nil {
		return nil, fmt.Errorf("discord session is not configured")
	}
	return c.Session.session, nil
}

func (c SideEffectClient) botUserID(ctx context.Context, session *dgo.Session) (string, error) {
	if session.State != nil && session.State.User != nil && session.State.User.ID != "" {
		return session.State.User.ID, nil
	}
	user, err := session.User("@me", dgo.WithContext(ctx))
	if err != nil {
		return "", fmt.Errorf("fetch discord bot user: %w", err)
	}
	if user == nil || user.ID == "" {
		return "", fmt.Errorf("discord bot user is unavailable")
	}
	return user.ID, ctx.Err()
}

func rolePosition(roles []*dgo.Role, roleID string) (int, bool) {
	for _, role := range roles {
		if role != nil && role.ID == roleID {
			return role.Position, true
		}
	}
	return 0, false
}

func highestRolePosition(roles []*dgo.Role, roleIDs []string) int {
	highest := -1
	for _, roleID := range roleIDs {
		if position, ok := rolePosition(roles, roleID); ok && position > highest {
			highest = position
		}
	}
	return highest
}

func outboundEmbeds(embeds []ports.OutboundEmbed) []*dgo.MessageEmbed {
	converted := make([]*dgo.MessageEmbed, 0, len(embeds))
	for _, embed := range embeds {
		out := &dgo.MessageEmbed{
			Title:       embed.Title,
			Description: embed.Description,
			Color:       embed.Color,
		}
		if embed.AuthorName != "" || embed.AuthorIconURL != "" || embed.AuthorURL != "" {
			out.Author = &dgo.MessageEmbedAuthor{
				Name:    embed.AuthorName,
				IconURL: embed.AuthorIconURL,
				URL:     embed.AuthorURL,
			}
		}
		if embed.FooterText != "" || embed.FooterIconURL != "" {
			out.Footer = &dgo.MessageEmbedFooter{
				Text:    embed.FooterText,
				IconURL: embed.FooterIconURL,
			}
		}
		if embed.ThumbnailURL != "" {
			out.Thumbnail = &dgo.MessageEmbedThumbnail{URL: embed.ThumbnailURL}
		}
		if embed.ImageURL != "" {
			out.Image = &dgo.MessageEmbedImage{URL: embed.ImageURL}
		}
		if !embed.Timestamp.IsZero() {
			out.Timestamp = embed.Timestamp.Format(time.RFC3339)
		}
		for _, field := range embed.Fields {
			out.Fields = append(out.Fields, &dgo.MessageEmbedField{
				Name:   field.Name,
				Value:  field.Value,
				Inline: field.Inline,
			})
		}
		converted = append(converted, out)
	}
	return converted
}

func outboundComponents(rows []ports.OutboundComponentRow) []dgo.MessageComponent {
	converted := make([]dgo.MessageComponent, 0, len(rows))
	for _, row := range rows {
		actionRow := dgo.ActionsRow{Components: make([]dgo.MessageComponent, 0, len(row.Components))}
		for _, component := range row.Components {
			switch component.Type {
			case "button":
				actionRow.Components = append(actionRow.Components, dgo.Button{
					CustomID: component.CustomID,
					Label:    component.Label,
					Emoji:    toDiscordEmoji(component.Emoji),
					Style:    outboundButtonStyle(component.Style),
				})
			case "select":
				options := make([]dgo.SelectMenuOption, 0, len(component.Options))
				for _, option := range component.Options {
					options = append(options, dgo.SelectMenuOption{
						Label:       option.Label,
						Value:       option.Value,
						Description: option.Description,
						Emoji:       toDiscordEmoji(option.Emoji),
						Default:     option.Default,
					})
				}
				selectMenu := dgo.SelectMenu{
					MenuType:    dgo.StringSelectMenu,
					CustomID:    component.CustomID,
					Placeholder: component.Placeholder,
					Options:     options,
				}
				if component.MinValues > 0 {
					minValues := component.MinValues
					selectMenu.MinValues = &minValues
				}
				if component.MaxValues > 0 {
					selectMenu.MaxValues = component.MaxValues
				}
				actionRow.Components = append(actionRow.Components, selectMenu)
			}
		}
		if len(actionRow.Components) > 0 {
			converted = append(converted, actionRow)
		}
	}
	return converted
}

func outboundButtonStyle(style string) dgo.ButtonStyle {
	switch style {
	case "primary":
		return dgo.PrimaryButton
	case "secondary":
		return dgo.SecondaryButton
	case "success":
		return dgo.SuccessButton
	case "danger":
		return dgo.DangerButton
	case "link":
		return dgo.LinkButton
	default:
		return dgo.SecondaryButton
	}
}

func coreAllowedMentions(allowed ports.AllowedMentions) *dgo.MessageAllowedMentions {
	out := &dgo.MessageAllowedMentions{}
	if allowed.ParseEveryone {
		out.Parse = append(out.Parse, dgo.AllowedMentionTypeEveryone)
	}
	if allowed.ParseUsers {
		out.Parse = append(out.Parse, dgo.AllowedMentionTypeUsers)
	}
	if allowed.ParseRoles {
		out.Parse = append(out.Parse, dgo.AllowedMentionTypeRoles)
	}
	out.Users = append([]string(nil), allowed.UserIDs...)
	out.Roles = append([]string(nil), allowed.RoleIDs...)
	out.RepliedUser = allowed.RepliedUser
	return out
}

func outboundMessageSend(channelID string, msg ports.OutboundMessage) *dgo.MessageSend {
	send := &dgo.MessageSend{
		Content:         msg.Content,
		Embeds:          outboundEmbeds(msg.Embeds),
		Components:      outboundComponents(msg.Components),
		AllowedMentions: coreAllowedMentions(msg.AllowedMentions),
	}
	if messageID := strings.TrimSpace(msg.ReplyToMessageID); messageID != "" {
		failIfNotExists := true
		send.Reference = &dgo.MessageReference{
			MessageID:       messageID,
			ChannelID:       strings.TrimSpace(channelID),
			FailIfNotExists: &failIfNotExists,
		}
	}
	return send
}

var _ ports.DiscordChannelPort = SideEffectClient{}
var _ ports.DiscordMessagePort = SideEffectClient{}
var _ ports.DiscordTypingPort = SideEffectClient{}
var _ ports.DiscordReactionPort = SideEffectClient{}
var _ ports.DiscordMessageCleaner = SideEffectClient{}
var _ ports.DiscordDirectMessagePort = SideEffectClient{}
var _ ports.DiscordGuildMemberReader = SideEffectClient{}
var _ ports.DiscordGuildStatsReader = SideEffectClient{}
var _ ports.DiscordRoleInspector = SideEffectClient{}
var _ ports.DiscordMemberPort = SideEffectClient{}
var _ ports.DiscordMemberHierarchyInspector = SideEffectClient{}
