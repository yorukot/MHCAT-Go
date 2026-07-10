package fakediscord

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
)

type SideEffects struct {
	mu                sync.Mutex
	Channels          []ports.ChannelRef
	Created           []ports.ChannelCreateRequest
	Renamed           []ChannelRename
	Deleted           []string
	Sent              []SentMessage
	TypingChannels    []string
	DirectMessages    []DirectMessage
	Edited            []EditedMessage
	DeletedMessage    []ports.MessageRef
	AuditEntries      []ports.AuditLogEntry
	CleanupRequests   []ports.MessageCleanupRequest
	Reactions         []ReactionAdd
	AddedRoles        []RoleChange
	RemovedRoles      []RoleChange
	MovedMembers      []MemberMove
	Nicknames         []NicknameChange
	Kicked            []KickAction
	Banned            []BanAction
	NonBotMembers     int
	TotalMembers      int
	ChannelCount      int
	TextChannelCount  int
	VoiceChannelCount int
	RoleNames         map[string]string
	RoleMemberCounts  map[string]int
	VoiceMembers      map[string]int
	MemberTagValues   map[string]string
	AssignableRoles   map[string]bool
	MissingRoles      map[string]bool
	CachedEmojiIDs    map[string]bool
	ModerationAllowed map[string]bool
	Err               error
	KickErr           error
	BanErr            error
	AuditErr          error
	CleanupErr        error
	CleanupDeleted    int
	ModerationErr     error
	nextChannel       int
	nextMessage       int
}

type SentMessage struct {
	ChannelID string
	Message   ports.OutboundMessage
	Ref       ports.MessageRef
}

type ChannelRename struct {
	GuildID   string
	ChannelID string
	Name      string
}

type DirectMessage struct {
	UserID  string
	Message ports.OutboundMessage
	Ref     ports.MessageRef
}

type EditedMessage struct {
	Ref     ports.MessageRef
	Message ports.OutboundMessage
}

type RoleChange struct {
	GuildID string
	UserID  string
	RoleID  string
}

type MemberMove struct {
	GuildID   string
	UserID    string
	ChannelID *string
}

type ReactionAdd struct {
	ChannelID string
	MessageID string
	Emoji     string
}

type NicknameChange struct {
	GuildID  string
	UserID   string
	Nickname string
}

type KickAction struct {
	GuildID string
	UserID  string
	Reason  string
}

type BanAction struct {
	GuildID           string
	UserID            string
	Reason            string
	DeleteMessageDays int
}

func NewSideEffects() *SideEffects {
	return &SideEffects{RoleNames: map[string]string{}, RoleMemberCounts: map[string]int{}, VoiceMembers: map[string]int{}, MemberTagValues: map[string]string{}, AssignableRoles: map[string]bool{}, MissingRoles: map[string]bool{}, CachedEmojiIDs: map[string]bool{}, ModerationAllowed: map[string]bool{}, nextChannel: 1, nextMessage: 1}
}

func (s *SideEffects) FindChannelByID(ctx context.Context, guildID string, channelID string) (ports.ChannelRef, error) {
	if err := s.ready(ctx); err != nil {
		return ports.ChannelRef{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, channel := range s.Channels {
		if channel.GuildID == guildID && channel.ChannelID == channelID {
			return channel, nil
		}
	}
	return ports.ChannelRef{}, ports.ErrChannelNotFound
}

func (s *SideEffects) FindCachedChannelByID(ctx context.Context, guildID string, channelID string) (ports.ChannelRef, error) {
	return s.FindChannelByID(ctx, guildID, channelID)
}

func (s *SideEffects) FindChannelByName(ctx context.Context, guildID string, name string, channelType int) (ports.ChannelRef, error) {
	if err := s.ready(ctx); err != nil {
		return ports.ChannelRef{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, channel := range s.Channels {
		if channel.GuildID == guildID && channel.Name == name && (channelType < 0 || channel.Type == channelType) {
			return channel, nil
		}
	}
	return ports.ChannelRef{}, ports.ErrChannelNotFound
}

func (s *SideEffects) GuildStats(ctx context.Context, guildID string) (domain.StatsSnapshot, error) {
	if err := s.ready(ctx); err != nil {
		return domain.StatsSnapshot{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	snapshot := domain.StatsSnapshot{
		MemberCount:       s.TotalMembers,
		UserCount:         s.NonBotMembers,
		BotCount:          s.TotalMembers - s.NonBotMembers,
		ChannelCount:      s.ChannelCount,
		TextChannelCount:  s.TextChannelCount,
		VoiceChannelCount: s.VoiceChannelCount,
	}
	if snapshot.BotCount < 0 {
		snapshot.BotCount = 0
	}
	if snapshot.ChannelCount == 0 || snapshot.TextChannelCount == 0 || snapshot.VoiceChannelCount == 0 {
		for _, channel := range s.Channels {
			if channel.GuildID != guildID {
				continue
			}
			switch channel.Type {
			case 0:
				if s.TextChannelCount == 0 {
					snapshot.TextChannelCount++
				}
				if s.ChannelCount == 0 {
					snapshot.ChannelCount++
				}
			case 2:
				if s.VoiceChannelCount == 0 {
					snapshot.VoiceChannelCount++
				}
				if s.ChannelCount == 0 {
					snapshot.ChannelCount++
				}
			case 4:
			default:
				if s.ChannelCount == 0 {
					snapshot.ChannelCount++
				}
			}
		}
	}
	return snapshot, nil
}

func (s *SideEffects) RoleStats(ctx context.Context, guildID string, roleID string) (domain.StatsRoleSnapshot, error) {
	if err := s.ready(ctx); err != nil {
		return domain.StatsRoleSnapshot{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.MissingRoles != nil && s.MissingRoles[guildID+"/"+roleID] {
		return domain.StatsRoleSnapshot{}, ports.ErrDiscordRoleMissing
	}
	roleName := ""
	if s.RoleNames != nil {
		roleName = s.RoleNames[guildID+"/"+roleID]
		if roleName == "" {
			roleName = s.RoleNames[roleID]
		}
	}
	if roleName == "" {
		return domain.StatsRoleSnapshot{}, ports.ErrDiscordRoleMissing
	}
	count := 0
	if s.RoleMemberCounts != nil {
		count = s.RoleMemberCounts[guildID+"/"+roleID]
		if count == 0 {
			count = s.RoleMemberCounts[roleID]
		}
	}
	return domain.StatsRoleSnapshot{RoleID: roleID, RoleName: roleName, MemberCount: count}, nil
}

func (s *SideEffects) CountNonBotMembers(ctx context.Context, guildID string) (int, error) {
	if err := s.ready(ctx); err != nil {
		return 0, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.NonBotMembers <= 0 {
		return 0, nil
	}
	return s.NonBotMembers, nil
}

func (s *SideEffects) MemberTag(ctx context.Context, guildID string, userID string) (string, bool, error) {
	if err := s.ready(ctx); err != nil {
		return "", false, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.MemberTagValues == nil {
		return "", false, nil
	}
	tag, ok := s.MemberTagValues[userID]
	return tag, ok, nil
}

func (s *SideEffects) MemberTags(ctx context.Context, guildID string, userIDs []string) (map[string]string, error) {
	if err := s.ready(ctx); err != nil {
		return nil, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make(map[string]string, len(userIDs))
	for _, userID := range userIDs {
		if tag, ok := s.MemberTagValues[userID]; ok {
			out[userID] = tag
		}
	}
	return out, nil
}

func (s *SideEffects) CanAssignRole(ctx context.Context, guildID string, roleID string) (bool, error) {
	if err := s.ready(ctx); err != nil {
		return false, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.MissingRoles != nil && s.MissingRoles[guildID+"/"+roleID] {
		return false, ports.ErrDiscordRoleMissing
	}
	if s.AssignableRoles == nil {
		return false, nil
	}
	return s.AssignableRoles[guildID+"/"+roleID], nil
}

func (s *SideEffects) CachedRoleExists(ctx context.Context, guildID string, roleID string) (bool, error) {
	if err := s.ready(ctx); err != nil {
		return false, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.MissingRoles == nil || !s.MissingRoles[guildID+"/"+roleID], nil
}

func (s *SideEffects) ActorCanModerate(ctx context.Context, guildID string, actorRoleIDs []string, targetUserID string) (bool, error) {
	if err := s.ready(ctx); err != nil {
		return false, err
	}
	if s.ModerationErr != nil {
		return false, s.ModerationErr
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	key := guildID + "/" + targetUserID
	allowed, ok := s.ModerationAllowed[key]
	if !ok {
		return true, nil
	}
	return allowed, nil
}

func (s *SideEffects) CreateChannel(ctx context.Context, req ports.ChannelCreateRequest) (ports.ChannelRef, error) {
	if err := s.ready(ctx); err != nil {
		return ports.ChannelRef{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if req.GuildID == "" || req.Name == "" {
		return ports.ChannelRef{}, errors.New("fake create channel requires guild and name")
	}
	ref := ports.ChannelRef{
		GuildID:              req.GuildID,
		ChannelID:            fmt.Sprintf("created-channel-%d", s.nextChannel),
		ParentID:             req.ParentID,
		Name:                 req.Name,
		Type:                 req.Type,
		PermissionOverwrites: append([]ports.PermissionOverwrite(nil), req.PermissionOverwrites...),
	}
	s.nextChannel++
	s.Created = append(s.Created, req)
	s.Channels = append(s.Channels, ref)
	return ref, nil
}

func (s *SideEffects) RenameChannel(ctx context.Context, guildID string, channelID string, name string) (ports.ChannelRef, error) {
	if err := s.ready(ctx); err != nil {
		return ports.ChannelRef{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for index, channel := range s.Channels {
		if channel.GuildID == guildID && channel.ChannelID == channelID {
			channel.Name = name
			s.Channels[index] = channel
			s.Renamed = append(s.Renamed, ChannelRename{GuildID: guildID, ChannelID: channelID, Name: name})
			return channel, nil
		}
	}
	return ports.ChannelRef{}, ports.ErrChannelNotFound
}

func (s *SideEffects) DeleteChannel(ctx context.Context, channelID string) error {
	if err := s.ready(ctx); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Deleted = append(s.Deleted, channelID)
	for index, channel := range s.Channels {
		if channel.ChannelID == channelID {
			s.Channels = append(s.Channels[:index], s.Channels[index+1:]...)
			break
		}
	}
	return nil
}

func (s *SideEffects) VoiceChannelMemberCount(ctx context.Context, guildID string, channelID string) (int, error) {
	if err := s.ready(ctx); err != nil {
		return 0, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.VoiceMembers == nil {
		return 0, nil
	}
	if count, ok := s.VoiceMembers[guildID+"/"+channelID]; ok {
		return count, nil
	}
	return s.VoiceMembers[channelID], nil
}

func (s *SideEffects) AddReaction(ctx context.Context, channelID string, messageID string, emoji string) error {
	if err := s.ready(ctx); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Reactions = append(s.Reactions, ReactionAdd{ChannelID: channelID, MessageID: messageID, Emoji: emoji})
	return nil
}

func (s *SideEffects) CachedEmojiExists(ctx context.Context, emojiID string) (bool, error) {
	if err := s.ready(ctx); err != nil {
		return false, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.CachedEmojiIDs[emojiID], nil
}

func (s *SideEffects) AddRole(ctx context.Context, guildID string, userID string, roleID string) error {
	if err := s.ready(ctx); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.AddedRoles = append(s.AddedRoles, RoleChange{GuildID: guildID, UserID: userID, RoleID: roleID})
	return nil
}

func (s *SideEffects) RemoveRole(ctx context.Context, guildID string, userID string, roleID string) error {
	if err := s.ready(ctx); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.RemovedRoles = append(s.RemovedRoles, RoleChange{GuildID: guildID, UserID: userID, RoleID: roleID})
	return nil
}

func (s *SideEffects) MoveMember(ctx context.Context, guildID string, userID string, channelID *string) error {
	if err := s.ready(ctx); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	var copied *string
	if channelID != nil {
		value := *channelID
		copied = &value
	}
	s.MovedMembers = append(s.MovedMembers, MemberMove{GuildID: guildID, UserID: userID, ChannelID: copied})
	return nil
}

func (s *SideEffects) SetNickname(ctx context.Context, guildID string, userID string, nickname string) error {
	if err := s.ready(ctx); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Nicknames = append(s.Nicknames, NicknameChange{GuildID: guildID, UserID: userID, Nickname: nickname})
	return nil
}

func (s *SideEffects) KickMember(ctx context.Context, guildID string, userID string, reason string) error {
	if err := s.ready(ctx); err != nil {
		return err
	}
	if s.KickErr != nil {
		return s.KickErr
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Kicked = append(s.Kicked, KickAction{GuildID: guildID, UserID: userID, Reason: reason})
	return nil
}

func (s *SideEffects) BanMember(ctx context.Context, guildID string, userID string, reason string, deleteMessageDays int) error {
	if err := s.ready(ctx); err != nil {
		return err
	}
	if s.BanErr != nil {
		return s.BanErr
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Banned = append(s.Banned, BanAction{GuildID: guildID, UserID: userID, Reason: reason, DeleteMessageDays: deleteMessageDays})
	return nil
}

func (s *SideEffects) SendMessage(ctx context.Context, channelID string, msg ports.OutboundMessage) (ports.MessageRef, error) {
	if err := s.ready(ctx); err != nil {
		return ports.MessageRef{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	ref := ports.MessageRef{ChannelID: channelID, MessageID: fmt.Sprintf("sent-message-%d", s.nextMessage)}
	s.nextMessage++
	s.Sent = append(s.Sent, SentMessage{ChannelID: channelID, Message: msg, Ref: ref})
	return ref, nil
}

func (s *SideEffects) SendTyping(ctx context.Context, channelID string) error {
	if err := s.ready(ctx); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.TypingChannels = append(s.TypingChannels, channelID)
	return nil
}

func (s *SideEffects) SendDirectMessage(ctx context.Context, userID string, msg ports.OutboundMessage) (ports.MessageRef, error) {
	if err := s.ready(ctx); err != nil {
		return ports.MessageRef{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	ref := ports.MessageRef{ChannelID: "dm-" + userID, MessageID: fmt.Sprintf("sent-message-%d", s.nextMessage)}
	s.nextMessage++
	s.DirectMessages = append(s.DirectMessages, DirectMessage{UserID: userID, Message: msg, Ref: ref})
	return ref, nil
}

func (s *SideEffects) EditMessage(ctx context.Context, ref ports.MessageRef, msg ports.OutboundMessage) error {
	if err := s.ready(ctx); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Edited = append(s.Edited, EditedMessage{Ref: ref, Message: msg})
	return nil
}

func (s *SideEffects) DeleteMessage(ctx context.Context, ref ports.MessageRef) error {
	if err := s.ready(ctx); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.DeletedMessage = append(s.DeletedMessage, ref)
	return nil
}

func (s *SideEffects) CleanupMessages(ctx context.Context, req ports.MessageCleanupRequest) (int, error) {
	if err := s.ready(ctx); err != nil {
		return 0, err
	}
	if s.CleanupErr != nil {
		return 0, s.CleanupErr
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.CleanupRequests = append(s.CleanupRequests, req)
	if s.CleanupDeleted > 0 {
		return s.CleanupDeleted, nil
	}
	if req.Limit < 0 {
		return 0, nil
	}
	return req.Limit, nil
}

func (s *SideEffects) AuditLog(ctx context.Context, query ports.AuditLogQuery) ([]ports.AuditLogEntry, error) {
	if err := s.ready(ctx); err != nil {
		return nil, err
	}
	if s.AuditErr != nil {
		return nil, s.AuditErr
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	entries := []ports.AuditLogEntry{}
	for _, entry := range s.AuditEntries {
		if query.Action != 0 && entry.Action != query.Action {
			continue
		}
		if query.UserID != "" && entry.UserID != query.UserID {
			continue
		}
		entries = append(entries, entry)
		if query.Limit > 0 && len(entries) >= query.Limit {
			break
		}
	}
	return entries, nil
}

func (s *SideEffects) ready(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if s.Err != nil {
		return s.Err
	}
	return nil
}

var _ ports.DiscordChannelPort = (*SideEffects)(nil)
var _ ports.DiscordMessagePort = (*SideEffects)(nil)
var _ ports.DiscordTypingPort = (*SideEffects)(nil)
var _ ports.DiscordReactionPort = (*SideEffects)(nil)
var _ ports.DiscordMessageCleaner = (*SideEffects)(nil)
var _ ports.DiscordDirectMessagePort = (*SideEffects)(nil)
var _ ports.DiscordAuditLogPort = (*SideEffects)(nil)
var _ ports.DiscordRolePort = (*SideEffects)(nil)
var _ ports.DiscordMemberPort = (*SideEffects)(nil)
var _ ports.DiscordMemberHierarchyInspector = (*SideEffects)(nil)
var _ ports.DiscordGuildStatsReader = (*SideEffects)(nil)
var _ ports.DiscordRoleStatsReader = (*SideEffects)(nil)
var _ ports.DiscordGuildMemberReader = (*SideEffects)(nil)
var _ ports.DiscordRoleInspector = (*SideEffects)(nil)
var _ ports.DiscordCachedRoleReader = (*SideEffects)(nil)
