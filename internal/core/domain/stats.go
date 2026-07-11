package domain

import (
	"errors"
	"strconv"
	"strings"
)

var ErrInvalidStatsConfigRequest = errors.New("invalid stats config request")
var ErrInvalidStatsChannelType = errors.New("invalid stats channel type")
var ErrInvalidStatsOption = errors.New("invalid stats option")
var ErrStatsOptionRequired = errors.New("stats option is required")
var ErrStatsChannelAlreadyExists = errors.New("stats channel already exists")
var ErrStatsRoleRequired = errors.New("stats role is required")

const (
	StatsChannelTypeText  = "文字頻道"
	StatsChannelTypeVoice = "語音頻道"

	StatsOptionChannelCount = "頻道數量"
	StatsOptionTextCount    = "文字頻道數量"
	StatsOptionVoiceCount   = "語音頻道數量"
)

type StatsSnapshot struct {
	MemberCount       int
	UserCount         int
	BotCount          int
	ChannelCount      int
	TextChannelCount  int
	VoiceChannelCount int
}

type StatsRoleSnapshot struct {
	RoleID      string
	RoleName    string
	MemberCount int
}

type StatsConfig struct {
	GuildID           string
	ParentID          string
	MemberNumberID    string
	MemberNumberName  string
	UserNumberID      string
	UserNumberName    string
	BotNumberID       string
	BotNumberName     string
	ChannelNumberID   string
	ChannelNumberName string
	TextNumberID      string
	TextNumberName    string
	VoiceNumberID     string
	VoiceNumberName   string
}

type StatsConfigCounterUpdate struct {
	MemberNumberName  *string
	UserNumberName    *string
	BotNumberName     *string
	ChannelNumberName *string
	TextNumberName    *string
	VoiceNumberName   *string
}

type StatsRoleConfig struct {
	GuildID     string
	ChannelID   string
	ChannelName string
	RoleID      string
}

func StatsCounterValue(value int) *string {
	text := strconv.Itoa(value)
	return &text
}

func StatsRoleCounterValue(value int) string {
	return strconv.Itoa(value)
}

func (u StatsConfigCounterUpdate) IsZero() bool {
	return u.MemberNumberName == nil &&
		u.UserNumberName == nil &&
		u.BotNumberName == nil &&
		u.ChannelNumberName == nil &&
		u.TextNumberName == nil &&
		u.VoiceNumberName == nil
}

func (c StatsRoleConfig) Normalize() StatsRoleConfig {
	return StatsRoleConfig{
		GuildID:     strings.TrimSpace(c.GuildID),
		ChannelID:   strings.TrimSpace(c.ChannelID),
		ChannelName: strings.TrimSpace(c.ChannelName),
		RoleID:      strings.TrimSpace(c.RoleID),
	}
}

func (c StatsConfig) Normalize() StatsConfig {
	return StatsConfig{
		GuildID:           strings.TrimSpace(c.GuildID),
		ParentID:          strings.TrimSpace(c.ParentID),
		MemberNumberID:    strings.TrimSpace(c.MemberNumberID),
		MemberNumberName:  strings.TrimSpace(c.MemberNumberName),
		UserNumberID:      strings.TrimSpace(c.UserNumberID),
		UserNumberName:    strings.TrimSpace(c.UserNumberName),
		BotNumberID:       strings.TrimSpace(c.BotNumberID),
		BotNumberName:     strings.TrimSpace(c.BotNumberName),
		ChannelNumberID:   strings.TrimSpace(c.ChannelNumberID),
		ChannelNumberName: strings.TrimSpace(c.ChannelNumberName),
		TextNumberID:      strings.TrimSpace(c.TextNumberID),
		TextNumberName:    strings.TrimSpace(c.TextNumberName),
		VoiceNumberID:     strings.TrimSpace(c.VoiceNumberID),
		VoiceNumberName:   strings.TrimSpace(c.VoiceNumberName),
	}
}

func ParseStatsChannelType(value string) (string, bool) {
	value = strings.TrimSpace(value)
	switch value {
	case StatsChannelTypeText, StatsChannelTypeVoice:
		return value, true
	default:
		return "", false
	}
}

func ParseStatsOption(value string) (string, bool) {
	value = strings.TrimSpace(value)
	switch value {
	case StatsOptionChannelCount, StatsOptionTextCount, StatsOptionVoiceCount:
		return value, true
	default:
		return "", false
	}
}

func (c StatsConfig) HasOptionalChannel(option string) bool {
	switch option {
	case StatsOptionChannelCount:
		return c.ChannelNumberID != ""
	case StatsOptionTextCount:
		return c.TextNumberID != ""
	case StatsOptionVoiceCount:
		return c.VoiceNumberID != ""
	default:
		return false
	}
}

func (c StatsConfig) WithOptionalChannel(option string, channelID string, currentValue int) (StatsConfig, error) {
	channelID = strings.TrimSpace(channelID)
	if channelID == "" {
		return StatsConfig{}, ErrInvalidStatsConfigRequest
	}
	value := strconv.Itoa(currentValue)
	switch option {
	case StatsOptionChannelCount:
		c.ChannelNumberID = channelID
		c.ChannelNumberName = value
	case StatsOptionTextCount:
		c.TextNumberID = channelID
		c.TextNumberName = value
	case StatsOptionVoiceCount:
		c.VoiceNumberID = channelID
		c.VoiceNumberName = value
	default:
		return StatsConfig{}, ErrInvalidStatsOption
	}
	return c, nil
}
