package fakediscord

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
)

type SideEffects struct {
	mu              sync.Mutex
	Channels        []ports.ChannelRef
	Created         []ports.ChannelCreateRequest
	Deleted         []string
	Sent            []SentMessage
	DirectMessages  []DirectMessage
	Edited          []EditedMessage
	DeletedMessage  []ports.MessageRef
	AddedRoles      []RoleChange
	RemovedRoles    []RoleChange
	Nicknames       []NicknameChange
	Kicked          []KickAction
	NonBotMembers   int
	MemberTagValues map[string]string
	AssignableRoles map[string]bool
	MissingRoles    map[string]bool
	Err             error
	KickErr         error
	nextChannel     int
	nextMessage     int
}

type SentMessage struct {
	ChannelID string
	Message   ports.OutboundMessage
	Ref       ports.MessageRef
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

func NewSideEffects() *SideEffects {
	return &SideEffects{MemberTagValues: map[string]string{}, AssignableRoles: map[string]bool{}, MissingRoles: map[string]bool{}, nextChannel: 1, nextMessage: 1}
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
		GuildID:   req.GuildID,
		ChannelID: fmt.Sprintf("created-channel-%d", s.nextChannel),
		Name:      req.Name,
		Type:      req.Type,
	}
	s.nextChannel++
	s.Created = append(s.Created, req)
	s.Channels = append(s.Channels, ref)
	return ref, nil
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
	return s.ready(ctx)
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
	return s.ready(ctx)
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
var _ ports.DiscordDirectMessagePort = (*SideEffects)(nil)
var _ ports.DiscordRolePort = (*SideEffects)(nil)
var _ ports.DiscordMemberPort = (*SideEffects)(nil)
var _ ports.DiscordGuildMemberReader = (*SideEffects)(nil)
var _ ports.DiscordRoleInspector = (*SideEffects)(nil)
