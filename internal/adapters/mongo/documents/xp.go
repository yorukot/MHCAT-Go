package documents

import (
	"strconv"
	"strings"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type TextXPChannelDocument struct {
	Guild      string `bson:"guild" json:"guild"`
	Channel    string `bson:"channel" json:"channel"`
	Background string `bson:"background,omitempty" json:"background,omitempty"`
	Color      string `bson:"color,omitempty" json:"color,omitempty"`
	Message    string `bson:"message,omitempty" json:"message,omitempty"`
}

type VoiceXPChannelDocument struct {
	Guild      string `bson:"guild" json:"guild"`
	Channel    string `bson:"channel" json:"channel"`
	Background string `bson:"background,omitempty" json:"background,omitempty"`
	Color      string `bson:"color,omitempty" json:"color,omitempty"`
	Message    string `bson:"message,omitempty" json:"message,omitempty"`
}

type XPProfileDocument struct {
	Guild     string        `bson:"guild" json:"guild"`
	Member    string        `bson:"member" json:"member"`
	XP        bson.RawValue `bson:"xp" json:"xp"`
	Leavel    bson.RawValue `bson:"leavel" json:"leavel"`
	LeaveJoin string        `bson:"leavejoin,omitempty" json:"leavejoin,omitempty"`
}

type XPRewardRoleDocument struct {
	Guild         string `bson:"guild" json:"guild"`
	Leavel        string `bson:"leavel" json:"leavel"`
	Role          string `bson:"role" json:"role"`
	DeleteWhenNot bool   `bson:"delete_when_not" json:"delete_when_not"`
}

func TextXPChannelDocumentFromDomain(config domain.TextXPConfig) TextXPChannelDocument {
	return TextXPChannelDocument{
		Guild:   config.GuildID,
		Channel: config.ChannelID,
		Color:   config.Color,
		Message: config.Message,
	}
}

func VoiceXPChannelDocumentFromDomain(config domain.VoiceXPConfig) VoiceXPChannelDocument {
	return VoiceXPChannelDocument{
		Guild:   config.GuildID,
		Channel: config.ChannelID,
		Color:   config.Color,
		Message: config.Message,
	}
}

func (d TextXPChannelDocument) ToDomain() domain.TextXPConfig {
	return domain.TextXPConfig{
		GuildID:   d.Guild,
		ChannelID: d.Channel,
		Color:     d.Color,
		Message:   d.Message,
	}
}

func (d VoiceXPChannelDocument) ToDomain() domain.VoiceXPConfig {
	return domain.VoiceXPConfig{
		GuildID:   d.Guild,
		ChannelID: d.Channel,
		Color:     d.Color,
		Message:   d.Message,
	}
}

func (d XPProfileDocument) ToDomain() domain.XPProfile {
	return domain.XPProfile{
		GuildID:   d.Guild,
		UserID:    d.Member,
		XP:        legacyInt64(d.XP),
		Level:     legacyInt64(d.Leavel),
		LeaveJoin: d.LeaveJoin,
	}
}

func XPRewardRoleDocumentFromDomain(config domain.XPRewardRoleConfig) XPRewardRoleDocument {
	config = config.Normalize()
	return XPRewardRoleDocument{
		Guild:         config.GuildID,
		Leavel:        strconv.FormatInt(config.Level, 10),
		Role:          config.RoleID,
		DeleteWhenNot: config.DeleteWhenNot,
	}
}

func (d XPRewardRoleDocument) ToDomain() domain.XPRewardRoleConfig {
	level, _ := strconv.ParseInt(strings.TrimSpace(d.Leavel), 10, 64)
	return domain.XPRewardRoleConfig{
		GuildID:       d.Guild,
		Level:         level,
		RoleID:        d.Role,
		DeleteWhenNot: d.DeleteWhenNot,
	}.Normalize()
}
