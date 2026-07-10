package documents

import (
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type BirthdayConfigDocument struct {
	Guild                      string  `bson:"guild" json:"guild"`
	Message                    string  `bson:"msg" json:"msg"`
	UTCOffset                  string  `bson:"utc" json:"utc"`
	Channel                    string  `bson:"channel" json:"channel"`
	EveryoneCanSetBirthdayDate bool    `bson:"everyone_can_set_birthday_date" json:"everyone_can_set_birthday_date"`
	Role                       *string `bson:"role" json:"role"`
}

type BirthdayProfileDocument struct {
	Guild         string `bson:"guild" json:"guild"`
	User          string `bson:"user" json:"user"`
	BirthdayYear  *int   `bson:"birthday_year" json:"birthday_year"`
	BirthdayMonth *int   `bson:"birthday_month" json:"birthday_month"`
	BirthdayDay   *int   `bson:"birthday_day" json:"birthday_day"`
	SendHour      *int   `bson:"send_msg_hour" json:"send_msg_hour"`
	SendMinute    *int   `bson:"send_msg_min" json:"send_msg_min"`
	Allow         bool   `bson:"allow" json:"allow"`
}

type BirthdayConfigReadDocument struct {
	Guild                      bson.RawValue `bson:"guild" json:"guild"`
	Message                    bson.RawValue `bson:"msg" json:"msg"`
	UTCOffset                  bson.RawValue `bson:"utc" json:"utc"`
	Channel                    bson.RawValue `bson:"channel" json:"channel"`
	EveryoneCanSetBirthdayDate bson.RawValue `bson:"everyone_can_set_birthday_date" json:"everyone_can_set_birthday_date"`
	Role                       bson.RawValue `bson:"role" json:"role"`
}

type BirthdayProfileReadDocument struct {
	Guild         bson.RawValue `bson:"guild" json:"guild"`
	User          bson.RawValue `bson:"user" json:"user"`
	BirthdayYear  bson.RawValue `bson:"birthday_year" json:"birthday_year"`
	BirthdayMonth bson.RawValue `bson:"birthday_month" json:"birthday_month"`
	BirthdayDay   bson.RawValue `bson:"birthday_day" json:"birthday_day"`
	SendHour      bson.RawValue `bson:"send_msg_hour" json:"send_msg_hour"`
	SendMinute    bson.RawValue `bson:"send_msg_min" json:"send_msg_min"`
	Allow         bson.RawValue `bson:"allow" json:"allow"`
}

func BirthdayConfigDocumentFromDomain(config domain.BirthdayConfig) BirthdayConfigDocument {
	return BirthdayConfigDocument{
		Guild:                      config.GuildID,
		Message:                    config.Message,
		UTCOffset:                  config.UTCOffset,
		Channel:                    config.ChannelID,
		EveryoneCanSetBirthdayDate: config.EveryoneCanSetBirthdayDate,
		Role:                       stringPointerOrNil(config.RoleID),
	}
}

func BirthdayProfileDocumentFromDomain(profile domain.BirthdayProfile) BirthdayProfileDocument {
	return BirthdayProfileDocument{
		Guild:         profile.GuildID,
		User:          profile.UserID,
		BirthdayYear:  intPointerOrNil(profile.BirthdayYear),
		BirthdayMonth: intPointerOrNil(profile.BirthdayMonth),
		BirthdayDay:   intPointerOrNil(profile.BirthdayDay),
		SendHour:      intPointerOrNil(profile.SendHour),
		SendMinute:    intPointerOrNil(profile.SendMinute),
		Allow:         profile.AllowAdmin,
	}
}

func (d BirthdayConfigDocument) ToDomain() domain.BirthdayConfig {
	return domain.BirthdayConfig{
		GuildID:                    d.Guild,
		Message:                    d.Message,
		UTCOffset:                  d.UTCOffset,
		ChannelID:                  d.Channel,
		EveryoneCanSetBirthdayDate: d.EveryoneCanSetBirthdayDate,
		RoleID:                     stringValue(d.Role),
	}
}

func (d BirthdayProfileDocument) ToDomain() domain.BirthdayProfile {
	return domain.BirthdayProfile{
		GuildID:       d.Guild,
		UserID:        d.User,
		BirthdayYear:  intPointerOrNil(d.BirthdayYear),
		BirthdayMonth: intPointerOrNil(d.BirthdayMonth),
		BirthdayDay:   intPointerOrNil(d.BirthdayDay),
		SendHour:      intPointerOrNil(d.SendHour),
		SendMinute:    intPointerOrNil(d.SendMinute),
		AllowAdmin:    d.Allow,
	}
}

func (d BirthdayConfigReadDocument) ToDomain() domain.BirthdayConfig {
	guild, _ := legacyMongooseString(d.Guild)
	message, _ := legacyMongooseString(d.Message)
	utcOffset, _ := legacyMongooseString(d.UTCOffset)
	channel, _ := legacyMongooseString(d.Channel)
	role, _ := legacyMongooseString(d.Role)
	return domain.BirthdayConfig{
		GuildID:                    guild,
		Message:                    message,
		UTCOffset:                  utcOffset,
		ChannelID:                  channel,
		EveryoneCanSetBirthdayDate: legacyMongooseBoolean(d.EveryoneCanSetBirthdayDate),
		RoleID:                     role,
	}
}

func (d BirthdayProfileReadDocument) ToDomain() domain.BirthdayProfile {
	guild, _ := legacyMongooseString(d.Guild)
	user, _ := legacyMongooseString(d.User)
	return domain.BirthdayProfile{
		GuildID:       guild,
		UserID:        user,
		BirthdayYear:  birthdayNullableInt(d.BirthdayYear),
		BirthdayMonth: birthdayNullableInt(d.BirthdayMonth),
		BirthdayDay:   birthdayNullableInt(d.BirthdayDay),
		SendHour:      birthdayNullableInt(d.SendHour),
		SendMinute:    birthdayNullableInt(d.SendMinute),
		AllowAdmin:    legacyMongooseBoolean(d.Allow),
	}
}

func birthdayNullableInt(value bson.RawValue) *int {
	if value.Type == 0 || value.Type == bson.TypeUndefined || value.Type == bson.TypeNull {
		return nil
	}
	parsed, ok := LegacyExactInt64(value)
	if !ok {
		return nil
	}
	converted := int(parsed)
	if int64(converted) != parsed {
		return nil
	}
	return &converted
}

func intPointerOrNil(value *int) *int {
	if value == nil {
		return nil
	}
	copied := *value
	return &copied
}
