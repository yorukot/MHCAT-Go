package domain

import (
	"errors"
	"strings"
)

var ErrInvalidJoinRoleConfig = errors.New("invalid join role config")
var ErrInvalidJoinMessageConfig = errors.New("invalid join message config")
var ErrInvalidLeaveMessageConfig = errors.New("invalid leave message config")
var ErrInvalidVerificationConfig = errors.New("invalid verification config")
var ErrInvalidVerificationChallenge = errors.New("invalid verification challenge")
var ErrInvalidAccountAgeConfig = errors.New("invalid account age config")

const (
	JoinRoleGiveAllUsers = "all_user"
	JoinRoleGiveBots     = "all_bot"
	JoinRoleGiveMembers  = "all_member"
)

type JoinRoleConfig struct {
	GuildID string
	RoleID  string
	GiveTo  string
}

type JoinMessageConfig struct {
	GuildID        string
	Enabled        bool
	ChannelID      string
	MessageContent string
	Color          string
	ImageURL       string
}

type LeaveMessageConfig struct {
	GuildID        string
	ChannelID      string
	MessageContent string
	Title          string
	Color          string
}

type VerificationConfig struct {
	GuildID        string
	RoleID         string
	RenameTemplate string
}

type VerificationChallenge struct {
	StateID string
	GuildID string
	UserID  string
	Answer  string
}

type AccountAgeConfig struct {
	GuildID         string
	RequiredSeconds int64
	ChannelID       string
}

func (c JoinRoleConfig) Validate() error {
	if strings.TrimSpace(c.GuildID) == "" || strings.TrimSpace(c.RoleID) == "" {
		return ErrInvalidJoinRoleConfig
	}
	switch strings.TrimSpace(c.GiveTo) {
	case "", JoinRoleGiveAllUsers, JoinRoleGiveBots, JoinRoleGiveMembers:
		return nil
	default:
		return ErrInvalidJoinRoleConfig
	}
}

func (c JoinMessageConfig) Deliverable() bool {
	return strings.TrimSpace(c.GuildID) != "" &&
		c.Enabled &&
		strings.TrimSpace(c.ChannelID) != "" &&
		strings.TrimSpace(c.MessageContent) != "" &&
		strings.TrimSpace(c.Color) != ""
}

func NormalizeJoinRoleGiveTo(value string) string {
	switch strings.TrimSpace(value) {
	case JoinRoleGiveBots:
		return JoinRoleGiveBots
	case JoinRoleGiveMembers:
		return JoinRoleGiveMembers
	default:
		return JoinRoleGiveAllUsers
	}
}

func (c VerificationConfig) Validate() error {
	if strings.TrimSpace(c.GuildID) == "" || strings.TrimSpace(c.RoleID) == "" {
		return ErrInvalidVerificationConfig
	}
	return nil
}

func (c VerificationChallenge) Validate() error {
	if strings.TrimSpace(c.StateID) == "" || strings.TrimSpace(c.GuildID) == "" || strings.TrimSpace(c.UserID) == "" || strings.TrimSpace(c.Answer) == "" {
		return ErrInvalidVerificationChallenge
	}
	return nil
}

func (c AccountAgeConfig) Validate() error {
	if strings.TrimSpace(c.GuildID) == "" || c.RequiredSeconds <= 0 {
		return ErrInvalidAccountAgeConfig
	}
	return nil
}

func (c LeaveMessageConfig) ValidateChannel() error {
	if strings.TrimSpace(c.GuildID) == "" || strings.TrimSpace(c.ChannelID) == "" {
		return ErrInvalidLeaveMessageConfig
	}
	return nil
}

func (c LeaveMessageConfig) ValidateContent() error {
	if strings.TrimSpace(c.GuildID) == "" {
		return ErrInvalidLeaveMessageConfig
	}
	if strings.TrimSpace(c.MessageContent) == "" || strings.TrimSpace(c.Title) == "" || strings.TrimSpace(c.Color) == "" {
		return ErrInvalidLeaveMessageConfig
	}
	if !ValidLegacyColor(c.Color) && strings.TrimSpace(c.Color) != "Random" {
		return ErrInvalidLeaveMessageConfig
	}
	return nil
}
