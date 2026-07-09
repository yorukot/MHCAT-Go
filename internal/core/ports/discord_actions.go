package ports

import (
	"context"
	"errors"
	"time"
)

var ErrChannelNotFound = errors.New("channel not found")

type OutboundMessage struct {
	Content         string
	Embeds          []OutboundEmbed
	Components      []OutboundComponentRow
	AllowedMentions AllowedMentions
}

type OutboundEmbed struct {
	AuthorName    string
	AuthorIconURL string
	AuthorURL     string
	Title         string
	Description   string
	Color         int
	Fields        []OutboundEmbedField
	FooterText    string
	FooterIconURL string
	ThumbnailURL  string
	ImageURL      string
	Timestamp     time.Time
}

type OutboundEmbedField struct {
	Name   string
	Value  string
	Inline bool
}

type OutboundComponentRow struct {
	Components []OutboundComponent
}

type OutboundComponent struct {
	Type        string
	CustomID    string
	Label       string
	Style       string
	Emoji       string
	Placeholder string
	MinValues   int
	MaxValues   int
	Options     []OutboundSelectOption
}

type OutboundSelectOption struct {
	Label       string
	Value       string
	Description string
	Emoji       string
	Default     bool
}

type AllowedMentions struct {
	ParseEveryone bool
	ParseUsers    bool
	ParseRoles    bool
	UserIDs       []string
	RoleIDs       []string
	RepliedUser   bool
}

type MessageRef struct {
	ChannelID string
	MessageID string
}

type MessageCleanupRequest struct {
	ChannelID string
	Limit     int
	UserID    string
}

type PermissionOverwrite struct {
	ID    string
	Type  int
	Allow int64
	Deny  int64
}

type ChannelCreateRequest struct {
	GuildID              string
	ParentID             string
	Name                 string
	Type                 int
	UserLimit            int
	PermissionOverwrites []PermissionOverwrite
}

type ChannelRef struct {
	GuildID   string
	ChannelID string
	Name      string
	Type      int
}

type AuditLogQuery struct {
	GuildID string
	UserID  string
	Before  string
	Action  int
	Limit   int
}

type AuditLogEntry struct {
	ID       string
	UserID   string
	TargetID string
	Reason   string
	Action   int
}

type DiscordMessagePort interface {
	SendMessage(ctx context.Context, channelID string, msg OutboundMessage) (MessageRef, error)
	EditMessage(ctx context.Context, ref MessageRef, msg OutboundMessage) error
	DeleteMessage(ctx context.Context, ref MessageRef) error
}

type DiscordMessageCleaner interface {
	CleanupMessages(ctx context.Context, req MessageCleanupRequest) (int, error)
}

type DiscordDirectMessagePort interface {
	SendDirectMessage(ctx context.Context, userID string, msg OutboundMessage) (MessageRef, error)
}

type DiscordChannelPort interface {
	FindChannelByName(ctx context.Context, guildID string, name string, channelType int) (ChannelRef, error)
	CreateChannel(ctx context.Context, req ChannelCreateRequest) (ChannelRef, error)
	DeleteChannel(ctx context.Context, channelID string) error
}

type DiscordRolePort interface {
	AddRole(ctx context.Context, guildID string, userID string, roleID string) error
	RemoveRole(ctx context.Context, guildID string, userID string, roleID string) error
}

type DiscordMemberPort interface {
	MoveMember(ctx context.Context, guildID string, userID string, channelID *string) error
	SetNickname(ctx context.Context, guildID string, userID string, nickname string) error
	KickMember(ctx context.Context, guildID string, userID string, reason string) error
	BanMember(ctx context.Context, guildID string, userID string, reason string, deleteMessageDays int) error
}

type DiscordMemberHierarchyInspector interface {
	ActorCanModerate(ctx context.Context, guildID string, actorRoleIDs []string, targetUserID string) (bool, error)
}

type DiscordAuditLogPort interface {
	AuditLog(ctx context.Context, query AuditLogQuery) ([]AuditLogEntry, error)
}
