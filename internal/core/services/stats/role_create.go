package stats

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
)

const permissionManageChannels = 1 << 4

type RoleCreateRequest struct {
	GuildID     string
	ChannelType string
	RoleID      string
	BotUserID   string
}

type RoleCreateService struct {
	StatsRepository ports.StatsConfigRepository
	RoleRepository  ports.StatsRoleConfigRepository
	Channels        ports.DiscordChannelPort
	Roles           ports.DiscordRoleStatsReader
}

func (s RoleCreateService) Create(ctx context.Context, req RoleCreateRequest) (domain.StatsRoleConfig, error) {
	if err := ctx.Err(); err != nil {
		return domain.StatsRoleConfig{}, err
	}
	req.GuildID = strings.TrimSpace(req.GuildID)
	req.ChannelType = strings.TrimSpace(req.ChannelType)
	req.RoleID = strings.TrimSpace(req.RoleID)
	req.BotUserID = strings.TrimSpace(req.BotUserID)
	if req.GuildID == "" || s.StatsRepository == nil || s.RoleRepository == nil || s.Channels == nil || s.Roles == nil {
		return domain.StatsRoleConfig{}, domain.ErrInvalidStatsConfigRequest
	}
	if req.RoleID == "" {
		return domain.StatsRoleConfig{}, domain.ErrStatsRoleRequired
	}
	if _, ok := domain.ParseStatsChannelType(req.ChannelType); !ok {
		return domain.StatsRoleConfig{}, domain.ErrInvalidStatsChannelType
	}

	statsConfig, err := s.StatsRepository.GetStatsConfig(ctx, req.GuildID)
	if err != nil {
		return domain.StatsRoleConfig{}, err
	}
	statsConfig = statsConfig.Normalize()

	role, err := s.Roles.RoleStats(ctx, req.GuildID, req.RoleID)
	if err != nil {
		return domain.StatsRoleConfig{}, err
	}
	role.RoleID = strings.TrimSpace(role.RoleID)
	role.RoleName = strings.TrimSpace(role.RoleName)
	if role.RoleID == "" || role.RoleName == "" {
		return domain.StatsRoleConfig{}, ports.ErrDiscordRoleMissing
	}

	parentID := ""
	if statsConfig.ParentID != "" {
		if _, err := s.Channels.FindChannelByID(ctx, req.GuildID, statsConfig.ParentID); err == nil {
			parentID = statsConfig.ParentID
		} else if !errors.Is(err, ports.ErrChannelNotFound) {
			return domain.StatsRoleConfig{}, err
		}
	}

	channelType := discordChannelTypeGuildText
	if req.ChannelType == domain.StatsChannelTypeVoice {
		channelType = discordChannelTypeGuildVoice
	}
	created, err := s.Channels.CreateChannel(ctx, roleStatsChannelRequest(req, parentID, channelType, role.RoleName+": "+strconv.Itoa(role.MemberCount)))
	if err != nil {
		return domain.StatsRoleConfig{}, err
	}
	config := domain.StatsRoleConfig{
		GuildID:     req.GuildID,
		ChannelID:   created.ChannelID,
		ChannelName: strconv.Itoa(role.MemberCount),
		RoleID:      role.RoleID,
	}.Normalize()
	if err := s.RoleRepository.SaveStatsRoleConfig(ctx, config); err != nil {
		return domain.StatsRoleConfig{}, err
	}
	return config, ctx.Err()
}

func roleStatsChannelRequest(req RoleCreateRequest, parentID string, channelType int, name string) ports.ChannelCreateRequest {
	return ports.ChannelCreateRequest{
		GuildID:              req.GuildID,
		ParentID:             parentID,
		Name:                 name,
		Type:                 channelType,
		PermissionOverwrites: roleStatsPermissionOverwrites(req.GuildID, req.BotUserID, channelType),
	}
}

func roleStatsPermissionOverwrites(guildID string, botUserID string, channelType int) []ports.PermissionOverwrite {
	allow := int64(permissionViewChannel | permissionManageChannels | permissionSendMessages)
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
