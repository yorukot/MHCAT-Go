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
	GuildID string
	UserID  string
	XP      int64
	Level   int64
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
	if strings.TrimSpace(c.Color) != "" && !ValidLegacyColor(c.Color) {
		return ErrInvalidTextXPConfig
	}
	return nil
}

func (c VoiceXPConfig) Validate() error {
	if strings.TrimSpace(c.GuildID) == "" || strings.TrimSpace(c.ChannelID) == "" {
		return ErrInvalidVoiceXPConfig
	}
	if strings.TrimSpace(c.Color) != "" && !ValidLegacyColor(c.Color) {
		return ErrInvalidVoiceXPConfig
	}
	return nil
}

func (p XPProfile) Normalize() XPProfile {
	p.GuildID = strings.TrimSpace(p.GuildID)
	p.UserID = strings.TrimSpace(p.UserID)
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
	return int64(float64(level)*float64(level)/3*100 + 100)
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

var legacyHexColorPattern = regexp.MustCompile(`^#?([0-9a-fA-F]{3}|[0-9a-fA-F]{6})$`)

func ValidLegacyColor(value string) bool {
	_, ok := ParseLegacyColorValue(value)
	return strings.TrimSpace(value) == "" || ok
}

func ParseLegacyColorValue(value string) (int, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, false
	}
	if legacyHexColorPattern.MatchString(value) {
		raw := strings.TrimPrefix(value, "#")
		if len(raw) == 3 {
			raw = string([]byte{raw[0], raw[0], raw[1], raw[1], raw[2], raw[2]})
		}
		parsed, err := strconv.ParseInt(raw, 16, 32)
		if err != nil {
			return 0, false
		}
		return int(parsed), true
	}
	parsed, ok := legacyCSSColorValues[strings.ToLower(value)]
	return parsed, ok
}

var legacyCSSColorValues = map[string]int{
	"black":       0x000000,
	"blue":        0x0000FF,
	"brown":       0xA52A2A,
	"cyan":        0x00FFFF,
	"gray":        0x808080,
	"green":       0x008000,
	"grey":        0x808080,
	"lime":        0x00FF00,
	"magenta":     0xFF00FF,
	"orange":      0xFFA500,
	"pink":        0xFFC0CB,
	"purple":      0x800080,
	"red":         0xFF0000,
	"transparent": 0x000000,
	"white":       0xFFFFFF,
	"yellow":      0xFFFF00,
}
