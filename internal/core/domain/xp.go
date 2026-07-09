package domain

import (
	"errors"
	"regexp"
	"strconv"
	"strings"
)

var ErrInvalidTextXPConfig = errors.New("invalid text xp config")
var ErrInvalidVoiceXPConfig = errors.New("invalid voice xp config")

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
