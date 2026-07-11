package documents

import (
	"math"
	"strconv"
	"strings"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type JoinRoleDocument struct {
	Guild     string `bson:"guild" json:"guild"`
	Role      string `bson:"role" json:"role"`
	GiveToWho string `bson:"give_to_who" json:"give_to_who"`
}

type JoinMessageDocument struct {
	Guild          string  `bson:"guild" json:"guild"`
	Enable         *bool   `bson:"enable,omitempty" json:"enable,omitempty"`
	MessageContent *string `bson:"message_content,omitempty" json:"message_content,omitempty"`
	Color          *string `bson:"color,omitempty" json:"color,omitempty"`
	Channel        string  `bson:"channel" json:"channel"`
	Image          *string `bson:"img,omitempty" json:"img,omitempty"`
}

type LeaveMessageDocument struct {
	Guild          string  `bson:"guild" json:"guild"`
	MessageContent *string `bson:"message_content,omitempty" json:"message_content,omitempty"`
	Title          *string `bson:"title,omitempty" json:"title,omitempty"`
	Color          *string `bson:"color,omitempty" json:"color,omitempty"`
	Channel        string  `bson:"channel" json:"channel"`
}

type VerificationDocument struct {
	Guild string  `bson:"guild" json:"guild"`
	Role  string  `bson:"role" json:"role"`
	Name  *string `bson:"name,omitempty" json:"name,omitempty"`
}

type AccountAgeDocument struct {
	Guild   string  `bson:"guild" json:"guild"`
	Hours   string  `bson:"hours" json:"hours"`
	Channel *string `bson:"channel,omitempty" json:"channel,omitempty"`
}

type AccountAgeReadDocument struct {
	Guild   bson.RawValue `bson:"guild" json:"guild"`
	Hours   bson.RawValue `bson:"hours" json:"hours"`
	Channel bson.RawValue `bson:"channel" json:"channel"`
}

func JoinRoleDocumentFromDomain(config domain.JoinRoleConfig) JoinRoleDocument {
	return JoinRoleDocument{
		Guild:     config.GuildID,
		Role:      config.RoleID,
		GiveToWho: domain.NormalizeJoinRoleGiveTo(config.GiveTo),
	}
}

func (d JoinRoleDocument) ToDomain() domain.JoinRoleConfig {
	giveTo := strings.TrimSpace(d.GiveToWho)
	if giveTo == "" {
		giveTo = domain.JoinRoleGiveAllUsers
	}
	return domain.JoinRoleConfig{
		GuildID: d.Guild,
		RoleID:  d.Role,
		GiveTo:  giveTo,
	}
}

func (d JoinMessageDocument) ToDomain() domain.JoinMessageConfig {
	enabled := true
	if d.Enable != nil {
		enabled = *d.Enable
	}
	return domain.JoinMessageConfig{
		GuildID:        d.Guild,
		Enabled:        enabled,
		ChannelID:      d.Channel,
		MessageContent: stringValue(d.MessageContent),
		Color:          stringValue(d.Color),
		ImageURL:       stringValue(d.Image),
	}
}

func LeaveMessageDocumentFromDomain(config domain.LeaveMessageConfig) LeaveMessageDocument {
	return LeaveMessageDocument{
		Guild:          config.GuildID,
		MessageContent: stringPointerOrNil(config.MessageContent),
		Title:          stringPointerOrNil(config.Title),
		Color:          stringPointerOrNil(config.Color),
		Channel:        config.ChannelID,
	}
}

func (d LeaveMessageDocument) ToDomain() domain.LeaveMessageConfig {
	return domain.LeaveMessageConfig{
		GuildID:        d.Guild,
		ChannelID:      d.Channel,
		MessageContent: stringValue(d.MessageContent),
		Title:          stringValue(d.Title),
		Color:          stringValue(d.Color),
	}
}

func VerificationDocumentFromDomain(config domain.VerificationConfig) VerificationDocument {
	return VerificationDocument{
		Guild: config.GuildID,
		Role:  config.RoleID,
		Name:  stringPointerOrNil(config.RenameTemplate),
	}
}

func (d VerificationDocument) ToDomain() domain.VerificationConfig {
	return domain.VerificationConfig{
		GuildID:        d.Guild,
		RoleID:         d.Role,
		RenameTemplate: stringValue(d.Name),
	}
}

func AccountAgeDocumentFromDomain(config domain.AccountAgeConfig) AccountAgeDocument {
	return AccountAgeDocument{
		Guild:   config.GuildID,
		Hours:   strconv.FormatFloat(config.RequiredSeconds, 'f', -1, 64),
		Channel: stringPointerOrNil(config.ChannelID),
	}
}

func (d AccountAgeReadDocument) ToDomain() (domain.AccountAgeConfig, error) {
	guild, _ := legacyMongooseString(d.Guild)
	hoursText, ok := legacyMongooseString(d.Hours)
	if !ok {
		return domain.AccountAgeConfig{}, domain.ErrInvalidAccountAgeConfig
	}
	seconds, ok := legacyJavaScriptNumericString(hoursText)
	if !ok || seconds <= 0 || math.IsInf(seconds, 0) || math.IsNaN(seconds) {
		return domain.AccountAgeConfig{}, domain.ErrInvalidAccountAgeConfig
	}
	channel, _ := legacyMongooseString(d.Channel)
	config := domain.AccountAgeConfig{GuildID: guild, RequiredSeconds: seconds, ChannelID: channel}
	if err := config.Validate(); err != nil {
		return domain.AccountAgeConfig{}, err
	}
	return config, nil
}

func (d AccountAgeReadDocument) ChannelID() string {
	channel, _ := legacyMongooseString(d.Channel)
	return channel
}

func (d AccountAgeDocument) ToDomain() (domain.AccountAgeConfig, error) {
	seconds, err := parsePositiveInt64(d.Hours)
	if err != nil {
		return domain.AccountAgeConfig{}, domain.ErrInvalidAccountAgeConfig
	}
	config := domain.AccountAgeConfig{
		GuildID:         d.Guild,
		RequiredSeconds: float64(seconds),
		ChannelID:       stringValue(d.Channel),
	}
	if err := config.Validate(); err != nil {
		return domain.AccountAgeConfig{}, err
	}
	return config, nil
}

func stringPointerOrNil(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func parsePositiveInt64(value string) (int64, error) {
	parsed, err := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
	if err != nil || parsed <= 0 {
		return 0, domain.ErrInvalidAccountAgeConfig
	}
	return parsed, nil
}
