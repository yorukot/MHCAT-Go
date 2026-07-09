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
