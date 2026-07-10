package domain

import (
	"errors"
	"regexp"
	"strconv"
	"strings"
)

var ErrInvalidTextXPConfig = errors.New("invalid text xp config")
var ErrInvalidVoiceXPConfig = errors.New("invalid voice xp config")
var ErrInvalidXPRewardRoleConfig = errors.New("invalid xp reward role config")
var ErrInvalidXPAdjustment = errors.New("invalid xp adjustment")
var ErrInvalidXPRankQuery = errors.New("invalid xp rank query")

const (
	VoiceXPSessionJoined = "join"
	VoiceXPSessionLeft   = "leave"
)

type TextXPConfig struct {
	GuildID   string
	ChannelID string
	Color     string
	Message   string
}

type VoiceXPConfig struct {
	GuildID   string
	ChannelID string
	Color     string
	Message   string
}

type XPProfile struct {
	GuildID   string
	UserID    string
	XP        int64
	Level     int64
	LeaveJoin string
}

type XPAdjustment struct {
	GuildID string
	UserID  string
	Delta   int64
}

type XPRewardRoleConfig struct {
	GuildID       string
	Level         int64
	RoleID        string
	DeleteWhenNot bool
}

func (c TextXPConfig) Validate() error {
	if strings.TrimSpace(c.GuildID) == "" || strings.TrimSpace(c.ChannelID) == "" {
		return ErrInvalidTextXPConfig
	}
	if c.Color != "" && !ValidLegacyXPColor(c.Color) {
		return ErrInvalidTextXPConfig
	}
	return nil
}

func (c VoiceXPConfig) Validate() error {
	if strings.TrimSpace(c.GuildID) == "" || strings.TrimSpace(c.ChannelID) == "" {
		return ErrInvalidVoiceXPConfig
	}
	if c.Color != "" && !ValidLegacyXPColor(c.Color) {
		return ErrInvalidVoiceXPConfig
	}
	return nil
}

func (p XPProfile) Normalize() XPProfile {
	p.GuildID = strings.TrimSpace(p.GuildID)
	p.UserID = strings.TrimSpace(p.UserID)
	p.LeaveJoin = strings.TrimSpace(p.LeaveJoin)
	return p
}

func (p XPProfile) Validate() error {
	p = p.Normalize()
	if p.GuildID == "" || p.UserID == "" {
		return ErrInvalidXPAdjustment
	}
	return nil
}

func (a XPAdjustment) Normalize() XPAdjustment {
	a.GuildID = strings.TrimSpace(a.GuildID)
	a.UserID = strings.TrimSpace(a.UserID)
	return a
}

func (a XPAdjustment) Validate() error {
	a = a.Normalize()
	if a.GuildID == "" || a.UserID == "" {
		return ErrInvalidXPAdjustment
	}
	return nil
}

func ApplyTextXPAdjustment(profile XPProfile, delta int64) XPProfile {
	return applyXPAdjustment(profile.Normalize(), delta, textXPRequiredForLevel)
}

func ApplyVoiceXPAdjustment(profile XPProfile, delta int64) XPProfile {
	return applyXPAdjustment(profile.Normalize(), delta, voiceXPRequiredForLevel)
}

func ApplyTextXPMessage(profile XPProfile, gained int64) (XPProfile, bool) {
	profile = profile.Normalize()
	if gained < 0 {
		gained = 0
	}
	if profile.XP+gained > textXPRequiredForLevel(profile.Level) {
		profile.Level++
		profile.XP = 0
		return profile, true
	}
	profile.XP += gained
	return profile, false
}

func ApplyVoiceXPTick(profile XPProfile, gained int64) (XPProfile, bool) {
	profile = profile.Normalize()
	if gained < 0 {
		gained = 0
	}
	if profile.XP+gained > voiceXPRequiredForLevel(profile.Level) {
		profile.Level++
		profile.XP = gained
		return profile, true
	}
	profile.XP += gained
	return profile, false
}

func LegacyTextXPCoinReward(level int64, xpMultiple float64) int64 {
	return int64(float64(level) * xpMultiple)
}

func LegacyVoiceXPCoinReward(level int64, xpMultiple float64) int64 {
	return LegacyTextXPCoinReward(level, xpMultiple)
}

func applyXPAdjustment(profile XPProfile, delta int64, required func(int64) int64) XPProfile {
	lessXP := delta
	allXP := int64(0)
	level := profile.Level
	if delta > 0 {
		for lessXP > 0 {
			needed := required(level)
			if level == profile.Level {
				needed -= profile.XP
			}
			lessXP -= needed
			if lessXP <= 0 {
				allXP = lessXP + needed
			}
			level++
		}
		level--
	} else {
		for lessXP < 0 {
			needed := required(level)
			if level == profile.Level {
				needed = profile.XP
			}
			lessXP += needed
			if lessXP >= 0 {
				allXP = lessXP
			}
			level--
		}
		level++
	}
	profile.Level = level
	if allXP == profile.XP {
		profile.XP += delta
	} else {
		profile.XP = allXP
	}
	return profile
}

func textXPRequiredForLevel(level int64) int64 {
	return int64(float64(level)*(float64(level)/3)*100 + 100)
}

func voiceXPRequiredForLevel(level int64) int64 {
	return int64(float64(level)*float64(level)/2*100 + 100)
}

func (c XPRewardRoleConfig) Normalize() XPRewardRoleConfig {
	c.GuildID = strings.TrimSpace(c.GuildID)
	c.RoleID = strings.TrimSpace(c.RoleID)
	return c
}

func (c XPRewardRoleConfig) Validate() error {
	c = c.Normalize()
	if c.GuildID == "" || c.RoleID == "" {
		return ErrInvalidXPRewardRoleConfig
	}
	return nil
}

var legacyHexColorPattern = regexp.MustCompile(`^#?[0-9a-fA-F]{6}$`)

func ValidLegacyColor(value string) bool {
	_, ok := ParseLegacyColorValue(value)
	return strings.TrimSpace(value) == "" || ok
}

func ParseLegacyColorValue(value string) (int, bool) {
	if value == "" {
		return 0, false
	}
	if legacyHexColorPattern.MatchString(value) {
		raw := strings.TrimPrefix(value, "#")
		parsed, err := strconv.ParseInt(raw, 16, 32)
		if err != nil {
			return 0, false
		}
		return int(parsed), true
	}
	parsed, ok := legacyDiscordColorValues[value]
	return parsed, ok
}

var legacyDiscordColorValues = map[string]int{
	"Default":           0x000000,
	"White":             0xFFFFFF,
	"Aqua":              0x1ABC9C,
	"Green":             0x57F287,
	"Blue":              0x3498DB,
	"Yellow":            0xFEE75C,
	"Purple":            0x9B59B6,
	"LuminousVividPink": 0xE91E63,
	"Fuchsia":           0xEB459E,
	"Gold":              0xF1C40F,
	"Orange":            0xE67E22,
	"Red":               0xED4245,
	"Grey":              0x95A5A6,
	"Navy":              0x34495E,
	"DarkAqua":          0x11806A,
	"DarkGreen":         0x1F8B4C,
	"DarkBlue":          0x206694,
	"DarkPurple":        0x71368A,
	"DarkVividPink":     0xAD1457,
	"DarkGold":          0xC27C0E,
	"DarkOrange":        0xA84300,
	"DarkRed":           0x992D22,
	"DarkGrey":          0x979C9F,
	"DarkerGrey":        0x7F8C8D,
	"LightGrey":         0xBCC0C0,
	"DarkNavy":          0x2C3E50,
	"Blurple":           0x5865F2,
	"Greyple":           0x99AAB5,
	"DarkButNotBlack":   0x2C2F33,
	"NotQuiteBlack":     0x23272A,
}
