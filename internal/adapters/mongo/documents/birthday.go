package documents

import "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"

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

func intPointerOrNil(value *int) *int {
	if value == nil {
		return nil
	}
	copied := *value
	return &copied
}
