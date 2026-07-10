package documents

import (
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type LoggingConfigDocument struct {
	Guild             string `bson:"guild" json:"guild"`
	ChannelID         string `bson:"channel_id" json:"channel_id"`
	MessageUpdate     bool   `bson:"message_update" json:"message_update"`
	MessageDelete     bool   `bson:"message_delete" json:"message_delete"`
	ChannelUpdate     bool   `bson:"channel_update" json:"channel_update"`
	MemberVoiceUpdate bool   `bson:"member_voice_update" json:"member_voice_update"`
}

type LoggingConfigReadDocument struct {
	Guild             string        `bson:"guild" json:"guild"`
	ChannelID         bson.RawValue `bson:"channel_id" json:"channel_id"`
	MessageUpdate     bson.RawValue `bson:"message_update" json:"message_update"`
	MessageDelete     bson.RawValue `bson:"message_delete" json:"message_delete"`
	ChannelUpdate     bson.RawValue `bson:"channel_update" json:"channel_update"`
	MemberVoiceUpdate bson.RawValue `bson:"member_voice_update" json:"member_voice_update"`
}

func LoggingConfigDocumentFromDomain(config domain.LoggingConfig) LoggingConfigDocument {
	return LoggingConfigDocument{
		Guild:             config.GuildID,
		ChannelID:         config.ChannelID,
		MessageUpdate:     config.MessageUpdate,
		MessageDelete:     config.MessageDelete,
		ChannelUpdate:     config.ChannelUpdate,
		MemberVoiceUpdate: config.MemberVoiceUpdate,
	}
}

func (d LoggingConfigDocument) ToDomain() domain.LoggingConfig {
	return domain.LoggingConfig{
		GuildID:           d.Guild,
		ChannelID:         d.ChannelID,
		MessageUpdate:     d.MessageUpdate,
		MessageDelete:     d.MessageDelete,
		ChannelUpdate:     d.ChannelUpdate,
		MemberVoiceUpdate: d.MemberVoiceUpdate,
	}
}

func (d LoggingConfigReadDocument) ToDomain() domain.LoggingConfig {
	channelID, _ := legacyMongooseString(d.ChannelID)
	return domain.LoggingConfig{
		GuildID:           d.Guild,
		ChannelID:         channelID,
		MessageUpdate:     legacyMongooseBoolean(d.MessageUpdate),
		MessageDelete:     legacyMongooseBoolean(d.MessageDelete),
		ChannelUpdate:     legacyMongooseBoolean(d.ChannelUpdate),
		MemberVoiceUpdate: legacyMongooseBoolean(d.MemberVoiceUpdate),
	}
}
