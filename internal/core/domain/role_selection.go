package domain

import (
	"errors"
	"regexp"
	"strings"
)

var ErrInvalidRoleSelectionConfig = errors.New("invalid role selection config")
var ErrInvalidRoleSelectionEmoji = errors.New("invalid role selection emoji")

var (
	customEmojiRe = regexp.MustCompile(`^<a?:([A-Za-z0-9_]{1,32}):([0-9]{17,20})>$`)
)

const (
	discordMessageURLPrefix    = "https://discord.com/channels/"
	discordAppMessageURLPrefix = "https://discordapp.com/channels/"
)

type RoleReactionConfig struct {
	GuildID   string
	MessageID string
	React     string
	RoleID    string
}

type RoleButtonConfig struct {
	GuildID string
	Number  string
	RoleID  string
}

type DiscordMessageTarget struct {
	GuildID   string
	ChannelID string
	MessageID string
}

type LegacyReaction struct {
	Stored        string
	API           string
	CustomEmojiID string
}

func (c RoleReactionConfig) Normalize() RoleReactionConfig {
	return RoleReactionConfig{
		GuildID:   strings.TrimSpace(c.GuildID),
		MessageID: strings.TrimSpace(c.MessageID),
		React:     strings.TrimSpace(c.React),
		RoleID:    strings.TrimSpace(c.RoleID),
	}
}

func (c RoleReactionConfig) Validate() error {
	c = c.Normalize()
	if c.GuildID == "" || c.MessageID == "" || c.React == "" || c.RoleID == "" {
		return ErrInvalidRoleSelectionConfig
	}
	return nil
}

func (c RoleButtonConfig) Normalize() RoleButtonConfig {
	return RoleButtonConfig{
		GuildID: strings.TrimSpace(c.GuildID),
		Number:  strings.TrimSpace(c.Number),
		RoleID:  strings.TrimSpace(c.RoleID),
	}
}

func (c RoleButtonConfig) Validate() error {
	c = c.Normalize()
	if c.GuildID == "" || c.Number == "" || c.RoleID == "" {
		return ErrInvalidRoleSelectionConfig
	}
	return nil
}

func ParseLegacyDiscordMessageURL(raw string, allowDiscordApp bool) (DiscordMessageTarget, error) {
	value := raw + "{"
	if !strings.Contains(value, discordMessageURLPrefix) && (!allowDiscordApp || !strings.Contains(value, discordAppMessageURLPrefix)) {
		return DiscordMessageTarget{}, ErrInvalidRoleSelectionConfig
	}
	value = strings.Replace(value, discordMessageURLPrefix, "", 1)
	if allowDiscordApp {
		value = strings.Replace(value, discordAppMessageURLPrefix, "", 1)
	}
	firstSlash := strings.Index(value, "/")
	lastSlash := strings.LastIndex(value, "/")
	messageSource := legacySubstring(value, firstSlash+1, strings.LastIndex(value, "{")) + "{"
	target := DiscordMessageTarget{
		GuildID:   legacySubstring(value, 0, firstSlash),
		ChannelID: legacySubstring(value, firstSlash+1, lastSlash),
		MessageID: legacySubstring(messageSource, strings.Index(messageSource, "/")+1, strings.LastIndex(messageSource, "{")),
	}
	return target, nil
}

func legacySubstring(value string, start int, end int) string {
	if start < 0 {
		start = 0
	}
	if end < 0 {
		end = 0
	}
	if start > len(value) {
		start = len(value)
	}
	if end > len(value) {
		end = len(value)
	}
	if start > end {
		start, end = end, start
	}
	return value[start:end]
}

func NormalizeLegacyReaction(raw string) (LegacyReaction, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return LegacyReaction{}, ErrInvalidRoleSelectionConfig
	}
	if matches := customEmojiRe.FindStringSubmatch(raw); matches != nil {
		return LegacyReaction{
			Stored:        matches[2],
			API:           matches[1] + ":" + matches[2],
			CustomEmojiID: matches[2],
		}, nil
	}
	return LegacyReaction{Stored: raw, API: raw}, nil
}

// IsLegacyUnicodeEmoji mirrors the unanchored UTF-16 ranges used by the
// legacy JavaScript command, including its broad surrogate-pair match.
func IsLegacyUnicodeEmoji(raw string) bool {
	runes := []rune(raw)
	for index, current := range runes {
		if current >= 0x10000 || isLegacyEmojiRune(current) {
			return true
		}
		if current < '#' || current > '9' {
			continue
		}
		next := index + 1
		if next < len(runes) && runes[next] == 0xFE0F {
			next++
		}
		if next < len(runes) && runes[next] == 0x20E3 {
			return true
		}
	}
	return false
}

func isLegacyEmojiRune(value rune) bool {
	switch {
	case value == '[', // Accepted by a malformed character class in the legacy regex.
		value == 0x00A9,
		value == 0x00AE,
		value == 0x203C,
		value == 0x2049,
		value >= 0x2190 && value <= 0x21FF,
		value == 0x2122,
		value == 0x2139,
		value == 0x231A,
		value == 0x231B,
		value == 0x2328,
		value == 0x23CF,
		value >= 0x23E9 && value <= 0x23F3,
		value >= 0x23F8 && value <= 0x23FA,
		value == 0x24C2,
		value >= 0x25AA && value <= 0x25AB,
		value == 0x25B6,
		value == 0x25C0,
		value >= 0x25FB && value <= 0x25FE,
		value >= 0x2600 && value <= 0x26FF,
		value >= 0x2700 && value <= 0x27BF,
		value == 0x2934,
		value == 0x2935,
		value == 0x2B05,
		value == 0x2B06,
		value == 0x2B07,
		value == 0x2B1B,
		value == 0x2B1C,
		value == 0x2B50,
		value == 0x2B55,
		value == 0x3030,
		value == 0x303D,
		value == 0x3297,
		value == 0x3299:
		return true
	default:
		return false
	}
}
