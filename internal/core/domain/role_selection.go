package domain

import (
	"errors"
	"regexp"
	"strings"
)

var ErrInvalidRoleSelectionConfig = errors.New("invalid role selection config")

var (
	discordMessageURLRe = regexp.MustCompile(`^https://(?:discordapp\.com|discord\.com)/channels/([^/]+)/([^/]+)/([^/?#]+)`)
	customEmojiRe       = regexp.MustCompile(`^<a?:([A-Za-z0-9_]{1,32}):([0-9]{17,20})>$`)
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
	Stored string
	API    string
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

func ParseDiscordMessageURL(raw string) (DiscordMessageTarget, error) {
	raw = strings.Trim(strings.TrimSpace(raw), "<>")
	matches := discordMessageURLRe.FindStringSubmatch(raw)
	if matches == nil {
		return DiscordMessageTarget{}, ErrInvalidRoleSelectionConfig
	}
	target := DiscordMessageTarget{
		GuildID:   strings.TrimSpace(matches[1]),
		ChannelID: strings.TrimSpace(matches[2]),
		MessageID: strings.TrimSpace(matches[3]),
	}
	if target.GuildID == "" || target.ChannelID == "" || target.MessageID == "" {
		return DiscordMessageTarget{}, ErrInvalidRoleSelectionConfig
	}
	return target, nil
}

func NormalizeLegacyReaction(raw string) (LegacyReaction, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return LegacyReaction{}, ErrInvalidRoleSelectionConfig
	}
	if matches := customEmojiRe.FindStringSubmatch(raw); matches != nil {
		return LegacyReaction{
			Stored: matches[2],
			API:    matches[1] + ":" + matches[2],
		}, nil
	}
	return LegacyReaction{Stored: raw, API: raw}, nil
}
