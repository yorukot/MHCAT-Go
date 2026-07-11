package stats

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
)

const (
	discordChannelTypeGuildText     = 0
	discordChannelTypeGuildVoice    = 2
	discordChannelTypeGuildCategory = 4

	permissionOverwriteRole   = 0
	permissionOverwriteMember = 1

	permissionViewChannel    = 1 << 10
	permissionSendMessages   = 1 << 11
	permissionManageMessages = 1 << 13
	permissionConnect        = 1 << 20
)

type CreateRequest struct {
	GuildID          string
	ChannelType      string
	Option           string
	BotUserID        string
	BeforeBaseCreate func(context.Context) error
}

type CreateService struct {
	Repository ports.StatsConfigRepository
	Channels   ports.DiscordChannelPort
	GuildStats ports.DiscordGuildStatsReader
}

func (s CreateService) Create(ctx context.Context, req CreateRequest) (domain.StatsConfig, error) {
	if err := ctx.Err(); err != nil {
		return domain.StatsConfig{}, err
	}
	req.GuildID = strings.TrimSpace(req.GuildID)
	req.ChannelType = strings.TrimSpace(req.ChannelType)
	req.Option = strings.TrimSpace(req.Option)
	req.BotUserID = strings.TrimSpace(req.BotUserID)
	if req.GuildID == "" || s.Repository == nil || s.Channels == nil || s.GuildStats == nil {
		return domain.StatsConfig{}, domain.ErrInvalidStatsConfigRequest
	}
	if _, ok := domain.ParseStatsChannelType(req.ChannelType); !ok {
		return domain.StatsConfig{}, domain.ErrInvalidStatsChannelType
	}
	snapshot, err := s.GuildStats.GuildStats(ctx, req.GuildID)
	if err != nil {
		return domain.StatsConfig{}, err
	}
	existing, err := s.Repository.GetStatsConfig(ctx, req.GuildID)
	if err != nil && !errors.Is(err, ports.ErrStatsConfigMissing) {
		return domain.StatsConfig{}, err
	}
	if errors.Is(err, ports.ErrStatsConfigMissing) {
		if req.BeforeBaseCreate != nil {
			if err := req.BeforeBaseCreate(ctx); err != nil {
				return domain.StatsConfig{}, err
			}
		}
		return s.createBase(ctx, req, snapshot)
	}
	return s.createOptional(ctx, req, existing, snapshot)
}

func (s CreateService) createBase(ctx context.Context, req CreateRequest, snapshot domain.StatsSnapshot) (domain.StatsConfig, error) {
	parent, err := s.Channels.CreateChannel(ctx, ports.ChannelCreateRequest{
		GuildID: req.GuildID,
		Name:    "伺服器統計數據(這串可隨便改)",
		Type:    discordChannelTypeGuildCategory,
	})
	if err != nil {
		return domain.StatsConfig{}, err
	}
	channelType := discordChannelTypeGuildText
	if req.ChannelType == domain.StatsChannelTypeVoice {
		channelType = discordChannelTypeGuildVoice
	}
	member, err := s.Channels.CreateChannel(ctx, statsChannelRequest(req, parent.ChannelID, channelType, baseStatsChannelName("總人數", req.ChannelType, snapshot.MemberCount)))
	if err != nil {
		return domain.StatsConfig{}, err
	}
	user, err := s.Channels.CreateChannel(ctx, statsChannelRequest(req, parent.ChannelID, channelType, baseStatsChannelName("總成員", req.ChannelType, snapshot.UserCount)))
	if err != nil {
		return domain.StatsConfig{}, err
	}
	bot, err := s.Channels.CreateChannel(ctx, statsChannelRequest(req, parent.ChannelID, channelType, baseStatsChannelName("總BOT數", req.ChannelType, snapshot.BotCount)))
	if err != nil {
		return domain.StatsConfig{}, err
	}
	config := domain.StatsConfig{
		GuildID:          req.GuildID,
		ParentID:         parent.ChannelID,
		MemberNumberID:   member.ChannelID,
		MemberNumberName: strconv.Itoa(snapshot.MemberCount),
		UserNumberID:     user.ChannelID,
		UserNumberName:   strconv.Itoa(snapshot.UserCount),
		BotNumberID:      bot.ChannelID,
		BotNumberName:    strconv.Itoa(snapshot.BotCount),
	}
	if err := s.Repository.SaveStatsConfig(ctx, config); err != nil {
		return domain.StatsConfig{}, err
	}
	return config.Normalize(), ctx.Err()
}

func (s CreateService) createOptional(ctx context.Context, req CreateRequest, existing domain.StatsConfig, snapshot domain.StatsSnapshot) (domain.StatsConfig, error) {
	if req.Option == "" {
		return domain.StatsConfig{}, domain.ErrStatsOptionRequired
	}
	option, ok := domain.ParseStatsOption(req.Option)
	if !ok {
		return domain.StatsConfig{}, domain.ErrInvalidStatsOption
	}
	if existing.HasOptionalChannel(option) {
		return domain.StatsConfig{}, domain.ErrStatsChannelAlreadyExists
	}
	parentID, err := statsParentID(ctx, s.Channels, req.GuildID, existing.ParentID)
	if err != nil {
		return domain.StatsConfig{}, err
	}
	displayValue := optionalStatsValue(option, snapshot)
	storedValue := displayValue
	if option == domain.StatsOptionVoiceCount {
		storedValue = snapshot.TextChannelCount
	}
	channelType := discordChannelTypeGuildText
	if req.ChannelType == domain.StatsChannelTypeVoice {
		channelType = discordChannelTypeGuildVoice
	}
	created, err := s.Channels.CreateChannel(ctx, statsChannelRequest(req, parentID, channelType, optionalStatsChannelName(option, displayValue)))
	if err != nil {
		return domain.StatsConfig{}, err
	}
	return s.Repository.AddStatsConfigChannel(ctx, req.GuildID, option, created.ChannelID, storedValue)
}

func statsParentID(ctx context.Context, channels ports.DiscordChannelPort, guildID string, configuredParentID string) (string, error) {
	if configuredParentID == "" {
		return "", nil
	}
	if _, err := channels.FindCachedChannelByID(ctx, guildID, configuredParentID); err == nil {
		return configuredParentID, nil
	} else if !errors.Is(err, ports.ErrChannelNotFound) {
		return "", err
	}
	return "", nil
}

func statsChannelRequest(req CreateRequest, parentID string, channelType int, name string) ports.ChannelCreateRequest {
	return ports.ChannelCreateRequest{
		GuildID:              req.GuildID,
		ParentID:             parentID,
		Name:                 name,
		Type:                 channelType,
		PermissionOverwrites: statsPermissionOverwrites(req.GuildID, req.BotUserID, channelType),
	}
}

func statsPermissionOverwrites(guildID string, botUserID string, channelType int) []ports.PermissionOverwrite {
	allow := int64(permissionViewChannel | permissionManageMessages | permissionSendMessages)
	deny := int64(permissionSendMessages)
	if channelType == discordChannelTypeGuildVoice {
		allow = int64(permissionViewChannel | permissionManageMessages | permissionConnect)
		deny = int64(permissionConnect)
	}
	overwrites := []ports.PermissionOverwrite{}
	if botUserID != "" {
		overwrites = append(overwrites, ports.PermissionOverwrite{ID: botUserID, Type: permissionOverwriteMember, Allow: allow})
	}
	overwrites = append(overwrites, ports.PermissionOverwrite{ID: guildID, Type: permissionOverwriteRole, Allow: int64(permissionViewChannel), Deny: deny})
	return overwrites
}

func baseStatsChannelName(label string, channelType string, value int) string {
	if channelType == domain.StatsChannelTypeVoice {
		return label + ":" + strconv.Itoa(value)
	}
	return label + ": " + strconv.Itoa(value)
}

func optionalStatsChannelName(option string, value int) string {
	switch option {
	case domain.StatsOptionChannelCount:
		return "總頻道數: " + strconv.Itoa(value)
	case domain.StatsOptionTextCount:
		return "總文字頻道數: " + strconv.Itoa(value)
	case domain.StatsOptionVoiceCount:
		return "總語音頻道數: " + strconv.Itoa(value)
	default:
		return strconv.Itoa(value)
	}
}

func optionalStatsValue(option string, snapshot domain.StatsSnapshot) int {
	switch option {
	case domain.StatsOptionChannelCount:
		return snapshot.ChannelCount
	case domain.StatsOptionTextCount:
		return snapshot.TextChannelCount
	case domain.StatsOptionVoiceCount:
		return snapshot.VoiceChannelCount
	default:
		return 0
	}
}
